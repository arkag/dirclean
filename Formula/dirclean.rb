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

  # Define all possible binary combinations and config paths
  def self.config_path
    if OS.mac?
      "/usr/local/share/dirclean"
    else
      "/usr/share/dirclean"
    end
  end

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
    
    # Create and install example config
    config_dir = etc/"dirclean"
    config_dir.mkpath
    
    # Create example config content
    example_config = <<~EOS
      defaults:
        delete_older_than_days: 30
        mode: dry-run
        log_level: INFO
        log_file: dirclean.log
      
      rules:
        - paths:
            - ~/Downloads
            - ~/Documents/temp
          delete_older_than_days: 7
          min_file_size: 1MB
          mode: dry-run
      
        - paths:
            - /tmp/*
            - /var/tmp/*
          delete_older_than_days: 1
          max_file_size: 100MB
          mode: dry-run
    EOS
    
    # Write example config
    (config_dir/"example.config.yaml").write(example_config)
    
    # Create share directory and symlink
    share_dir = "#{HOMEBREW_PREFIX}/share/dirclean"
    system "mkdir", "-p", share_dir unless Dir.exist?(share_dir)
    system "ln", "-sf", "#{config_dir}/example.config.yaml", "#{share_dir}/example.config.yaml"
  end

  test do
    system "#{bin}/dirclean", "--version"
  end
end
