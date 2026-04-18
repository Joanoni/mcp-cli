package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var hashRegex = regexp.MustCompile(`\b([0-9a-f]{40})\b`)
var trailingWhitespace = regexp.MustCompile(`[ \t]+$`)
var progressLine = regexp.MustCompile(`^(Counting|Compressing|Writing|Resolving|Enumerating|remote:)\s`)
var percentageLine = regexp.MustCompile(`^[\d\s%|=>\-]+$`)
var binaryDiffLine = regexp.MustCompile(`^Binary files .* differ`)

// Process applies token-reduction rules to git command output based on the subcommand.
func Process(subcommand string, output string) string {
	lines := strings.Split(output, "\n")

	// Remove trailing whitespace from every line (applied to all)
	for i, line := range lines {
		lines[i] = trailingWhitespace.ReplaceAllString(line, "")
	}

	switch subcommand {
	case "log":
		lines = filterBlankLines(lines)
		lines = truncateHashes(lines)
		lines = limitCommits(lines)

	case "diff", "show":
		lines = truncateHashes(lines)
		lines = filterBinaryDiffLines(lines)
		lines = collapseDiffHunks(lines)
		lines = limitDiffFiles(lines, 20)
		lines = removeConsecutiveBlanks(lines)

	case "fetch", "pull", "push":
		lines = removeProgressLines(lines)
		lines = removeConsecutiveBlanks(lines)

	case "branch":
		lines = filterRemoteBranches(lines)
		lines = trimBranchLines(lines)
		lines = truncateHashes(lines)
		lines = removeConsecutiveBlanks(lines)

	case "stash":
		lines = truncateHashes(lines)
		lines = removeConsecutiveBlanks(lines)

	default:
		lines = removeConsecutiveBlanks(lines)
	}

	result := strings.Join(lines, "\n")

	// Hard cap: truncate output exceeding 8000 characters, cut at last newline boundary
	const maxChars = 8000
	if len(result) > maxChars {
		cutAt := strings.LastIndex(result[:maxChars], "\n")
		if cutAt < 0 {
			cutAt = maxChars
		}
		omitted := len(result) - cutAt
		result = result[:cutAt] + fmt.Sprintf("\n[output truncated: %d chars omitted]", omitted)
	}

	return result
}

// truncateHashes replaces 40-char hex hashes with 8-char prefixes.
func truncateHashes(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = hashRegex.ReplaceAllStringFunc(line, func(hash string) string {
			return hash[:8]
		})
	}
	return result
}

// limitCommits limits log output to 50 commits (one line per commit in oneline format)
// unless -n or --max-count was passed in os.Args.
func limitCommits(lines []string) []string {
	if hasMaxCountFlag() {
		return lines
	}

	const maxCommits = 50
	commitCount := 0
	result := []string{}

	for _, line := range lines {
		if line == "" {
			result = append(result, line)
		} else {
			commitCount++
			if commitCount > maxCommits {
				break
			}
			result = append(result, line)
		}
	}

	return result
}

// hasMaxCountFlag checks if -n or --max-count was passed in os.Args.
func hasMaxCountFlag() bool {
	for _, arg := range os.Args {
		if arg == "-n" || strings.HasPrefix(arg, "--max-count") || strings.HasPrefix(arg, "-n=") {
			return true
		}
	}
	return false
}

// collapseDiffHunks collapses diff context blocks with more than 5 unchanged lines.
func collapseDiffHunks(lines []string) []string {
	result := []string{}
	contextLines := []string{}
	inContext := false

	flushContext := func() {
		if len(contextLines) > 5 {
			result = append(result, fmt.Sprintf("[... %d lines omitted ...]", len(contextLines)))
		} else {
			result = append(result, contextLines...)
		}
		contextLines = []string{}
		inContext = false
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "diff ") ||
			strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "+++") {
			if inContext {
				flushContext()
			}
			result = append(result, line)
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			if inContext {
				flushContext()
			}
			result = append(result, line)
		} else {
			// Unchanged context line
			inContext = true
			contextLines = append(contextLines, line)
		}
	}

	if inContext {
		flushContext()
	}

	return result
}

// filterBinaryDiffLines removes lines matching "Binary files .* differ" and their
// preceding diff header lines (diff --git, index, ---, +++).
func filterBinaryDiffLines(lines []string) []string {
	// First pass: collect indices of binary lines
	binaryIndices := map[int]bool{}
	for i, line := range lines {
		if binaryDiffLine.MatchString(line) {
			binaryIndices[i] = true
		}
	}

	if len(binaryIndices) == 0 {
		return lines
	}

	// Second pass: for each binary line, walk backwards to mark header lines for removal
	removeIndices := map[int]bool{}
	for idx := range binaryIndices {
		removeIndices[idx] = true
		// Walk backwards removing contiguous header lines
		for j := idx - 1; j >= 0; j-- {
			line := lines[j]
			if strings.HasPrefix(line, "diff --git") ||
				strings.HasPrefix(line, "index ") ||
				strings.HasPrefix(line, "--- ") ||
				strings.HasPrefix(line, "+++ ") {
				removeIndices[j] = true
			} else {
				break
			}
		}
	}

	result := []string{}
	for i, line := range lines {
		if removeIndices[i] {
			continue
		}
		result = append(result, line)
	}
	return result
}

// limitDiffFiles limits diff output to maxFiles files, appending a summary for the rest.
func limitDiffFiles(lines []string, maxFiles int) []string {
	fileCount := 0
	result := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			fileCount++
			if fileCount > maxFiles {
				// Count remaining diff --git headers to report omitted count
				remaining := 1 // current one already counted
				for _, l := range lines[len(result)+1:] {
					if strings.HasPrefix(l, "diff --git") {
						remaining++
					}
				}
				result = append(result, fmt.Sprintf("[%d more files omitted]", remaining))
				return result
			}
		}
		result = append(result, line)
	}

	return result
}

// filterBlankLines removes all blank lines from the slice.
func filterBlankLines(lines []string) []string {
	result := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

// filterRemoteBranches removes lines starting with "  remotes/" when -a or -r is not in args.
func filterRemoteBranches(lines []string) []string {
	showRemotes := false
	for _, arg := range os.Args {
		if arg == "-a" || arg == "-r" {
			showRemotes = true
			break
		}
	}
	if showRemotes {
		return lines
	}

	result := []string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "  remotes/") {
			continue
		}
		result = append(result, line)
	}
	return result
}

// removeProgressLines removes fetch/pull/push progress noise.
func removeProgressLines(lines []string) []string {
	result := []string{}
	for _, line := range lines {
		if strings.ContainsRune(line, '\r') {
			continue
		}
		if progressLine.MatchString(line) {
			continue
		}
		if percentageLine.MatchString(strings.TrimSpace(line)) && strings.TrimSpace(line) != "" {
			continue
		}
		result = append(result, line)
	}
	return result
}

// trimBranchLines removes extra leading/trailing whitespace per line.
func trimBranchLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		// Preserve the leading * for current branch indicator
		trimmed := strings.TrimRight(line, " \t")
		result[i] = trimmed
	}
	return result
}

// removeConsecutiveBlanks removes consecutive blank lines (keeps at most one).
func removeConsecutiveBlanks(lines []string) []string {
	result := []string{}
	prevBlank := false
	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}
	return result
}
