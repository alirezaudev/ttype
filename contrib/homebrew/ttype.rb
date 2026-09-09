# This file written by AI :)
#
# Homebrew formula for ttype. The version and sha256 values are rewritten by
# the release workflow; do not edit them by hand.
class Ttype < Formula
  desc "Terminal-first typing practice"
  homepage "https://github.com/alirezaudev/ttype"
  version "0.0.0"

  on_macos do
    on_arm do
      url "https://github.com/alirezaudev/ttype/releases/download/v#{version}/ttype_#{version}_darwin_arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/alirezaudev/ttype/releases/download/v#{version}/ttype_#{version}_darwin_amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/alirezaudev/ttype/releases/download/v#{version}/ttype_#{version}_linux_arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/alirezaudev/ttype/releases/download/v#{version}/ttype_#{version}_linux_amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  def install
    bin.install "ttype"
    man1.install "ttype.1" if File.exist?("ttype.1")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ttype --version")
  end
end
