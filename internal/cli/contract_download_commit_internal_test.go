package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCommitDownloadedFileFallsBackWhenHardLinksAreUnavailable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	temporaryPath := filepath.Join(dir, "download.part")
	outputPath := filepath.Join(dir, "download.pdf")
	if err := os.WriteFile(temporaryPath, []byte("complete download"), 0o600); err != nil {
		t.Fatalf("WriteFile(temporary) error = %v", err)
	}

	err := commitDownloadedFileWithLink(temporaryPath, outputPath, false, func(_, _ string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: errors.ErrUnsupported}
	})
	if err != nil {
		t.Fatalf("commitDownloadedFileWithLink() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(output) error = %v", err)
	}
	if string(content) != "complete download" {
		t.Fatalf("output content = %q", string(content))
	}
}

func TestCommitDownloadedFileDoesNotOverwriteConcurrentOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	temporaryPath := filepath.Join(dir, "download.part")
	outputPath := filepath.Join(dir, "download.pdf")
	if err := os.WriteFile(temporaryPath, []byte("new download"), 0o600); err != nil {
		t.Fatalf("WriteFile(temporary) error = %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("concurrent output"), 0o600); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}

	err := commitDownloadedFileWithLink(temporaryPath, outputPath, false, func(_, _ string) error {
		return &os.LinkError{Op: "link", Old: temporaryPath, New: outputPath, Err: errors.ErrUnsupported}
	})
	if err == nil {
		t.Fatal("commitDownloadedFileWithLink() error = nil")
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile(output) error = %v", readErr)
	}
	if string(content) != "concurrent output" {
		t.Fatalf("output content = %q", string(content))
	}
}
