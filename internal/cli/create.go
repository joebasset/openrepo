package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	list              bool
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
		Long:  "Scaffold a new monorepo with frontend and/or backend stacks.",
		Example: `  openrepo create
  openrepo create --list
  openrepo create --project-name my-app --mode fullstack
  openrepo create --project-name api --mode backend --backend hono-workers --no-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.list {
				registry := catalog.MustDefaultRegistry()
				addonRegistry := catalog.MustDefaultAddonRegistry()
				cmd.Println(renderAvailableOptions(registry, addonRegistry))
				return nil
			}

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
	createCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Println(renderCreateHelp())
	})

	createCmd.Flags().StringVar(&options.projectName, "project-name", "", "Name for the generated repository")
	createCmd.Flags().StringVar(&options.mode, "mode", "", "Project mode (frontend, backend, fullstack)")
	createCmd.Flags().StringVar(&options.frontend, "frontend", "", "Frontend stack (nextjs, expo)")
	createCmd.Flags().StringVar(&options.backend, "backend", "", "Backend stack (hono-node, hono-workers, fastapi, gin)")
	createCmd.Flags().StringVar(&options.packageManager, "package-manager", "", "JS package manager (npm, pnpm, bun, yarn)")
	createCmd.Flags().StringVar(&options.database, "database", "", "Database (postgres, sqlite, supabase, d1)")
	createCmd.Flags().StringVar(&options.auth, "auth", "", "Auth provider (better-auth, supabase-auth)")
	createCmd.Flags().StringVar(&options.storage, "storage", "", "Object storage (r2, s3, supabase-storage)")
	createCmd.Flags().StringVar(&options.email, "email", "", "Email provider (resend)")
	createCmd.Flags().StringVar(&options.outputDir, "output-dir", "", "Output directory (default: ./<project-name>)")
	createCmd.Flags().BoolVar(&options.gitInit, "git-init", true, "Initialize a git repository")
	createCmd.Flags().BoolVar(&options.install, "install", true, "Install dependencies after scaffolding")
	createCmd.Flags().BoolVar(&options.recommendedSkills, "recommended-skills", false, "Copy recommended skill bundles into .agents/skills")
	createCmd.Flags().BoolVar(&options.list, "list", false, "List available packs and addon options")
	createCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip prompts and require all values as flags")

	return createCmd
}

func runCreate(cmd *cobra.Command, options createOptions, flagState commandFlagState) error {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
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

	result, err := generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
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
	overview := []string{
		fmt.Sprintf("project=%s", spec.ProjectName),
		fmt.Sprintf("mode=%s", spec.Mode),
		fmt.Sprintf("workspace=%s", plan.WorkspaceStrategy),
	}

	stacks := make([]string, 0, 2)
	if spec.FrontendPackID != "" {
		stacks = append(stacks, fmt.Sprintf("frontend=%s", registry.MustGet(spec.FrontendPackID).DisplayName))
	}
	if spec.BackendPackID != "" {
		stacks = append(stacks, fmt.Sprintf("backend=%s", registry.MustGet(spec.BackendPackID).DisplayName))
	}

	tooling := []string{
		fmt.Sprintf("git=%s", yesNo(selections.InitializeGit)),
		fmt.Sprintf("install=%s", yesNo(selections.InstallDependencies)),
		fmt.Sprintf("skills=%s", yesNo(selections.IncludeRecommendedSkills)),
	}
	if spec.PackageManager != catalog.PackageManagerNone {
		tooling = append(tooling, fmt.Sprintf("package-manager=%s", spec.PackageManager))
	}

	integrations := make([]string, 0, 4)
	if spec.Mode != catalog.ProjectModeFrontend {
		integrations = append(integrations,
			fmt.Sprintf("database=%s", displayValue(string(spec.Database), "none")),
			fmt.Sprintf("auth=%s", displayValue(string(spec.Auth), "none")),
			fmt.Sprintf("storage=%s", displayValue(string(spec.Storage), "none")),
			fmt.Sprintf("email=%s", displayValue(string(spec.Email), "none")),
		)
	}

	extras := []string{fmt.Sprintf("shared-types=%s", yesNo(plan.CreateSharedTypes))}
	if spec.BackendPackID == catalog.PackIDHonoWorkers {
		extras = append(extras, "cloudflare-bindings=dev|staging|production with d1|kv|r2")
	}

	lines := []string{
		"Summary",
		fmt.Sprintf("  %s", strings.Join(overview, "  ")),
	}
	if len(stacks) > 0 {
		lines = append(lines, fmt.Sprintf("  %s", strings.Join(stacks, "  ")))
	}
	lines = append(lines, fmt.Sprintf("  %s", strings.Join(tooling, "  ")))
	if len(integrations) > 0 {
		lines = append(lines, fmt.Sprintf("  %s", strings.Join(integrations, "  ")))
	}
	lines = append(lines, fmt.Sprintf("  %s", strings.Join(extras, "  ")))

	return strings.Join(lines, "\n")
}

func displayValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

func renderCreateHelp() string {
	sections := []string{
		"openrepo create",
		"  Scaffold a new project.",
		"",
		"Usage",
		"  openrepo create [flags]",
		"",
		"Quick start",
		"  openrepo create",
		"  openrepo create --list",
		"  openrepo create --project-name my-app --mode fullstack --frontend nextjs --backend hono-node",
		"",
		"Packs",
		"  frontend: nextjs, expo",
		"  backend:  nextjs, hono-node, hono-workers, fastapi, gin",
		"",
		"Options",
		"  mode:            frontend, backend, fullstack",
		"  package-manager: npm, pnpm, bun, yarn",
		"  database:        postgres, sqlite, supabase, d1, none",
		"  auth:            better-auth, supabase-auth, none",
		"  storage:         r2, s3, supabase-storage, none",
		"  email:           resend, none",
		"",
		"Flags",
	}

	type helpFlag struct {
		name        string
		description string
	}

	flags := []helpFlag{
		{name: "--project-name", description: "Repository name"},
		{name: "--mode", description: "Project mode"},
		{name: "--frontend", description: "Frontend pack"},
		{name: "--backend", description: "Backend pack"},
		{name: "--package-manager", description: "JS package manager"},
		{name: "--database", description: "Database option"},
		{name: "--auth", description: "Auth option"},
		{name: "--storage", description: "Storage option"},
		{name: "--email", description: "Email option"},
		{name: "--output-dir", description: "Output path"},
		{name: "--recommended-skills", description: "Copy recommended skill bundles"},
		{name: "--git-init", description: "Initialize git (default: true)"},
		{name: "--install", description: "Install dependencies (default: true)"},
		{name: "--no-interactive", description: "Disable prompts"},
		{name: "--list", description: "Show available packs and addon support"},
		{name: "--help", description: "Show help"},
	}

	slices.SortFunc(flags, func(a, b helpFlag) int {
		return strings.Compare(a.name, b.name)
	})

	for _, flag := range flags {
		sections = append(sections, fmt.Sprintf("  %-20s %s", flag.name, flag.description))
	}

	return strings.Join(sections, "\n")
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
