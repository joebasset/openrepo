package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/generator"
	"github.com/joebasset/openrepo/internal/resolver"
	"github.com/spf13/cobra"
)

type createOptions struct {
	projectName       string
	mode              string
	frontend          string
	backend           string
	packageManager    string
	database          string
	auth              string
	storage           string
	email             string
	outputDir         string
	interactive       bool
	gitInit           bool
	install           bool
	recommendedSkills bool
}

type createSelections struct {
	InitializeGit            bool
	InstallDependencies      bool
	IncludeRecommendedSkills bool
}

func newCreateCmd() *cobra.Command {
	options := &createOptions{
		interactive: true,
		gitInit:     true,
		install:     true,
	}
	noInteractive := false

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Scaffold a new monorepo with frontend and/or backend stacks.

By default the command runs interactively, prompting for any choices you
don't supply via flags. Pass --no-interactive to skip prompts and require
all values as flags.

Available stacks:
  Frontend: nextjs, expo
  Backend:  hono-node, hono-workers, fastapi, gin`,
		Example: `  openrepo create
  openrepo create --project-name my-app --mode fullstack
  openrepo create --project-name api --mode backend --backend hono-workers --no-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noInteractive {
				options.interactive = false
			}

			flagState := commandFlagState{
				gitInitSet:           cmd.Flags().Changed("git-init"),
				installSet:           cmd.Flags().Changed("install"),
				recommendedSkillsSet: cmd.Flags().Changed("recommended-skills"),
			}

			return runCreate(cmd, *options, flagState)
		},
	}

	createCmd.Flags().StringVar(&options.projectName, "project-name", "", "Name for the generated repository")
	createCmd.Flags().StringVar(&options.mode, "mode", "", "Project mode (frontend, backend, fullstack)")
	createCmd.Flags().StringVar(&options.frontend, "frontend", "", "Frontend stack (nextjs, expo)")
	createCmd.Flags().StringVar(&options.backend, "backend", "", "Backend stack (hono-node, hono-workers, fastapi, gin)")
	createCmd.Flags().StringVar(&options.packageManager, "package-manager", "", "JS package manager (npm, pnpm, bun, yarn)")
	createCmd.Flags().StringVar(&options.database, "database", "", "Database (postgres, sqlite, supabase, d1)")
	createCmd.Flags().StringVar(&options.auth, "auth", "", "Auth provider (better-auth, supabase-auth)")
	createCmd.Flags().StringVar(&options.storage, "storage", "", "Object storage (r2, s3)")
	createCmd.Flags().StringVar(&options.email, "email", "", "Email provider (resend)")
	createCmd.Flags().StringVar(&options.outputDir, "output-dir", "", "Output directory (default: ./<project-name>)")
	createCmd.Flags().BoolVar(&options.gitInit, "git-init", true, "Initialize a git repository")
	createCmd.Flags().BoolVar(&options.install, "install", true, "Install dependencies after scaffolding")
	createCmd.Flags().BoolVar(&options.recommendedSkills, "recommended-skills", false, "Copy recommended skill bundles into .agents/skills")
	createCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip prompts and require all values as flags")

	return createCmd
}

func runCreate(cmd *cobra.Command, options createOptions, flagState commandFlagState) error {
	registry := catalog.MustDefaultRegistry()
	input := newCreateInput(options)

	if options.interactive {
		if err := promptForMissingValues(cmd, registry, &input, flagState); err != nil {
			return err
		}
	}

	applyDerivedDefaults(registry, &input)

	spec, selections, err := input.toSpec()
	if err != nil {
		return err
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		return err
	}

	targetDir, devMode, err := resolveTargetDir(options.outputDir, spec.ProjectName)
	if err != nil {
		return err
	}

	result, err := generator.Generate(spec, plan, registry, generator.Options{
		TargetDir:                targetDir,
		InitializeGit:            selections.InitializeGit,
		InstallDependencies:      selections.InstallDependencies,
		IncludeRecommendedSkills: selections.IncludeRecommendedSkills,
		DevMode:                  devMode,
	})
	if err != nil {
		cmd.Println(renderCreateSummary(spec, selections, plan, registry))
		cmd.Println()
		return err
	}

	cmd.Println(renderCreateSummary(spec, selections, plan, registry))
	cmd.Println()
	cmd.Printf("Created project at %s\n", result.TargetDir)

	for _, note := range result.Notes {
		cmd.Printf("Note: %s\n", note)
	}

	return nil
}

func renderCreateSummary(spec resolver.ProjectSpec, selections createSelections, plan resolver.ResolvedPlan, registry catalog.Registry) string {
	lines := []string{
		fmt.Sprintf("Project: %s", spec.ProjectName),
		fmt.Sprintf("Mode: %s", spec.Mode),
		fmt.Sprintf("Workspace strategy: %s", plan.WorkspaceStrategy),
		fmt.Sprintf("Initialize git: %t", selections.InitializeGit),
		fmt.Sprintf("Install dependencies: %t", selections.InstallDependencies),
		fmt.Sprintf("Recommended skills: %t", selections.IncludeRecommendedSkills),
	}

	if spec.FrontendPackID != "" {
		lines = append(lines, fmt.Sprintf("Frontend: %s", registry.MustGet(spec.FrontendPackID).DisplayName))
	}

	if spec.BackendPackID != "" {
		lines = append(lines, fmt.Sprintf("Backend: %s", registry.MustGet(spec.BackendPackID).DisplayName))
	}

	if spec.PackageManager != catalog.PackageManagerNone {
		lines = append(lines, fmt.Sprintf("Package manager: %s", spec.PackageManager))
	}

	if spec.Mode != catalog.ProjectModeFrontend {
		lines = append(lines,
			fmt.Sprintf("Database: %s", displayValue(string(spec.Database), "none")),
			fmt.Sprintf("Auth: %s", displayValue(string(spec.Auth), "none")),
			fmt.Sprintf("Storage: %s", displayValue(string(spec.Storage), "none")),
			fmt.Sprintf("Email: %s", displayValue(string(spec.Email), "none")),
		)
	}

	lines = append(lines, fmt.Sprintf("Shared types package: %t", plan.CreateSharedTypes))

	if spec.BackendPackID == catalog.PackIDHonoWorkers {
		lines = append(lines, "Cloudflare bindings: Wrangler dev/staging/production auto-provisioned D1 + KV + R2")
	}

	return strings.Join(lines, "\n")
}

func displayValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func runPromptForm(cmd *cobra.Command, fields ...huh.Field) error {
	form := huh.NewForm(huh.NewGroup(fields...))
	form.WithAccessible(os.Getenv("ACCESSIBLE") != "")
	return form.Run()
}

func resolveTargetDir(outputDir string, projectName string) (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("determine working directory: %w", err)
	}

	if strings.TrimSpace(outputDir) != "" {
		if filepath.IsAbs(outputDir) {
			return outputDir, false, nil
		}

		return filepath.Join(cwd, outputDir), false, nil
	}

	if os.Getenv("OPENREPO_DEV_MODE") == "1" {
		return filepath.Join(cwd, ".openrepo-dev", projectName), true, nil
	}

	return filepath.Join(cwd, projectName), false, nil
}
