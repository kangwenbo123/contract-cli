package cli_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

func TestContractUploadFileCommandAcceptsLegacyBotIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	uploadPath := filepath.Join(dir, "财务合同.docx")
	if err := os.WriteFile(uploadPath, []byte("contract file bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Store:  store,
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s", req.Method)
				}
				if req.URL.Path != "/open-apis/contract/v1/files/upload" {
					t.Fatalf("path = %s", req.URL.Path)
				}
				if req.URL.Query().Get("user_id") != "ou_123" {
					t.Fatalf("user_id = %q", req.URL.Query().Get("user_id"))
				}
				if req.URL.Query().Get("user_id_type") != "employee_id" {
					t.Fatalf("user_id_type = %q", req.URL.Query().Get("user_id_type"))
				}
				if req.Header.Get("Authorization") != "Bearer app-token" {
					t.Fatalf("authorization = %q", req.Header.Get("Authorization"))
				}
				if got := req.Header.Get("Content-Type"); !strings.HasPrefix(got, "multipart/form-data; boundary=") {
					t.Fatalf("content-type = %q", got)
				}
				assertUploadMultipart(t, req, "财务合同.docx", "text", "contract file bytes")
				return jsonResponse(`{"code":0,"data":{"file_id":"file-123"},"msg":"success"}`), nil
			}),
		},
	})

	err := app.Run(context.Background(), []string{
		"contract", "upload-file",
		"--profile", "contract",
		"--as", "bot",
		"--file", uploadPath,
		"--file-type", "text",
		"--user-id", "ou_123",
		"--user-id-type", "employee_id",
		"--output", "json",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(stdout.String(), `"file_id": "file-123"`) {
		t.Fatalf("missing file_id in output: %s", stdout.String())
	}
}

func TestContractUploadFileCommandUsesDefaultAppIdentityAndOverridesFileName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	uploadPath := filepath.Join(dir, "local-name.pdf")
	if err := os.WriteFile(uploadPath, []byte("pdf bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store := config.NewStore(dir)
	if err := store.UpsertProfile(uploadProfile(config.IdentityApp), true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Store:  store,
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("Authorization") != "Bearer app-token" {
					t.Fatalf("authorization = %q", req.Header.Get("Authorization"))
				}
				assertUploadMultipart(t, req, "附件.pdf", "attachment", "pdf bytes")
				return jsonResponse(`{"code":0,"data":{"file_id":"file-456"},"msg":"success"}`), nil
			}),
		},
	})

	err := app.Run(context.Background(), []string{
		"contract", "upload-file",
		"--profile", "contract",
		"--file", uploadPath,
		"--file-type", "attachment",
		"--file-name", "附件.pdf",
		"--raw",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != `{"code":0,"data":{"file_id":"file-456"},"msg":"success"}` {
		t.Fatalf("raw output = %s", stdout.String())
	}
}

func TestContractUploadFileCommandUploadsAsUser(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		args           []string
		wantUserID     string
		wantUserIDType string
	}{
		{
			name:           "explicit user",
			args:           []string{"contract", "upload-file", "--profile", "contract", "--as", "user", "--user-id", "ou_user_1", "--user-id-type", "employee_id"},
			wantUserID:     "ou_user_1",
			wantUserIDType: "employee_id",
		},
		{
			name:           "default user",
			args:           []string{"contract", "upload-file", "--profile", "contract"},
			wantUserIDType: "user_id",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			uploadPath := filepath.Join(dir, "合同.pdf")
			if err := os.WriteFile(uploadPath, []byte("pdf bytes"), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			store := config.NewStore(dir)
			if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
				t.Fatalf("UpsertProfile() error = %v", err)
			}

			stdout := &bytes.Buffer{}
			app := cli.New(cli.Options{
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
				Store:  store,
				HTTPClient: &http.Client{
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodPost {
							t.Fatalf("method = %s", req.Method)
						}
						if req.URL.Path != "/open-apis/contract/v1/files/upload" {
							t.Fatalf("path = %s", req.URL.Path)
						}
						if req.Header.Get("Authorization") != "Bearer user-token" {
							t.Fatalf("authorization = %q", req.Header.Get("Authorization"))
						}
						if req.URL.Query().Get("user_id") != tc.wantUserID {
							t.Fatalf("user_id = %q", req.URL.Query().Get("user_id"))
						}
						if req.URL.Query().Get("user_id_type") != tc.wantUserIDType {
							t.Fatalf("user_id_type = %q", req.URL.Query().Get("user_id_type"))
						}
						if got := req.Header.Get("Content-Type"); !strings.HasPrefix(got, "multipart/form-data; boundary=") {
							t.Fatalf("content-type = %q", got)
						}
						assertUploadMultipart(t, req, "合同.pdf", "text", "pdf bytes")
						return jsonResponse(`{"code":0,"data":{"file_id":"user-file-123"},"msg":"success"}`), nil
					}),
				},
			})

			args := append([]string{}, tc.args...)
			args = append(args, "--file", uploadPath, "--file-type", "text", "--output", "json")
			err := app.Run(context.Background(), args)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if !strings.Contains(stdout.String(), `"file_id": "user-file-123"`) {
				t.Fatalf("missing file_id in output: %s", stdout.String())
			}
		})
	}
}

func TestContractUploadFileCommandValidationErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	uploadPath := filepath.Join(dir, "合同.pdf")
	if err := os.WriteFile(uploadPath, []byte("pdf bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	tooLargePath := filepath.Join(dir, "too-large.pdf")
	tooLarge, err := os.Create(tooLargePath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := tooLarge.Truncate(200*1024*1024 + 1); err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}
	if err := tooLarge.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	testCases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing file",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file-type", "text"},
			wantErr: "--file is required",
		},
		{
			name:    "missing file type",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file", uploadPath},
			wantErr: "--file-type is required",
		},
		{
			name:    "missing path",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file", filepath.Join(dir, "missing.pdf"), "--file-type", "text"},
			wantErr: "stat upload file",
		},
		{
			name:    "directory path",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file", dir, "--file-type", "text"},
			wantErr: "must be a regular file",
		},
		{
			name:    "too large",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file", tooLargePath, "--file-type", "text"},
			wantErr: "must be <= 200MB",
		},
		{
			name:    "json body flags",
			args:    []string{"contract", "upload-file", "--profile", "contract", "--as", "app", "--file", uploadPath, "--file-type", "text", "--input-file", uploadPath},
			wantErr: "does not accept --input-file or --data",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := config.NewStore(t.TempDir())
			if err := store.UpsertProfile(uploadProfile(config.IdentityApp), true); err != nil {
				t.Fatalf("UpsertProfile() error = %v", err)
			}
			app := cli.New(cli.Options{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Store:  store,
				HTTPClient: &http.Client{
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						t.Fatalf("request transport should not be used for validation error")
						return nil, nil
					}),
				},
			})

			err := app.Run(context.Background(), tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func assertUploadMultipart(t *testing.T, req *http.Request, wantFileName, wantFileType, wantContent string) {
	t.Helper()

	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}
	if got := req.MultipartForm.Value["file_name"]; len(got) != 1 || got[0] != wantFileName {
		t.Fatalf("file_name = %v, want %q", got, wantFileName)
	}
	if got := req.MultipartForm.Value["file_type"]; len(got) != 1 || got[0] != wantFileType {
		t.Fatalf("file_type = %v, want %q", got, wantFileType)
	}
	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("file parts = %d", len(files))
	}
	if files[0].Filename != wantFileName {
		t.Fatalf("filename = %q, want %q", files[0].Filename, wantFileName)
	}
	uploaded, err := files[0].Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = uploaded.Close() }()
	content, err := io.ReadAll(uploaded)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(content) != wantContent {
		t.Fatalf("uploaded content = %q, want %q", string(content), wantContent)
	}
}

func uploadProfile(defaultIdentity config.IdentityKind) config.Profile {
	return productionProfileFixture(config.Profile{
		Name:                "contract",
		Environment:         "prod",
		OpenPlatformBaseURL: "https://open.qfei.cn",
		Resource:            "https://open.qfei.cn",
		DefaultIdentity:     defaultIdentity,
		Identities: config.Identities{
			User: config.UserIdentity{
				Token: &config.Token{
					AccessToken: "user-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				},
			},
			App: config.AppIdentity{
				Token: &config.Token{
					AccessToken: "app-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				},
			},
		},
	})
}
