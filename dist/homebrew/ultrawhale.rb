class Ultrawhale < Formula
  desc "DeepSeek-native coding agent — 32 blocks engine, 7 plugins, AG-UI themes"
  homepage "https://ultrawhale.vaked.dev"
  url "https://github.com/peterlodri-sec/ultrawhale/archive/refs/tags/v8.1.0.tar.gz"
  sha256 "TBD" # Run: shasum -a 256 v8.1.0.tar.gz
  license "Apache-2.0"
  version "8.1.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags=-s -w",
           "-o", bin/"ultrawhale", "./cmd/whale"
    system "go", "build", "-o", bin/"ultrawhale-setup", "./cmd/setup"
    system "go", "build", "-o", bin/"ultrawhale-bench-tui", "./cmd/bench-tui"
  end

  test do
    system "#{bin}/ultrawhale", "--help"
  end
end
