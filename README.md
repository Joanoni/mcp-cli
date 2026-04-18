# mcps — MCP Servers Workspace

A monorepo containing Model Context Protocol (MCP) servers. The first server, **cli-runner**, reads CLI configuration JSON files and dynamically exposes each configured subcommand as an MCP tool callable by any MCP-compatible client (e.g. Claude Desktop, Roo).

## Prerequisites

- [Node.js](https://nodejs.org/) v18 or later
- npm v9 or later

## Quick Start

```bash
# 1. Install dependencies for the TypeScript source
cd src/mcp/cli-runner
npm install

# 2. Build — compiles TypeScript into mcp/cli-runner/dist/
npm run build

# 3. Point your MCP client at the production server
#    Entry point: mcp/cli-runner/dist/index.js
```

See [docs/cli-runner/getting-started.md](docs/cli-runner/getting-started.md) for full client configuration instructions.

## Table of Contents

| Document | Description |
|---|---|
| [docs/overview.md](docs/overview.md) | Architecture overview and directory roles |
| [docs/cli-runner/getting-started.md](docs/cli-runner/getting-started.md) | Installation, build, and first usage |
| [docs/cli-runner/config-reference.md](docs/cli-runner/config-reference.md) | Full schema reference for `config.json` and CLI config files |
| [docs/cli-runner/adding-clis.md](docs/cli-runner/adding-clis.md) | Step-by-step guide to adding new CLI tools |
| [docs/development/project-structure.md](docs/development/project-structure.md) | Annotated workspace directory tree |
| [docs/development/build-and-deploy.md](docs/development/build-and-deploy.md) | TypeScript compilation and deployment workflow |
