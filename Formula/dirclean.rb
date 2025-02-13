class Dirclean < Formula
  desc "Clean up old files from directories"
  homepage "https://github.com/arkag/dirclean"
  version "1.0.0"  # This will be overridden by the URL

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/dirclean-darwin-arm64.tar.gz"
      sha256 "sha256_for_darwin_arm64_tarball"  # Replace with the actual SHA-256 checksum
    else
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/dirclean-darwin-amd64.tar.gz"
      sha256 "sha256_for_darwin_amd64_tarball"  # Replace with the actual SHA-256 checksum
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/dirclean-linux-arm64.tar.gz"
      sha256 "sha256_for_linux_arm64_tarball"  # Replace with the actual SHA-256 checksum
    else
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/dirclean-linux-amd64.tar.gz"
      sha256 "sha256_for_linux_amd64_tarball"  # Replace with the actual SHA-256 checksum
    end
  end

  def install
    bin.install "dirclean"  # Installs the binary into Homebrew's bin directory
  end

  test do
    system "#{bin}/dirclean", "--version"  # Simple test to verify the installation
  end
end