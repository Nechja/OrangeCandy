# OrangeCandy

[![CI](https://github.com/Nechja/OrangeCandy/actions/workflows/ci.yml/badge.svg)](https://github.com/Nechja/OrangeCandy/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0%20%2B%20Commons%20Clause-blue)](LICENSE)

AI-assisted debugging suite for .NET. A single MCP server that provides two complementary tools for AI agents and a live dashboard for us humans. Still testing it out in my own every day workflows.

## Tools

### Debugger
Interactive debugging over the Debug Adapter Protocol. Launch .NET projects, set breakpoints, step through execution, inspect locals and call stacks, evaluate watches, and diagnose exceptions.

### Observer
AOP method tracing via DI interception. The [OrangeCandy.Observe](sdk/OrangeCandy.Observe) NuGet package is added to the target application. Every method call on DI-registered interfaces is captured (arguments, return values, timing, exceptions) and streamed to the debug server in real time.

## Components

| Path | Description |
|------|-------------|
| `mcp-server/` | Go MCP server. Binary includes embedded web UI. |
| `ui/` | Svelte dashboard served at `http://localhost:9119`. |
| `sdk/OrangeCandy.Observe/` | .NET NuGet package for AOP observation. |
| `test-app/` | Example .NET application demonstrating both tools. |

## Installation

### MCP server
Download the binary for your platform from the [releases page](https://github.com/Nechja/OrangeCandy/releases) and place it on your PATH.

Add to your MCP client configuration (e.g. `~/.yourBrandOfRobot.json`):

```json
{
  "mcpServers": {
    "orangecandy-debug": {
      "command": "/path/to/orangecandy-debug"
    }
  }
}
```

The server also requires [netcoredbg](https://github.com/Samsung/netcoredbg) on the PATH for DAP support.

### Observer SDK
See [sdk/OrangeCandy.Observe/README.md](sdk/OrangeCandy.Observe/README.md).

## Dashboard

The MCP server embeds a web dashboard at `http://localhost:9119`. It connects via WebSocket and shows the live debug session: timeline of events, source context, locals, watches, call stack, and method call flow.

## Requirements

- .NET 10 (for the SDK and target applications)
- Go 1.26+ (to build the MCP server from source)
- Node 24+ and pnpm 10+ (to build the UI from source)
- [netcoredbg](https://github.com/Samsung/netcoredbg) on PATH

## License

Apache 2.0 with Commons Clause. Free to use, modify, and redistribute. You may not sell the software or a product/service whose value derives substantially from it. See [LICENSE](LICENSE).
