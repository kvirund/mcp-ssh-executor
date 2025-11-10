FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssh-executor .

FROM scratch
COPY --from=builder /app/ssh-executor /ssh-executor
ENTRYPOINT ["/ssh-executor"]
