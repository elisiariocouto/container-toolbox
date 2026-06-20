package docker

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ComposeAvailable reports whether the `docker` binary and the `docker compose`
// v2 subcommand are usable. The returned error explains what is missing.
func ComposeAvailable() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("`docker` not found in PATH")
	}
	if out, err := exec.Command("docker", "compose", "version").CombinedOutput(); err != nil {
		reason := strings.TrimSpace(string(out))
		if reason == "" {
			reason = err.Error()
		}
		return fmt.Errorf("`docker compose` (v2) unavailable: %s", reason)
	}
	return nil
}

// composeCmd builds a `docker compose <args...>` command rooted in dir, so
// Compose auto-detects the compose file and derives the project name the same
// way the SDK labels report it.
func composeCmd(ctx context.Context, dir string, args ...string) *exec.Cmd {
	full := append([]string{"compose"}, args...)
	cmd := exec.CommandContext(ctx, "docker", full...)
	cmd.Dir = dir
	return cmd
}

// run executes a compose subcommand and returns combined stdout+stderr. On
// failure the output is wrapped into the error so callers can surface why.
func run(ctx context.Context, dir string, args ...string) error {
	out, err := composeCmd(ctx, dir, args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return err
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// Up runs `docker compose up -d`.
func Up(ctx context.Context, dir string) error { return run(ctx, dir, "up", "-d") }

// Down runs `docker compose down`.
func Down(ctx context.Context, dir string) error { return run(ctx, dir, "down") }

// Restart runs `docker compose restart`.
func Restart(ctx context.Context, dir string) error { return run(ctx, dir, "restart") }

// PullUp pulls newer images and recreates the stack: `pull` then `up -d`.
func PullUp(ctx context.Context, dir string) error {
	if err := run(ctx, dir, "pull"); err != nil {
		return err
	}
	return run(ctx, dir, "up", "-d")
}

// StreamLogs starts `docker compose logs -f --tail=200` and pushes each output
// line into ch. ch is closed when the stream ends (EOF or context cancel).
// Cancel the context passed in to stop streaming and kill the process.
func StreamLogs(ctx context.Context, dir string, ch chan<- string) error {
	cmd := composeCmd(ctx, dir, "logs", "-f", "--tail=200", "--no-color")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout // StdoutPipe set cmd.Stdout to the pipe; fold stderr into it too
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case ch <- scanner.Text():
			case <-ctx.Done():
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				return
			}
		}
		// A scan error (read failure, or a line over the 1MB cap) otherwise
		// ends the stream silently; surface it as a final line so the user
		// sees why logs stopped. Suppress it on a cancelled context, where the
		// read is expected to fail.
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			select {
			case ch <- fmt.Sprintf("[stream error: %s]", err):
			case <-ctx.Done():
			}
		}
		_ = cmd.Wait()
	}()

	return nil
}
