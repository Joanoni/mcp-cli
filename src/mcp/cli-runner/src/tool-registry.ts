import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { LoadedCliConfig, GlobalConfig } from "./config.js";
import { execute, resolveTimeout } from "./executor.js";

/**
 * Registers one MCP tool per subcommand entry across all CLI configs.
 * Each tool accepts a single `args` parameter (array of strings).
 */
export function registerTools(
  server: McpServer,
  cliConfigs: LoadedCliConfig[],
  globalConfig: GlobalConfig
): void {
  for (const cli of cliConfigs) {
    for (const cmd of cli.commands) {
      const timeoutSeconds = resolveTimeout(
        cmd.timeout,
        cli.timeout,
        globalConfig.defaultTimeout
      );

      // Capture loop variables for the async closure
      const cliCommand = cli.command;
      const cliConfigDir = cli.configDir;
      const subcommand = cmd.subcommand;
      const toolName = cmd.name;
      const toolDescription = `[${cli.description}] ${cmd.description}`;
      const timeout = timeoutSeconds;

      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore — TS2589: type instantiation depth issue with SDK's ZodRawShapeCompat inference
      server.tool(
        toolName,
        toolDescription,
        { args: z.array(z.string()) },
        async (input: { args?: string[] }) => {
          const args: string[] = input.args ?? [];

          const result = await execute({
            command: cliCommand,
            configDir: cliConfigDir,
            subcommand,
            args,
            timeoutSeconds: timeout,
          });

          const parts: string[] = [];

          if (result.timedOut) {
            parts.push(`[cli-runner] Command timed out after ${timeout}s`);
          }

          if (result.stdout.trim()) {
            parts.push(result.stdout.trim());
          }

          if (result.stderr.trim()) {
            parts.push(`STDERR:\n${result.stderr.trim()}`);
          }

          if (result.exitCode !== null && result.exitCode !== 0) {
            parts.push(`Exit code: ${result.exitCode}`);
          }

          const output = parts.join("\n\n") || "(no output)";

          return {
            content: [{ type: "text" as const, text: output }],
          };
        }
      );
    }
  }
}
