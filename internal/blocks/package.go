package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ── Package Primitive — Distribution Artifacts ────────────────────────
// Vaked layer: Materializes. Generates brew formula, Dockerfile, npm package.

// PackageType is the distribution format.
type PackageType string

const (
	PkgBrew   PackageType = "brew"
	PkgDocker PackageType = "docker"
	PkgNPM    PackageType = "npm"
	PkgBinary PackageType = "binary"
)

// PackageArtifact is a built distribution package.
type PackageArtifact struct {
	Type    PackageType
	Path    string
	Size    int64
	Ref     string
	BuiltAt string
}

// BuildPackage generates a distribution package.
func BuildPackage(pkgType PackageType, version, outputDir string) (*PackageArtifact, error) {
	os.MkdirAll(outputDir, 0o755)
	pov := CurrentPOV()

	switch pkgType {
	case PkgBrew:
		return buildBrew(version, outputDir, pov)
	case PkgDocker:
		return buildDocker(version, outputDir, pov)
	case PkgBinary:
		return buildBinary(version, outputDir, pov)
	default:
		return nil, fmt.Errorf("package: unsupported type %s", pkgType)
	}
}

func buildBrew(version, dir string, pov POV) (*PackageArtifact, error) {
	formula := filepath.Join(dir, "ultrawhale.rb")
	content := fmt.Sprintf(`class Ultrawhale < Formula
  desc "DeepSeek-native coding agent — %d blocks, 7 plugins"
  homepage "https://ultrawhale.vaked.dev"
  url "https://github.com/peterlodri-sec/ultrawhale/archive/refs/tags/v%s.tar.gz"
  version "%s"
  license "Apache-2.0"
  depends_on "go" => :build
  def install
    system "go", "build", "-o", bin/"ultrawhale", "./cmd/whale"
  end
end
`, 48, version, version)
	
	if err := os.WriteFile(formula, []byte(content), 0o644); err != nil {
		return nil, err
	}
	
	fi, _ := os.Stat(formula)
	Log(LogInfo, "package.brew", formula, "", "", 0, nil)
	_ = pov
	return &PackageArtifact{Type: PkgBrew, Path: formula, Size: fi.Size(), Ref: Ref([]byte(content))}, nil
}

func buildDocker(version, dir string, pov POV) (*PackageArtifact, error) {
	dockerfile := filepath.Join(dir, "Dockerfile")
	content := fmt.Sprintf(`FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o ultrawhale ./cmd/whale
FROM alpine:3.20
COPY --from=builder /app/ultrawhale /usr/local/bin/ultrawhale
ENTRYPOINT ["ultrawhale"]
`)
	os.WriteFile(dockerfile, []byte(content), 0o644)
	fi, _ := os.Stat(dockerfile)
	Log(LogInfo, "package.docker", dockerfile, "", "", 0, nil)
	_ = pov
	return &PackageArtifact{Type: PkgDocker, Path: dockerfile, Size: fi.Size(), Ref: Ref([]byte(content))}, nil
}

func buildBinary(version, dir string, pov POV) (*PackageArtifact, error) {
	binary := filepath.Join(dir, "ultrawhale")
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", binary, "./cmd/whale")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("build: %s", string(out))
	}
	fi, _ := os.Stat(binary)
	Log(LogInfo, "package.binary", binary, "", "", 0, nil)
	_ = pov
	return &PackageArtifact{Type: PkgBinary, Path: binary, Size: fi.Size(), Ref: Ref(nil)}, nil
}
