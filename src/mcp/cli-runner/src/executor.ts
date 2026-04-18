import { spawn } from "child_process";
import * as path from "path";

export interface ExecuteOptions {
  command: string;
  configDir: string;
  subcommand: string;
  args: string[];
  timeoutSeconds: number;
}

export interface ExecuteResult {
  stdout: string;
  stderr: string;
  exitCode: number | null;
  timedOut: boolean;
}

/**
 * Executes a CLI command using child_process.spawn.
 * Captures stdout and stderr, and enforces a timeout.
 * If command starts with './' or '../', it is resolved relative to configDir.
 */
export function execute(options: ExecuteOptions): Promise<ExecuteResult> {
  const { command, configDir, subcommand, args, timeoutSeconds } = options;

  const resolvedCommand =
    command.startsWith("./") || command.startsWith("../")
      ? path.resolve(configDir, command)
      : command;

  const fullArgs = [subcommand, ...args];

  return new Promise((resolve) => {
    const child = spawn(resolvedCommand, fullArgs, {
      shell: false,
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";
    let timedOut = false;

    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGTERM");

      // Force kill after 5 seconds if SIGTERM is not enough
      setTimeout(() => {
        try {
          child.kill("SIGKILL");
        } catch {
          // Process may have already exited
        }
      }, 5000);
    }, timeoutSeconds * 1000);

    child.stdout.on("data", (chunk: Buffer) => {
      stdout += chunk.toString();
    });

    child.stderr.on("data", (chunk: Buffer) => {
      stderr += chunk.toString();
    });

    child.on("close", (exitCode) => {
      clearTimeout(timer);
      resolve({ stdout, stderr, exitCode, timedOut });
    });

    child.on("error", (err) => {
      clearTimeout(timer);
      resolve({
        stdout,
        stderr: stderr + `\nProcess error: ${err.message}`,
        exitCode: null,
        timedOut,
      });
    });
  });
}

/**
 * Resolves the effective timeout using cascading priority:
 * command.timeout → cli.timeout → config.defaultTimeout
 */
export function resolveTimeout(
  commandTimeout: number | undefined,
  cliTimeout: number | undefined,
  defaultTimeout: number
): number {
  return commandTimeout ?? cliTimeout ?? defaultTimeout;
}
