class Ultrawhale < Formula
  desc "Coding agent that builds itself — 150 blocks, 7 recursions, 8 engines, 14 protocols"
  homepage "https://vaked.dev/ultrawhale"
  url "https://github.com/peterlodri-sec/ultrawhale/archive/refs/tags/v100.1.0.tar.gz"
  sha256 "3a9140201ff7e0144ce512dbe3e9d796dcde9917c7eeb50308502dd765e8a674"
  license "Apache-2.0"
  version "100.1.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags=-s -w",
           "-o", bin/"ultrawhale", "./cmd/whale"
  end

  test do
    system "#{bin}/ultrawhale", "--help"
  end
end
