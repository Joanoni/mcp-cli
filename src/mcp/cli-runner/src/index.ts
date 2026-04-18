import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import * as path from "path";
import { loadGlobalConfig, loadCliConfigs } from "./config.js";
import { registerTools } from "./tool-registry.js";

async function main(): Promise<void> {
  // Resolve config.json relative to this file's location (dist/index.js → project root)
  const projectRoot = path.resolve(__dirname, "..");
  const configPath = path.join(projectRoot, "config.json");

  // Load and validate global config
  const globalConfig = loadGlobalConfig(configPath);

  // Resolve CLIs directory relative to project root
  const clisDirectory = path.resolve(projectRoot, globalConfig.clisDirectory);

  // Load and validate CLI configs (invalid files are skipped with a warning)
  const cliConfigs = loadCliConfigs(clisDirectory);

  if (cliConfigs.length === 0) {
    console.warn("[cli-runner] WARNING: No valid CLI configs found. No tools will be registered.");
  }

  // Create MCP server
  const server = new McpServer({
    name: "cli-runner",
    version: "1.0.0",
  });

  // Register one tool per subcommand
  registerTools(server, cliConfigs, globalConfig);

  // Start server with stdio transport
  const transport = new StdioServerTransport();
  await server.connect(transport);

  console.error("[cli-runner] Server started successfully via stdio transport");
}

main().catch((err) => {
  console.error("[cli-runner] Fatal error:", err);
  process.exit(1);
});
