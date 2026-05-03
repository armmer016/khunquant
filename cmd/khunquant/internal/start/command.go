package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	var port string
	var public bool
	var noBrowser bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch gateway and open web UI",
		Long:  "Start khunquant-launcher: launches the gateway and opens the web management UI in your browser.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			launcherPath := findLauncherBinary()

			args := []string{}
			if port != "" {
				args = append(args, "-port", port)
			}
			if public {
				args = append(args, "-public")
			}
			if noBrowser {
				args = append(args, "-no-browser")
			}

			return execLauncher(launcherPath, args)
		},
	}

	cmd.Flags().StringVar(&port, "port", "", "Port for the web UI (default: 18800)")
	cmd.Flags().BoolVar(&public, "public", false, "Listen on all interfaces instead of localhost only")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Do not auto-open browser on startup")

	return cmd
}

// findLauncherBinary locates the khunquant-launcher executable.
// Search order:
//  1. KHUNQUANT_LAUNCHER environment variable
//  2. Same directory as the current executable
//  3. Falls back to "khunquant-launcher" on $PATH
func findLauncherBinary() string {
	name := "khunquant-launcher"
	if runtime.GOOS == "windows" {
		name = "khunquant-launcher.exe"
	}

	if p := os.Getenv("KHUNQUANT_LAUNCHER"); p != "" {
		if info, _ := os.Stat(p); info != nil && !info.IsDir() {
			return p
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	return name
}

// execLauncher replaces the current process with the launcher on Unix,
// or runs it as a child process on Windows.
func execLauncher(launcherPath string, args []string) error {
	resolved, err := exec.LookPath(launcherPath)
	if err != nil {
		return fmt.Errorf("khunquant-launcher not found: build it with 'make build-launcher'\n  looked for: %s", launcherPath)
	}

	cmd := exec.Command(resolved, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
