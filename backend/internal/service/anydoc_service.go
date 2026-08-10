package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const anydocTimeout = 30 * time.Second

type AnyDocService struct {
	Binary string
}

func NewAnyDocService() *AnyDocService {
	bin := os.Getenv("ANYDOC_BIN")
	if bin == "" {
		bin = "anydoc"
	}
	return &AnyDocService{Binary: bin}
}

func (s *AnyDocService) ConvertFileToMarkdown(ctx context.Context, srcPath string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, anydocTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.Binary, srcPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("anydoc conversion timed out: %w", ctx.Err())
		}
		return "", fmt.Errorf("anydoc conversion failed: %v: %s", err, stderr.String())
	}
	if len(stdout.Bytes()) == 0 {
		return "", fmt.Errorf("anydoc returned empty output for %s", srcPath)
	}
	return stdout.String(), nil
}