package catalog

func reactPack() Pack {
	return Pack{
		ID:          PackIDReact,
		DisplayName: "React",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeReact,
		OutputDir:   "apps/web",
		Strategy:    PackStrategyLocalTemplate,
		Description: "React web frontend with Vite, Vitest, and Biome defaults.",
		Files: []ManagedFile{
			{Path: "apps/web/package.json", Role: FileRoleLocalTemplate, AssetPath: "assets/react/package.json.tmpl"},
			{Path: "apps/web/tsconfig.json", Role: FileRoleLocalTemplate, AssetPath: "assets/react/tsconfig.json.tmpl"},
			{Path: "apps/web/tsconfig.node.json", Role: FileRoleLocalTemplate, AssetPath: "assets/react/tsconfig.node.json.tmpl"},
			{Path: "apps/web/vite.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/react/vite.config.ts.tmpl"},
			{Path: "apps/web/vitest.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/react/vitest.config.ts.tmpl"},
			{Path: "apps/web/index.html", Role: FileRoleLocalTemplate, AssetPath: "assets/react/index.html.tmpl"},
			{Path: "apps/web/src/main.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/react/src/main.tsx.tmpl"},
			{Path: "apps/web/src/App.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/react/src/App.tsx.tmpl"},
			{Path: "apps/web/src/styles.css", Role: FileRoleLocalTemplate, AssetPath: "assets/react/src/styles.css.tmpl"},
			{Path: "apps/web/.env.example", Role: FileRoleOverlay},
			{Path: "apps/web/AGENTS.md", Role: FileRoleOverlay},
		},
		EnvVars: []EnvVar{
			{Name: "VITE_API_URL", Example: "http://localhost:3001", Required: false, Description: "Optional API base URL for frontend data calls."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "vite"},
			{Name: "build", Command: "tsc --noEmit && vite build"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Client Boundaries", Instruction: "Keep browser-only code in UI components and isolate environment access behind explicit helpers."},
		},
		Capabilities: PackCapabilities{
			UsesTypeScript: true,
			ReactBased:     true,
		},
		Local: &LocalTemplate{TemplateRoot: "assets/react"},
	}
}

func vuePack() Pack {
	return Pack{
		ID:          PackIDVue,
		DisplayName: "Vue",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeVue,
		OutputDir:   "apps/web",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Vue web frontend with Vite and Vitest defaults.",
		Files: []ManagedFile{
			{Path: "apps/web/package.json", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/package.json.tmpl"},
			{Path: "apps/web/tsconfig.json", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/tsconfig.json.tmpl"},
			{Path: "apps/web/tsconfig.node.json", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/tsconfig.node.json.tmpl"},
			{Path: "apps/web/vite.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/vite.config.ts.tmpl"},
			{Path: "apps/web/vitest.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/vitest.config.ts.tmpl"},
			{Path: "apps/web/index.html", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/index.html.tmpl"},
			{Path: "apps/web/src/main.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/src/main.ts.tmpl"},
			{Path: "apps/web/src/App.vue", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/src/App.vue.tmpl"},
			{Path: "apps/web/src/styles.css", Role: FileRoleLocalTemplate, AssetPath: "assets/vue/src/styles.css.tmpl"},
			{Path: "apps/web/.env.example", Role: FileRoleOverlay},
			{Path: "apps/web/AGENTS.md", Role: FileRoleOverlay},
		},
		EnvVars: []EnvVar{
			{Name: "VITE_API_URL", Example: "http://localhost:3001", Required: false, Description: "Optional API base URL for frontend data calls."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "vite"},
			{Name: "build", Command: "vue-tsc --noEmit && vite build"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Single File Components", Instruction: "Keep component concerns localized and prefer small composables over broad global state."},
		},
		Capabilities: PackCapabilities{
			UsesTypeScript: true,
		},
		Local: &LocalTemplate{TemplateRoot: "assets/vue"},
	}
}

func ionicReactPack() Pack {
	return Pack{
		ID:          PackIDIonicReact,
		DisplayName: "Ionic React",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeIonicReact,
		OutputDir:   "apps/mobile",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Ionic React mobile frontend with Vite defaults.",
		Files: []ManagedFile{
			{Path: "apps/mobile/package.json", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/package.json.tmpl"},
			{Path: "apps/mobile/tsconfig.json", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/tsconfig.json.tmpl"},
			{Path: "apps/mobile/tsconfig.node.json", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/tsconfig.node.json.tmpl"},
			{Path: "apps/mobile/vite.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/vite.config.ts.tmpl"},
			{Path: "apps/mobile/index.html", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/index.html.tmpl"},
			{Path: "apps/mobile/src/main.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/src/main.tsx.tmpl"},
			{Path: "apps/mobile/src/App.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/src/App.tsx.tmpl"},
			{Path: "apps/mobile/src/styles.css", Role: FileRoleLocalTemplate, AssetPath: "assets/ionic-react/src/styles.css.tmpl"},
			{Path: "apps/mobile/.env.example", Role: FileRoleOverlay},
			{Path: "apps/mobile/AGENTS.md", Role: FileRoleOverlay},
		},
		EnvVars: []EnvVar{
			{Name: "VITE_API_URL", Example: "http://localhost:3001", Required: false, Description: "Optional API base URL for app data calls."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "vite"},
			{Name: "build", Command: "tsc --noEmit && vite build"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Ionic First", Instruction: "Use Ionic components and navigation primitives before introducing custom platform shells."},
		},
		Capabilities: PackCapabilities{
			UsesTypeScript: true,
			ReactBased:     true,
			Mobile:         true,
		},
		Local: &LocalTemplate{TemplateRoot: "assets/ionic-react"},
	}
}

func tanStackStartPack() Pack {
	return Pack{
		ID:          PackIDTanStack,
		DisplayName: "TanStack Start",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeTanStackStart,
		OutputDir:   "apps/web",
		Strategy:    PackStrategyLocalTemplate,
		Description: "TanStack Start fullstack-capable React app.",
		Files: []ManagedFile{
			{Path: "apps/web/package.json", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/package.json.tmpl"},
			{Path: "apps/web/tsconfig.json", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/tsconfig.json.tmpl"},
			{Path: "apps/web/tsconfig.node.json", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/tsconfig.node.json.tmpl"},
			{Path: "apps/web/vite.config.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/vite.config.ts.tmpl"},
			{Path: "apps/web/index.html", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/index.html.tmpl"},
			{Path: "apps/web/src/entry-client.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/src/entry-client.tsx.tmpl"},
			{Path: "apps/web/src/entry-server.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/src/entry-server.tsx.tmpl"},
			{Path: "apps/web/src/routes/__root.tsx", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/src/routes/__root.tsx.tmpl"},
			{Path: "apps/web/src/styles.css", Role: FileRoleLocalTemplate, AssetPath: "assets/tanstack-start/src/styles.css.tmpl"},
			{Path: "apps/web/.env.example", Role: FileRoleOverlay},
			{Path: "apps/web/AGENTS.md", Role: FileRoleOverlay},
		},
		EnvVars: []EnvVar{
			{Name: "APP_URL", Example: "http://localhost:3000", Required: true, Description: "Base URL for the app."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "vite"},
			{Name: "build", Command: "tsc --noEmit && vite build"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Route-Driven Structure", Instruction: "Keep route modules focused and move shared behavior into reusable hooks or services."},
		},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			UsesTypeScript:        true,
			SupportsDatabase:      true,
			SupportsBetterAuth:    true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
			ReactBased:            true,
			SupportsBackendMode:   true,
		},
		Local: &LocalTemplate{TemplateRoot: "assets/tanstack-start"},
	}
}

func laravelPack() Pack {
	return Pack{
		ID:          PackIDLaravel,
		DisplayName: "Laravel",
		Category:    PackCategoryBackend,
		Language:    LanguagePHP,
		Runtime:     RuntimeLaravel,
		OutputDir:   "apps/api",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Laravel backend scaffold with testing and API route defaults.",
		Files: []ManagedFile{
			{Path: "apps/api/composer.json", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/composer.json.tmpl"},
			{Path: "apps/api/artisan", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/artisan.tmpl"},
			{Path: "apps/api/bootstrap/app.php", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/bootstrap/app.php.tmpl"},
			{Path: "apps/api/public/index.php", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/public/index.php.tmpl"},
			{Path: "apps/api/routes/web.php", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/routes/web.php.tmpl"},
			{Path: "apps/api/phpunit.xml", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/phpunit.xml.tmpl"},
			{Path: "apps/api/tests/Feature/HealthTest.php", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/tests/Feature/HealthTest.php.tmpl"},
			{Path: "apps/api/.env.example", Role: FileRoleOverlay},
			{Path: "apps/api/AGENTS.md", Role: FileRoleOverlay},
		},
		EnvVars: []EnvVar{
			{Name: "APP_ENV", Example: "local", Required: true, Description: "Application environment."},
			{Name: "APP_URL", Example: "http://localhost:8000", Required: true, Description: "Base URL for the Laravel app."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "php artisan serve --host=127.0.0.1 --port=8000"},
			{Name: "test", Command: "php artisan test"},
		},
		AgentRules: []AgentRule{
			{Title: "Laravel Conventions", Instruction: "Prefer framework conventions and keep HTTP concerns inside controllers and form requests."},
		},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			SupportsDatabase:      true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		Local: &LocalTemplate{TemplateRoot: "assets/laravel"},
	}
}
