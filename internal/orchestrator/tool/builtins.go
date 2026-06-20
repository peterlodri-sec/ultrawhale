package tool

import "time"

// RegisterBuiltins adds all known orchestrator tools from session history.
func RegisterBuiltins() {
	// Docker tools
	Register(&OrchestratorTool{Name: "docker:ps", Command: "docker ps --format '{{.Names}} {{.Status}}'", Description: "List running containers", Version: "v1.0.0", Timeout: 5 * time.Second, Cacheable: false})
	Register(&OrchestratorTool{Name: "docker:compose-up", Command: "docker compose -f docker/supabase-compose.yml up -d", Description: "Start Supabase containers", Version: "v1.0.0", Timeout: 30 * time.Second, Cacheable: false})
	Register(&OrchestratorTool{Name: "docker:compose-down", Command: "docker compose -f docker/supabase-compose.yml down", Description: "Stop Supabase containers", Version: "v1.0.0", Timeout: 10 * time.Second, Cacheable: false})

	// Nix tools
	Register(&OrchestratorTool{Name: "nix:develop", Command: "nix develop .# --command bash -c 'cd /home/dev/whale && go build -o bin/ultrawhale ./cmd/whale'", Description: "Build ultrawhale in nix shell", Version: "v1.0.0", Timeout: 2 * time.Minute, Cacheable: true})
	Register(&OrchestratorTool{Name: "nix:shell", Command: "nix develop .#", Description: "Enter nix dev shell", Version: "v1.0.0", Timeout: 5 * time.Second, Cacheable: false})

	// Git tools
	Register(&OrchestratorTool{Name: "git:push", Command: "git push origin main", Description: "Push to main", Version: "v1.0.0", Timeout: 30 * time.Second, Cacheable: false})
	Register(&OrchestratorTool{Name: "git:pull", Command: "git pull origin main", Description: "Pull from main", Version: "v1.0.0", Timeout: 30 * time.Second, Cacheable: false})
	Register(&OrchestratorTool{Name: "git:status", Command: "git status --short", Description: "Working tree status", Version: "v1.0.0", Timeout: 5 * time.Second, Cacheable: false})

	// SSH tools
	Register(&OrchestratorTool{Name: "ssh:dev-cx53", Command: "ssh dev-cx53", Description: "SSH to dev-cx53", Version: "v1.0.0", Timeout: 10 * time.Second, Cacheable: false})
	Register(&OrchestratorTool{Name: "ssh:deploy", Command: "scp bin/ultrawhale-linux dev-cx53:~/.local/bin/ultrawhale", Description: "Deploy to dev-cx53", Version: "v1.0.0", Timeout: 30 * time.Second, Cacheable: false})

	// Build tools
	Register(&OrchestratorTool{Name: "build:macos", Command: "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o bin/ultrawhale ./cmd/whale", Description: "Build macOS binary", Version: "v1.0.0", Timeout: 2 * time.Minute, Cacheable: true})
	Register(&OrchestratorTool{Name: "build:linux", Command: "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -trimpath -ldflags='-s -w' -o bin/ultrawhale-linux ./cmd/whale", Description: "Build Linux binary", Version: "v1.0.0", Timeout: 2 * time.Minute, Cacheable: true})
	Register(&OrchestratorTool{Name: "build:shell", Command: "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -trimpath -ldflags='-s -w' -o bin/ultrawhale-shell-linux ./cmd/ultrawhale-shell/", Description: "Build shell daemon", Version: "v1.0.0", Timeout: 1 * time.Minute, Cacheable: true})
	Register(&OrchestratorTool{Name: "build:all", Command: "task build", Description: "Build all binaries via Taskfile", Version: "v1.0.0", Timeout: 5 * time.Minute, Cacheable: false})

	// Bench tools
	Register(&OrchestratorTool{Name: "bench:blocks", Command: "go test -bench=. -benchmem -benchtime=300ms ./internal/blocks/", Description: "Run blocks benchmarks", Version: "v1.0.0", Timeout: 2 * time.Minute, Cacheable: false})
	Register(&OrchestratorTool{Name: "bench:tui", Command: "./bin/ultrawhale-bench-tui", Description: "Run TUI load test", Version: "v1.0.0", Timeout: 1 * time.Minute, Cacheable: false})

	// Deploy tools
	Register(&OrchestratorTool{Name: "deploy:pages", Command: "rm -rf deploy-out && mkdir -p deploy-out/ultrawhale/docs && cp site/index.html deploy-out/ultrawhale/ && cp docs/index.html deploy-out/ultrawhale/docs/ && cp site/robots.txt site/sitemap.xml deploy-out/ultrawhale/ 2>/dev/null && wrangler pages deploy deploy-out --project-name=vaked-dev --branch=main --commit-dirty=true", Description: "Deploy docs site", Version: "v1.0.0", Timeout: 1 * time.Minute, Cacheable: false})
}
