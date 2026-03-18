package cli

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/resolver"
	"github.com/spf13/cobra"
)

type commandFlagState struct {
	gitInitSet           bool
	installSet           bool
	recommendedSkillsSet bool
}

type reviewAction string

const (
	noneSelectionValue                         = "none"
	reviewActionCreate            reviewAction = "create"
	reviewActionMode              reviewAction = "mode"
	reviewActionFrontend          reviewAction = "frontend"
	reviewActionBackend           reviewAction = "backend"
	reviewActionPackageManager    reviewAction = "package-manager"
	reviewActionAuth              reviewAction = "auth"
	reviewActionDatabase          reviewAction = "database"
	reviewActionStorage           reviewAction = "storage"
	reviewActionEmail             reviewAction = "email"
	reviewActionGitInit           reviewAction = "git-init"
	reviewActionInstall           reviewAction = "install"
	reviewActionRecommendedSkills reviewAction = "recommended-skills"
)

type createInput struct {
	ProjectName       string
	Mode              string
	Frontend          string
	Backend           string
	PackageManager    string
	Database          string
	Auth              string
	Storage           string
	Email             string
	GitInit           bool
	Install           bool
	RecommendedSkills bool
}

func newCreateInput(options createOptions) createInput {
	return createInput{
		ProjectName:       strings.TrimSpace(options.projectName),
		Mode:              normalizeValue(options.mode),
		Frontend:          normalizeValue(options.frontend),
		Backend:           normalizeValue(options.backend),
		PackageManager:    normalizeValue(options.packageManager),
		Database:          normalizeValue(options.database),
		Auth:              normalizeValue(options.auth),
		Storage:           normalizeValue(options.storage),
		Email:             normalizeValue(options.email),
		GitInit:           options.gitInit,
		Install:           options.install,
		RecommendedSkills: options.recommendedSkills,
	}
}

func promptForMissingValues(cmd *cobra.Command, registry catalog.Registry, input *createInput, flagState commandFlagState) error {
	if err := promptSelectionSteps(cmd, registry, input, flagState); err != nil {
		return err
	}

	reviewFlagState := commandFlagState{
		gitInitSet:           true,
		installSet:           true,
		recommendedSkillsSet: true,
	}

	for {
		action, err := promptReviewStep(cmd, registry, *input)
		if err != nil {
			return err
		}

		if action == reviewActionCreate {
			return nil
		}

		if action == reviewActionGitInit {
			if err := promptGitInit(cmd, input); err != nil {
				return err
			}
			continue
		}

		if action == reviewActionInstall {
			if err := promptInstall(cmd, input); err != nil {
				return err
			}
			continue
		}

		if action == reviewActionRecommendedSkills {
			if err := promptRecommendedSkills(cmd, registry, input); err != nil {
				return err
			}
			continue
		}

		resetInputForReview(input, action)

		if err := promptSelectionSteps(cmd, registry, input, reviewFlagState); err != nil {
			return err
		}
	}
}

func frontendOptions(registry catalog.Registry) []huh.Option[string] {
	return packOptions(registry, catalog.PackCategoryFrontend, catalog.PackIDNextJS, nil)
}

func backendOptions(registry catalog.Registry) []huh.Option[string] {
	return packOptions(registry, catalog.PackCategoryBackend, catalog.PackIDHonoNode, func(pack catalog.Pack) string {
		if pack.ID == catalog.PackIDHonoWorkers {
			return " (Wrangler envs: dev, staging, production + D1/KV/R2)"
		}
		return ""
	})
}

func packOptions(registry catalog.Registry, category catalog.PackCategory, recommendedID catalog.PackID, labelSuffix func(catalog.Pack) string) []huh.Option[string] {
	options := make([]huh.Option[string], 0)
	var recommended *huh.Option[string]

	for _, pack := range registry.All() {
		if !pack.SupportsCategory(category) {
			continue
		}

		label := pack.DisplayName
		if labelSuffix != nil {
			label += labelSuffix(pack)
		}

		option := huh.NewOption(label, string(pack.ID))
		if pack.ID == recommendedID {
			recommended = &option
			continue
		}

		options = append(options, option)
	}

	if recommended != nil {
		options = append([]huh.Option[string]{*recommended}, options...)
	}

	return options
}

func packageManagerOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	allowed := allowedPackageManagers(registry, input)
	options := make([]huh.Option[string], 0, len(allowed))
	recommended := recommendedPackageManager(registry, input, allowed)

	if recommended != catalog.PackageManagerNone {
		options = append(options, huh.NewOption(string(recommended)+" (recommended)", string(recommended)))
	}

	for _, manager := range preferredPackageManagerOrder() {
		if !slices.Contains(allowed, manager) || manager == recommended {
			continue
		}

		options = append(options, huh.NewOption(string(manager), string(manager)))
	}

	return options
}

func authPromptOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	if input.Mode == string(catalog.ProjectModeFrontend) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	packs := selectedPacks(registry, input)
	recommended := recommendedAuthOption(packs)
	options := make([]huh.Option[string], 0, 3)

	if recommended == catalog.AuthBetter {
		options = append(options, huh.NewOption("Better Auth (recommended)", string(catalog.AuthBetter)))
	}
	if recommended == catalog.AuthSupabase {
		options = append(options, huh.NewOption("Supabase Auth (recommended)", string(catalog.AuthSupabase)))
	}

	if hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsBetterAuth }) && recommended != catalog.AuthBetter {
		options = append(options, huh.NewOption("Better Auth", string(catalog.AuthBetter)))
	}
	if hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsSupabaseAuth }) && recommended != catalog.AuthSupabase {
		options = append(options, huh.NewOption("Supabase Auth", string(catalog.AuthSupabase)))
	}

	if recommended == catalog.AuthNone || len(options) == 0 {
		options = append([]huh.Option[string]{huh.NewOption("None", noneSelectionValue)}, options...)
	} else {
		options = append(options, huh.NewOption("None", noneSelectionValue))
	}

	return options
}

func databasePromptOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	if input.Mode == string(catalog.ProjectModeFrontend) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	packs := selectedPacks(registry, input)
	if usesCloudflareWorkers(input) {
		return []huh.Option[string]{
			huh.NewOption("Cloudflare D1 (required for Workers; KV + R2 are configured too)", string(catalog.DatabaseD1)),
		}
	}
	if !hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsDatabase }) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	options := []huh.Option[string]{
		huh.NewOption("Postgres (recommended)", string(catalog.DatabasePostgres)),
		huh.NewOption("SQLite", string(catalog.DatabaseSQLite)),
		huh.NewOption("Supabase", string(catalog.DatabaseSupabase)),
	}

	return append(options, huh.NewOption("None", noneSelectionValue))
}

func storagePromptOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	if input.Mode == string(catalog.ProjectModeFrontend) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	packs := selectedPacks(registry, input)
	if usesCloudflareWorkers(input) {
		return []huh.Option[string]{
			huh.NewOption("Cloudflare R2 (required for Workers; D1 + KV are configured too)", string(catalog.StorageR2)),
		}
	}
	if !hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsStorage }) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	options := []huh.Option[string]{
		huh.NewOption("Cloudflare R2 (recommended)", string(catalog.StorageR2)),
		huh.NewOption("Amazon S3", string(catalog.StorageS3)),
		huh.NewOption("Supabase Storage", string(catalog.StorageSupabase)),
	}

	return append(options, huh.NewOption("None", noneSelectionValue))
}

func emailPromptOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	if input.Mode == string(catalog.ProjectModeFrontend) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	packs := selectedPacks(registry, input)
	if !hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsEmail }) {
		return []huh.Option[string]{huh.NewOption("None", noneSelectionValue)}
	}

	return []huh.Option[string]{
		huh.NewOption("Resend (recommended)", string(catalog.EmailResend)),
		huh.NewOption("None", noneSelectionValue),
	}
}

func shouldPromptPackageManager(registry catalog.Registry, input createInput) bool {
	return len(allowedPackageManagers(registry, input)) > 0
}

func allowedPackageManagers(registry catalog.Registry, input createInput) []catalog.PackageManager {
	packs := selectedPacks(registry, input)
	var allowed []catalog.PackageManager

	for _, pack := range packs {
		if pack.Language != catalog.LanguageTypeScript {
			continue
		}

		if allowed == nil {
			allowed = append(allowed, supportedManagersForPack(pack)...)
			continue
		}

		allowed = intersectPackageManagers(allowed, supportedManagersForPack(pack))
	}

	return allowed
}

func recommendedPackageManager(registry catalog.Registry, input createInput, allowed []catalog.PackageManager) catalog.PackageManager {
	for _, pack := range selectedPacks(registry, input) {
		if pack.External == nil {
			continue
		}

		if slices.Contains(allowed, pack.External.RecommendedPackageManager) {
			return pack.External.RecommendedPackageManager
		}
	}

	if len(allowed) > 0 {
		return allowed[0]
	}

	return catalog.PackageManagerNone
}

func preferredPackageManagerOrder() []catalog.PackageManager {
	return []catalog.PackageManager{
		catalog.PackageManagerPNPM,
		catalog.PackageManagerNPM,
		catalog.PackageManagerBun,
		catalog.PackageManagerYarn,
	}
}

func supportedManagersForPack(pack catalog.Pack) []catalog.PackageManager {
	managers := make([]catalog.PackageManager, 0)

	if pack.External == nil {
		return managers
	}

	for _, command := range pack.External.Commands {
		managers = append(managers, command.PackageManager)
	}

	return managers
}

func intersectPackageManagers(left []catalog.PackageManager, right []catalog.PackageManager) []catalog.PackageManager {
	intersection := make([]catalog.PackageManager, 0)

	for _, candidate := range left {
		if slices.Contains(right, candidate) && !slices.Contains(intersection, candidate) {
			intersection = append(intersection, candidate)
		}
	}

	return intersection
}

func selectedPacks(registry catalog.Registry, input createInput) []catalog.Pack {
	packs := make([]catalog.Pack, 0, 2)

	if input.Frontend != "" {
		if pack, ok := registry.Get(catalog.PackID(input.Frontend)); ok {
			packs = append(packs, pack)
		}
	}

	if input.Backend != "" {
		if pack, ok := registry.Get(catalog.PackID(input.Backend)); ok {
			packs = append(packs, pack)
		}
	}

	return packs
}

func hasCapability(packs []catalog.Pack, supports func(pack catalog.Pack) bool) bool {
	for _, pack := range packs {
		if supports(pack) {
			return true
		}
	}

	return false
}

func hasRecommendedSkills(registry catalog.Registry, input createInput) bool {
	for _, pack := range selectedPacks(registry, input) {
		if pack.SkillAssets != nil {
			return true
		}
	}

	for _, addon := range selectedAddons(input) {
		if addon.SkillAssets != nil {
			return true
		}
	}

	return false
}

func selectedAddons(input createInput) []catalog.Addon {
	if input.Backend == "" {
		return nil
	}

	return catalog.MustDefaultAddonRegistry().Resolve(
		catalog.PackID(input.Backend),
		authOptionFromInput(input.Auth),
		databaseOptionFromInput(input.Database),
		storageOptionFromInput(input.Storage),
		emailOptionFromInput(input.Email),
	)
}

func authOptionFromInput(value string) catalog.AuthOption {
	option, err := parseAuthOption(value)
	if err != nil {
		return catalog.AuthNone
	}

	return option
}

func databaseOptionFromInput(value string) catalog.DatabaseOption {
	option, err := parseDatabaseOption(value)
	if err != nil {
		return catalog.DatabaseNone
	}

	return option
}

func storageOptionFromInput(value string) catalog.StorageOption {
	option, err := parseStorageOption(value)
	if err != nil {
		return catalog.StorageNone
	}

	return option
}

func emailOptionFromInput(value string) catalog.EmailOption {
	option, err := parseEmailOption(value)
	if err != nil {
		return catalog.EmailNone
	}

	return option
}

func requiresFrontendPack(mode string) bool {
	return mode == string(catalog.ProjectModeFrontend) || mode == string(catalog.ProjectModeFullStack)
}

func requiresBackendPack(mode string) bool {
	return mode == string(catalog.ProjectModeBackend) || mode == string(catalog.ProjectModeFullStack)
}

func normalizeValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func promptSelectionSteps(cmd *cobra.Command, registry catalog.Registry, input *createInput, flagState commandFlagState) error {
	applySelectionConstraints(registry, input)

	if input.ProjectName == "" {
		field := huh.NewInput().
			Title("Project name").
			Value(&input.ProjectName).
			Validate(func(value string) error {
				if strings.TrimSpace(value) == "" {
					return errors.New("project name is required")
				}

				return nil
			})

		if err := runPromptForm(cmd, field); err != nil {
			return err
		}

		input.ProjectName = strings.TrimSpace(input.ProjectName)
	}

	if input.Mode == "" {
		field := huh.NewSelect[string]().
			Title("What do you want to build?").
			Value(&input.Mode).
			Options(
				huh.NewOption("Fullstack", string(catalog.ProjectModeFullStack)),
				huh.NewOption("Frontend only", string(catalog.ProjectModeFrontend)),
				huh.NewOption("Backend only", string(catalog.ProjectModeBackend)),
			)

		if err := runPromptForm(cmd, field); err != nil {
			return err
		}
	}

	applySelectionConstraints(registry, input)

	if requiresFrontendPack(input.Mode) && input.Frontend == "" {
		input.Frontend = string(catalog.PackIDNextJS)
		field := huh.NewSelect[string]().
			Title("Frontend stack").
			Value(&input.Frontend).
			Options(frontendOptions(registry)...)

		if err := runPromptForm(cmd, field); err != nil {
			return err
		}
	}

	applySelectionConstraints(registry, input)

	if requiresBackendPack(input.Mode) && input.Backend == "" {
		input.Backend = string(catalog.PackIDHonoNode)
		field := huh.NewSelect[string]().
			Title("Backend stack").
			Description("Cloudflare Workers includes Wrangler dev, staging, and production environments with D1, KV, and R2 bindings.").
			Value(&input.Backend).
			Options(backendOptions(registry)...)

		if err := runPromptForm(cmd, field); err != nil {
			return err
		}
	}

	applySelectionConstraints(registry, input)
	applyWorkersLockedDefaults(input)

	if shouldPromptPackageManager(registry, *input) && input.PackageManager == "" {
		recommended := recommendedPackageManager(registry, *input, allowedPackageManagers(registry, *input))
		if recommended != catalog.PackageManagerNone {
			input.PackageManager = string(recommended)
		}

		field := huh.NewSelect[string]().
			Title("JavaScript package manager").
			Value(&input.PackageManager).
			Options(packageManagerOptions(registry, *input)...)

		if err := runPromptForm(cmd, field); err != nil {
			return err
		}
	}

	applySelectionConstraints(registry, input)

	if input.Auth == "" {
		authOptions := authPromptOptions(registry, *input)
		if len(authOptions) == 1 {
			input.Auth = authOptions[0].Value
		} else {
			recommended := recommendedAuthOption(selectedPacks(registry, *input))
			if recommended != catalog.AuthNone {
				input.Auth = string(recommended)
			}
			field := huh.NewSelect[string]().
				Title("Authentication").
				Value(&input.Auth).
				Options(authOptions...)

			if err := runPromptForm(cmd, field); err != nil {
				return err
			}
		}
	}

	applySelectionConstraints(registry, input)

	if input.Database == "" {
		databaseOptions := databasePromptOptions(registry, *input)
		if len(databaseOptions) == 1 {
			input.Database = databaseOptions[0].Value
		} else {
			input.Database = string(recommendedDatabaseOption(*input))
			field := huh.NewSelect[string]().
				Title("Database").
				Value(&input.Database).
				Options(databaseOptions...)

			if err := runPromptForm(cmd, field); err != nil {
				return err
			}
		}
	}

	applySelectionConstraints(registry, input)

	if input.Storage == "" {
		storageOptions := storagePromptOptions(registry, *input)
		if len(storageOptions) == 1 {
			input.Storage = storageOptions[0].Value
		} else {
			input.Storage = string(recommendedStorageOption(*input))
			field := huh.NewSelect[string]().
				Title("Storage").
				Value(&input.Storage).
				Options(storageOptions...)

			if err := runPromptForm(cmd, field); err != nil {
				return err
			}
		}
	}

	applySelectionConstraints(registry, input)

	if input.Email == "" {
		emailOptions := emailPromptOptions(registry, *input)
		if len(emailOptions) == 1 {
			input.Email = emailOptions[0].Value
		} else {
			input.Email = string(recommendedEmailOption(selectedPacks(registry, *input)))
			field := huh.NewSelect[string]().
				Title("Email").
				Value(&input.Email).
				Options(emailOptions...)

			if err := runPromptForm(cmd, field); err != nil {
				return err
			}
		}
	}

	if !flagState.gitInitSet {
		if err := promptGitInit(cmd, input); err != nil {
			return err
		}
	}

	if !flagState.installSet {
		if err := promptInstall(cmd, input); err != nil {
			return err
		}
	}

	if !flagState.recommendedSkillsSet && hasRecommendedSkills(registry, *input) {
		input.RecommendedSkills = true
		if err := promptRecommendedSkills(cmd, registry, input); err != nil {
			return err
		}
	}

	return nil
}

func promptReviewStep(cmd *cobra.Command, registry catalog.Registry, input createInput) (reviewAction, error) {
	action := string(reviewActionCreate)
	field := huh.NewSelect[string]().
		Title("Review selections").
		Description("Create the project now, or jump back and change one choice.").
		Value(&action).
		Options(reviewOptions(registry, input)...)

	if err := runPromptForm(cmd, field); err != nil {
		return "", err
	}

	return reviewAction(action), nil
}

func reviewOptions(registry catalog.Registry, input createInput) []huh.Option[string] {
	options := []huh.Option[string]{
		huh.NewOption("Create project", string(reviewActionCreate)),
		huh.NewOption("Change mode ("+displayValue(input.Mode, "unset")+")", string(reviewActionMode)),
		huh.NewOption("Change initialize git ("+boolLabel(input.GitInit)+")", string(reviewActionGitInit)),
		huh.NewOption("Change install dependencies ("+boolLabel(input.Install)+")", string(reviewActionInstall)),
	}

	if requiresFrontendPack(input.Mode) {
		options = append(options, huh.NewOption("Change frontend stack ("+packDisplayName(registry, input.Frontend)+")", string(reviewActionFrontend)))
	}
	if requiresBackendPack(input.Mode) {
		options = append(options, huh.NewOption("Change backend stack ("+packDisplayName(registry, input.Backend)+")", string(reviewActionBackend)))
	}
	if shouldPromptPackageManager(registry, input) {
		options = append(options, huh.NewOption("Change package manager ("+displayValue(input.PackageManager, "unset")+")", string(reviewActionPackageManager)))
	}
	if len(authPromptOptions(registry, input)) > 1 {
		options = append(options, huh.NewOption("Change auth ("+displayValue(input.Auth, "none")+")", string(reviewActionAuth)))
	}
	if len(databasePromptOptions(registry, input)) > 1 {
		options = append(options, huh.NewOption("Change database ("+displayValue(input.Database, "none")+")", string(reviewActionDatabase)))
	}
	if len(storagePromptOptions(registry, input)) > 1 {
		options = append(options, huh.NewOption("Change storage ("+displayValue(input.Storage, "none")+")", string(reviewActionStorage)))
	}
	if len(emailPromptOptions(registry, input)) > 1 {
		options = append(options, huh.NewOption("Change email ("+displayValue(input.Email, "none")+")", string(reviewActionEmail)))
	}
	if hasRecommendedSkills(registry, input) || input.RecommendedSkills {
		options = append(options, huh.NewOption("Change recommended skills ("+boolLabel(input.RecommendedSkills)+")", string(reviewActionRecommendedSkills)))
	}

	return options
}

func resetInputForReview(input *createInput, action reviewAction) {
	switch action {
	case reviewActionMode:
		input.Mode = ""
		input.Frontend = ""
		input.Backend = ""
		input.PackageManager = ""
		input.Auth = ""
		input.Database = ""
		input.Storage = ""
		input.Email = ""
	case reviewActionFrontend:
		input.Frontend = ""
		input.PackageManager = ""
		input.Auth = ""
	case reviewActionBackend:
		input.Backend = ""
		input.PackageManager = ""
		input.Auth = ""
		input.Database = ""
		input.Storage = ""
		input.Email = ""
	case reviewActionPackageManager:
		input.PackageManager = ""
	case reviewActionAuth:
		input.Auth = ""
	case reviewActionDatabase:
		input.Database = ""
	case reviewActionStorage:
		input.Storage = ""
	case reviewActionEmail:
		input.Email = ""
	case reviewActionGitInit, reviewActionInstall, reviewActionRecommendedSkills:
	}
}

func packDisplayName(registry catalog.Registry, packID string) string {
	if packID == "" {
		return "unset"
	}

	pack, ok := registry.Get(catalog.PackID(packID))
	if !ok {
		return packID
	}

	return pack.DisplayName
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

func promptGitInit(cmd *cobra.Command, input *createInput) error {
	field := huh.NewConfirm().
		Title("Initialize git?").
		Value(&input.GitInit)

	return runPromptForm(cmd, field)
}

func promptInstall(cmd *cobra.Command, input *createInput) error {
	field := huh.NewConfirm().
		Title("Install dependencies?").
		Value(&input.Install)

	return runPromptForm(cmd, field)
}

func promptRecommendedSkills(cmd *cobra.Command, registry catalog.Registry, input *createInput) error {
	if !hasRecommendedSkills(registry, *input) {
		input.RecommendedSkills = false
		return nil
	}

	field := huh.NewConfirm().
		Title("Copy recommended skills?").
		Description("Copy pack- and integration-specific skill bundles into .agents/skills for coding agents.").
		Value(&input.RecommendedSkills)

	return runPromptForm(cmd, field)
}

func (input createInput) toSpec() (resolver.ProjectSpec, createSelections, error) {
	mode, err := parseProjectMode(input.Mode)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	packageManager, err := parsePackageManager(input.PackageManager)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	database, err := parseDatabaseOption(input.Database)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	auth, err := parseAuthOption(input.Auth)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	storage, err := parseStorageOption(input.Storage)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	email, err := parseEmailOption(input.Email)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	spec := resolver.ProjectSpec{
		ProjectName:    strings.TrimSpace(input.ProjectName),
		Mode:           mode,
		FrontendPackID: catalog.PackID(input.Frontend),
		BackendPackID:  catalog.PackID(input.Backend),
		PackageManager: packageManager,
		Database:       database,
		Auth:           auth,
		Storage:        storage,
		Email:          email,
	}

	selections := createSelections{
		InitializeGit:            input.GitInit,
		InstallDependencies:      input.Install,
		IncludeRecommendedSkills: input.RecommendedSkills,
	}

	return spec, selections, nil
}

func parseProjectMode(value string) (catalog.ProjectMode, error) {
	switch value {
	case string(catalog.ProjectModeFrontend):
		return catalog.ProjectModeFrontend, nil
	case string(catalog.ProjectModeBackend):
		return catalog.ProjectModeBackend, nil
	case string(catalog.ProjectModeFullStack):
		return catalog.ProjectModeFullStack, nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("unsupported mode %q", value)
	}
}

func parsePackageManager(value string) (catalog.PackageManager, error) {
	switch value {
	case "":
		return catalog.PackageManagerNone, nil
	case string(catalog.PackageManagerNPM):
		return catalog.PackageManagerNPM, nil
	case string(catalog.PackageManagerPNPM):
		return catalog.PackageManagerPNPM, nil
	case string(catalog.PackageManagerBun):
		return catalog.PackageManagerBun, nil
	case string(catalog.PackageManagerYarn):
		return catalog.PackageManagerYarn, nil
	default:
		return catalog.PackageManagerNone, fmt.Errorf("unsupported package manager %q", value)
	}
}

func parseDatabaseOption(value string) (catalog.DatabaseOption, error) {
	switch value {
	case "", "none":
		return catalog.DatabaseNone, nil
	case string(catalog.DatabaseD1):
		return catalog.DatabaseD1, nil
	case string(catalog.DatabasePostgres):
		return catalog.DatabasePostgres, nil
	case string(catalog.DatabaseSQLite):
		return catalog.DatabaseSQLite, nil
	case string(catalog.DatabaseSupabase):
		return catalog.DatabaseSupabase, nil
	default:
		return catalog.DatabaseNone, fmt.Errorf("unsupported database %q", value)
	}
}

func parseAuthOption(value string) (catalog.AuthOption, error) {
	switch value {
	case "", "none":
		return catalog.AuthNone, nil
	case string(catalog.AuthBetter):
		return catalog.AuthBetter, nil
	case string(catalog.AuthSupabase):
		return catalog.AuthSupabase, nil
	default:
		return catalog.AuthNone, fmt.Errorf("unsupported auth option %q", value)
	}
}

func parseStorageOption(value string) (catalog.StorageOption, error) {
	switch value {
	case "", "none":
		return catalog.StorageNone, nil
	case string(catalog.StorageR2):
		return catalog.StorageR2, nil
	case string(catalog.StorageS3):
		return catalog.StorageS3, nil
	case string(catalog.StorageSupabase):
		return catalog.StorageSupabase, nil
	default:
		return catalog.StorageNone, fmt.Errorf("unsupported storage option %q", value)
	}
}

func parseEmailOption(value string) (catalog.EmailOption, error) {
	switch value {
	case "", "none":
		return catalog.EmailNone, nil
	case string(catalog.EmailResend):
		return catalog.EmailResend, nil
	default:
		return catalog.EmailNone, fmt.Errorf("unsupported email option %q", value)
	}
}

func applyDerivedDefaults(registry catalog.Registry, input *createInput) {
	if requiresFrontendPack(input.Mode) && input.Frontend == "" {
		input.Frontend = string(catalog.PackIDNextJS)
	}

	if requiresBackendPack(input.Mode) && input.Backend == "" {
		input.Backend = string(catalog.PackIDHonoNode)
	}

	if input.PackageManager == "" && shouldPromptPackageManager(registry, *input) {
		recommended := recommendedPackageManager(registry, *input, allowedPackageManagers(registry, *input))
		if recommended != catalog.PackageManagerNone {
			input.PackageManager = string(recommended)
		}
	}

	applyWorkersLockedDefaults(input)

	packs := selectedPacks(registry, *input)

	if input.Database == "" && input.Mode != string(catalog.ProjectModeFrontend) {
		input.Database = string(recommendedDatabaseOption(*input))
	}

	if input.Auth == "" && input.Mode != string(catalog.ProjectModeFrontend) {
		input.Auth = string(recommendedAuthOption(packs))
	}

	if input.Storage == "" && input.Mode != string(catalog.ProjectModeFrontend) {
		input.Storage = string(recommendedStorageOption(*input))
	}

	if input.Email == "" && input.Mode != string(catalog.ProjectModeFrontend) {
		input.Email = string(recommendedEmailOption(packs))
	}

	if !hasRecommendedSkills(registry, *input) {
		input.RecommendedSkills = false
	}
}

func applySelectionConstraints(registry catalog.Registry, input *createInput) {
	if !requiresFrontendPack(input.Mode) {
		input.Frontend = ""
	}

	if !requiresBackendPack(input.Mode) {
		input.Backend = ""
	}

	if input.Mode == string(catalog.ProjectModeFrontend) {
		input.Auth = ""
		input.Database = ""
		input.Storage = ""
		input.Email = ""
	}

	if !shouldPromptPackageManager(registry, *input) {
		input.PackageManager = ""
	} else if input.PackageManager != "" {
		allowed := allowedPackageManagers(registry, *input)
		if !slices.Contains(allowed, catalog.PackageManager(input.PackageManager)) {
			input.PackageManager = ""
		}
	}

	if shouldClearSelection(authPromptOptions(registry, *input), input.Auth) {
		input.Auth = ""
	}

	if shouldClearSelection(databasePromptOptions(registry, *input), input.Database) {
		input.Database = ""
	}

	if shouldClearSelection(storagePromptOptions(registry, *input), input.Storage) {
		input.Storage = ""
	}

	if shouldClearSelection(emailPromptOptions(registry, *input), input.Email) {
		input.Email = ""
	}

	if !hasRecommendedSkills(registry, *input) {
		input.RecommendedSkills = false
	}
}

func applyWorkersLockedDefaults(input *createInput) {
	if !usesCloudflareWorkers(*input) {
		return
	}

	if input.Database == "" {
		input.Database = string(catalog.DatabaseD1)
	}

	if input.Storage == "" {
		input.Storage = string(catalog.StorageR2)
	}
}

func usesCloudflareWorkers(input createInput) bool {
	return input.Backend == string(catalog.PackIDHonoWorkers)
}

func recommendedAuthOption(packs []catalog.Pack) catalog.AuthOption {
	if hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsBetterAuth }) {
		return catalog.AuthBetter
	}
	if hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsSupabaseAuth }) {
		return catalog.AuthSupabase
	}

	return catalog.AuthNone
}

func recommendedDatabaseOption(input createInput) catalog.DatabaseOption {
	if usesCloudflareWorkers(input) {
		return catalog.DatabaseD1
	}

	return catalog.DatabasePostgres
}

func recommendedStorageOption(input createInput) catalog.StorageOption {
	return catalog.StorageR2
}

func recommendedEmailOption(packs []catalog.Pack) catalog.EmailOption {
	if hasCapability(packs, func(pack catalog.Pack) bool { return pack.Capabilities.SupportsEmail }) {
		return catalog.EmailResend
	}

	return catalog.EmailNone
}

func optionValuesContain(options []huh.Option[string], value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}

	return false
}

func shouldClearSelection(options []huh.Option[string], value string) bool {
	if optionValuesContain(options, value) {
		return false
	}

	return !(value == "none" && optionValuesContain(options, ""))
}
