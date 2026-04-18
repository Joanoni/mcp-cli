package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: git-wrapper <subcommand> [args...]")
		os.Exit(1)
	}

	subcommand := args[0]
	gitArgs := args // pass all args including subcommand to git

	// Inject flags for status subcommand
	if subcommand == "status" {
		hasShort := false
		for _, a := range gitArgs {
			if a == "--short" || a == "-s" {
				hasShort = true
				break
			}
		}
		if !hasShort {
			gitArgs = append([]string{"status", "--short", "--branch"}, gitArgs[1:]...)
		}
	}

	// Inject flags for log subcommand
	if subcommand == "log" {
		hasFormat := false
		for _, a := range gitArgs {
			if a == "--oneline" || a == "--format" || a == "--pretty" ||
				strings.HasPrefix(a, "--format=") || strings.HasPrefix(a, "--pretty=") {
				hasFormat = true
				break
			}
		}
		if !hasFormat {
			gitArgs = append([]string{"log", "--oneline", "--decorate"}, gitArgs[1:]...)
		}
	}

	// Inject flags for show subcommand
	if subcommand == "show" {
		hasStat := false
		for _, a := range gitArgs {
			if a == "--stat" || a == "--format" || a == "--pretty" ||
				strings.HasPrefix(a, "--format=") || strings.HasPrefix(a, "--pretty=") ||
				a == "-p" || a == "--patch" || a == "--name-only" || a == "--name-status" {
				hasStat = true
				break
			}
		}
		if !hasStat {
			gitArgs = append([]string{"show", "--stat"}, gitArgs[1:]...)
		}
	}

	cmd := exec.Command("git", gitArgs...)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	// Process stderr: filter progress lines and append to stdout
	stderrLines := strings.Split(stderrBuf.String(), "\n")
	stderrLines = removeProgressLines(stderrLines)
	stderrOut := strings.Join(stderrLines, "\n")

	combined := stdoutBuf.String()
	if strings.TrimSpace(stderrOut) != "" {
		if combined != "" && !strings.HasSuffix(combined, "\n") {
			combined += "\n"
		}
		combined += stderrOut
	}

	processed := Process(subcommand, combined)
	fmt.Fprint(os.Stdout, processed)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ProcessState.ExitCode())
		}
		os.Exit(1)
	}
}
