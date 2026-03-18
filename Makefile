.PHONY: test refresh-all refresh-templates refresh-nextjs refresh-hono-node refresh-hono-workers refresh-skill-assets refresh-skills-nextjs refresh-skills-hono-node refresh-skills-hono-workers refresh-skills-better-auth refresh-skills-supabase refresh-skills-resend

test:
	go test ./...

refresh-all:
	$(MAKE) refresh-templates
	$(MAKE) refresh-skill-assets

refresh-templates:
	./scripts/templates/refresh/all.sh

refresh-nextjs:
	./scripts/templates/refresh/nextjs.sh

refresh-hono-node:
	./scripts/templates/refresh/hono-node.sh

refresh-hono-workers:
	./scripts/templates/refresh/hono-workers.sh

refresh-skill-assets:
	./scripts/skills/install-all.sh

refresh-skills-nextjs:
	./scripts/skills/install-nextjs.sh

refresh-skills-hono-node:
	./scripts/skills/install-hono-node.sh

refresh-skills-hono-workers:
	./scripts/skills/install-hono-workers.sh

refresh-skills-better-auth:
	./scripts/skills/install-better-auth.sh

refresh-skills-supabase:
	./scripts/skills/install-supabase.sh

refresh-skills-resend:
	./scripts/skills/install-resend.sh
