# MailSandbox — email testing & simulation for developers

> **Fork of [Mailpit](https://github.com/axllent/mailpit) adding Postmark API emulation and an MCP server for AI-assisted debugging.**  
> Inspired by (and grateful to) the original work on Mailpit and MailHog. See [Credits](#credits--acknowledgements).

<p align="center">
  <a href="https://github.com/btafoya/mailsandbox">Repository</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#api">API</a>
</p>

---

**MailSandbox** is a small, fast, low-memory, zero-dependency, multi‑platform **email testing tool & API** for developers.

It acts as an SMTP server, provides a modern web interface to view & test captured emails, and includes APIs for automated integration testing.  
This fork extends the original with **Postmark API emulation** and an **MCP server** so AI coding tools (e.g., Claude Code) can list, read, and analyze messages during development.


## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Postmark API Emulation](#postmark-api-emulation)
- [MCP Server for AI Assistants](#mcp-server-for-ai-assistants)
  - [MCP with Docker](#mcp-with-docker)
- [Troubleshooting](#troubleshooting)
- [Configuring sendmail](#configuring-sendmail)
- [Migration Guide (MailSandbox → MailSandbox)](#migration-guide-mailsandbox--mailsandbox)
- [Credits & Acknowledgements](#credits--acknowledgements)
- [License](#license)

## Features

Everything you love from MailSandbox, plus new simulation and AI hooks:

- Runs entirely from a single **static binary** or multi‑architecture **Docker images**
- **Modern web UI** with advanced **mail search** to view emails (formatted HTML, highlighted HTML source, text, headers, raw source, and MIME attachments including image thumbnails), including optional **HTTPS** & **authentication**
- **SMTP server** with optional STARTTLS or SSL/TLS, authentication (including an “accept any” mode)
- **REST API** for integration testing
- Real-time web UI updates using **WebSockets** for new mail & optional **browser notifications**
- Optional **POP3 server** to download captured messages directly into your email client
- **HTML check** to test & score mail client compatibility with HTML emails
- **Link check** to test message links (HTML & text) & linked images
- **Spam check** to test message “spamminess” using a running SpamAssassin server
- **Create screenshots** of HTML messages via web UI
- **Mobile and tablet HTML preview** toggle in desktop mode
- **Message tagging** including manual tagging or automated tagging using filtering and “plus addressing”
- **SMTP relaying** (message release) — relay messages via a different SMTP server including an optional allowlist of accepted recipients
- **SMTP forwarding** — automatically forward messages via a different SMTP server to predefined email addresses
- Fast message **storing & processing** — ingesting 100–200 emails/second over SMTP (hardware/size dependent), easily handling tens of thousands of emails, with automatic pruning (default: keep most recent 500 emails)
- **Chaos** feature to enable configurable SMTP errors to test application resilience
- `List-Unsubscribe` syntax validation
- Optional **webhook** for received messages
- **NEW: Postmark API emulation** — drop‑in replacement for Postmark API during development & testing
- **NEW: MCP server** — enables AI assistants (like Claude Code) to read and analyze messages for debugging workflows

> For deep links to the original MailSandbox docs covering search filters, HTTPS/auth, SMTP/POP3 config, storage, screenshots, tagging, relay/forward, chaos, webhooks, etc., see the MailSandbox documentation: https://MailSandbox.premdev.com/docs/

## Installation

The web UI listens by default on **`http://0.0.0.0:8025`** and SMTP on **`0.0.0.0:1025`**.

### Install via package managers

- **macOS**: `brew install MailSandbox` *(binary name remains `MailSandbox` if you use Homebrew; see Migration Guide for renaming/aliases)*  
- **Arch Linux**: AUR: `MailSandbox`  
- **FreeBSD**: `pkg install MailSandbox`

> If you prefer `mailsandbox` as the binary name, download the static binary or create a shell alias/symlink. See [Migration Guide](#migration-guide-MailSandbox--mailsandbox).

### Install via script (Linux & macOS)

```bash
# Installs the upstream binary to /usr/local/bin/MailSandbox
sudo sh < <(curl -sL https://raw.githubusercontent.com/btafoya/MailSandbox/develop/install.sh)
```

Customize install path:

```bash
INSTALL_PATH=/usr/bin sudo sh < <(curl -sL https://raw.githubusercontent.com/btafoya/MailSandbox/develop/install.sh)
```

### Download static binary (Windows, Linux, macOS)

Grab the latest static binaries from upstream releases and rename to `mailsandbox` if desired:  
https://github.com/axllent/MailSandbox/releases/latest

### Docker

Multi-arch images are available upstream; for MCP/Postmark additions you can run the **official image** for core features or build a **custom image** with the fork’s extras (see below).

```bash
# Core features (official upstream image, binary is /MailSandbox)
docker run -d --name mailsandbox \
  -p 8025:8025 -p 1025:1025 \
  axllent/MailSandbox
```

### Compile from source

See upstream instructions: https://MailSandbox.axllent.org/docs/install/source/

## Usage

List runtime flags:

```bash
MailSandbox -h
```

If installed via Homebrew, run in the background:

```bash
brew services start MailSandbox
```

### Quick test

See upstream quick test docs: https://MailSandbox.axllent.org/docs/install/testing/

## Postmark API Emulation

Emulate the Postmark API locally so apps using Postmark can be tested without external calls.

### Enable Postmark API

```bash
MailSandbox --postmark-api --postmark-token "your-secret-token"
```

Or with environment variables:

```bash
export MP_POSTMARK_API=true
export MP_POSTMARK_TOKEN="your-secret-token"
MailSandbox
```

### Endpoints

- `POST /email` — Send a single email  
- `POST /email/batch` — Send multiple emails  
- `POST /email/withTemplate` — Send template-based email

### SDK examples

**Node.js**

```js
const postmark = require("postmark");
const client = new postmark.ServerClient("your-secret-token");
client.apiUrl = "http://localhost:8025";
client.sendEmail({
  From: "sender@example.com",
  To: "recipient@example.com",
  Subject: "Test Email",
  TextBody: "Hello from MailSandbox!",
  HtmlBody: "<p>Hello from <strong>MailSandbox</strong>!</p>"
});
```

**Python**

```python
from postmarker.core import PostmarkClient

client = PostmarkClient(
    server_token='your-secret-token',
    api_base='http://localhost:8025'
)

client.emails.send(
    From='sender@example.com',
    To='recipient@example.com',
    Subject='Test Email',
    HtmlBody='<p>Hello from <strong>MailSandbox</strong>!</p>'
)
```

**PHP**

```php
use Postmark\PostmarkClient;

$client = new PostmarkClient('your-secret-token');
$client->setApiUrl('http://localhost:8025');

$client->sendEmail([
    'From' => 'sender@example.com',
    'To' => 'recipient@example.com',
    'Subject' => 'Test Email',
    'HtmlBody' => '<p>Hello from <strong>MailSandbox</strong>!</p>'
]);
```

### Options

```bash
--postmark-api                 # Enable Postmark API emulation
--postmark-token string        # Auth token (required)
--postmark-accept-any          # Accept any token (dev mode)
```

## MCP Server for AI Assistants

Enable an [MCP](https://spec.modelcontextprotocol.io/) server so AI coding tools can search, read, and analyze messages.

### Enable MCP

```bash
MailSandbox --mcp-server --mcp-transport stdio
```

Or with env vars:

```bash
export MP_MCP_SERVER=true
export MP_MCP_TRANSPORT=stdio  # or websocket/http if supported by your client
MailSandbox
```

### Claude Code integration

**Docker (custom image recommended for MCP support):**

```bash
# Build custom image with MCP extras
docker build -t mailsandbox-with-mcp .

# Run with MCP + Postmark API
docker run -d --name MailSandbox-prod \
  -p 127.0.0.1:8025:8025 \
  -p 1025:1025 \
  -e MP_MCP_SERVER=true \
  -e MP_MCP_TRANSPORT=stdio \
  -e MP_POSTMARK_API=true \
  -e MP_POSTMARK_TOKEN=your-secure-token \
  -e MP_POSTMARK_ACCEPT_ANY=true \
  --restart unless-stopped \
  mailsandbox-with-mcp
```

**Register as a global MCP server in Claude Code:**

```bash
claude mcp add --scope user mailsandbox -- \
  docker exec -i MailSandbox-prod /MailSandbox \
  --mcp-server --database=/data/MailSandbox.db --mcp-transport=stdio \
  --smtp=:0 --listen=:0
```

### Available MCP tools

1. `list_messages` — List and filter messages with optional search/tags  
2. `get_message` — Retrieve full message (headers, parts, attachments)  
3. `search_messages` — Advanced search with date/content filters  
4. `analyze_message` — HTML compatibility, link validation, spam scoring

### Transport options

```bash
# stdio (recommended for local)
MailSandbox --mcp-server --mcp-transport stdio

# http/websocket (only if your client supports it)
MailSandbox --mcp-server --mcp-transport http --mcp-http-addr :8026
```

Secure HTTP transport with an auth token:

```bash
MailSandbox --mcp-server --mcp-transport http \
  --mcp-http-addr :8026 \
  --mcp-auth-token $(openssl rand -hex 32)
```

## MCP with Docker

### Minimal stdio setup

```bash
docker build -t mailsandbox-with-mcp .

docker run -d --name MailSandbox-prod \
  -p 127.0.0.1:8025:8025 \
  -p 1025:1025 \
  -e MP_MCP_SERVER=true \
  -e MP_MCP_TRANSPORT=stdio \
  -e MP_POSTMARK_API=true \
  -e MP_POSTMARK_TOKEN=dev-token-123 \
  --restart unless-stopped \
  mailsandbox-with-mcp
```

Register in Claude Code (global):

```bash
claude mcp add --scope user mailsandbox -- \
  docker exec -i MailSandbox-prod /MailSandbox \
  --mcp-server --database=/data/MailSandbox.db --mcp-transport=stdio \
  --smtp=:0 --listen=:0
```

### Docker Compose

```yaml
version: '3.8'
services:
  mailsandbox:
    image: axllent/MailSandbox
    ports:
      - "8025:8025"
      - "1025:1025"
    environment:
      MP_MCP_SERVER: "true"
      MP_MCP_TRANSPORT: "stdio"
      MP_POSTMARK_API: "true"
      MP_POSTMARK_TOKEN: "dev-token-123"
      MP_POSTMARK_ACCEPT_ANY: "true"
    volumes:
      - mailsandbox-data:/data

volumes:
  mailsandbox-data:
```

Register the Compose service in Claude Code:

```bash
claude mcp add mailsandbox-compose -- \
  docker-compose exec -T mailsandbox MailSandbox \
  --mcp-server --mcp-transport stdio --database /data/MailSandbox.db
```

### Security tips

- Bind UI & MCP to localhost for dev, or protect behind a reverse proxy with auth/TLS
- Use random tokens for `MP_POSTMARK_TOKEN` and `MP_MCP_AUTH_TOKEN`
- Prefer Docker bridge networks and avoid exposing MCP ports publicly

## Troubleshooting

### MCP server issues

**“command not found” / flag not recognized**  
You’re likely running an upstream image without the MCP extras. Build the custom image shown above.

**Can’t connect to MCP**  
- Check container is running: `docker ps --filter name=MailSandbox-prod`  
- Check logs: `docker logs MailSandbox-prod`  
- Verify the binary: `docker exec MailSandbox-prod /MailSandbox --version`

**Database path errors**  
Use `/data/MailSandbox.db` inside the container and mount a volume for persistence.

**Port conflicts**  
Run MCP with `--smtp=:0 --listen=:0` to avoid binding SMTP/HTTP for the MCP-only process.

### Postmark API problems

- Verify it’s enabled: `curl http://localhost:8025/api/v1/server` (look for `PostmarkAPIEnabled: true`)  
- Ensure tokens match between your app and MailSandbox  
- Point SDKs to `http://localhost:8025` (not postmarkapp.com)

### Docker checks

```bash
docker logs MailSandbox-prod -f
curl http://localhost:8025/api/v1/messages
telnet localhost 1025
```

## Configuring sendmail

MailSandbox’s SMTP server defaults to port **1025**. Configure your sending application or MTA accordingly.  
MailSandbox’s sendmail replacement docs apply here: https://MailSandbox.axllent.org/docs/install/sendmail/

## Migration Guide (MailSandbox → MailSandbox)

- **Binary name**: You can continue using `MailSandbox` or rename to `mailsandbox`.  
  - Example symlink: `ln -s /usr/local/bin/MailSandbox /usr/local/bin/mailsandbox`
- **Flags & ports**: Defaults are unchanged (HTTP 8025, SMTP 1025).  
- **Docker images**: For MCP features, build the custom image in this repo; for base features, upstream `axllent/MailSandbox` works.  
- **APIs**: Existing REST API remains; **new** Postmark emulation endpoints are opt‑in via flags/env.

## Documentation

This README summarizes both upstream and fork-specific features. For detailed configuration of core features (search filters, HTTPS/auth, POP3, storage backends, screenshots, tagging, relaying, forwarding, chaos, webhooks, etc.), consult the upstream docs: https://MailSandbox.axllent.org/docs/

## API

- **Upstream API v1**: https://MailSandbox.axllent.org/docs/api-v1/  
- **Postmark emulation**: Enabled via `--postmark-api` (see examples above).

## Credits & Acknowledgements

- **MailSandbox** — © [Axel Lenferna de la Motte](https://github.com/axllent) and contributors. Original project, codebase, docs, and ongoing inspiration.  
  Repository: https://github.com/axllent/MailSandbox  
  Website/Docs: https://MailSandbox.axllent.org

- **MailHog** — The original inspiration for developer email testing tools.  
  Repository: https://github.com/mailhog/MailHog

**MailSandbox** is a fork maintained by Brian Tafoya to add **Postmark API emulation** and **MCP server** capabilities while preserving compatibility with upstream MailSandbox features wherever possible.

## License

MailSandbox inherits the upstream license from MailSandbox. See `LICENSE` for details.
