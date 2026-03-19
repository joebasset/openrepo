package cli

import (
	"errors"
	"fmt"
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

type createInput struct {
	ProjectName       string
	Frontend          string
	Backend           string
	PackageManager    string
	Database          string
	ORM               string
	Lint              string
	Tests             string
	Tailwind          string
	AddOns            []string
	GitInit           bool
	Install           bool
	RecommendedSkills bool
}

type optionValue struct {
	Value       string
	Label       string
	Description string
}

type packGroup string

type promptStep int

const (
	packGroupWeb        packGroup = "web"
	packGroupMobile     packGroup = "mobile"
	packGroupTypeScript packGroup = "typescript"
	packGroupPython     packGroup = "python"
	packGroupGo         packGroup = "go"
	packGroupPHP        packGroup = "php"
	backValue           string    = "__back__"

	promptStepFrontendGroup promptStep = iota
	promptStepFrontendPack
	promptStepBackendGroup
	promptStepBackendPack
	promptStepPackageManagerChoice
	promptStepDatabaseChoice
	promptStepORMChoice
	promptStepLintChoice
	promptStepRecommendedSkillsChoice
	promptStepShowAddonsChoice
	promptStepOptionalAddonsChoice
	promptStepGitInitChoice
	promptStepInstallChoice
	promptStepReviewChoice
)

func newCreateInput(options createOptions) (createInput, error) {
	addOns := normalizeAddonIDs(options.addAddons)
	for _, addonID := range addOns {
		if _, _, err := parseAddonSelection(addonID); err != nil {
			return createInput{}, err
		}
	}

	return createInput{
		ProjectName:       strings.TrimSpace(options.projectName),
		Frontend:          strings.TrimSpace(options.fe),
		Backend:           strings.TrimSpace(options.be),
		PackageManager:    strings.TrimSpace(options.packageManager),
		Database:          strings.TrimSpace(options.db),
		ORM:               strings.TrimSpace(options.orm),
		Lint:              strings.TrimSpace(options.lint),
		Tests:             strings.TrimSpace(options.tests),
		Tailwind:          strings.TrimSpace(options.tailwind),
		AddOns:            addOns,
		GitInit:           options.gitInit,
		Install:           options.install,
		RecommendedSkills: options.recommendedSkills,
	}, nil
}

func promptForMissingValues(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, flagState commandFlagState) error {
	if strings.TrimSpace(input.ProjectName) == "" {
		field := huh.NewInput().
			Title("What is the name of your project?").
			Value(&input.ProjectName)
		if err := runCreatePrompt(cmd, "Project", "", field); err != nil {
			return err
		}
		input.ProjectName = strings.TrimSpace(input.ProjectName)
	}

	frontendGroup := recommendedPackGroup(catalog.PackCategoryFrontend)
	backendGroup := recommendedPackGroup(catalog.PackCategoryBackend)
	showAddons := len(input.AddOns) > 0

	for step := promptStepFrontendGroup; step <= promptStepReviewChoice; {
		if !promptStepVisible(step, registry, addonRegistry, *input, showAddons, flagState) {
			step++
			continue
		}

		back, err := runPromptStep(cmd, registry, addonRegistry, input, &frontendGroup, &backendGroup, &showAddons, flagState, step)
		if err != nil {
			return err
		}
		if back {
			for step--; step >= promptStepFrontendGroup; step-- {
				if promptStepVisible(step, registry, addonRegistry, *input, showAddons, flagState) {
					break
				}
			}
			if step < promptStepFrontendGroup {
				step = promptStepFrontendGroup
			}
			continue
		}

		applyAutomaticSelections(registry, addonRegistry, input)

		if step == promptStepReviewChoice {
			return nil
		}

		step++
	}

	return nil
}

func packPromptOptions(registry catalog.Registry, category catalog.PackCategory, group packGroup) []huh.Option[string] {
	options := make([]huh.Option[string], 0)
	recommended := recommendedPackValueForGroup(category, group)

	for _, pack := range registry.All() {
		if !pack.SupportsCategory(category) {
			continue
		}
		if group != "" && packGroupForPack(pack) != group {
			continue
		}

		label := pack.DisplayName
		if string(pack.ID) == recommended {
			label += " (recommended)"
		}

		options = append(options, huh.NewOption(label, string(pack.ID)))
	}

	return options
}

func promptStepVisible(step promptStep, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, showAddons bool, flagState commandFlagState) bool {
	switch step {
	case promptStepFrontendGroup, promptStepFrontendPack:
		return input.Frontend == ""
	case promptStepBackendGroup, promptStepBackendPack:
		return input.Backend == ""
	case promptStepPackageManagerChoice:
		return input.PackageManager == "" && shouldPromptPackageManager(registry, input)
	case promptStepDatabaseChoice:
		return input.Database == "" && shouldPromptSelectionKind(registry, addonRegistry, input, catalog.SelectionKindDatabase)
	case promptStepORMChoice:
		return input.ORM == "" && shouldPromptSelectionKind(registry, addonRegistry, input, catalog.SelectionKindORM)
	case promptStepLintChoice:
		return input.Lint == "" && shouldPromptSelectionKind(registry, addonRegistry, input, catalog.SelectionKindLint)
	case promptStepRecommendedSkillsChoice:
		return !flagState.recommendedSkillsSet && hasRecommendedSkills(registry, addonRegistry, input)
	case promptStepShowAddonsChoice:
		return len(input.AddOns) == 0 && len(optionalAddonOptions(registry, addonRegistry, input)) > 0
	case promptStepOptionalAddonsChoice:
		return showAddons && len(optionalAddonOptions(registry, addonRegistry, input)) > 0
	case promptStepGitInitChoice:
		return !flagState.gitInitSet
	case promptStepInstallChoice:
		return !flagState.installSet
	case promptStepReviewChoice:
		return true
	default:
		return false
	}
}

func runPromptStep(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState, step promptStep) (bool, error) {
	switch step {
	case promptStepFrontendGroup:
		return stepFrontendGroup(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepFrontendPack:
		return stepFrontendPack(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepBackendGroup:
		return stepBackendGroup(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepBackendPack:
		return stepBackendPack(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepPackageManagerChoice:
		return stepPackageManager(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepDatabaseChoice:
		return promptRequiredSelection(cmd, registry, addonRegistry, input, catalog.SelectionKindDatabase)
	case promptStepORMChoice:
		return promptRequiredSelection(cmd, registry, addonRegistry, input, catalog.SelectionKindORM)
	case promptStepLintChoice:
		return promptRequiredSelection(cmd, registry, addonRegistry, input, catalog.SelectionKindLint)
	case promptStepRecommendedSkillsChoice:
		return stepRecommendedSkills(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepShowAddonsChoice:
		return stepShowAddons(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepOptionalAddonsChoice:
		return stepOptionalAddons(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepGitInitChoice:
		return stepGitInit(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepInstallChoice:
		return stepInstall(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	case promptStepReviewChoice:
		return stepReview(cmd, registry, addonRegistry, input, frontendGroup, backendGroup, showAddons, flagState)
	default:
		return false, nil
	}
}

func optionPromptValues(options []optionValue) []huh.Option[string] {
	promptOptions := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		label := option.Label
		if option.Description != "" {
			label += "\n" + option.Description
		}
		promptOptions = append(promptOptions, huh.NewOption(label, option.Value))
	}

	return promptOptions
}

func promptRequiredSelection(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, kind catalog.SelectionKind) (bool, error) {
	values := visibleSelectionValues(registry, addonRegistry, *input, kind)
	options := make([]optionValue, 0, len(values))
	for _, value := range values {
		label := catalog.SelectionValueLabel(kind, value)
		if value == recommendedSelectionValue(registry, addonRegistry, *input, kind) {
			label += " (recommended)"
		}
		options = append(options, optionValue{
			Value:       value,
			Label:       label,
			Description: "",
		})
	}

	selected, back, err := promptSelectValue(
		cmd,
		"Foundations",
		catalog.SelectionDefinitionFor(kind).Label,
		optionPromptValues(options),
		true,
		8,
	)
	if err != nil {
		return false, err
	}
	if back {
		clearForRequiredSelectionBack(registry, addonRegistry, input, kind)
		return true, nil
	}

	input.setSelection(kind, selected)
	return false, nil
}

func promptReview(cmd *cobra.Command, registry catalog.Registry, input *createInput) error {
	spec, selections, err := input.toSpec()
	if err != nil {
		return err
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		return err
	}

	confirmed := true
	field := huh.NewConfirm().
		Title("Create this project?").
		Description(renderCreateSummary(spec, selections, plan, registry)).
		Value(&confirmed)
	if err := runCreatePrompt(cmd, "Review", "", field); err != nil {
		return err
	}
	if !confirmed {
		return errors.New("create cancelled")
	}

	return nil
}

func (input createInput) toSpec() (resolver.ProjectSpec, createSelections, error) {
	packageManager, err := parsePackageManager(input.PackageManager)
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	selections, err := input.selectionSet()
	if err != nil {
		return resolver.ProjectSpec{}, createSelections{}, err
	}

	spec := resolver.ProjectSpec{
		ProjectName:    strings.TrimSpace(input.ProjectName),
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackID(input.Frontend),
		BackendPackID:  catalog.PackID(input.Backend),
		PackageManager: packageManager,
		Selections:     selections,
		AddonIDs:       append([]string(nil), input.AddOns...),
	}

	return spec, createSelections{
		AddOns:                   append([]string(nil), input.AddOns...),
		InitializeGit:            input.GitInit,
		InstallDependencies:      input.Install,
		IncludeRecommendedSkills: input.RecommendedSkills,
	}, nil
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

func selectionDescription(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, kind catalog.SelectionKind, value string) string {
	pack, ok := selectedPackForKind(registry, input, kind)
	if !ok {
		return ""
	}

	selections, err := input.selectionSet()
	if err != nil {
		return ""
	}
	selections.Set(kind, value)

	for _, addon := range addonRegistry.ResolveSelections(pack, catalog.SelectionDefinitionFor(kind).Target, selections) {
		if addon.Kind == kind && addon.Value == value {
			return addon.DisplayName
		}
	}

	return catalog.SelectionValueLabel(kind, value)
}

func optionalAddonOptions(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput) []optionValue {
	options := make([]optionValue, 0)

	for _, kind := range optionalAddonKinds() {
		values := visibleSelectionValues(registry, addonRegistry, input, kind)
		for _, value := range values {
			addonID := addonSelectionID(kind, value)
			options = append(options, optionValue{
				Value:       addonID,
				Label:       catalog.SelectionValueLabel(kind, value),
				Description: "",
			})
		}
	}

	return options
}

func promptSelectValue(cmd *cobra.Command, section string, title string, options []huh.Option[string], allowBack bool, maxHeight int) (string, bool, error) {
	visibleOptions := append([]huh.Option[string]{}, options...)
	if allowBack {
		visibleOptions = append(visibleOptions, huh.NewOption("Go Back", backValue))
	}

	selected := ""
	field := huh.NewSelect[string]().
		Title(title).
		Height(promptListHeight(len(visibleOptions), maxHeight)).
		Value(&selected).
		Options(visibleOptions...)
	if err := runCreatePrompt(cmd, section, "", field); err != nil {
		return "", false, err
	}
	if selected == backValue {
		return "", true, nil
	}
	return selected, false, nil
}

func promptBooleanValue(cmd *cobra.Command, section string, title string, allowBack bool) (bool, bool, error) {
	selected, back, err := promptSelectValue(
		cmd,
		section,
		title,
		[]huh.Option[string]{
			huh.NewOption("Yes", "yes"),
			huh.NewOption("No", "no"),
		},
		allowBack,
		6,
	)
	if err != nil {
		return false, false, err
	}
	if back {
		return false, true, nil
	}
	return selected == "yes", false, nil
}

func stepFrontendGroup(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	selected, _, err := promptSelectValue(cmd, "Frontend", "Frontend type", packGroupPromptOptions(catalog.PackCategoryFrontend), false, 6)
	if err != nil {
		return false, err
	}
	*frontendGroup = selected
	input.Frontend = ""
	return false, nil
}

func stepFrontendPack(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	selected, back, err := promptSelectValue(cmd, "Frontend", "Pick your frontend pack", packPromptOptions(registry, catalog.PackCategoryFrontend, packGroup(*frontendGroup)), true, 8)
	if err != nil {
		return false, err
	}
	if back {
		input.Frontend = ""
		return true, nil
	}
	input.Frontend = selected
	return false, nil
}

func stepBackendGroup(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	selected, back, err := promptSelectValue(cmd, "Backend", "Backend type", packGroupPromptOptions(catalog.PackCategoryBackend), true, 8)
	if err != nil {
		return false, err
	}
	if back {
		input.Frontend = ""
		return true, nil
	}
	*backendGroup = selected
	input.Backend = ""
	return false, nil
}

func stepBackendPack(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	selected, back, err := promptSelectValue(cmd, "Backend", "Pick your backend pack", packPromptOptions(registry, catalog.PackCategoryBackend, packGroup(*backendGroup)), true, 8)
	if err != nil {
		return false, err
	}
	if back {
		input.Backend = ""
		return true, nil
	}
	input.Backend = selected
	return false, nil
}

func stepPackageManager(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	selected, back, err := promptSelectValue(cmd, "Foundations", "Pick your package manager", optionPromptValues(packageManagerOptionLabels(registry, *input)), true, 8)
	if err != nil {
		return false, err
	}
	if back {
		input.Backend = ""
		return true, nil
	}
	input.PackageManager = selected
	return false, nil
}

func stepRecommendedSkills(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	if !hasRecommendedSkills(registry, addonRegistry, *input) {
		input.RecommendedSkills = false
		return false, nil
	}
	value, back, err := promptBooleanValue(cmd, "Skills", "Copy recommended skills?", true)
	if err != nil {
		return false, err
	}
	if back {
		input.Lint = ""
		return true, nil
	}
	input.RecommendedSkills = value
	return false, nil
}

func stepShowAddons(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	value, back, err := promptBooleanValue(cmd, "Addons", "Show optional addons?", true)
	if err != nil {
		return false, err
	}
	if back {
		return true, nil
	}
	*showAddons = value
	if !value {
		input.AddOns = nil
	}
	return false, nil
}

func stepOptionalAddons(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	if !*showAddons {
		return false, nil
	}
	options := optionalAddonOptions(registry, addonRegistry, *input)
	if len(options) == 0 {
		return false, nil
	}

	selected := append([]string(nil), input.AddOns...)
	field := huh.NewMultiSelect[string]().
		Title("Optional addons").
		Filtering(true).
		Height(promptListHeight(len(options), 10)).
		Value(&selected).
		Options(optionPromptValues(options)...)
	if err := runCreatePrompt(cmd, "Addons", "", field); err != nil {
		return false, err
	}
	input.AddOns = normalizeAddonIDs(selected)
	return false, nil
}

func stepGitInit(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	value, back, err := promptBooleanValue(cmd, "Finishing Touches", "Initialize git?", true)
	if err != nil {
		return false, err
	}
	if back {
		return true, nil
	}
	input.GitInit = value
	return false, nil
}

func stepInstall(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	value, back, err := promptBooleanValue(cmd, "Finishing Touches", "Install dependencies?", true)
	if err != nil {
		return false, err
	}
	if back {
		return true, nil
	}
	input.Install = value
	return false, nil
}

func stepReview(cmd *cobra.Command, registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, frontendGroup *string, backendGroup *string, showAddons *bool, flagState commandFlagState) (bool, error) {
	return false, promptReview(cmd, registry, input)
}

func clearForRequiredSelectionBack(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput, kind catalog.SelectionKind) {
	switch kind {
	case catalog.SelectionKindDatabase:
		if shouldPromptPackageManager(registry, *input) {
			input.PackageManager = ""
			return
		}
		input.Backend = ""
	case catalog.SelectionKindORM:
		input.Database = ""
	case catalog.SelectionKindLint:
		input.ORM = ""
	}
}

func runCreatePrompt(cmd *cobra.Command, section string, description string, fields ...huh.Field) error {
	banner := huh.NewNote().
		Title(openrepoBanner()).
		Description(createPromptMeta(section, description))

	allFields := make([]huh.Field, 0, len(fields)+1)
	allFields = append(allFields, banner)
	allFields = append(allFields, fields...)
	return runPromptForm(cmd, allFields...)
}

func openrepoBanner() string {
	return `
 ██████╗ ██████╗ ███████╗███╗   ██╗██████╗ ███████╗██████╗  ██████╗
██╔═══██╗██╔══██╗██╔════╝████╗  ██║██╔══██╗██╔════╝██╔══██╗██╔═══██╗
██║   ██║██████╔╝█████╗  ██╔██╗ ██║██████╔╝█████╗  ██████╔╝██║   ██║
██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║
╚██████╔╝██║     ███████╗██║ ╚████║██║  ██║███████╗██║     ╚██████╔╝
 ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝
`
}

func createPromptMeta(section string, description string) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(section) != "" {
		parts = append(parts, strings.ToUpper(strings.TrimSpace(section)))
	}
	if strings.TrimSpace(description) != "" {
		parts = append(parts, description)
	}
	return strings.Join(parts, "\n")
}

func recommendedPackValue(category catalog.PackCategory) string {
	if category == catalog.PackCategoryBackend {
		return string(catalog.PackIDHonoNode)
	}
	return string(catalog.PackIDNextJS)
}

func recommendedPackGroup(category catalog.PackCategory) string {
	if category == catalog.PackCategoryBackend {
		return string(packGroupTypeScript)
	}
	return string(packGroupWeb)
}

func recommendedPackValueForGroup(category catalog.PackCategory, group packGroup) string {
	switch {
	case category == catalog.PackCategoryFrontend && group == packGroupMobile:
		return string(catalog.PackIDExpo)
	case category == catalog.PackCategoryBackend && group == packGroupPython:
		return string(catalog.PackIDFastAPI)
	case category == catalog.PackCategoryBackend && group == packGroupGo:
		return string(catalog.PackIDGin)
	case category == catalog.PackCategoryBackend && group == packGroupPHP:
		return string(catalog.PackIDLaravel)
	default:
		return recommendedPackValue(category)
	}
}

func packGroupPromptOptions(category catalog.PackCategory) []huh.Option[string] {
	if category == catalog.PackCategoryFrontend {
		return []huh.Option[string]{
			huh.NewOption("Web", string(packGroupWeb)),
			huh.NewOption("Mobile", string(packGroupMobile)),
		}
	}

	return []huh.Option[string]{
		huh.NewOption("TypeScript", string(packGroupTypeScript)),
		huh.NewOption("Python", string(packGroupPython)),
		huh.NewOption("Go", string(packGroupGo)),
		huh.NewOption("PHP", string(packGroupPHP)),
	}
}

func packGroupForPack(pack catalog.Pack) packGroup {
	if pack.Category == catalog.PackCategoryFrontend {
		if pack.Capabilities.Mobile {
			return packGroupMobile
		}
		return packGroupWeb
	}

	switch pack.Language {
	case catalog.LanguagePython:
		return packGroupPython
	case catalog.LanguageGo:
		return packGroupGo
	case catalog.LanguagePHP:
		return packGroupPHP
	default:
		return packGroupTypeScript
	}
}

func promptListHeight(count int, max int) int {
	if count <= 0 {
		return 4
	}
	height := count + 2
	if height < 4 {
		height = 4
	}
	if height > max {
		return max
	}
	return height
}

func applyAutomaticSelections(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input *createInput) {
	for _, kind := range []catalog.SelectionKind{
		catalog.SelectionKindTests,
		catalog.SelectionKindTailwind,
	} {
		if input.selectionValue(kind) != "" {
			continue
		}
		if !shouldPromptSelectionKind(registry, addonRegistry, *input, kind) {
			continue
		}

		input.setSelection(kind, recommendedSelectionValue(registry, addonRegistry, *input, kind))
	}
}

func addonSelectionID(kind catalog.SelectionKind, value string) string {
	switch kind {
	case catalog.SelectionKindAuth:
		return "auth:" + value
	case catalog.SelectionKindStorage:
		return "storage:" + value
	case catalog.SelectionKindEmail:
		return "email:" + value
	case catalog.SelectionKindIcons:
		return "icons:" + value
	case catalog.SelectionKindComponents:
		return "components:" + value
	default:
		return value
	}
}

func hasRecommendedSkills(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput) bool {
	for _, pack := range selectedPacks(registry, input) {
		if pack.SkillAssets != nil {
			return true
		}
	}

	for _, addon := range selectedAddons(registry, addonRegistry, input) {
		if addon.SkillAssets != nil {
			return true
		}
	}

	return false
}

func selectedAddons(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput) []catalog.Addon {
	selections, err := input.selectionSet()
	if err != nil {
		return nil
	}

	addons := make([]catalog.Addon, 0)
	for _, definition := range catalog.AllSelectionDefinitions() {
		pack, ok := selectedPackForKind(registry, input, definition.Kind)
		if !ok {
			continue
		}

		addons = append(addons, addonRegistry.ResolveSelections(pack, definition.Target, selections)...)
	}

	return addons
}
