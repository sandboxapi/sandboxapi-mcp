# SandboxAPI MCP Server

[![MCP](https://img.shields.io/badge/MCP-1.0-blue)](https://modelcontextprotocol.io)
[![Languages](https://img.shields.io/badge/languages-8-green)](https://sandboxapi.dev)
[![License](https://img.shields.io/badge/license-MIT-orange)](LICENSE)

**Give your AI agent the ability to execute code in 8 programming languages, safely.**

SandboxAPI MCP Server connects any MCP-compatible AI client (Claude, Cursor, VS Code, Windsurf, etc.) to secure code execution. Every execution runs inside a gVisor-sandboxed Docker container with no network access, strict resource limits, and ephemeral filesystems.

## Quick Start

### Option 1: Remote Server (No Setup)

Connect directly to the hosted endpoint. No local installation required.

**Claude Desktop** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "sandboxapi": {
      "url": "https://mcp.sandboxapi.dev/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_API_KEY"
      }
    }
  }
}
```

**VS Code** (`.vscode/mcp.json`):

```json
{
  "servers": {
    "sandboxapi": {
      "url": "https://mcp.sandboxapi.dev/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_API_KEY"
      }
    }
  }
}
```

**Cursor** (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "sandboxapi": {
      "url": "https://mcp.sandboxapi.dev/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_API_KEY"
      }
    }
  }
}
```

### Option 2: Docker

```bash
docker run -d \
  -p 8081:8081 \
  -e SANDBOXAPI_API_KEY=your_sandboxapi_key \
  -e MCP_API_KEY=your_mcp_auth_key \
  sandboxapi/mcp:latest
```

Then point your client to `http://localhost:8081/mcp`.

### Option 3: Build from Source

```bash
git clone https://github.com/sandboxapi/sandboxapi-mcp.git
cd sandboxapi-mcp
go build -o sandboxapi-mcp .

export SANDBOXAPI_API_KEY=your_key
export MCP_API_KEY=optional_auth_key
./sandboxapi-mcp
```

## Available Tools

### `execute_code`

Execute code in a sandboxed container.

| Parameter  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `language` | string | Yes      | `python3`, `javascript`, `typescript`, `bash`, `java`, `cpp`, `c`, `go` |
| `code`     | string | Yes      | Source code to execute (max 1MB) |
| `timeout`  | number | No       | Timeout in seconds (default: 10, max: 300) |
| `stdin`    | string | No       | Standard input to pass to the program |

### `execute_batch`

Execute multiple code snippets. Each runs in its own isolated sandbox.

| Parameter    | Type  | Required | Description |
|--------------|-------|----------|-------------|
| `executions` | array | Yes      | Array of `{language, code, timeout?, stdin?}` objects |

### `list_languages`

List all supported programming languages with versions and example code. No parameters.

## Supported Languages

| Language   | Version  | Aliases |
|------------|----------|---------|
| Python     | 3.12     | `python3`, `python`, `py` |
| JavaScript | Node 22  | `javascript`, `js`, `node` |
| TypeScript | 5.4      | `typescript`, `ts` |
| Go         | 1.22     | `go`, `golang` |
| Java       | 21       | `java`, `jdk` |
| C++        | GCC 14   | `cpp`, `c++` |
| C          | GCC 14   | `c`, `gcc` |
| Bash       | 5.2      | `bash`, `sh`, `shell` |

## Environment Variables

| Variable            | Required | Description |
|---------------------|----------|-------------|
| `SANDBOXAPI_API_KEY` | Yes     | API key for the SandboxAPI backend |
| `MCP_API_KEY`       | No       | Auth key for the MCP endpoint (omit for open access) |
| `MCP_PORT`          | No       | Port to listen on (default: `8081`) |
| `SANDBOXAPI_URL`    | No       | API base URL (default: `https://api.sandboxapi.dev`) |

## Security

Every code execution is isolated with defense-in-depth:

- **gVisor (runsc)** — User-space kernel intercepts all syscalls
- **Network isolation** — Executed code cannot make outbound connections
- **Resource limits** — CPU, memory, and disk usage are capped
- **Ephemeral containers** — Destroyed immediately after execution
- **Read-only filesystem** — No persistent writes between executions
- **Code size limits** — Source code capped at 1MB
- **Timeout enforcement** — Hard kill after configured timeout

## API Key

Get your API key at [sandboxapi.dev](https://sandboxapi.dev) or through [RapidAPI](https://rapidapi.com/sandboxapidev/api/sandboxapi).

## Links

- [Website](https://sandboxapi.dev)
- [Playground](https://sandboxapi.dev/playground)
- [RapidAPI Listing](https://rapidapi.com/sandboxapidev/api/sandboxapi)
- [MCP Specification](https://modelcontextprotocol.io)

## License

MIT
