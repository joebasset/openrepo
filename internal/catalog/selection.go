package catalog

import "sort"

type SelectionKind string

const (
	SelectionKindDatabase   SelectionKind = "database"
	SelectionKindORM        SelectionKind = "orm"
	SelectionKindLint       SelectionKind = "lint"
	SelectionKindTests      SelectionKind = "tests"
	SelectionKindTailwind   SelectionKind = "tailwind"
	SelectionKindAuth       SelectionKind = "auth"
	SelectionKindStorage    SelectionKind = "storage"
	SelectionKindEmail      SelectionKind = "email"
	SelectionKindIcons      SelectionKind = "icons"
	SelectionKindComponents SelectionKind = "components"
)

type SelectionTarget string

const (
	SelectionTargetFrontend SelectionTarget = "frontend"
	SelectionTargetBackend  SelectionTarget = "backend"
)

type SelectionSet map[SelectionKind]string

func NewSelectionSet() SelectionSet {
	return make(SelectionSet)
}

func (s SelectionSet) Clone() SelectionSet {
	cloned := make(SelectionSet, len(s))
	for kind, value := range s {
		cloned[kind] = value
	}

	return cloned
}

func (s SelectionSet) Get(kind SelectionKind) string {
	if s == nil {
		return ""
	}

	return s[kind]
}

func (s SelectionSet) Set(kind SelectionKind, value string) {
	if s == nil {
		return
	}

	if value == "" {
		delete(s, kind)
		return
	}

	s[kind] = value
}

func (s SelectionSet) Kinds() []SelectionKind {
	kinds := make([]SelectionKind, 0, len(s))
	for kind, value := range s {
		if value == "" {
			continue
		}
		kinds = append(kinds, kind)
	}

	sort.Slice(kinds, func(i, j int) bool {
		return SelectionDefinitionFor(kinds[i]).PromptOrder < SelectionDefinitionFor(kinds[j]).PromptOrder
	})

	return kinds
}

type SelectionDefinition struct {
	Kind         SelectionKind
	Label        string
	ReviewLabel  string
	FlagName     string
	Target       SelectionTarget
	PromptOrder  int
	AllowsNone   bool
	DefaultValue string
}

var selectionDefinitions = map[SelectionKind]SelectionDefinition{
	SelectionKindDatabase: {
		Kind:         SelectionKindDatabase,
		Label:        "Database",
		ReviewLabel:  "database",
		FlagName:     "db",
		Target:       SelectionTargetBackend,
		PromptOrder:  10,
		AllowsNone:   false,
		DefaultValue: string(DatabasePostgres),
	},
	SelectionKindORM: {
		Kind:         SelectionKindORM,
		Label:        "ORM",
		ReviewLabel:  "orm",
		FlagName:     "orm",
		Target:       SelectionTargetBackend,
		PromptOrder:  20,
		AllowsNone:   false,
		DefaultValue: string(ORMDrizzle),
	},
	SelectionKindLint: {
		Kind:         SelectionKindLint,
		Label:        "Lint / Format",
		ReviewLabel:  "lint",
		FlagName:     "lint",
		Target:       SelectionTargetBackend,
		PromptOrder:  30,
		AllowsNone:   false,
		DefaultValue: string(LintBiome),
	},
	SelectionKindTests: {
		Kind:         SelectionKindTests,
		Label:        "Tests",
		ReviewLabel:  "tests",
		FlagName:     "tests",
		Target:       SelectionTargetBackend,
		PromptOrder:  40,
		AllowsNone:   false,
		DefaultValue: string(TestsVitest),
	},
	SelectionKindTailwind: {
		Kind:         SelectionKindTailwind,
		Label:        "Tailwind",
		ReviewLabel:  "tailwind",
		FlagName:     "tailwind",
		Target:       SelectionTargetFrontend,
		PromptOrder:  50,
		AllowsNone:   false,
		DefaultValue: string(TailwindCSS),
	},
	SelectionKindAuth: {
		Kind:         SelectionKindAuth,
		Label:        "Authentication",
		ReviewLabel:  "auth",
		FlagName:     "auth",
		Target:       SelectionTargetBackend,
		PromptOrder:  60,
		AllowsNone:   true,
		DefaultValue: string(AuthBetter),
	},
	SelectionKindStorage: {
		Kind:         SelectionKindStorage,
		Label:        "Storage",
		ReviewLabel:  "storage",
		FlagName:     "storage",
		Target:       SelectionTargetBackend,
		PromptOrder:  70,
		AllowsNone:   true,
		DefaultValue: string(StorageR2),
	},
	SelectionKindEmail: {
		Kind:         SelectionKindEmail,
		Label:        "Email",
		ReviewLabel:  "email",
		FlagName:     "email",
		Target:       SelectionTargetBackend,
		PromptOrder:  80,
		AllowsNone:   true,
		DefaultValue: string(EmailResend),
	},
	SelectionKindIcons: {
		Kind:         SelectionKindIcons,
		Label:        "Icons",
		ReviewLabel:  "icons",
		FlagName:     "icons",
		Target:       SelectionTargetFrontend,
		PromptOrder:  90,
		AllowsNone:   true,
		DefaultValue: string(IconsLucideReact),
	},
	SelectionKindComponents: {
		Kind:         SelectionKindComponents,
		Label:        "Components",
		ReviewLabel:  "components",
		FlagName:     "components",
		Target:       SelectionTargetFrontend,
		PromptOrder:  100,
		AllowsNone:   true,
		DefaultValue: string(ComponentsShadcn),
	},
}

func AllSelectionDefinitions() []SelectionDefinition {
	definitions := make([]SelectionDefinition, 0, len(selectionDefinitions))
	for _, definition := range selectionDefinitions {
		definitions = append(definitions, definition)
	}

	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].PromptOrder < definitions[j].PromptOrder
	})

	return definitions
}

func SelectionDefinitionFor(kind SelectionKind) SelectionDefinition {
	return selectionDefinitions[kind]
}

func SelectionKindsForTarget(target SelectionTarget) []SelectionKind {
	kinds := make([]SelectionKind, 0)
	for _, definition := range AllSelectionDefinitions() {
		if definition.Target == target {
			kinds = append(kinds, definition.Kind)
		}
	}

	return kinds
}

func SelectionValueLabel(kind SelectionKind, value string) string {
	if value == "" {
		return "None"
	}

	switch kind {
	case SelectionKindDatabase:
		switch DatabaseOption(value) {
		case DatabasePostgres:
			return "Postgres"
		case DatabaseMySQL:
			return "MySQL"
		case DatabaseSQLite:
			return "SQLite"
		case DatabaseSupabase:
			return "Supabase"
		case DatabaseMongoDB:
			return "MongoDB"
		case DatabaseFirebase:
			return "Firebase Firestore"
		case DatabaseD1:
			return "Cloudflare D1"
		}
	case SelectionKindORM:
		switch ORMOption(value) {
		case ORMDrizzle:
			return "Drizzle"
		case ORMPrisma:
			return "Prisma"
		case ORMSQLAlchemy:
			return "SQLAlchemy"
		case ORMGORM:
			return "GORM"
		case ORMEloquent:
			return "Eloquent"
		}
	case SelectionKindLint:
		switch LintOption(value) {
		case LintBiome:
			return "Biome"
		case LintRuff:
			return "Ruff"
		case LintGoFmt:
			return "gofmt"
		case LintPint:
			return "Laravel Pint"
		}
	case SelectionKindTests:
		switch TestsOption(value) {
		case TestsVitest:
			return "Vitest"
		case TestsPytest:
			return "Pytest"
		case TestsGoTest:
			return "go test"
		case TestsPHPUnit:
			return "PHPUnit"
		}
	case SelectionKindTailwind:
		if TailwindOption(value) == TailwindCSS {
			return "Tailwind CSS"
		}
	case SelectionKindAuth:
		switch AuthOption(value) {
		case AuthBetter:
			return "Better Auth"
		case AuthSupabase:
			return "Supabase Auth"
		case AuthFirebase:
			return "Firebase Auth"
		case AuthSanctum:
			return "Laravel Sanctum"
		case AuthPassport:
			return "Laravel Passport"
		}
	case SelectionKindStorage:
		switch StorageOption(value) {
		case StorageS3:
			return "Amazon S3"
		case StorageR2:
			return "Cloudflare R2"
		case StorageSupabase:
			return "Supabase Storage"
		case StorageFirebase:
			return "Firebase Storage"
		}
	case SelectionKindEmail:
		if EmailOption(value) == EmailResend {
			return "Resend"
		}
	case SelectionKindIcons:
		switch IconsOption(value) {
		case IconsLucideReact:
			return "Lucide React"
		case IconsReactIcons:
			return "React Icons"
		}
	case SelectionKindComponents:
		switch ComponentsOption(value) {
		case ComponentsShadcn:
			return "shadcn/ui"
		case ComponentsMUI:
			return "Material UI"
		}
	}

	return value
}
