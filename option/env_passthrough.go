// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package option

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The environment-variable passthrough logic here is ported from the Finch CLI:
// https://github.com/runfinch/finch/blob/main/cmd/finch/nerdctl_remote.go

func resolveEnvPassthrough(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--env-file"):
			var (
				filename     string
				consumedNext bool
			)
			switch {
			case arg == "--env-file":
				if i+1 >= len(args) {
					out = append(out, arg)
					continue
				}
				filename = args[i+1]
				consumedNext = true
			case strings.HasPrefix(arg, "--env-file="):
				filename = arg[len("--env-file="):]
			default:
				out = append(out, arg)
				continue
			}

			expanded, err := expandEnvFile(filename)
			if err != nil {
				out = append(out, arg)
				if consumedNext {
					out = append(out, args[i+1])
					i++
				}
				continue
			}
			out = append(out, expanded...)
			if consumedNext {
				i++
			}
		case arg == "-e" || arg == "--env":
			if i+1 >= len(args) {
				out = append(out, arg)
				continue
			}
			val := args[i+1]
			i++
			if resolved, ok := resolveEnvValue(val); ok {
				out = append(out, arg, resolved)
			}
		case strings.HasPrefix(arg, "--env="):
			if resolved, ok := resolveEnvValue(arg[len("--env="):]); ok {
				out = append(out, "--env="+resolved)
			}
		case strings.HasPrefix(arg, "-e") && arg != "-e":
			// inline form: -eVAR or -eVAR=value
			if resolved, ok := resolveEnvValue(arg[len("-e"):]); ok {
				out = append(out, "-e"+resolved)
			}
		default:
			out = append(out, arg)
		}
	}
	return out
}

func resolveEnvValue(v string) (string, bool) {
	if strings.Contains(v, "=") {
		return v, true
	}
	if val, ok := os.LookupEnv(v); ok {
		return fmt.Sprintf("%s=%s", v, val), true
	}
	return "", false
}

func expandEnvFile(filename string) ([]string, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck // read-only; closed on process exit anyway.

	var envs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case len(line) == 0, strings.HasPrefix(line, "#"):
			continue
		case strings.Contains(line, "="):
			envs = append(envs, "-e", line)
		default:
			if val, ok := os.LookupEnv(line); ok {
				envs = append(envs, "-e", fmt.Sprintf("%s=%s", line, val))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return envs, nil
}
