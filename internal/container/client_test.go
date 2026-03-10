package container

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerHostFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         *dockerContextInfo
		wantHost    string
		wantErrText string
	}{
		{
			name: "unix context stays supported",
			ctx: &dockerContextInfo{
				Name: "desktop-linux",
				Host: "unix:///Users/test/.docker/run/docker.sock",
			},
			wantHost: "unix:///Users/test/.docker/run/docker.sock",
		},
		{
			name: "plain tcp context stays supported",
			ctx: &dockerContextInfo{
				Name: "remote-http",
				Host: "tcp://127.0.0.1:2375",
			},
			wantHost: "tcp://127.0.0.1:2375",
		},
		{
			name: "tls backed tcp context is rejected explicitly",
			ctx: &dockerContextInfo{
				Name:           "remote-tls",
				Host:           "tcp://docker.example.com:2376",
				HasTLSMaterial: true,
			},
			wantErrText: "TLS-backed TCP endpoint",
		},
		{
			name: "ssh context is rejected explicitly",
			ctx: &dockerContextInfo{
				Name: "remote-ssh",
				Host: "ssh://ubuntu@example.com",
			},
			wantErrText: "SSH endpoint",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dockerHostFromContext(tt.ctx)
			if tt.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrText)
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrText, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantHost {
				t.Fatalf("expected host %q, got %q", tt.wantHost, got)
			}
		})
	}
}

func TestCurrentDockerContextHonorsEnvOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOCKER_CONTEXT", "default")

	if err := os.MkdirAll(filepath.Join(home, ".docker"), 0755); err != nil {
		t.Fatalf("mkdir .docker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".docker", "config.json"), []byte(`{"currentContext":"desktop-linux"}`), 0644); err != nil {
		t.Fatalf("write config.json: %v", err)
	}

	ctx, err := currentDockerContext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx == nil {
		t.Fatal("expected context, got nil")
	}
	if ctx.Name != "default" {
		t.Fatalf("expected context name %q, got %q", "default", ctx.Name)
	}
	if ctx.Host != "unix:///var/run/docker.sock" {
		t.Fatalf("expected host %q, got %q", "unix:///var/run/docker.sock", ctx.Host)
	}
}
