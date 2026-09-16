package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// Build the real entrypoint; only native identity inspection is stalled by the overlay.
func TestInterruptPreservesInputAndInspectionCleanup(t *testing.T) {
	directory := t.TempDir()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	replacement := filepath.Join(directory, "identity_other.go")
	const stalledIdentity = `package invocation
import ("os"; "strconv"; "time")
func platformApplicationIdentities([]Process) ([]ApplicationIdentity, []string) {
	if err := os.WriteFile(os.Getenv("QFEI_INSPECTION_TEST_PID"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil { panic(err) }
	time.Sleep(time.Minute)
	return nil, nil
}`
	if err := os.WriteFile(replacement, []byte(stalledIdentity), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(root, "internal/invocation/identity_other.go"): replacement,
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	binary := filepath.Join(directory, "contract-cli")
	build := exec.CommandContext(ctx, "go", "build", "-overlay", overlayPath, "-o", binary, "./cmd/contract-cli")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	t.Setenv("CONTRACT_CLI_CONFIG_DIR", filepath.Join(directory, "profile"))
	pidFile := filepath.Join(directory, "helper.pid")
	t.Setenv("QFEI_INSPECTION_TEST_PID", pidFile)

	t.Run("first-interrupt-while-reading-input-pipe", func(t *testing.T) {
		command := exec.CommandContext(ctx, binary, "contract", "search", "--input-file", "/dev/stdin")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		input, err := command.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		defer input.Close()
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- command.Wait() }()
		exited := false
		defer func() {
			if !exited {
				_ = command.Process.Kill()
				<-done
			}
		}()
		// Confirm os.ReadFile has opened the same pipe as stdin before interrupting.
		fds := filepath.Join("/proc", strconv.Itoa(command.Process.Pid), "fd")
		ready := false
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) && !ready {
			stdin, err := os.Readlink(filepath.Join(fds, "0"))
			if err == nil {
				entries, _ := os.ReadDir(fds)
				for _, entry := range entries {
					fd, _ := strconv.Atoi(entry.Name())
					if target, err := os.Readlink(filepath.Join(fds, entry.Name())); fd > 2 && err == nil && target == stdin {
						ready = true
						break
					}
				}
			}
			if !ready {
				time.Sleep(10 * time.Millisecond)
			}
		}
		if !ready {
			t.Fatal("command did not open its JSON input pipe")
		}
		if err := syscall.Kill(-command.Process.Pid, syscall.SIGINT); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-done:
			exited = true
			status := command.ProcessState.Sys().(syscall.WaitStatus)
			if err == nil || !status.Signaled() || status.Signal() != syscall.SIGINT {
				t.Fatalf("first interrupt should terminate blocked input: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("first interrupt was swallowed while reading JSON input")
		}
	})

	t.Run("interrupt-reaps-stalled-inspection", func(t *testing.T) {
		command := exec.CommandContext(ctx, binary, "environment", "inspect")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- command.Wait() }()
		helperPID, exited := 0, false
		defer func() {
			if !exited {
				_ = command.Process.Kill()
				<-done
			}
			if helperPID > 0 {
				_ = syscall.Kill(-helperPID, syscall.SIGKILL)
			}
		}()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if data, err := os.ReadFile(pidFile); err == nil {
				helperPID, _ = strconv.Atoi(string(data))
				if helperPID > 0 {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		if helperPID <= 0 {
			t.Fatal("stalled inspection helper did not start")
		}
		if err := syscall.Kill(-command.Process.Pid, syscall.SIGINT); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-done:
			exited = true
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) || exitError.ExitCode() != 130 {
				t.Fatalf("inspection interrupt should exit after helper cleanup: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("interrupt did not cancel the stalled inspection")
		}
		if err := syscall.Kill(helperPID, 0); !errors.Is(err, syscall.ESRCH) {
			t.Fatalf("inspection helper was not reaped: %v", err)
		}
		helperPID = 0
	})
}
