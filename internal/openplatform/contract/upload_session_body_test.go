package contract_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/openplatform/contract"
)

type failedUploadSource struct{}

func (failedUploadSource) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestUploadSessionContentBodyPropagatesSourceFailure(t *testing.T) {
	body, _ := contract.UploadSessionContentBody("file.pdf", failedUploadSource{})
	defer func() { _ = body.Close() }()
	if _, err := io.Copy(io.Discard, body); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("lost file read error: %v", err)
	}
}

func TestUploadSessionContentBodyClosesBeforeAndDuringReading(t *testing.T) {
	for _, startReading := range []bool{false, true} {
		body, _ := contract.UploadSessionContentBody("file.pdf", strings.NewReader("file bytes"))
		if startReading {
			if _, err := body.Read(make([]byte, 1)); err != nil {
				t.Fatal(err)
			}
		}
		if err := body.Close(); err != nil {
			t.Fatal(err)
		}
		finished := make(chan error, 1)
		go func() { _, err := io.Copy(io.Discard, body); finished <- err }()
		select {
		case err := <-finished:
			if !errors.Is(err, io.ErrClosedPipe) {
				t.Fatalf("closed body error = %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("multipart reader remained blocked after close")
		}
	}
}
