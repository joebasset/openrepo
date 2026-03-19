package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/cli"
)

func TestCreateCommandCreatesProjectWithFoundations(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "acme")

	cmd := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--no-interactive",
		"--project-name", "acme",
		"--fe", "nextjs",
		"--be", "hono-node",
		"--package-manager", "pnpm",
		"--db", "postgres",
		"--orm", "drizzle",
		"--lint", "biome",
		"--tests", "vitest",
		"--tailwind", "tailwindcss",
		"--add-addon", "auth:better-auth,email:resend",
		"--git-init=false",
		"--install=false",
		"--output-dir", outputDir,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Summary",
		"fe=Next.js",
		"be=Hono (Node.js)",
		"db=Postgres",
		"orm=Drizzle",
		"addons=auth:better-auth,email:resend",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got %q", expected, output)
		}
	}

	for _, path := range []string{
		filepath.Join(outputDir, "README.md"),
		filepath.Join(outputDir, ".env.example"),
		filepath.Join(outputDir, "apps", "web", "package.json"),
		filepath.Join(outputDir, "apps", "api", "src", "lib", "auth.ts"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %q to exist: %v", path, err)
		}
	}
}

func TestCreateCommandListsContextAwareAddons(t *testing.T) {
	cmd := cli.NewRootCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"create",
		"--list-addons",
		"--fe", "expo",
		"--be", "fastapi",
		"--db", "postgres",
		"--orm", "sqlalchemy",
		"--lint", "ruff",
		"--tests", "pytest",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected list command to succeed, got %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "auth:better-auth") {
		t.Fatalf("did not expect better-auth in fastapi addon list, got %q", output)
	}
	if !strings.Contains(output, "email:resend") {
		t.Fatalf("expected resend in addon list, got %q", output)
	}
}
