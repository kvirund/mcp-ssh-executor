# MCP SSH Executor

A Model Context Protocol (MCP) server that enables remote command execution over SSH connections. This server maintains persistent SSH tunnels and provides tools for executing shell commands on remote machines.

## Features

- **Persistent SSH Connections**: Maintains open SSH connections across multiple command executions
- **Docker Containerized**: Runs in a minimal Docker container for easy deployment
- **Shell Operator Support**: Full support for pipes (`|`), logical AND (`&&`), logical OR (`||`), and other shell operators
- **Authentication Methods**: Supports both SSH key and password authentication
- **MCP Protocol**: Implements the official MCP stdio protocol for seamless integration

## Architecture

The server consists of:
- **Go-based MCP Server**: Implements the MCP protocol over stdio using JSON-RPC 2.0
- **SSH Connection Manager**: Handles SSH connection lifecycle and command execution
- **Docker Container**: Provides isolated execution environment with SSH key mounting

## Installation

### Prerequisites
- Docker
- SSH access to target machine
- SSH private key (recommended) or password

### Build and Run

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kvirund/mcp-ssh-executor.git
   cd mcp-ssh-executor
   ```

2. **Build the Docker image**:
   ```bash
   docker build -t mcp-ssh-executor .
   ```

3. **Configure MCP Settings**:
   Add to your MCP configuration (e.g., `cline_mcp_settings.json`):
   ```json
   {
     "mcpServers": {
       "ssh-executor": {
         "command": "docker",
         "args": ["run", "-i", "--rm", "-v", "C:\\Users\\username\\.ssh:/root/.ssh", "mcp-ssh-executor"],
         "env": {
           "SSH_HOST": "your-server-ip",
           "SSH_USER": "your-username",
           "SSH_PRIVATE_KEY_PATH": "/root/.ssh/id_rsa",
           "SSH_PORT": "22"
         },
         "disabled": false,
         "autoApprove": []
       }
     }
   }
   ```

## Usage

### Available Tools

1. **connect_ssh**
   - Establishes SSH connection to the configured server
   - No parameters required
   - Returns success/error status

2. **execute_command**
   - Executes shell commands on the remote server
   - Parameters: `command` (string) - The command to execute
   - Returns command output and exit status

3. **disconnect_ssh**
   - Closes the SSH connection
   - No parameters required
   - Returns confirmation

### Example Usage Flow

1. **Connect**: Call `connect_ssh` to establish connection
2. **Execute Commands**:
   - `ls -la` - List directory contents
   - `ps aux | grep nginx` - Find nginx processes
   - `cd /var/log && tail -f syslog` - Monitor system logs
   - `mkdir backup && cp *.log backup/ && tar czf backup.tar.gz backup` - Create backup
3. **Disconnect**: Call `disconnect_ssh` when done

### Shell Operators

All standard shell operators work:
- **Pipes**: `command1 | command2`
- **Logical AND**: `command1 && command2`
- **Logical OR**: `command1 || command2`
- **Redirection**: `command > file.txt`
- **Background**: `command &`

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `SSH_HOST` | Target server hostname/IP | Yes |
| `SSH_USER` | SSH username | Yes |
| `SSH_PRIVATE_KEY_PATH` | Path to SSH private key | Yes (if not using password) |
| `SSH_PASSWORD` | SSH password | Yes (if not using key) |
| `SSH_PORT` | SSH port (default: 22) | No |

### SSH Key Setup

For key-based authentication:
1. Generate SSH key pair if needed: `ssh-keygen`
2. Copy public key to server: `ssh-copy-id user@host`
3. Mount the `.ssh` directory in Docker: `-v /path/to/.ssh:/root/.ssh`

## Development

### Local Development

1. **Install Go 1.23+**
2. **Clone and build**:
   ```bash
   git clone https://github.com/kvirund/mcp-ssh-executor.git
   cd mcp-ssh-executor
   go mod tidy
   go build -o ssh-executor .
   ```

3. **Test locally**:
   ```bash
   export SSH_HOST=your-host
   export SSH_USER=your-user
   export SSH_PRIVATE_KEY_PATH=/path/to/key
   ./ssh-executor
   ```

### Testing the MCP Protocol

Send JSON-RPC messages to stdin:

```bash
# Initialize
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize"}' | ./ssh-executor

# List tools
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | ./ssh-executor
```

## Security Considerations

- SSH keys are more secure than passwords
- Use SSH port forwarding for additional security
- Limit SSH user permissions on target server
- Regularly rotate SSH keys
- Use `ssh-keygen -t ed25519` for modern key types

## Troubleshooting

### Connection Issues
- Verify SSH credentials and server accessibility
- Check SSH key permissions (`chmod 600 ~/.ssh/id_rsa`)
- Ensure SSH service is running on target server

### Docker Issues
- Verify Docker daemon is running
- Check volume mounting paths
- Ensure SSH directory has correct permissions

### MCP Integration
- Confirm MCP settings are correctly configured
- Check that the Docker image is built and available
- Verify environment variables are set

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and test
4. Submit a pull request

## Related Projects

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
