# Configuration Reference — cli-runner

The server uses two levels of configuration:

1. **Global config** (`mcp/cli-runner/config.json`) — server-wide settings.
2. **CLI config files** (`mcp/cli-runner/clis/<name>/config.json`) — one subdirectory per CLI tool.

All files are validated at startup using [Zod](https://zod.dev/) schemas defined in [`src/config.ts`](../../src/mcp/cli-runner/src/config.ts). Invalid files are skipped with a warning; they do not crash the server.

---

## Global Config — `config.json`

### Schema (`GlobalConfig`)

| Field | Type | Required | Description |
|---|---|---|---|
| `defaultTimeout` | `number` (positive integer) | ✅ | Default command timeout in **seconds**, applied when neither the CLI config nor the command config specifies a timeout. |
| `clisDirectory` | `string` (non-empty) | ✅ | Path to the directory containing CLI subdirectories. Resolved relative to the `mcp/cli-runner/` directory. Each subdirectory must contain a `config.json`. |

### Example

```json
{
  "defaultTimeout": 30,
  "clisDirectory": "./clis"
}
```

---

## CLI Config Files — `clis/<name>/config.json`

Each subdirectory inside `clisDirectory` describes one CLI tool. The subdirectory must contain a `config.json` file that follows the schema below. The subdirectory name is used only for organisation — the `name` field inside `config.json` is what appears in tool descriptions.

### Top-Level Schema (`CliConfig`)

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` (non-empty) | ✅ | Human-readable identifier for the CLI (used in tool descriptions). |
| `command` | `string` (non-empty) | ✅ | The executable name or path invoked by the server. Can be a bare name (looked up on `PATH`), an absolute path, or a relative path starting with `./` or `../` (resolved relative to the CLI's own subdirectory — useful for bundled binaries). |
| `description` | `string` (non-empty) | ✅ | Short description of the CLI tool, prepended to each tool's description. |
| `timeout` | `number` (positive integer) | ❌ | CLI-level timeout in seconds. Overrides `defaultTimeout` for all commands in this file unless a command-level timeout is set. |
| `commands` | `CommandConfig[]` (min 1) | ✅ | Array of subcommand definitions. At least one entry is required. |

### Command Schema (`CommandConfig`)

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` (non-empty) | ✅ | MCP tool name. Must be unique across all loaded CLI configs. Used by the client to invoke the tool. |
| `description` | `string` (non-empty) | ✅ | Human-readable description of what this subcommand does. |
| `subcommand` | `string` (non-empty) | ✅ | The subcommand string passed as the first argument to `command` (e.g. `"status"` → `git status`). |
| `timeout` | `number` (positive integer) | ❌ | Command-level timeout in seconds. Takes highest priority over CLI-level and global timeouts. |

### Timeout Resolution Priority

```
command.timeout  →  cli.timeout  →  config.defaultTimeout
   (highest)                            (lowest)
```

---

## Annotated Examples

### `clis/git-wrapper/config.json` — bundled binary

```json
{
  "name": "git",                          // identifier shown in tool descriptions
  "command": "./git-wrapper.exe",         // relative path: resolved from clis/git-wrapper/
  "description": "Git version control CLI (token-optimized wrapper)",
  "timeout": 60,
  "commands": [
    {
      "name": "git_status",               // MCP tool name the client calls
      "description": "Show the working tree status",
      "subcommand": "status"              // executes: ./git-wrapper.exe status [args...]
    },
    {
      "name": "git_log",
      "description": "Show commit logs (hashes truncated, limited to 50 commits)",
      "subcommand": "log"                 // executes: ./git-wrapper.exe log [args...]
    }
  ]
}
```

The `git-wrapper.exe` binary lives alongside `config.json` in `clis/git-wrapper/`. Because `command` starts with `./`, the server resolves it relative to that subdirectory — no `PATH` entry required.

### `clis/docker/config.json` — system binary on PATH

```json
{
  "name": "docker",
  "command": "docker",                    // bare name: looked up on PATH
  "description": "Docker CLI for managing containers and images",
  "commands": [
    {
      "name": "docker_ps",
      "description": "List running Docker containers",
      "subcommand": "ps"                  // executes: docker ps [args...]
    },
    {
      "name": "docker_images",
      "description": "List Docker images",
      "subcommand": "images"              // executes: docker images [args...]
    }
  ]
}
```

### Example with custom timeouts

```json
{
  "name": "kubectl",
  "command": "kubectl",
  "description": "Kubernetes CLI",
  "timeout": 60,                          // CLI-level: 60s for all commands
  "commands": [
    {
      "name": "kubectl_get_pods",
      "description": "List pods in the current namespace",
      "subcommand": "get pods"
    },
    {
      "name": "kubectl_apply",
      "description": "Apply a configuration to a resource",
      "subcommand": "apply",
      "timeout": 120                      // command-level: overrides the 60s CLI timeout
    }
  ]
}
```

---

## Tool Input Schema

Every registered MCP tool accepts a single input parameter:

| Parameter | Type | Description |
|---|---|---|
| `args` | `string[]` | Additional arguments appended after the subcommand when the CLI is invoked. Pass an empty array `[]` for no extra arguments. |

**Invocation example** (pseudo-call from an MCP client):

```json
{
  "tool": "git_log",
  "input": {
    "args": ["--oneline", "-10"]
  }
}
```

This executes: `git log --oneline -10`
