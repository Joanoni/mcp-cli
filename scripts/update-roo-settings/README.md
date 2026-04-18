# update-roo-settings sync script

## Purpose

Syncs `scripts/update-roo-settings/mcp_settings.json` to `.roo/mcp.json`, keeping Roo's active MCP configuration up to date.

## Usage

```bash
node scripts/update-roo-settings/update.js
```

## When to Run

Run this script after adding or removing any MCP server entry in `mcp_settings.json`.

## What It Does

1. Copies `mcp_settings.json` to `.roo/mcp.json`
2. Reads both files back from disk
3. Compares their contents (string equality)
4. Prints `✅ Success: .roo/mcp.json is up to date` and exits 0 if they match
5. Prints `❌ Error: files differ after copy` and exits 1 if they differ

No interactive input is required — the script is fully non-interactive.

## Source → Destination

| Role | Path |
|---|---|
| Source | `scripts/update-roo-settings/mcp_settings.json` |
| Destination | `.roo/mcp.json` |

## Conventions

### `alwaysAllow`

Every MCP entry in `mcp_settings.json` **must** include an `alwaysAllow` array listing **all tool names** exposed by that MCP. This prevents Roo from prompting for confirmation on every tool call.

**Example:**

```json
"hello-world": {
  "command": "node",
  "args": ["C:/Projects/mcps/mcp/hello-world/index.js"],
  "alwaysAllow": [
    "greet"
  ]
}
```

When adding a new MCP:
1. List every tool name the MCP exposes in the `alwaysAllow` array.
2. Run `node scripts/update-roo-settings/update.js` to sync to `.roo/mcp.json`.
