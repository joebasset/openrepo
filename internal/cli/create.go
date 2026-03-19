package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/generator"
	"github.com/joebasset/openrepo/internal/resolver"
	"github.com/spf13/cobra"
)

type createOptions struct {
	projectName       string
	fe                string
	be                string
	packageManager    string
	db                string
	orm               string
	lint              string
	tests             string
	tailwind          string
	addAddons         []string
	outputDir         string
	interactive       bool
	gitInit           bool
	install           bool
	recommendedSkills bool
	list              bool
	listFE            bool
	listBE            bool
	listDB            bool
	listORMs          bool
	listLint          bool
	listTests         bool
	listTailwind      bool
	listAddons        bool
}

type createSelections struct {
	AddOns                   []string
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
		Short: "Create a new full-stack project",
		Long:  "Scaffold a new full-stack project by choosing frontend and backend packs, then layering foundations and optional addons.",
		Example: `  openrepo create
  openrepo create --list
  openrepo create --project-name acme --fe nextjs --be hono-node --package-manager pnpm --db postgres --orm drizzle --lint biome --tests vitest --tailwind tailwindcss
  openrepo create --project-name acme --fe expo --be hono-workers --package-manager npm --db d1 --orm drizzle --lint biome --tests vitest --add-addon storage:r2 --add-addon email:resend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noInteractive {
				options.interactive = false
			}

			registry := catalog.MustDefaultRegistry()
			addonRegistry := catalog.MustDefaultAddonRegistry()

			if hasAnyListFlag(*options) {
				cmd.Println(renderAvailableOptions(registry, addonRegistry, *options))
				return nil
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
	createCmd.Flags().StringVar(&options.fe, "fe", "", "Frontend pack")
	createCmd.Flags().StringVar(&options.be, "be", "", "Backend pack")
	createCmd.Flags().StringVar(&options.packageManager, "package-manager", "", "Package manager for JavaScript/TypeScript stacks")
	createCmd.Flags().StringVar(&options.db, "db", "", "Database foundation")
	createCmd.Flags().StringVar(&options.orm, "orm", "", "ORM foundation")
	createCmd.Flags().StringVar(&options.lint, "lint", "", "Lint / formatter foundation")
	createCmd.Flags().StringVar(&options.tests, "tests", "", "Testing foundation")
	createCmd.Flags().StringVar(&options.tailwind, "tailwind", "", "Tailwind foundation for supported frontend packs")
	createCmd.Flags().StringArrayVar(&options.addAddons, "add-addon", nil, "Optional addon id, repeatable or comma-separated (for example auth:better-auth,email:resend)")
	createCmd.Flags().StringVar(&options.outputDir, "output-dir", "", "Output directory (default: ./<project-name>)")
	createCmd.Flags().BoolVar(&options.gitInit, "git-init", true, "Initialize a git repository")
	createCmd.Flags().BoolVar(&options.install, "install", true, "Install dependencies after scaffolding")
	createCmd.Flags().BoolVar(&options.recommendedSkills, "recommended-skills", false, "Copy recommended skill bundles into .agents/skills")
	createCmd.Flags().BoolVar(&options.list, "list", false, "List supported packs, foundations, and context-aware addons")
	createCmd.Flags().BoolVar(&options.listFE, "list-fe", false, "List supported frontend packs")
	createCmd.Flags().BoolVar(&options.listBE, "list-be", false, "List supported backend packs")
	createCmd.Flags().BoolVar(&options.listDB, "list-db", false, "List supported databases")
	createCmd.Flags().BoolVar(&options.listORMs, "list-orms", false, "List supported ORMs")
	createCmd.Flags().BoolVar(&options.listLint, "list-lint", false, "List supported lint / formatter options")
	createCmd.Flags().BoolVar(&options.listTests, "list-tests", false, "List supported test options")
	createCmd.Flags().BoolVar(&options.listTailwind, "list-tailwind", false, "List supported Tailwind options")
	createCmd.Flags().BoolVar(&options.listAddons, "list-addons", false, "List context-aware optional addons")
	createCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip prompts and require all values as flags")

	return createCmd
}

func hasAnyListFlag(options createOptions) bool {
	return options.list ||
		options.listFE ||
		options.listBE ||
		options.listDB ||
		options.listORMs ||
		options.listLint ||
		options.listTests ||
		options.listTailwind ||
		options.listAddons
}

func runCreate(cmd *cobra.Command, options createOptions, flagState commandFlagState) error {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	input, err := newCreateInput(options)
	if err != nil {
		return err
	}

	if options.interactive {
		if err := promptForMissingValues(cmd, registry, addonRegistry, &input, flagState); err != nil {
			return err
		}
	}

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
		fmt.Sprintf("workspace=%s", plan.WorkspaceStrategy),
	}

	stacks := []string{
		fmt.Sprintf("fe=%s", registry.MustGet(spec.FrontendPackID).DisplayName),
		fmt.Sprintf("be=%s", registry.MustGet(spec.BackendPackID).DisplayName),
	}

	foundations := []string{
		fmt.Sprintf("package-manager=%s", spec.PackageManager),
		fmt.Sprintf("db=%s", catalog.SelectionValueLabel(catalog.SelectionKindDatabase, string(spec.DatabaseOption()))),
		fmt.Sprintf("orm=%s", catalog.SelectionValueLabel(catalog.SelectionKindORM, string(spec.OrmOption()))),
		fmt.Sprintf("lint=%s", catalog.SelectionValueLabel(catalog.SelectionKindLint, string(spec.LintOption()))),
		fmt.Sprintf("tests=%s", catalog.SelectionValueLabel(catalog.SelectionKindTests, string(spec.TestsOption()))),
	}
	if spec.TailwindOption() != catalog.TailwindNone {
		foundations = append(foundations, fmt.Sprintf("tailwind=%s", catalog.SelectionValueLabel(catalog.SelectionKindTailwind, string(spec.TailwindOption()))))
	}

	optional := []string{
		fmt.Sprintf("git=%s", yesNo(selections.InitializeGit)),
		fmt.Sprintf("install=%s", yesNo(selections.InstallDependencies)),
		fmt.Sprintf("skills=%s", yesNo(selections.IncludeRecommendedSkills)),
	}
	if len(selections.AddOns) > 0 {
		optional = append(optional, fmt.Sprintf("addons=%s", strings.Join(selections.AddOns, ",")))
	}

	extras := []string{fmt.Sprintf("shared-types=%s", yesNo(plan.CreateSharedTypes))}
	if spec.BackendPackID == catalog.PackIDHonoWorkers {
		extras = append(extras, "workers-bindings=dev|staging|production")
	}

	lines := []string{
		"Summary",
		fmt.Sprintf("  %s", strings.Join(overview, "  ")),
		fmt.Sprintf("  %s", strings.Join(stacks, "  ")),
		fmt.Sprintf("  %s", strings.Join(foundations, "  ")),
		fmt.Sprintf("  %s", strings.Join(optional, "  ")),
		fmt.Sprintf("  %s", strings.Join(extras, "  ")),
	}

	return strings.Join(lines, "\n")
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
		"  Scaffold a full-stack project.",
		"",
		"Usage",
		"  openrepo create [flags]",
		"",
		"Quick start",
		"  openrepo create",
		"  openrepo create --list",
		"  openrepo create --project-name acme --fe nextjs --be hono-node --package-manager pnpm --db postgres --orm drizzle --lint biome --tests vitest --tailwind tailwindcss",
		"  openrepo create --project-name acme --fe expo --be hono-workers --package-manager npm --db d1 --orm drizzle --lint biome --tests vitest --add-addon storage:r2 --add-addon email:resend",
		"",
		"Addon ids",
		"  auth:better-auth",
		"  auth:supabase-auth",
		"  auth:firebase-auth",
		"  storage:s3",
		"  storage:r2",
		"  storage:supabase-storage",
		"  email:resend",
		"  icons:lucide-react",
		"  icons:react-icons",
		"  components:shadcn",
		"  components:mui",
		"",
		"Flags",
	}

	type helpFlag struct {
		name        string
		description string
	}

	flags := []helpFlag{
		{name: "--project-name", description: "Repository name"},
		{name: "--fe", description: "Frontend pack"},
		{name: "--be", description: "Backend pack"},
		{name: "--package-manager", description: "Package manager foundation"},
		{name: "--db", description: "Database foundation"},
		{name: "--orm", description: "ORM foundation"},
		{name: "--lint", description: "Lint / formatter foundation"},
		{name: "--tests", description: "Test foundation"},
		{name: "--tailwind", description: "Tailwind foundation for supported frontend packs"},
		{name: "--add-addon", description: "Optional addon id, repeatable or comma-separated"},
		{name: "--output-dir", description: "Output path"},
		{name: "--recommended-skills", description: "Copy recommended skill bundles"},
		{name: "--git-init", description: "Initialize git (default: true)"},
		{name: "--install", description: "Install dependencies (default: true)"},
		{name: "--list", description: "Show grouped FE / BE / foundations / addons"},
		{name: "--list-fe", description: "Show frontend packs"},
		{name: "--list-be", description: "Show backend packs"},
		{name: "--list-db", description: "Show database foundations"},
		{name: "--list-orms", description: "Show ORM foundations"},
		{name: "--list-lint", description: "Show lint foundations"},
		{name: "--list-tests", description: "Show test foundations"},
		{name: "--list-tailwind", description: "Show Tailwind foundations"},
		{name: "--list-addons", description: "Show compatible optional addons"},
		{name: "--no-interactive", description: "Disable prompts"},
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
	form.WithTheme(openrepoPromptTheme())
	return form.Run()
}

func openrepoPromptTheme() *huh.Theme {
	theme := huh.ThemeBase()

	background := lipgloss.AdaptiveColor{Light: "#F4F7FF", Dark: "#262A3D"}
	border := lipgloss.AdaptiveColor{Light: "#A2ABC7", Dark: "#525B7D"}
	highlight := lipgloss.AdaptiveColor{Light: "#0AB7A7", Dark: "#17F0D5"}
	muted := lipgloss.AdaptiveColor{Light: "#50607E", Dark: "#A8B0CC"}
	text := lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F3F7FF"}
	warn := lipgloss.AdaptiveColor{Light: "#D455FF", Dark: "#F37DFF"}
	errorColor := lipgloss.AdaptiveColor{Light: "#C6285D", Dark: "#FF6E9F"}

	theme.Form.Base = theme.Form.Base.
		Background(background).
		Padding(1, 2)
	theme.Group.Base = theme.Group.Base.Background(background)
	theme.FieldSeparator = lipgloss.NewStyle().Background(background).SetString("\n")

	theme.Focused.Base = theme.Focused.Base.
		Background(background).
		BorderForeground(border).
		PaddingLeft(1)
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = theme.Focused.Title.
		Foreground(background).
		Background(highlight).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)
	theme.Focused.NoteTitle = theme.Focused.NoteTitle.
		Foreground(highlight).
		Bold(true).
		MarginBottom(1)
	theme.Focused.Description = theme.Focused.Description.Foreground(muted)
	theme.Focused.Option = theme.Focused.Option.Foreground(text)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(highlight).Bold(true)
	theme.Focused.MultiSelectSelector = theme.Focused.MultiSelectSelector.Foreground(highlight).Bold(true)
	theme.Focused.SelectedOption = theme.Focused.SelectedOption.Foreground(highlight).Bold(true)
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(highlight).SetString("[x] ")
	theme.Focused.UnselectedOption = theme.Focused.UnselectedOption.Foreground(text)
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("[ ] ")
	theme.Focused.NextIndicator = theme.Focused.NextIndicator.Foreground(warn).Bold(true)
	theme.Focused.PrevIndicator = theme.Focused.PrevIndicator.Foreground(warn).Bold(true)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(highlight).Bold(true)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(highlight)
	theme.Focused.TextInput.CursorText = theme.Focused.TextInput.CursorText.Foreground(text)
	theme.Focused.TextInput.Text = theme.Focused.TextInput.Text.Foreground(text)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(muted)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Foreground(background).Background(highlight).Bold(true)
	theme.Focused.BlurredButton = theme.Focused.BlurredButton.Foreground(text).Background(border)
	theme.Focused.Next = theme.Focused.FocusedButton
	theme.Focused.ErrorIndicator = theme.Focused.ErrorIndicator.Foreground(errorColor)
	theme.Focused.ErrorMessage = theme.Focused.ErrorMessage.Foreground(errorColor)

	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Blurred.Base.Background(background).BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.Card = theme.Blurred.Base
	theme.Blurred.Option = theme.Blurred.Option.Foreground(text)
	theme.Blurred.UnselectedOption = theme.Blurred.UnselectedOption.Foreground(text)
	theme.Blurred.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("[ ] ")
	theme.Blurred.SelectedOption = theme.Blurred.SelectedOption.Foreground(highlight)
	theme.Blurred.SelectedPrefix = lipgloss.NewStyle().Foreground(highlight).SetString("[x] ")
	theme.Blurred.NextIndicator = lipgloss.NewStyle()
	theme.Blurred.PrevIndicator = lipgloss.NewStyle()
	theme.Blurred.Title = theme.Focused.Title
	theme.Blurred.NoteTitle = theme.Focused.NoteTitle
	theme.Blurred.Description = theme.Focused.Description

	theme.Group.Title = theme.Focused.Title
	theme.Group.Description = theme.Focused.Description

	return theme
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
