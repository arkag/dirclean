class Dirclean < Formula
  desc "Clean up old files from directories"
  homepage "https://github.com/arkag/dirclean"
  
  # Fetch latest release version
  def self.latest_version
    uri = URI("https://api.github.com/repos/arkag/dirclean/releases/latest")
    response = Net::HTTP.get(uri)
    JSON.parse(response)["tag_name"].sub(/^v/, "")
  rescue
    "1.0.0" # Fallback version if API call fails
  end

  version latest_version

  # Define all possible binary combinations
  def self.binary_info
    {
      darwin_arm64: "dirclean-darwin-arm64.tar.gz",
      darwin_amd64: "dirclean-darwin-amd64.tar.gz",
      linux_arm64: "dirclean-linux-arm64.tar.gz",
      linux_amd64: "dirclean-linux-amd64.tar.gz"
    }
  end

  # Fetch SHA256 for a specific binary
  def self.fetch_checksum(version, binary)
    uri = URI("https://github.com/arkag/dirclean/releases/download/v#{version}/checksums.txt")
    response = Net::HTTP.get(uri)
    response.lines.each do |line|
      checksum, file = line.split
      return checksum if file.end_with?(binary)
    end
    raise "Checksum not found for #{binary}"
  rescue
    "0" * 64 # Return dummy SHA if fetch fails
  end

  on_macos do
    if Hardware::CPU.arm?
      binary = binary_info[:darwin_arm64]
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/#{binary}"
      sha256 fetch_checksum(version, binary)
    else
      binary = binary_info[:darwin_amd64]
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/#{binary}"
      sha256 fetch_checksum(version, binary)
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      binary = binary_info[:linux_arm64]
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/#{binary}"
      sha256 fetch_checksum(version, binary)
    else
      binary = binary_info[:linux_amd64]
      url "https://github.com/arkag/dirclean/releases/download/v#{version}/#{binary}"
      sha256 fetch_checksum(version, binary)
    end
  end

  def install
    bin.install "dirclean"
  end

  test do
    system "#{bin}/dirclean", "--version"
  end
end
