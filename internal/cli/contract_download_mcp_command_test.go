package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

func TestContractDownloadFileAsUserWritesOnlyAfterValidation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "contract.pdf")
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			switch requests {
			case 1:
				if req.URL.Path != "/open-apis/contract/v1/mcp/contracts/contract-1/files/file-1/download" {
					t.Fatalf("metadata path = %q", req.URL.Path)
				}
				if req.URL.RawQuery != "" {
					t.Fatalf("metadata query = %q", req.URL.RawQuery)
				}
				if req.Header.Get("Authorization") != "Bearer user-token" {
					t.Fatalf("metadata authorization = %q", req.Header.Get("Authorization"))
				}
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","file_name":"contract.pdf","file_size":14,"download_url":"https://files.example.test/object?signature=secret"}}`), nil
			case 2:
				if req.URL.String() != "https://files.example.test/object?signature=secret" {
					t.Fatalf("signed URL = %q", req.URL.String())
				}
				if req.Header.Get("Authorization") != "" {
					t.Fatalf("signed download leaked authorization = %q", req.Header.Get("Authorization"))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("download bytes")),
				}, nil
			default:
				t.Fatalf("unexpected request %d: %s", requests, req.URL.String())
				return nil, nil
			}
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--output-file", outputPath,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(output) error = %v", err)
	}
	if string(content) != "download bytes" {
		t.Fatalf("downloaded content = %q", string(content))
	}
	if matches, err := filepath.Glob(outputPath + ".part-*"); err != nil || len(matches) != 0 {
		t.Fatalf("temporary files = %v, error = %v", matches, err)
	}
	if !strings.Contains(stdout.String(), "Downloaded file to "+outputPath) {
		t.Fatalf("missing download message: %s", stdout.String())
	}
}

func TestContractDownloadFileAsUserKeepsExistingFileWhenSignedDownloadFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "contract.pdf")
	if err := os.WriteFile(outputPath, []byte("original bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if requests == 1 {
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","file_size":100,"download_url":"https://files.example.test/object?signature=secret"}}`), nil
			}
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("storage unavailable")),
			}, nil
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--output-file", outputPath,
		"--force",
	})
	if err == nil || !strings.Contains(err.Error(), "signed file download failed with status 502") {
		t.Fatalf("Run() error = %v", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile(output) error = %v", readErr)
	}
	if string(content) != "original bytes" {
		t.Fatalf("existing content changed to %q", string(content))
	}
	if matches, globErr := filepath.Glob(outputPath + ".part-*"); globErr != nil || len(matches) != 0 {
		t.Fatalf("temporary files = %v, error = %v", matches, globErr)
	}
}

func TestContractDownloadFileAsUserReplacesExistingFileAfterSuccessfulDownload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "contract.pdf")
	if err := os.WriteFile(outputPath, []byte("original bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if requests == 1 {
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","file_size":14,"download_url":"https://files.example.test/object?signature=secret"}}`), nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("download bytes")),
			}, nil
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--output-file", outputPath,
		"--force",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile(output) error = %v", readErr)
	}
	if string(content) != "download bytes" {
		t.Fatalf("output content = %q", string(content))
	}
	if matches, globErr := filepath.Glob(outputPath + ".part-*"); globErr != nil || len(matches) != 0 {
		t.Fatalf("temporary files = %v, error = %v", matches, globErr)
	}
}

func TestContractDownloadFileAsUserRequiresContractBeforeHTTP(t *testing.T) {
	t.Parallel()

	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			return jsonResponse(`{"code":0}`), nil
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--profile", "contract",
		"--as", "user",
		"--output-file", filepath.Join(t.TempDir(), "contract.pdf"),
	})
	if err == nil || !strings.Contains(err.Error(), "--contract is required with --as user") {
		t.Fatalf("Run() error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("missing contract sent %d HTTP requests", requests)
	}
}

func TestContractDownloadFileAsUserStopsOnMetadataBusinessError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "contract.pdf")
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			return jsonResponse(`{"code":110125,"msg":"not visible","data":null,"success":false}`), nil
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--output-file", outputPath,
	})
	if err == nil || !strings.Contains(err.Error(), "code 110125") {
		t.Fatalf("Run() error = %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("output should not exist, stat error = %v", statErr)
	}
}

func TestContractDownloadFileAsUserDoesNotExposeSignedURLOnTransportFailure(t *testing.T) {
	t.Parallel()

	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	stderr := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: stderr,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if requests == 1 {
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","file_size":14,"download_url":"https://files.example.test/object?signature=do-not-log"}}`), nil
			}
			return nil, errors.New(`Get "https://files.example.test/object?signature=do-not-log": connection failed`)
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--raw",
	})
	if err == nil {
		t.Fatal("Run() error = nil")
	}
	if strings.Contains(err.Error(), "do-not-log") || strings.Contains(stderr.String(), "do-not-log") {
		t.Fatalf("signed URL leaked; error = %v, logs = %s", err, stderr.String())
	}
}

func TestContractDownloadFileAsUserRejectsInsecureRedirect(t *testing.T) {
	t.Parallel()

	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			switch requests {
			case 1:
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","download_url":"https://files.example.test/object?signature=secret"}}`), nil
			case 2:
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": {"http://insecure.example.test/object"}},
					Body:       io.NopCloser(strings.NewReader("redirect")),
				}, nil
			default:
				t.Fatalf("insecure redirect was followed: %s", req.URL.String())
				return nil, nil
			}
		})},
	})

	err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--raw",
	})
	if err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("Run() error = %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestContractDownloadFileAsUserDoesNotForwardSignedURLAsRedirectReferer(t *testing.T) {
	t.Parallel()

	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	requests := 0
	app := cli.New(cli.Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			switch requests {
			case 1:
				return jsonResponse(`{"code":0,"success":true,"data":{"file_id":"file-1","file_size":14,"download_url":"https://files.example.test/object?signature=do-not-forward"}}`), nil
			case 2:
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": {"https://cdn.example.test/object"}},
					Body:       io.NopCloser(strings.NewReader("redirect")),
				}, nil
			case 3:
				if referer := req.Header.Get("Referer"); referer != "" {
					t.Fatalf("redirect referer = %q, want empty", referer)
				}
				if req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" {
					t.Fatalf("redirect forwarded credentials: %v", req.Header)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("download bytes")),
				}, nil
			default:
				t.Fatalf("unexpected request %d", requests)
				return nil, nil
			}
		})},
	})

	if err := app.Run(context.Background(), []string{
		"contract", "download-file", "file-1",
		"--contract", "contract-1",
		"--profile", "contract",
		"--as", "user",
		"--raw",
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want 3", requests)
	}
}
