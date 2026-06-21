class Ultrawhale < Formula
  desc "DeepSeek-native coding agent — 97-block engine, 6 recursions, Council of LLMs"
  homepage "https://vaked.dev/ultrawhale"
  url "https://github.com/peterlodri-sec/ultrawhale/archive/refs/tags/v64.0.0.tar.gz"
  sha256 "TBD" # Run: curl -L https://github.com/peterlodri-sec/ultrawhale/archive/refs/tags/v64.0.0.tar.gz | shasum -a 256
  license "Apache-2.0"
  version "64.0.0"

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
