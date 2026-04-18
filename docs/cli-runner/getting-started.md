# Getting Started — cli-runner

## Prerequisites

- Node.js v18 or later
- npm v9 or later
- CLI tools that use system binaries must be installed and available on the system `PATH` (e.g. `docker`). CLIs that ship a local binary (e.g. `git-wrapper`) do **not** need to be on `PATH` — the binary lives inside the CLI's subdirectory under `clis/`.

---

## 1. Install Dependencies

The TypeScript source lives in `src/mcp/cli-runner/`. Install its dependencies once:

```bash
cd src/mcp/cli-runner
npm install
```

---

## 2. Build

Compile the TypeScript source into the production directory:

```bash
# still inside src/mcp/cli-runner/
npm run build
```

This runs `tsc` using [`tsconfig.json`](../../src/mcp/cli-runner/tsconfig.json), which outputs compiled JavaScript to `mcp/cli-runner/dist/`.

After a successful build the production tree looks like:

```
mcp/cli-runner/
├── config.json          ← global server config
├── dist/
│   ├── index.js         ← compiled entry point
│   ├── config.js
│   ├── executor.js
│   └── tool-registry.js
├── clis/
│   ├── docker/
│   │   └── config.json
│   └── git-wrapper/
│       ├── config.json
│       └── git-wrapper.exe
└── package.json
```

---

## 3. Configure Your MCP Client

### Claude Desktop

Add the following block to your Claude Desktop MCP settings file (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "cli-runner": {
      "command": "node",
      "args": ["<absolute-path-to-workspace>/mcp/cli-runner/dist/index.js"]
    }
  }
}
```

Replace `<absolute-path-to-workspace>` with the absolute path to the `mcps` workspace root.

### Roo (VS Code extension)

Add the server to your Roo MCP settings (`.roo/mcp_settings.json` or the workspace settings file):

```json
{
  "mcpServers": {
    "cli-runner": {
      "command": "node",
      "args": ["mcp/cli-runner/dist/index.js"],
      "cwd": "<absolute-path-to-workspace>"
    }
  }
}
```

---

## 4. First Usage

Once the client is configured and restarted, the tools registered from the default CLI configs will be available immediately.

**Example — ask your MCP client:**

> "Run `git_status` with no arguments."

The server will execute `git-wrapper.exe status` in the working directory and return the output as text.

**Example — ask your MCP client:**

> "Run `docker_ps` to list running containers."

The server will execute `docker ps` and return the container list.

---

## 5. Verify the Server Starts Manually

You can test the server outside of a client by running it directly:

```bash
node mcp/cli-runner/dist/index.js
```

If the server starts successfully you will see on stderr:

```
[cli-runner] Server started successfully via stdio transport
```

Press `Ctrl+C` to stop it.
