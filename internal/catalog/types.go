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
	PackIDReact       PackID = "react"
	PackIDVue         PackID = "vue"
	PackIDExpo        PackID = "expo"
	PackIDIonicReact  PackID = "ionic-react"
	PackIDTanStack    PackID = "tanstack-start"
	PackIDHonoNode    PackID = "hono-node"
	PackIDExpress     PackID = "express"
	PackIDHonoWorkers PackID = "hono-workers"
	PackIDFastAPI     PackID = "fastapi"
	PackIDGin         PackID = "gin"
	PackIDLaravel     PackID = "laravel"
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
	LanguagePHP        Language = "php"
)

type Runtime string

const (
	RuntimeNextJS            Runtime = "nextjs"
	RuntimeReact             Runtime = "react"
	RuntimeVue               Runtime = "vue"
	RuntimeExpo              Runtime = "expo"
	RuntimeIonicReact        Runtime = "ionic-react"
	RuntimeTanStackStart     Runtime = "tanstack-start"
	RuntimeNodeJS            Runtime = "nodejs"
	RuntimeExpress           Runtime = "express"
	RuntimeCloudflareWorkers Runtime = "cloudflare-workers"
	RuntimeFastAPI           Runtime = "fastapi"
	RuntimeGin               Runtime = "gin"
	RuntimeLaravel           Runtime = "laravel"
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
	DatabaseMySQL    DatabaseOption = "mysql"
	DatabaseSQLite   DatabaseOption = "sqlite"
	DatabaseSupabase DatabaseOption = "supabase"
	DatabaseMongoDB  DatabaseOption = "mongodb"
	DatabaseFirebase DatabaseOption = "firebase"
)

type AuthOption string

const (
	AuthNone     AuthOption = ""
	AuthBetter   AuthOption = "better-auth"
	AuthSupabase AuthOption = "supabase-auth"
	AuthFirebase AuthOption = "firebase-auth"
	AuthSanctum  AuthOption = "sanctum"
	AuthPassport AuthOption = "passport"
)

type StorageOption string

const (
	StorageNone     StorageOption = ""
	StorageS3       StorageOption = "s3"
	StorageR2       StorageOption = "r2"
	StorageSupabase StorageOption = "supabase-storage"
	StorageFirebase StorageOption = "firebase-storage"
)

type EmailOption string

const (
	EmailNone   EmailOption = ""
	EmailResend EmailOption = "resend"
)

type ORMOption string

const (
	ORMNone       ORMOption = ""
	ORMDrizzle    ORMOption = "drizzle"
	ORMPrisma     ORMOption = "prisma"
	ORMSQLAlchemy ORMOption = "sqlalchemy"
	ORMGORM       ORMOption = "gorm"
	ORMEloquent   ORMOption = "eloquent"
)

type LintOption string

const (
	LintNone  LintOption = ""
	LintBiome LintOption = "biome"
	LintRuff  LintOption = "ruff"
	LintGoFmt LintOption = "gofmt"
	LintPint  LintOption = "pint"
)

type TestsOption string

const (
	TestsNone    TestsOption = ""
	TestsVitest  TestsOption = "vitest"
	TestsPytest  TestsOption = "pytest"
	TestsGoTest  TestsOption = "go-test"
	TestsPHPUnit TestsOption = "phpunit"
)

type TailwindOption string

const (
	TailwindNone TailwindOption = ""
	TailwindCSS  TailwindOption = "tailwindcss"
)

type IconsOption string

const (
	IconsNone        IconsOption = ""
	IconsLucideReact IconsOption = "lucide-react"
	IconsReactIcons  IconsOption = "react-icons"
)

type ComponentsOption string

const (
	ComponentsNone   ComponentsOption = ""
	ComponentsShadcn ComponentsOption = "shadcn"
	ComponentsMUI    ComponentsOption = "mui"
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
	ReactBased            bool
	Mobile                bool
	WorkersRuntime        bool
	SupportsTailwind      bool
	SupportsBackendMode   bool // frontend pack that can also serve as backend (e.g. Next.js API routes)
}

type PackTrait string

const (
	PackTraitServerRuntime PackTrait = "server-runtime"
	PackTraitTypeScript    PackTrait = "typescript"
	PackTraitReact         PackTrait = "react"
	PackTraitMobile        PackTrait = "mobile"
	PackTraitWorkers       PackTrait = "workers"
	PackTraitBackendMode   PackTrait = "backend-mode"
)

type Pack struct {
	ID           PackID
	DisplayName  string
	Category     PackCategory
	Language     Language
	Runtime      Runtime
	OutputDir    string
	Strategy     PackStrategy
	Description  string
	Files        []ManagedFile
	EnvVars      []EnvVar
	Scripts      []Script
	AgentRules   []AgentRule
	SkillAssets  *SkillAssetBundle
	Capabilities PackCapabilities
	External     *ExternalScaffold
	Local        *LocalTemplate
}

type IntegrationKind = SelectionKind

const (
	IntegrationAuth       = SelectionKindAuth
	IntegrationDatabase   = SelectionKindDatabase
	IntegrationStorage    = SelectionKindStorage
	IntegrationEmail      = SelectionKindEmail
	IntegrationIcons      = SelectionKindIcons
	IntegrationComponents = SelectionKindComponents
)

type AddonID string

func NewAddonID(kind SelectionKind, value string, packID PackID) AddonID {
	return AddonID(string(kind) + ":" + value + ":" + string(packID))
}

type AddonWhen struct {
	RequiredSelections  map[SelectionKind][]string
	ForbiddenSelections map[SelectionKind][]string
	RequiredPackTraits  []PackTrait
	ForbiddenPackTraits []PackTrait
}

type Addon struct {
	ID               AddonID
	Kind             SelectionKind
	Value            string
	Target           SelectionTarget
	Integration      SelectionKind
	IntegrationValue string
	PackID           PackID
	DisplayName      string
	When             AddonWhen
	Files            []ManagedFile
	Dependencies     map[string]string
	DevDependencies  map[string]string
	EnvVars          []EnvVar
	AgentRules       []AgentRule
	Scripts          map[string]string
	SkillAssets      *SkillAssetBundle
}

func (p Pack) SupportsCategory(cat PackCategory) bool {
	if p.Category == cat {
		return true
	}

	if cat == PackCategoryBackend && p.Capabilities.SupportsBackendMode {
		return true
	}

	return false
}

func (p Pack) HasTrait(trait PackTrait) bool {
	switch trait {
	case PackTraitServerRuntime:
		return p.Capabilities.ProvidesServerRuntime
	case PackTraitTypeScript:
		return p.Capabilities.UsesTypeScript
	case PackTraitReact:
		return p.Capabilities.ReactBased
	case PackTraitMobile:
		return p.Capabilities.Mobile
	case PackTraitWorkers:
		return p.Capabilities.WorkersRuntime
	case PackTraitBackendMode:
		return p.Capabilities.SupportsBackendMode
	default:
		return false
	}
}

func (p Pack) AllowsPackageManager(manager PackageManager) bool {
	if p.External == nil {
		if p.Language == LanguageTypeScript {
			switch manager {
			case PackageManagerNPM, PackageManagerPNPM, PackageManagerBun, PackageManagerYarn:
				return true
			default:
				return false
			}
		}

		return manager == PackageManagerNone
	}

	for _, command := range p.External.Commands {
		if command.PackageManager == manager {
			return true
		}
	}

	return false
}
