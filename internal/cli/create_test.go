package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/cli"
)

func TestCreateCommandCreatesLocalTemplateProject(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "acme")

	cmd := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "acme",
		"--mode", "backend",
		"--backend", "gin",
		"--database", "postgres",
		"--git-init=false",
		"--install=false",
		"--output-dir", outputDir,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Project: acme",
		"Mode: backend",
		"Workspace strategy: native",
		"Backend: Gin",
		"Created project at " + outputDir,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got %q", expected, output)
		}
	}

	for _, path := range []string{
		filepath.Join(outputDir, "README.md"),
		filepath.Join(outputDir, ".env.example"),
		filepath.Join(outputDir, "apps", "api", "go.mod"),
		filepath.Join(outputDir, "apps", "api", "AGENTS.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %q to exist: %v", path, err)
		}
	}
}

func TestCreateCommandRejectsUnsupportedExpoPackageManager(t *testing.T) {
	cmd := cli.NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "mobile",
		"--mode", "frontend",
		"--frontend", "expo",
		"--package-manager", "pnpm",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected command to fail for unsupported expo package manager")
	}
}

func TestCreateCommandRejectsUnknownMode(t *testing.T) {
	cmd := cli.NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "acme",
		"--mode", "desktop",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected command to fail for unsupported mode")
	}
}

func TestCreateCommandCreatesSnapshotFullstackProject(t *testing.T) {
	installFakeToolchain(t)
	outputDir := filepath.Join(t.TempDir(), "test-repo")

	cmd := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "test-repo",
		"--mode", "fullstack",
		"--frontend", "nextjs",
		"--backend", "hono-workers",
		"--auth", "better-auth",
		"--email", "resend",
		"--output-dir", outputDir,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Project: test-repo",
		"Mode: fullstack",
		"Workspace strategy: turbo",
		"Initialize git: true",
		"Install dependencies: true",
		"Frontend: Next.js",
		"Backend: Hono (Cloudflare Workers)",
		"Package manager: pnpm",
		"Database: d1",
		"Auth: better-auth",
		"Storage: r2",
		"Email: resend",
		"Shared types package: true",
		"Cloudflare bindings: Wrangler dev/staging/production auto-provisioned D1 + KV + R2",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got %q", expected, output)
		}
	}

	for _, path := range []string{
		filepath.Join(outputDir, ".git", "HEAD"),
		filepath.Join(outputDir, "pnpm-lock.yaml"),
		filepath.Join(outputDir, "apps", "web", "package.json"),
		filepath.Join(outputDir, "apps", "api", "wrangler.jsonc"),
		filepath.Join(outputDir, "packages", "shared-types", "package.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %q to exist: %v", path, err)
		}
	}
}

func TestCreateCommandCopiesRecommendedSkillsWhenRequested(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "test-repo")

	cmd := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "test-repo",
		"--mode", "fullstack",
		"--frontend", "nextjs",
		"--backend", "hono-workers",
		"--recommended-skills",
		"--git-init=false",
		"--install=false",
		"--output-dir", outputDir,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Recommended skills: true") {
		t.Fatalf("expected output to mention recommended skills, got %q", output)
	}

	for _, path := range []string{
		filepath.Join(outputDir, ".agents", "skills", "web-perf", "SKILL.md"),
		filepath.Join(outputDir, ".agents", "skills", "wrangler", "SKILL.md"),
		filepath.Join(outputDir, ".agents", "skills", "workers-best-practices", "SKILL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %q to exist: %v", path, err)
		}
	}
}

func installFakeToolchain(t *testing.T) {
	t.Helper()

	binDir := t.TempDir()
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	writeExecutable(t, filepath.Join(binDir, "pnpm"), `#!/bin/sh
set -eu
if [ "$1" = "create" ] && { [ "$2" = "next-app" ] || [ "$2" = "next-app@latest" ]; }; then
  dir="$3"
  mkdir -p "$dir/src/app"
  printf '{"name":"web"}\n' > "$dir/package.json"
  printf 'export default function Page() { return null }\n' > "$dir/src/app/page.tsx"
  exit 0
fi
if [ "$1" = "create" ] && [ "$2" = "hono@latest" ]; then
  dir="$3"
  mkdir -p "$dir/src"
  printf '{"name":"api"}\n' > "$dir/package.json"
  printf 'export default {};\n' > "$dir/src/index.ts"
  printf '{}\n' > "$dir/wrangler.jsonc"
  exit 0
fi
if [ "$1" = "install" ]; then
  : > "$PWD/pnpm-lock.yaml"
  exit 0
fi
exit 0
`)

	writeExecutable(t, filepath.Join(binDir, "git"), `#!/bin/sh
set -eu
if [ "$1" = "init" ]; then
  mkdir -p "$PWD/.git"
  printf 'ref: refs/heads/main\n' > "$PWD/.git/HEAD"
  exit 0
fi
exit 0
`)
}

func writeExecutable(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatalf("write executable %q: %v", path, err)
	}
}
