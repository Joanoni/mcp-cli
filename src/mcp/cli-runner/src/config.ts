import { z } from "zod";
import * as fs from "fs";
import * as path from "path";

// ── Zod Schemas ──────────────────────────────────────────────────────────────

export const GlobalConfigSchema = z.object({
  defaultTimeout: z.number().positive(),
  clisDirectory: z.string().min(1),
});

export const CommandConfigSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(1),
  subcommand: z.string().min(1),
  timeout: z.number().positive().optional(),
});

export const CliConfigSchema = z.object({
  name: z.string().min(1),
  command: z.string().min(1),
  description: z.string().min(1),
  timeout: z.number().positive().optional(),
  commands: z.array(CommandConfigSchema).min(1),
});

// ── Types ─────────────────────────────────────────────────────────────────────

export type GlobalConfig = z.infer<typeof GlobalConfigSchema>;
export type CommandConfig = z.infer<typeof CommandConfigSchema>;
export type CliConfig = z.infer<typeof CliConfigSchema>;

/**
 * Extends CliConfig with the absolute path of the CLI's own subdirectory.
 * Injected programmatically after validation — not part of the JSON schema.
 */
export interface LoadedCliConfig extends CliConfig {
  configDir: string;
}

// ── Loaders ───────────────────────────────────────────────────────────────────

/**
 * Reads and validates the global config.json file.
 * Throws a descriptive error if validation fails.
 */
export function loadGlobalConfig(configPath: string): GlobalConfig {
  const raw = fs.readFileSync(configPath, "utf-8");
  const parsed = JSON.parse(raw);
  const result = GlobalConfigSchema.safeParse(parsed);

  if (!result.success) {
    throw new Error(
      `Invalid global config at "${configPath}":\n${result.error.toString()}`
    );
  }

  return result.data;
}

/**
 * Scans subdirectories of the given directory and reads config.json from each.
 * Validates each config against CliConfigSchema and injects configDir.
 * Invalid or missing config files are logged as warnings and skipped.
 */
export function loadCliConfigs(clisDirectory: string): LoadedCliConfig[] {
  const resolvedDir = path.resolve(clisDirectory);

  if (!fs.existsSync(resolvedDir)) {
    throw new Error(`CLIs directory not found: "${resolvedDir}"`);
  }

  const entries = fs.readdirSync(resolvedDir, { withFileTypes: true });
  const subdirs = entries.filter((e) => e.isDirectory()).map((e) => e.name);

  const configs: LoadedCliConfig[] = [];

  for (const subdir of subdirs) {
    const subdirPath = path.join(resolvedDir, subdir);
    const configFilePath = path.join(subdirPath, "config.json");

    try {
      if (!fs.existsSync(configFilePath)) {
        console.warn(
          `[cli-runner] WARNING: No config.json found in "${subdirPath}", skipping.`
        );
        continue;
      }

      const raw = fs.readFileSync(configFilePath, "utf-8");
      const parsed = JSON.parse(raw);
      const result = CliConfigSchema.safeParse(parsed);

      if (!result.success) {
        console.warn(
          `[cli-runner] WARNING: Skipping invalid CLI config in "${subdir}":\n${result.error.toString()}`
        );
        continue;
      }

      configs.push({ ...result.data, configDir: path.resolve(subdirPath) });
    } catch (err) {
      console.warn(
        `[cli-runner] WARNING: Failed to read/parse CLI config in "${subdir}": ${(err as Error).message}`
      );
    }
  }

  return configs;
}
