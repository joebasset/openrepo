package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/resolver"
	"github.com/joebasset/openrepo/internal/templates"
)

type Options struct {
	TargetDir                string
	InitializeGit            bool
	InstallDependencies      bool
	IncludeRecommendedSkills bool
	DevMode                  bool
}

type Result struct {
	TargetDir string
	Notes     []string
}

type templateData struct {
	ProjectName string
	ProjectSlug string
	ModulePath  string
	Runtime     string
	Database    string
}

func Generate(spec resolver.ProjectSpec, plan resolver.ResolvedPlan, registry catalog.Registry, options Options) (Result, error) {
	if strings.TrimSpace(options.TargetDir) == "" {
		return Result{}, errors.New("target directory is required")
	}

	packs, err := selectedPacks(spec, registry)
	if err != nil {
		return Result{}, err
	}

	if err := prepareTargetDir(options.TargetDir, options.DevMode); err != nil {
		return Result{}, err
	}

	if err := createProjectDirectories(options.TargetDir, plan, packs); err != nil {
		return Result{}, err
	}

	if err := scaffoldPacks(options.TargetDir, spec, packs); err != nil {
		return Result{}, err
	}

	if err := writeRootFiles(options.TargetDir, spec, plan, packs, options); err != nil {
		return Result{}, err
	}

	if err := writePackOverlays(options.TargetDir, spec, packs); err != nil {
		return Result{}, err
	}

	result := Result{TargetDir: options.TargetDir}
	if options.InstallDependencies {
		notes, err := installDependencies(options.TargetDir, spec, plan)
		if err != nil {
			return Result{}, err
		}
		result.Notes = append(result.Notes, notes...)
	}

	if options.InitializeGit {
		if err := initializeGitRepository(options.TargetDir); err != nil {
			return Result{}, err
		}
	}

	return result, nil
}

func selectedPacks(spec resolver.ProjectSpec, registry catalog.Registry) ([]catalog.Pack, error) {
	packs := make([]catalog.Pack, 0, 2)

	if spec.FrontendPackID != "" {
		pack, ok := registry.Get(spec.FrontendPackID)
		if !ok {
			return nil, fmt.Errorf("unknown pack %q", spec.FrontendPackID)
		}
		packs = append(packs, pack)
	}

	if spec.BackendPackID != "" {
		pack, ok := registry.Get(spec.BackendPackID)
		if !ok {
			return nil, fmt.Errorf("unknown pack %q", spec.BackendPackID)
		}
		packs = append(packs, pack)
	}

	return packs, nil
}

func prepareTargetDir(targetDir string, devMode bool) error {
	if devMode {
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("reset dev target directory: %w", err)
		}
	}

	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("inspect target directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("target path %q must be a directory", targetDir)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("read target directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == ".agents" || entry.Name() == "AGENTS.md" {
			continue
		}

		return fmt.Errorf("target directory %q must be empty or contain only .agents and AGENTS.md", targetDir)
	}

	return nil
}

func createProjectDirectories(targetDir string, plan resolver.ResolvedPlan, packs []catalog.Pack) error {
	directories := map[string]struct{}{
		targetDir:                                     {},
		filepath.Join(targetDir, "apps"):              {},
		filepath.Join(targetDir, "packages"):          {},
		filepath.Join(targetDir, ".agents"):           {},
		filepath.Join(targetDir, ".agents", "skills"): {},
	}

	for _, pack := range packs {
		outputDir := filepath.Join(targetDir, filepath.FromSlash(pack.OutputDir))
		if pack.Strategy == catalog.PackStrategyExternalScaffold {
			directories[filepath.Dir(outputDir)] = struct{}{}
			continue
		}

		directories[outputDir] = struct{}{}
	}

	if plan.CreateSharedTypes {
		directories[filepath.Join(targetDir, "packages", "shared-types", "src")] = struct{}{}
	}

	paths := make([]string, 0, len(directories))
	for path := range directories {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create directory %q: %w", path, err)
		}
	}

	return nil
}

func writeRootFiles(targetDir string, spec resolver.ProjectSpec, plan resolver.ResolvedPlan, packs []catalog.Pack, options Options) error {
	rootFiles := []struct {
		path           string
		contents       string
		writeIfMissing bool
	}{
		{path: filepath.Join(targetDir, "README.md"), contents: renderRootReadme(spec, plan, packs)},
		{path: filepath.Join(targetDir, ".gitignore"), contents: renderGitignore(plan)},
		{path: filepath.Join(targetDir, ".env.example"), contents: renderRootEnvExample(spec, packs)},
		{path: filepath.Join(targetDir, "AGENTS.md"), contents: renderRootAgentsFile(spec, plan, packs, options.IncludeRecommendedSkills), writeIfMissing: true},
	}

	if hasTypeScriptPack(packs) {
		rootFiles = append(rootFiles, struct {
			path           string
			contents       string
			writeIfMissing bool
		}{
			path:     filepath.Join(targetDir, "biome.json"),
			contents: renderBiomeConfig(),
		})
	}

	for _, file := range rootFiles {
		var err error
		if file.writeIfMissing {
			err = writeFileIfMissing(file.path, file.contents)
		} else {
			err = writeFile(file.path, file.contents)
		}
		if err != nil {
			return err
		}
	}

	if plan.WorkspaceStrategy == catalog.WorkspaceStrategyTurbo {
		packageJSON, err := renderTurboPackageJSON(spec)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(targetDir, "package.json"), packageJSON); err != nil {
			return err
		}
		if err := writeFile(filepath.Join(targetDir, "turbo.json"), renderTurboConfig()); err != nil {
			return err
		}
		if spec.PackageManager == catalog.PackageManagerPNPM {
			if err := writeFile(filepath.Join(targetDir, "pnpm-workspace.yaml"), renderPNPMWorkspaceConfig()); err != nil {
				return err
			}
		}
	}

	if plan.WorkspaceStrategy == catalog.WorkspaceStrategyNative {
		makefile := renderMakefile(packs)
		if strings.TrimSpace(makefile) != "" {
			if err := writeFile(filepath.Join(targetDir, "Makefile"), makefile); err != nil {
				return err
			}
		}
	}

	if plan.CreateSharedTypes {
		if err := writeSharedTypesPackage(targetDir, spec); err != nil {
			return err
		}
	}

	if err := clearGeneratedSkillAssets(targetDir, packs); err != nil {
		return err
	}

	if options.IncludeRecommendedSkills {
		if err := writeRecommendedSkillAssets(targetDir, packs); err != nil {
			return err
		}
	}

	return nil
}

func scaffoldPacks(targetDir string, spec resolver.ProjectSpec, packs []catalog.Pack) error {
	for _, pack := range packs {
		data := templateData{
			ProjectName: spec.ProjectName,
			ProjectSlug: projectSlug(spec.ProjectName),
			ModulePath:  goModulePath(spec.ProjectName, pack.OutputDir),
			Runtime:     string(pack.Runtime),
			Database:    string(spec.Database),
		}

		if pack.Strategy == catalog.PackStrategyExternalScaffold {
			if err := executeExternalScaffold(targetDir, spec, pack); err != nil {
				return err
			}
			continue
		}

		for _, file := range pack.Files {
			if file.Role != catalog.FileRoleLocalTemplate {
				continue
			}

			outputPath := filepath.Join(targetDir, filepath.FromSlash(file.Path))
			contents, err := renderTemplateAsset(file.AssetPath, data)
			if err != nil {
				return err
			}
			if err := writeFile(outputPath, contents); err != nil {
				return err
			}
		}
	}

	return nil
}

func writePackOverlays(targetDir string, spec resolver.ProjectSpec, packs []catalog.Pack) error {
	for _, pack := range packs {
		for _, file := range pack.Files {
			if file.Role != catalog.FileRoleOverlay {
				continue
			}

			outputPath := filepath.Join(targetDir, filepath.FromSlash(file.Path))
			switch filepath.Base(file.Path) {
			case ".env.example":
				if err := writeFile(outputPath, renderPackEnvExample(pack, spec)); err != nil {
					return err
				}
			case "AGENTS.md":
				if err := writeFile(outputPath, renderPackAgentsFile(pack, spec)); err != nil {
					return err
				}
			case "wrangler.jsonc":
				if err := writeFile(outputPath, renderWorkersWranglerConfig(spec)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func writeSharedTypesPackage(targetDir string, spec resolver.ProjectSpec) error {
	packageJSON, err := renderSharedTypesPackageJSON(spec)
	if err != nil {
		return err
	}

	files := []struct {
		path     string
		contents string
	}{
		{path: filepath.Join(targetDir, "packages", "shared-types", "package.json"), contents: packageJSON},
		{path: filepath.Join(targetDir, "packages", "shared-types", "tsconfig.json"), contents: renderSharedTypesTSConfig()},
		{path: filepath.Join(targetDir, "packages", "shared-types", "src", "index.ts"), contents: "export type HealthStatus = \"ok\";\n"},
		{path: filepath.Join(targetDir, "packages", "shared-types", "README.md"), contents: renderSharedTypesReadme(spec)},
	}

	for _, file := range files {
		if err := writeFile(file.path, file.contents); err != nil {
			return err
		}
	}

	return nil
}

func initializeGitRepository(targetDir string) error {
	command := exec.Command("git", "init")
	command.Dir = targetDir
	command.Env = os.Environ()

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("initialize git repository: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func executeExternalScaffold(targetDir string, spec resolver.ProjectSpec, pack catalog.Pack) error {
	commandSpec, err := externalScaffoldCommand(targetDir, spec, pack)
	if err != nil {
		return err
	}

	command := exec.Command(commandSpec.name, commandSpec.args...)
	command.Dir = targetDir
	command.Env = append(os.Environ(), "CI=1")

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"run external scaffold for %s: %w: %s",
			pack.DisplayName,
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

type commandSpec struct {
	name string
	args []string
}

func externalScaffoldCommand(targetDir string, spec resolver.ProjectSpec, pack catalog.Pack) (commandSpec, error) {
	command, ok := externalCommandForPackageManager(pack, spec.PackageManager)
	if !ok {
		return commandSpec{}, fmt.Errorf("pack %q does not support package manager %q", pack.ID, spec.PackageManager)
	}

	args := make([]string, 0, len(command.Args)+8)
	projectDir := filepath.Join(targetDir, filepath.FromSlash(pack.OutputDir))
	for _, arg := range command.Args {
		args = append(args, strings.ReplaceAll(arg, "{{project_dir}}", projectDir))
	}

	args = append(args, packSpecificExternalArgs(pack.ID, spec.PackageManager)...)

	return commandSpec{
		name: args[0],
		args: args[1:],
	}, nil
}

func externalCommandForPackageManager(pack catalog.Pack, manager catalog.PackageManager) (catalog.ExternalCommand, bool) {
	if pack.External == nil {
		return catalog.ExternalCommand{}, false
	}

	for _, command := range pack.External.Commands {
		if command.PackageManager == manager {
			return command, true
		}
	}

	return catalog.ExternalCommand{}, false
}

func packSpecificExternalArgs(packID catalog.PackID, manager catalog.PackageManager) []string {
	switch packID {
	case catalog.PackIDNextJS:
		args := []string{
			"--ts",
			"--biome",
			"--tailwind",
			"--app",
			"--src-dir",
			"--import-alias",
			"@/*",
			"--yes",
			"--disable-git",
			"--skip-install",
		}

		switch manager {
		case catalog.PackageManagerNPM:
			args = append(args, "--use-npm")
		case catalog.PackageManagerPNPM:
			args = append(args, "--use-pnpm")
		case catalog.PackageManagerBun:
			args = append(args, "--use-bun")
		case catalog.PackageManagerYarn:
			args = append(args, "--use-yarn")
		}

		return args
	case catalog.PackIDHonoNode, catalog.PackIDHonoWorkers:
		return []string{"--pm", string(manager)}
	default:
		return nil
	}
}

func installDependencies(targetDir string, spec resolver.ProjectSpec, plan resolver.ResolvedPlan) ([]string, error) {
	if plan.WorkspaceStrategy != catalog.WorkspaceStrategyTurbo || spec.PackageManager == catalog.PackageManagerNone {
		return []string{"Dependency installation is not automated yet for native workspaces."}, nil
	}

	installCommand, ok := workspaceInstallCommand(spec.PackageManager)
	if !ok {
		return nil, fmt.Errorf("unsupported install command for package manager %q", spec.PackageManager)
	}

	command := exec.Command(installCommand[0], installCommand[1:]...)
	command.Dir = targetDir
	command.Env = os.Environ()

	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("install workspace dependencies: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil, nil
}

func workspaceInstallCommand(manager catalog.PackageManager) ([]string, bool) {
	switch manager {
	case catalog.PackageManagerNPM:
		return []string{"npm", "install"}, true
	case catalog.PackageManagerPNPM:
		return []string{"pnpm", "install"}, true
	case catalog.PackageManagerBun:
		return []string{"bun", "install"}, true
	case catalog.PackageManagerYarn:
		return []string{"yarn", "install"}, true
	default:
		return nil, false
	}
}

func renderTemplateAsset(assetPath string, data templateData) (string, error) {
	source, err := templates.Assets.ReadFile(assetPath)
	if err != nil {
		return "", fmt.Errorf("read template asset %q: %w", assetPath, err)
	}

	tmpl, err := texttemplate.New(assetPath).Parse(string(source))
	if err != nil {
		return "", fmt.Errorf("parse template asset %q: %w", assetPath, err)
	}

	buffer := &bytes.Buffer{}
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", fmt.Errorf("render template asset %q: %w", assetPath, err)
	}

	return buffer.String(), nil
}

func renderRootReadme(spec resolver.ProjectSpec, plan resolver.ResolvedPlan, packs []catalog.Pack) string {
	builder := &strings.Builder{}
	builder.WriteString("# ")
	builder.WriteString(spec.ProjectName)
	builder.WriteString("\n\n")
	builder.WriteString("Scaffolded with openrepo.\n\n")
	builder.WriteString("## Apps\n\n")

	for _, pack := range packs {
		builder.WriteString("- `")
		builder.WriteString(pack.OutputDir)
		builder.WriteString("`: ")
		builder.WriteString(pack.DisplayName)
		builder.WriteString("\n")
	}

	if plan.CreateSharedTypes {
		builder.WriteString("- `packages/shared-types`: shared TypeScript types for cross-app contracts\n")
	}

	builder.WriteString("\n## Workspace\n\n")
	builder.WriteString("- Strategy: `")
	builder.WriteString(string(plan.WorkspaceStrategy))
	builder.WriteString("`\n")

	if plan.WorkspaceStrategy == catalog.WorkspaceStrategyTurbo {
		runCommand := packageManagerRunCommand(spec.PackageManager, "dev")
		if runCommand != "" {
			builder.WriteString("- Root dev command: `")
			builder.WriteString(runCommand)
			builder.WriteString("`\n")
		}
	} else if makeCommand := "make test"; len(packs) > 0 {
		builder.WriteString("- Root task runner: `")
		builder.WriteString(makeCommand)
		builder.WriteString("`\n")
	}

	builder.WriteString("\n## Next Steps\n\n")

	if plan.WorkspaceStrategy == catalog.WorkspaceStrategyTurbo {
		installCommand := packageManagerInstallCommand(spec.PackageManager)
		if installCommand == "" {
			builder.WriteString("1. Fill in `.env.example` values for your local environment.\n")
			builder.WriteString("2. Run the app-specific setup commands documented in each generated app.\n")
			builder.WriteString("3. Start building inside the generated app directories.\n")
			return builder.String()
		}

		builder.WriteString("1. Install workspace dependencies with `")
		builder.WriteString(installCommand)
		builder.WriteString("` if you skipped automatic dependency installation.\n")
		builder.WriteString("2. Fill in `.env.example` values for your local environment.\n")
		builder.WriteString("3. Start building inside the generated app directories.\n")
		return builder.String()
	}

	builder.WriteString("1. Fill in `.env.example` values for your local environment.\n")
	builder.WriteString("2. Run the app-specific setup commands inside each generated app directory instead of installing from the repo root.\n")
	builder.WriteString("3. Use the generated root `Makefile` for shared tasks like `make dev` or `make test` when it exists.\n")

	return builder.String()
}

func renderGitignore(plan resolver.ResolvedPlan) string {
	lines := []string{
		".DS_Store",
		".env",
		".env.local",
		".venv",
		"node_modules",
		"dist",
		"coverage",
		".pytest_cache",
		".ruff_cache",
		"__pycache__",
		"bin",
	}

	if plan.WorkspaceStrategy == catalog.WorkspaceStrategyTurbo {
		lines = append(lines, ".turbo")
	}

	return strings.Join(lines, "\n") + "\n"
}

func renderRootEnvExample(spec resolver.ProjectSpec, packs []catalog.Pack) string {
	variables := make([]catalog.EnvVar, 0)
	seen := make(map[string]struct{})

	appendVars := func(items []catalog.EnvVar) {
		for _, item := range items {
			if _, ok := seen[item.Name]; ok {
				continue
			}
			seen[item.Name] = struct{}{}
			variables = append(variables, item)
		}
	}

	for _, pack := range packs {
		appendVars(pack.EnvVars)
	}

	appendVars(integrationEnvVars(spec))

	if len(variables) == 0 {
		return "# No environment variables declared for this selection yet.\n"
	}

	builder := &strings.Builder{}
	builder.WriteString("# Root environment values for ")
	builder.WriteString(spec.ProjectName)
	builder.WriteString("\n")

	for _, variable := range variables {
		builder.WriteString("\n")
		builder.WriteString("# ")
		builder.WriteString(variable.Description)
		if variable.Required {
			builder.WriteString(" (required)")
		}
		builder.WriteString("\n")
		builder.WriteString(variable.Name)
		builder.WriteString("=")
		builder.WriteString(variable.Example)
		builder.WriteString("\n")
	}

	return builder.String()
}

func renderRootAgentsFile(spec resolver.ProjectSpec, plan resolver.ResolvedPlan, packs []catalog.Pack, includeRecommendedSkills bool) string {
	builder := &strings.Builder{}
	builder.WriteString("# AGENTS.md\n\n")
	builder.WriteString("## Project\n\n")
	builder.WriteString("- Name: ")
	builder.WriteString(spec.ProjectName)
	builder.WriteString("\n")
	builder.WriteString("- Mode: ")
	builder.WriteString(string(spec.Mode))
	builder.WriteString("\n")
	builder.WriteString("- Workspace strategy: ")
	builder.WriteString(string(plan.WorkspaceStrategy))
	builder.WriteString("\n")

	builder.WriteString("\n## Apps\n\n")
	for _, pack := range packs {
		builder.WriteString("- ")
		builder.WriteString(pack.OutputDir)
		builder.WriteString(": ")
		builder.WriteString(pack.DisplayName)
		builder.WriteString("\n")
	}

	builder.WriteString("\n## Coding Standards\n\n")
	builder.WriteString("- Keep changes typed and explicit.\n")
	builder.WriteString("- Reuse existing implementations before adding new abstractions.\n")
	builder.WriteString("- Prefer simple extensions to the generated structure over broad rewrites.\n")

	builder.WriteString("\n## Stack Notes\n\n")
	for _, pack := range packs {
		builder.WriteString("- ")
		builder.WriteString(pack.DisplayName)
		builder.WriteString(": ")
		rules := make([]string, 0, len(pack.AgentRules))
		for _, rule := range pack.AgentRules {
			rules = append(rules, rule.Instruction)
		}
		builder.WriteString(strings.Join(rules, " "))
		builder.WriteString("\n")
	}

	builder.WriteString("\n## Skills\n\n")
	if includeRecommendedSkills {
		builder.WriteString("The recommended Codex skills for this scaffold have been copied into `.agents/skills/`.\n")
	} else {
		builder.WriteString("Use `.agents/skills/` for optional project-specific skill bundles when you choose to add them.\n")
	}

	return builder.String()
}

func clearGeneratedSkillAssets(targetDir string, packs []catalog.Pack) error {
	for _, pack := range packs {
		if pack.SkillAssets == nil {
			continue
		}

		entries, err := fs.ReadDir(templates.Assets, pack.SkillAssets.Path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("read skill asset bundle %q: %w", pack.SkillAssets.Path, err)
		}

		for _, entry := range entries {
			path := filepath.Join(targetDir, ".agents", "skills", entry.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("remove skill asset %q: %w", path, err)
			}
		}
	}

	return nil
}

func writeRecommendedSkillAssets(targetDir string, packs []catalog.Pack) error {
	for _, pack := range packs {
		if pack.SkillAssets == nil {
			continue
		}

		if err := copySkillAssetBundle(pack.SkillAssets.Path, filepath.Join(targetDir, ".agents", "skills")); err != nil {
			return err
		}
	}

	return nil
}

func copySkillAssetBundle(assetPath string, targetDir string) error {
	return fs.WalkDir(templates.Assets, assetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk skill asset path %q: %w", path, err)
		}

		relativePath, err := filepath.Rel(assetPath, path)
		if err != nil {
			return fmt.Errorf("compute relative skill asset path for %q: %w", path, err)
		}
		if relativePath == "." {
			return nil
		}

		outputPath := filepath.Join(targetDir, filepath.FromSlash(relativePath))
		if d.IsDir() {
			if err := os.MkdirAll(outputPath, 0o755); err != nil {
				return fmt.Errorf("create skill asset directory %q: %w", outputPath, err)
			}
			return nil
		}

		contents, err := templates.Assets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read skill asset %q: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("create parent directory for skill asset %q: %w", outputPath, err)
		}
		if err := os.WriteFile(outputPath, contents, 0o644); err != nil {
			return fmt.Errorf("write skill asset %q: %w", outputPath, err)
		}

		return nil
	})
}

func renderPackEnvExample(pack catalog.Pack, spec resolver.ProjectSpec) string {
	if len(pack.EnvVars) == 0 {
		return "# No app-specific environment variables declared for this pack yet.\n"
	}

	builder := &strings.Builder{}
	builder.WriteString("# ")
	builder.WriteString(pack.DisplayName)
	builder.WriteString(" environment values\n")

	if pack.ID == catalog.PackIDHonoWorkers {
		builder.WriteString("\n# Wrangler bindings for D1, KV, and R2 are configured in wrangler.jsonc.\n")
		builder.WriteString("# Wrangler 4.45+ can auto-provision those resources on first deploy and write the generated IDs back into that file.\n")
	}

	for _, variable := range pack.EnvVars {
		builder.WriteString("\n")
		builder.WriteString("# ")
		builder.WriteString(variable.Description)
		if variable.Required {
			builder.WriteString(" (required)")
		}
		builder.WriteString("\n")
		builder.WriteString(variable.Name)
		builder.WriteString("=")
		builder.WriteString(variable.Example)
		builder.WriteString("\n")
	}

	if pack.ID == catalog.PackIDHonoWorkers && spec.Database == catalog.DatabaseD1 {
		builder.WriteString("\n# Cloudflare D1, KV, and R2 are bound in wrangler.jsonc rather than exported as environment variables.\n")
		builder.WriteString("# Keep the generated binding stubs if you want Wrangler to auto-provision them, or replace them with fixed IDs and names from an existing account.\n")
	}

	return builder.String()
}

func renderPackAgentsFile(pack catalog.Pack, spec resolver.ProjectSpec) string {
	builder := &strings.Builder{}
	builder.WriteString("# AGENTS.md\n\n")
	builder.WriteString("## Stack\n\n")
	builder.WriteString("- ")
	builder.WriteString(pack.DisplayName)
	builder.WriteString("\n")
	builder.WriteString("- Runtime: ")
	builder.WriteString(string(pack.Runtime))
	builder.WriteString("\n")
	builder.WriteString("- Language: ")
	builder.WriteString(string(pack.Language))
	builder.WriteString("\n")

	if len(pack.Scripts) > 0 {
		builder.WriteString("\n## Scripts\n\n")
		for _, script := range pack.Scripts {
			builder.WriteString("- ")
			builder.WriteString(script.Name)
			builder.WriteString(": `")
			builder.WriteString(script.Command)
			builder.WriteString("`\n")
		}
	}

	if len(pack.AgentRules) > 0 {
		builder.WriteString("\n## Rules\n\n")
		for _, rule := range pack.AgentRules {
			builder.WriteString("- ")
			builder.WriteString(rule.Title)
			builder.WriteString(": ")
			builder.WriteString(rule.Instruction)
			builder.WriteString("\n")
		}
	}

	if pack.ID == catalog.PackIDHonoWorkers {
		builder.WriteString("\n## Cloudflare Setup\n\n")
		builder.WriteString("- Use Wrangler bindings for D1, KV, and R2 instead of external REST APIs.\n")
		builder.WriteString("- The generated wrangler.jsonc keeps shareable binding stubs so Wrangler can auto-provision and backfill IDs on first deploy.\n")
		builder.WriteString("- Run `wrangler types` after Wrangler config changes so `worker-configuration.d.ts` stays aligned with your bindings.\n")
		if spec.Database == catalog.DatabaseD1 {
			builder.WriteString("- The generated Workers config assumes D1 is the primary database binding.\n")
		}
	}

	return builder.String()
}

func renderWorkersWranglerConfig(spec resolver.ProjectSpec) string {
	projectSlug := projectSlug(spec.ProjectName)
	return fmt.Sprintf("{\n  \"$schema\": \"node_modules/wrangler/config-schema.json\",\n  \"name\": \"%s-api\",\n  \"main\": \"src/index.ts\",\n  \"compatibility_date\": \"%s\",\n  \"compatibility_flags\": [\"nodejs_compat\"],\n  \"observability\": {\n    \"enabled\": true,\n    \"head_sampling_rate\": 1\n  },\n  \"vars\": {\n    \"APP_ENV\": \"development\"\n  },\n  // Shareable starter bindings: Wrangler 4.45+ can provision D1, KV, and R2 on first deploy\n  // and then write the generated IDs back into this file. Replace these stubs with fixed IDs\n  // and names if you are pointing at pre-existing Cloudflare resources.\n  \"kv_namespaces\": [\n    {\n      \"binding\": \"CACHE\"\n    }\n  ],\n  \"r2_buckets\": [\n    {\n      \"binding\": \"ASSETS\"\n    }\n  ],\n  \"d1_databases\": [\n    {\n      \"binding\": \"DB\"\n    }\n  ],\n  \"env\": {\n    \"staging\": {\n      \"name\": \"%s-api-staging\",\n      \"vars\": {\n        \"APP_ENV\": \"staging\"\n      },\n      \"kv_namespaces\": [\n        {\n          \"binding\": \"CACHE\"\n        }\n      ],\n      \"r2_buckets\": [\n        {\n          \"binding\": \"ASSETS\"\n        }\n      ],\n      \"d1_databases\": [\n        {\n          \"binding\": \"DB\"\n        }\n      ]\n    },\n    \"production\": {\n      \"name\": \"%s-api-production\",\n      \"vars\": {\n        \"APP_ENV\": \"production\"\n      },\n      \"kv_namespaces\": [\n        {\n          \"binding\": \"CACHE\"\n        }\n      ],\n      \"r2_buckets\": [\n        {\n          \"binding\": \"ASSETS\"\n        }\n      ],\n      \"d1_databases\": [\n        {\n          \"binding\": \"DB\"\n        }\n      ]\n    }\n  }\n}\n", projectSlug, time.Now().Format("2006-01-02"), projectSlug, projectSlug)
}

func renderTurboPackageJSON(spec resolver.ProjectSpec) (string, error) {
	payload := map[string]any{
		"name":       projectSlug(spec.ProjectName),
		"private":    true,
		"workspaces": []string{"apps/*", "packages/*"},
		"scripts": map[string]string{
			"dev":         "turbo run dev --parallel",
			"build":       "turbo run build",
			"lint":        "turbo run lint",
			"test":        "turbo run test",
			"biome":       "biome check .",
			"biome:write": "biome format --write .",
		},
		"devDependencies": map[string]string{
			"@biomejs/biome": "latest",
			"turbo":          "latest",
		},
	}

	contents, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("render root package.json: %w", err)
	}

	return string(contents) + "\n", nil
}

func renderTurboConfig() string {
	return "{\n  \"$schema\": \"https://turbo.build/schema.json\",\n  \"tasks\": {\n    \"build\": {\n      \"dependsOn\": [\"^build\"],\n      \"outputs\": [\"dist/**\", \".next/**\"]\n    },\n    \"dev\": {\n      \"cache\": false,\n      \"persistent\": true\n    },\n    \"lint\": {\n      \"outputs\": []\n    },\n    \"format\": {\n      \"cache\": false,\n      \"outputs\": []\n    },\n    \"test\": {\n      \"outputs\": []\n    }\n  }\n}\n"
}

func renderBiomeConfig() string {
	return "{\n  \"$schema\": \"https://biomejs.dev/schemas/latest/schema.json\",\n  \"formatter\": {\n    \"indentStyle\": \"space\"\n  },\n  \"linter\": {\n    \"enabled\": true\n  },\n  \"files\": {\n    \"ignore\": [\"**/dist/**\", \"**/.next/**\", \"**/coverage/**\"]\n  }\n}\n"
}

func hasTypeScriptPack(packs []catalog.Pack) bool {
	for _, pack := range packs {
		if pack.Language == catalog.LanguageTypeScript {
			return true
		}
	}

	return false
}

func renderPNPMWorkspaceConfig() string {
	return "packages:\n  - apps/*\n  - packages/*\n"
}

func renderSharedTypesPackageJSON(spec resolver.ProjectSpec) (string, error) {
	payload := map[string]any{
		"name":    "@" + projectSlug(spec.ProjectName) + "/shared-types",
		"private": true,
		"type":    "module",
		"version": "0.0.0",
		"exports": map[string]string{
			".": "./src/index.ts",
		},
	}

	contents, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("render shared-types package.json: %w", err)
	}

	return string(contents) + "\n", nil
}

func renderSharedTypesTSConfig() string {
	return "{\n  \"compilerOptions\": {\n    \"composite\": true,\n    \"declaration\": true,\n    \"emitDeclarationOnly\": true,\n    \"module\": \"ESNext\",\n    \"moduleResolution\": \"Bundler\",\n    \"target\": \"ES2022\",\n    \"strict\": true\n  },\n  \"include\": [\"src\"]\n}\n"
}

func renderSharedTypesReadme(spec resolver.ProjectSpec) string {
	return fmt.Sprintf("# Shared Types\n\nCommon TypeScript contracts for the `%s` workspace.\n", spec.ProjectName)
}

func renderMakefile(packs []catalog.Pack) string {
	targetOrder := []string{"dev", "build", "start", "lint", "test", "fmt"}
	commandsByTarget := make(map[string][]string)

	for _, target := range targetOrder {
		for _, pack := range packs {
			for _, script := range pack.Scripts {
				if script.Name != target {
					continue
				}

				commandsByTarget[target] = append(
					commandsByTarget[target],
					fmt.Sprintf("cd %s && %s", pack.OutputDir, script.Command),
				)
			}
		}
	}

	availableTargets := make([]string, 0)
	for _, target := range targetOrder {
		if len(commandsByTarget[target]) > 0 {
			availableTargets = append(availableTargets, target)
		}
	}

	if len(availableTargets) == 0 {
		return ""
	}

	builder := &strings.Builder{}
	builder.WriteString(".PHONY:")
	for _, target := range availableTargets {
		builder.WriteString(" ")
		builder.WriteString(target)
	}
	builder.WriteString("\n\n")

	for _, target := range availableTargets {
		builder.WriteString(target)
		builder.WriteString(":\n")

		if target == "dev" && len(commandsByTarget[target]) > 1 {
			builder.WriteString("\t@set -e; \\\n")
			builder.WriteString("\ttrap 'kill 0' EXIT INT TERM; \\\n")
			for _, command := range commandsByTarget[target] {
				builder.WriteString("\t(")
				builder.WriteString(command)
				builder.WriteString(") & \\\n")
			}
			builder.WriteString("\twait\n\n")
			continue
		}

		for _, command := range commandsByTarget[target] {
			builder.WriteString("\t")
			builder.WriteString(command)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func integrationEnvVars(spec resolver.ProjectSpec) []catalog.EnvVar {
	var variables []catalog.EnvVar

	switch spec.Database {
	case catalog.DatabasePostgres:
		variables = append(variables, catalog.EnvVar{Name: "DATABASE_URL", Example: "postgres://postgres:postgres@localhost:5432/app", Required: true, Description: "Connection string for the primary Postgres database."})
	case catalog.DatabaseSQLite:
		variables = append(variables, catalog.EnvVar{Name: "DATABASE_URL", Example: "sqlite:///./app.db", Required: true, Description: "Connection string for the local SQLite database."})
	case catalog.DatabaseSupabase:
		variables = append(variables,
			catalog.EnvVar{Name: "DATABASE_URL", Example: "postgres://postgres:postgres@db.your-project.supabase.co:5432/postgres", Required: true, Description: "Connection string for the Supabase Postgres database used by the backend."},
			catalog.EnvVar{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: false, Description: "Optional Supabase project URL for client or auth integrations."},
			catalog.EnvVar{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: false, Description: "Optional Supabase anonymous client key for client or auth integrations."},
		)
	}

	switch spec.Auth {
	case catalog.AuthBetter:
		variables = append(variables,
			catalog.EnvVar{Name: "BETTER_AUTH_SECRET", Example: "replace-me", Required: true, Description: "Application secret used by Better Auth."},
			catalog.EnvVar{Name: "BETTER_AUTH_URL", Example: "http://localhost:3000", Required: true, Description: "Base URL used by Better Auth callbacks."},
		)
	case catalog.AuthSupabase:
		variables = append(variables,
			catalog.EnvVar{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL used for auth flows."},
			catalog.EnvVar{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous client key used for auth flows."},
		)
	}

	switch spec.Storage {
	case catalog.StorageR2:
		variables = append(variables,
			catalog.EnvVar{Name: "R2_BUCKET", Example: "app-assets", Required: true, Description: "Cloudflare R2 bucket name."},
			catalog.EnvVar{Name: "R2_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "R2 access key id."},
			catalog.EnvVar{Name: "R2_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "R2 secret access key."},
		)
	case catalog.StorageS3:
		variables = append(variables,
			catalog.EnvVar{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
			catalog.EnvVar{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
			catalog.EnvVar{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
			catalog.EnvVar{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
		)
	}

	if spec.Email == catalog.EmailResend {
		variables = append(variables, catalog.EnvVar{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."})
	}

	return variables
}

func packageManagerInstallCommand(manager catalog.PackageManager) string {
	switch manager {
	case catalog.PackageManagerNPM:
		return "npm install"
	case catalog.PackageManagerPNPM:
		return "pnpm install"
	case catalog.PackageManagerBun:
		return "bun install"
	case catalog.PackageManagerYarn:
		return "yarn install"
	default:
		return ""
	}
}

func packageManagerRunCommand(manager catalog.PackageManager, script string) string {
	switch manager {
	case catalog.PackageManagerNPM:
		return "npm run " + script
	case catalog.PackageManagerPNPM:
		return "pnpm " + script
	case catalog.PackageManagerBun:
		return "bun run " + script
	case catalog.PackageManagerYarn:
		return "yarn " + script
	default:
		return ""
	}
}

func projectSlug(projectName string) string {
	projectName = strings.TrimSpace(strings.ToLower(projectName))
	if projectName == "" {
		return "openrepo-app"
	}

	var builder strings.Builder
	lastWasDash := false

	for _, character := range projectName {
		switch {
		case character >= 'a' && character <= 'z':
			builder.WriteRune(character)
			lastWasDash = false
		case character >= '0' && character <= '9':
			builder.WriteRune(character)
			lastWasDash = false
		default:
			if builder.Len() == 0 || lastWasDash {
				continue
			}
			builder.WriteByte('-')
			lastWasDash = true
		}
	}

	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "openrepo-app"
	}

	return slug
}

func goModulePath(projectName string, outputDir string) string {
	return projectSlug(projectName) + "/" + strings.TrimPrefix(outputDir, "./")
}

func writeFile(path string, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %q: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}

	return nil
}

func writeFileIfMissing(path string, contents string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect existing file %q: %w", path, err)
	}

	return writeFile(path, contents)
}
