package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
}

func (s *SSHExecutor) Connect() error {
	host := os.Getenv("SSH_HOST")
	user := os.Getenv("SSH_USER")
	password := os.Getenv("SSH_PASSWORD")
	keyPath := os.Getenv("SSH_PRIVATE_KEY_PATH")
	portStr := os.Getenv("SSH_PORT")
	if portStr == "" {
		portStr = "22"
	}
	port, _ := strconv.Atoi(portStr)

	if host == "" || user == "" {
		return fmt.Errorf("SSH_HOST and SSH_USER required")
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return err
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return err
	}

	s.client = client
	return nil
}

func (s *SSHExecutor) ExecuteCommand(cmd string) (string, string, error) {
	if s.client == nil {
		return "", "", fmt.Errorf("not connected")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", string(output), err
	}

	return string(output), "", nil
}

func (s *SSHExecutor) Disconnect() {
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
}

type JSONRPCRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Id *int `json:"id,omitempty"`
	Method string `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id int `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name string `json:"name"`
	Description string `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	executor := &SSHExecutor{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}
		if req.Id == nil {
			continue
		}
		var result interface{}
		var rpcErr *JSONRPCError
		switch req.Method {
		case "initialize":
			result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name": "ssh-executor",
					"version": "1.0.0",
				},
			}
		case "tools/list":
			result = ListToolsResult{
				Tools: []Tool{
					{
						Name: "connect_ssh",
						Description: "Connect to SSH server",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						Name: "execute_command",
						Description: "Execute command on remote server",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"command": map[string]interface{}{
									"type": "string",
								},
							},
							"required": []string{"command"},
						},
					},
					{
						Name: "disconnect_ssh",
						Description: "Disconnect from SSH server",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{},
						},
					},
				},
			}
		case "tools/call":
			params, ok := req.Params.(map[string]interface{})
			if !ok {
				rpcErr = &JSONRPCError{Code: -32602, Message: "Invalid params"}
				break
			}
			name, ok := params["name"].(string)
			if !ok {
				rpcErr = &JSONRPCError{Code: -32602, Message: "Invalid params"}
				break
			}
			args, _ := params["arguments"].(map[string]interface{})
			switch name {
			case "connect_ssh":
				err := executor.Connect()
				if err != nil {
					result = CallToolResult{
						Content: []Content{{Type: "text", Text: fmt.Sprintf("Connection failed: %v", err)}},
						IsError: true,
					}
				} else {
					result = CallToolResult{
						Content: []Content{{Type: "text", Text: "Connected to SSH server"}},
					}
				}
			case "execute_command":
				cmd, ok := args["command"].(string)
				if !ok {
					result = CallToolResult{
						Content: []Content{{Type: "text", Text: "command parameter required"}},
						IsError: true,
					}
				} else {
					output, stderr, err := executor.ExecuteCommand(cmd)
					if err != nil {
						result = CallToolResult{
							Content: []Content{{Type: "text", Text: fmt.Sprintf("Command failed: %v\nOutput: %s\nError: %s", err, output, stderr)}},
							IsError: true,
						}
					} else {
						result = CallToolResult{
							Content: []Content{{Type: "text", Text: fmt.Sprintf("Output:\n%s", output)}},
						}
					}
				}
			case "disconnect_ssh":
				executor.Disconnect()
				result = CallToolResult{
					Content: []Content{{Type: "text", Text: "Disconnected from SSH server"}},
				}
			default:
				rpcErr = &JSONRPCError{Code: -32601, Message: "Method not found"}
			}
		default:
			rpcErr = &JSONRPCError{Code: -32601, Message: "Method not found"}
		}
		resp := JSONRPCResponse{
			Jsonrpc: "2.0",
			Id: *req.Id,
			Result: result,
			Error: rpcErr,
		}
		respBytes, _ := json.Marshal(resp)
		fmt.Println(string(respBytes))
	}
}
