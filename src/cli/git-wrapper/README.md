# git-wrapper

A token-optimized Git CLI wrapper written in Go. It proxies all `git` commands while compressing verbose output to reduce token usage when used with AI tools like MCP.

## What it does

`git-wrapper` acts as a transparent proxy to `git`, forwarding all arguments and exit codes exactly. Before returning output, it applies subcommand-specific compression rules:

| Subcommand(s)         | Compression applied |
|-----------------------|---------------------|
| `log`                 | Truncates 40-char hashes to 8 chars; limits to 50 commits unless `-n`/`--max-count` is set; removes consecutive blank lines |
| `diff`, `show`        | Truncates hashes; collapses unchanged context blocks >20 lines into `[... N lines omitted ...]`; removes consecutive blank lines |
| `fetch`, `pull`, `push` | Removes progress/noise lines (percentages, `remote:`, `Counting`, etc.) |
| `branch`              | Trims trailing whitespace; truncates hashes |
| `stash`               | Truncates hashes |
| All                   | Removes trailing whitespace; collapses 3+ consecutive blank lines into one |

stderr is always passed through unchanged. Exit codes from `git` are forwarded exactly.

## How to build

### Windows
```powershell
cd src/cli/git-wrapper
go build -o bin/git-wrapper.exe .
```

### Linux / macOS
```bash
cd src/cli/git-wrapper
go build -o bin/git-wrapper .
```

Requires Go 1.21 or later.

## How to install

After building, add the `bin/` directory to your system PATH, **or** copy the binary to a directory already on PATH.

### Windows (PowerShell, permanent)
```powershell
$env:PATH += ";$PWD\bin"
# Or add via System Properties > Environment Variables
```

### Linux / macOS
```bash
export PATH="$PATH:/path/to/src/cli/git-wrapper/bin"
# Or copy to /usr/local/bin:
cp bin/git-wrapper /usr/local/bin/
```

## Integration with cli-runner

The MCP `cli-runner` server reads `mcp/cli-runner/clis/git-wrapper.json`, which registers each git subcommand as an MCP tool (e.g. `git_status`, `git_log`, `git_diff`).

The `command` field is set to `"git-wrapper"` (no path), so the binary must be on the system PATH before starting the MCP server.

Each tool call passes the subcommand and any additional arguments to `git-wrapper`, which forwards them to `git` and returns compressed output.
