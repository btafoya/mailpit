# MailSandbox — email testing & simulation for mainers

> **Fork of [Mailpit](https://github.com/axllent/mailpit)** adding Postmark API emulation and an MCP server for AI-assisted debugging.  
> Inspired by and grateful to the original work on **Mailpit** and **MailHog**. See [Credits](#credits--acknowledgements).

<p align="center">
  <a href="https://github.com/btafoya/mailsandbox">Repository</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#postmark-api-emulation">Postmark API</a> •
  <a href="#mcp-server-for-ai-assistants">MCP Server</a>
</p>

---

**MailSandbox** is a small, fast, low-memory, zero-dependency, multi-platform **email testing tool & API** for mainers.  
It provides an SMTP server, modern web interface, REST API, Postmark API emulation, and an MCP server for AI-powered workflows.

---

## Features

- 🚀 **Lightweight** — single static binary or multi-arch Docker image  
- 🌐 **Modern Web UI** — view HTML, source, headers, attachments, search & filter  
- 📬 **SMTP & POP3 servers** — STARTTLS/SSL/TLS, authentication, forwarding, relaying  
- 🔍 **Testing tools** — HTML check, link check, spam score, screenshots, mobile preview  
- 🏷 **Tagging & filtering** — manual or automated (incl. plus addressing)  
- ⚡ **Performance** — 100–200 emails/sec, automatic pruning (default 500 emails)  
- 🧪 **Chaos mode** — inject SMTP errors to test resilience  
- 🔔 **Integrations** — webhooks, browser notifications, `List-Unsubscribe` validation  
- 🆕 **Postmark API emulation** — drop-in replacement for Postmark during mainment  
- 🤖 **MCP server** — AI assistants (e.g., Claude Code) can list, search, and analyze messages  

---

## Installation

The web UI runs on `http://0.0.0.0:8025` and SMTP on `0.0.0.0:1025`.

### Package managers
```bash
# macOS
brew install mailpit

# Arch Linux
yay -S mailpit

# FreeBSD
pkg install mailpit
```

### Script (Linux & macOS)
```bash
sudo sh < <(curl -sL https://raw.githubusercontent.com/axllent/mailpit/main/install.sh)
```

### Static binary
Download from [releases](https://github.com/btafoya/mailsandbox/releases/latest), rename to `mailsandbox`, and add to `$PATH`.

### Docker
```bash
docker run -d --name mailsandbox   -p 8025:8025 -p 1025:1025   btafoya/mailsandbox
```

---

## Usage

List options:
```bash
mailsandbox -h
```

Run in background (Homebrew):
```bash
brew services start mailpit
```

Quick test: [docs](https://mailpit.axllent.org/docs/install/testing/)

---

## Postmark API Emulation

Enable Postmark API emulation:
```bash
mailsandbox --postmark-api --postmark-token "your-secret-token"
```

Environment variables:
```bash
export MP_POSTMARK_API=true
export MP_POSTMARK_TOKEN=your-secret-token
mailsandbox
```

**Example: Node.js**
```js
const postmark = require("postmark");
const client = new postmark.ServerClient("your-secret-token");
client.apiUrl = "http://localhost:8025";
client.sendEmail({ From:"a@x.com", To:"b@y.com", Subject:"Test", TextBody:"Hello" });
```

Endpoints:
- `POST /email`
- `POST /email/batch`
- `POST /email/withTemplate`

---

## MCP Server for AI Assistants

Enable MCP server:
```bash
mailsandbox --mcp-server --mcp-transport stdio
```

Environment variables:
```bash
export MP_MCP_SERVER=true
export MP_MCP_TRANSPORT=stdio
```

### Docker with MCP
```bash
docker build -t mailsandbox-with-mcp .
docker run -d --name mailsandbox   -p 127.0.0.1:8025:8025 -p 1025:1025   -e MP_MCP_SERVER=true   -e MP_MCP_TRANSPORT=stdio   -e MP_POSTMARK_API=true   -e MP_POSTMARK_TOKEN=dev-token-123   mailsandbox-with-mcp
```

### MCP tools
- `list_messages` — list/filter emails  
- `get_message` — retrieve full message  
- `search_messages` — advanced search  
- `analyze_message` — HTML, links, spam scoring  

---

## Migration Guide

- Binary name remains `mailpit`; create a symlink if you prefer `mailsandbox`.  
- Default ports unchanged: HTTP `8025`, SMTP `1025`.  
- APIs: REST API unchanged; Postmark & MCP are opt-in.  
- Docker: use forked image for MCP/Postmark features.

---

## Credits & Acknowledgements

- **[Mailpit](https://github.com/axllent/mailpit)** — by Axel Lenferna de la Motte and contributors. Original project and active mainment.  
- **[MailHog](https://github.com/mailhog/MailHog)** — the original inspiration for mainer email testing tools.  
- **MailSandbox** — maintained by Brian Tafoya, adding Postmark API emulation and MCP server support.  

---

## License
MailSandbox inherits the upstream license from Mailpit. See `LICENSE` for details.
