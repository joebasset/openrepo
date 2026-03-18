package catalog

type ProjectMode string

const (
	ProjectModeFrontend  ProjectMode = "frontend"
	ProjectModeBackend   ProjectMode = "backend"
	ProjectModeFullStack ProjectMode = "fullstack"
)

type PackID string

const (
	PackIDNextJS      PackID = "nextjs"
	PackIDExpo        PackID = "expo"
	PackIDHonoNode    PackID = "hono-node"
	PackIDHonoWorkers PackID = "hono-workers"
	PackIDFastAPI     PackID = "fastapi"
	PackIDGin         PackID = "gin"
)

type PackCategory string

const (
	PackCategoryFrontend PackCategory = "frontend"
	PackCategoryBackend  PackCategory = "backend"
)

type Language string

const (
	LanguageTypeScript Language = "typescript"
	LanguagePython     Language = "python"
	LanguageGo         Language = "go"
)

type Runtime string

const (
	RuntimeNextJS            Runtime = "nextjs"
	RuntimeExpo              Runtime = "expo"
	RuntimeNodeJS            Runtime = "nodejs"
	RuntimeCloudflareWorkers Runtime = "cloudflare-workers"
	RuntimeFastAPI           Runtime = "fastapi"
	RuntimeGin               Runtime = "gin"
)

type PackageManager string

const (
	PackageManagerNone PackageManager = ""
	PackageManagerNPM  PackageManager = "npm"
	PackageManagerPNPM PackageManager = "pnpm"
	PackageManagerBun  PackageManager = "bun"
	PackageManagerYarn PackageManager = "yarn"
)

type WorkspaceStrategy string

const (
	WorkspaceStrategyTurbo  WorkspaceStrategy = "turbo"
	WorkspaceStrategyNative WorkspaceStrategy = "native"
)

type DatabaseOption string

const (
	DatabaseNone     DatabaseOption = ""
	DatabaseD1       DatabaseOption = "d1"
	DatabasePostgres DatabaseOption = "postgres"
	DatabaseSQLite   DatabaseOption = "sqlite"
	DatabaseSupabase DatabaseOption = "supabase"
)

type AuthOption string

const (
	AuthNone     AuthOption = ""
	AuthBetter   AuthOption = "better-auth"
	AuthSupabase AuthOption = "supabase-auth"
)

type StorageOption string

const (
	StorageNone StorageOption = ""
	StorageS3   StorageOption = "s3"
	StorageR2   StorageOption = "r2"
)

type EmailOption string

const (
	EmailNone   EmailOption = ""
	EmailResend EmailOption = "resend"
)

type PackStrategy string

const (
	PackStrategyExternalScaffold PackStrategy = "external_scaffold"
	PackStrategyLocalTemplate    PackStrategy = "local_template"
)

type FileRole string

const (
	FileRoleUpstreamGenerated FileRole = "upstream_generated"
	FileRoleLocalTemplate     FileRole = "local_template"
	FileRoleOverlay           FileRole = "overlay"
)

type ManagedFile struct {
	Path        string
	Role        FileRole
	Description string
	AssetPath   string
}

type EnvVar struct {
	Name        string
	Example     string
	Required    bool
	Description string
}

type Script struct {
	Name    string
	Command string
}

type AgentRule struct {
	Title       string
	Instruction string
}

type SkillRequirement struct {
	Name        string
	InstallHint string
}

type SkillAssetBundle struct {
	Path string
}

type ExternalCommand struct {
	PackageManager PackageManager
	Args           []string
}

type ExternalScaffold struct {
	Tool                      string
	RecommendedPackageManager PackageManager
	Commands                  []ExternalCommand
}

type LocalTemplate struct {
	TemplateRoot string
}

type PackCapabilities struct {
	ProvidesServerRuntime bool
	UsesTypeScript        bool
	SupportsDatabase      bool
	SupportsBetterAuth    bool
	SupportsSupabaseAuth  bool
	SupportsStorage       bool
	SupportsEmail         bool
}

type Pack struct {
	ID             PackID
	DisplayName    string
	Category       PackCategory
	Language       Language
	Runtime        Runtime
	OutputDir      string
	Strategy       PackStrategy
	Description    string
	Files          []ManagedFile
	EnvVars        []EnvVar
	Scripts        []Script
	AgentRules     []AgentRule
	RequiredSkills []SkillRequirement
	SkillAssets    *SkillAssetBundle
	Capabilities   PackCapabilities
	External       *ExternalScaffold
	Local          *LocalTemplate
}

func (p Pack) AllowsPackageManager(manager PackageManager) bool {
	if p.External == nil {
		return manager == PackageManagerNone
	}

	for _, command := range p.External.Commands {
		if command.PackageManager == manager {
			return true
		}
	}

	return false
}
