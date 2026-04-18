# Adding New CLI Tools

This guide walks through the full lifecycle of adding a new CLI tool to cli-runner: authoring the config in the staging area, validating it, and promoting it to production.

---

## Overview

```
src/cli/<tool>/config.json   →   validate   →   mcp/cli-runner/clis/<tool>/config.json
      (staging)                                          (production)
```

The `src/cli/` directory is the staging area where you develop and test new CLI configs before they are served by the running MCP server. Each CLI lives in its own subdirectory containing a `config.json` and, optionally, a local binary.

---

## Step-by-Step Guide

### Step 1 — Create a subdirectory with `config.json` in `src/cli/`

Create a new subdirectory named after the CLI tool you want to expose and add a `config.json` inside it:

```bash
# Example: adding npm as a CLI tool
mkdir src\cli\npm
```

Populate `src/cli/npm/config.json` following the [CLI config schema](config-reference.md#cli-config-files--clisnameconfigjson):

```json
{
  "name": "npm",
  "command": "npm",
  "description": "Node.js package manager CLI",
  "commands": [
    {
      "name": "npm_list",
      "description": "List installed packages",
      "subcommand": "list"
    },
    {
      "name": "npm_outdated",
      "description": "Check for outdated packages",
      "subcommand": "outdated"
    },
    {
      "name": "npm_audit",
      "description": "Run a security audit on installed packages",
      "subcommand": "audit",
      "timeout": 60
    }
  ]
}
```

**Naming conventions:**
- `name` (top-level): lowercase, matches the CLI binary name.
- `command`: the exact executable name as it appears on `PATH`, or `./binary-name` for a local binary (see [Adding a CLI with a local binary](#adding-a-cli-with-a-local-binary)).
- `commands[].name`: use `<cli>_<subcommand>` format (e.g. `npm_list`) — this becomes the MCP tool name.
- `commands[].subcommand`: the literal subcommand string passed to the binary.

---

### Step 2 — Validate the JSON structure

Before promoting, verify the file is valid JSON and matches the schema.

**Quick JSON syntax check:**

```bash
node -e "JSON.parse(require('fs').readFileSync('src/cli/npm/config.json','utf8')); console.log('JSON is valid')"
```

**Schema validation using the server's own loader:**

```bash
node -e "
const path = require('path');
const { loadCliConfigs } = require('./mcp/cli-runner/dist/config.js');
const configs = loadCliConfigs(path.resolve('src/cli'));
console.log('Loaded configs:', configs.map(c => c.name));
"
```

> **Note:** This requires the project to be built first (`npm run build` in `src/mcp/cli-runner/`). Invalid files are reported as warnings and skipped — they do not throw.

---

### Step 3 — Test locally

Point a temporary `config.json` at `src/cli/` to test without touching production:

```bash
# Create a temporary config pointing at the staging directory
node -e "
const fs = require('fs');
fs.writeFileSync('mcp/cli-runner/config.test.json', JSON.stringify({
  defaultTimeout: 30,
  clisDirectory: '../../src/cli'
}, null, 2));
console.log('Test config written');
"

# Start the server with the test config (override config path via env or manual edit)
node mcp/cli-runner/dist/index.js
```

Alternatively, temporarily edit `mcp/cli-runner/config.json` to point `clisDirectory` at `../../src/cli`, start the server, test via your MCP client, then revert.

---

### Step 4 — Promote to production

Once validated, copy the entire subdirectory to the production `clis/` directory:

**Windows:**

```bash
xcopy src\cli\npm mcp\cli-runner\clis\npm /E /I
```

**macOS/Linux:**

```bash
cp -r src/cli/npm mcp/cli-runner/clis/npm
```

---

### Step 5 — Restart the MCP server

The server reads CLI configs only at startup. Restart your MCP client (or the server process) to pick up the new tool.

After restart, the new tools (e.g. `npm_list`, `npm_outdated`, `npm_audit`) will be available to the MCP client.

---

## Adding a CLI with a local binary

Some CLIs ship a custom binary (e.g. a wrapper executable) that should live alongside the config rather than being installed system-wide. To do this:

1. Place the binary inside the CLI's subdirectory:

   ```
   mcp/cli-runner/clis/my-tool/
   ├── config.json
   └── my-tool.exe        ← local binary
   ```

2. Set `"command"` to a relative path starting with `./`:

   ```json
   {
     "name": "my-tool",
     "command": "./my-tool.exe",
     "description": "My custom CLI tool",
     "commands": [...]
   }
   ```

   The server resolves `./my-tool.exe` relative to the CLI's own subdirectory (`clis/my-tool/`), so the binary does **not** need to be on `PATH`.

3. Commit the binary to the repository alongside `config.json` so the server can run without any additional installation step.

---

## Complete Example — Adding `curl`

**`src/cli/curl/config.json`:**

```json
{
  "name": "curl",
  "command": "curl",
  "description": "Command-line tool for transferring data with URLs",
  "timeout": 30,
  "commands": [
    {
      "name": "curl_get",
      "description": "Perform an HTTP GET request",
      "subcommand": "-s"
    },
    {
      "name": "curl_head",
      "description": "Fetch HTTP headers only",
      "subcommand": "-I"
    }
  ]
}
```

**Promote (Windows):**

```bash
xcopy src\cli\curl mcp\cli-runner\clis\curl /E /I
```

**Usage from MCP client:**

```
Tool: curl_get
Args: ["https://api.github.com"]
```

Executes: `curl -s https://api.github.com`

---

## Checklist

- [ ] Subdirectory created at `src/cli/<tool>/`
- [ ] `config.json` created inside the subdirectory
- [ ] JSON syntax is valid
- [ ] All required fields are present (`name`, `command`, `description`, `commands`)
- [ ] Each command has a unique `name` across all CLI configs
- [ ] If using a local binary: binary placed in the subdirectory and `command` set to `./binary-name`
- [ ] If using a system binary: CLI is installed and on `PATH` on the target machine
- [ ] Subdirectory copied to `mcp/cli-runner/clis/<tool>/`
- [ ] MCP server restarted
