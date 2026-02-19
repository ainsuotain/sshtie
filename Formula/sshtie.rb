class Sshtie < Formula
  desc "SSH + mosh + tmux profiles — one command to connect"
  homepage "https://github.com/ainsuotain/sshtie"
  url "https://github.com/ainsuotain/sshtie/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "958afd5d8e121d0d430dfdef6d521b6d3790ff020c2630e7d1951c960891a3ed"
  license "Apache-2.0"
  head "https://github.com/ainsuotain/sshtie.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "."
  end

  test do
    assert_match "sshtie", shell_output("#{bin}/sshtie --help")
  end
end
