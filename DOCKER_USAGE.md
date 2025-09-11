# Docker Usage with New Features

## Quick Start

### Basic MailSandbox with all features enabled:

```bash
docker run -d \
  --name mailsandbox \
  -p 8025:8025 \
  -p 1025:1025 \
  -p 8026:8026 \
  -e MP_POSTMARK_API=true \
  -e MP_POSTMARK_TOKEN=your-secret-token \
  -e MP_MCP_SERVER=true \
  -e MP_MCP_TRANSPORT=websocket \
  axllent/mailsandbox
```

## Port Configuration

| Port | Service | Description |
|------|---------|-------------|
| 1025 | SMTP | Mail receiving |
| 1110 | POP3 | Mail retrieval |
| 8025 | HTTP | Web UI & API |
| 8026 | MCP  | AI assistant integration |

## Environment Variables

### Postmark API Emulation
```bash
MP_POSTMARK_API=true                    # Enable Postmark API
MP_POSTMARK_TOKEN=your-secret-token     # Authentication token
MP_POSTMARK_ACCEPT_ANY=true             # Accept any token (dev mode)
```

### MCP Server
```bash
MP_MCP_SERVER=true                      # Enable MCP server
MP_MCP_TRANSPORT=stdio|websocket        # Transport type
MP_MCP_HTTP_ADDR=:8026                  # HTTP/WebSocket address
MP_MCP_AUTH_TOKEN=your-mcp-token        # Authentication token
```

## Docker Compose Examples

### mainment Setup
```yaml
services:
  mailsandbox:
    image: axllent/mailsandbox
    ports:
      - "8025:8025"
      - "1025:1025"
      - "8026:8026"
    environment:
      MP_POSTMARK_API: true
      MP_POSTMARK_ACCEPT_ANY: true  # Accept any token
      MP_MCP_SERVER: true
      MP_MCP_TRANSPORT: websocket
    volumes:
      - ./data:/data
```

### Production Setup
```yaml
services:
  mailsandbox:
    image: axllent/mailsandbox
    ports:
      - "8025:8025"
      - "1025:1025"
      - "8026:8026"
    environment:
      MP_POSTMARK_API: true
      MP_POSTMARK_TOKEN: ${POSTMARK_TOKEN}
      MP_MCP_SERVER: true
      MP_MCP_TRANSPORT: websocket
      MP_MCP_AUTH_TOKEN: ${MCP_AUTH_TOKEN}
    volumes:
      - mailsandbox-data:/data
    restart: unless-stopped

volumes:
  mailsandbox-data:
```

## Testing the Container

### Test Postmark API:
```bash
curl -X POST http://localhost:8025/email \
  -H "Content-Type: application/json" \
  -H "X-Postmark-Server-Token: your-secret-token" \
  -d '{
    "From": "test@example.com",
    "To": "recipient@example.com", 
    "Subject": "Docker Test",
    "TextBody": "Test from Docker container"
  }'
```

### Test MCP WebSocket:
```bash
# Connect to WebSocket endpoint
wscat -c ws://localhost:8026/mcp
```

## Security Considerations

### Production Deployment
- Always set specific authentication tokens
- Use HTTPS/TLS termination with reverse proxy
- Restrict network access to trusted sources
- Regularly rotate authentication tokens

### Network Security
```yaml
services:
  mailsandbox:
    image: axllent/mailsandbox
    networks:
      - internal
    # Don't expose ports externally in production
    expose:
      - "8025"
      - "8026"

networks:
  internal:
    driver: bridge
```

## Health Checks

The container includes built-in health checks:
```bash
docker ps  # Shows health status
docker inspect mailsandbox | grep Health -A 10
```

## Logs and Debugging

View container logs:
```bash
docker logs mailsandbox

# Follow logs
docker logs -f mailsandbox

# Show only new feature logs
docker logs mailsandbox 2>&1 | grep -E "postmark|mcp"
```

## Volume Mounts

### Persistent Data
```bash
docker run -d \
  -v mailsandbox-data:/data \
  -e MP_DATA_FILE=/data/mailsandbox.db \
  axllent/mailsandbox
```

### Configuration Files
```bash
docker run -d \
  -v ./config:/config \
  -e MP_UI_CONFIG=/config/ui.json \
  axllent/mailsandbox
```