class Sshtie < Formula
  desc "SSH + mosh + tmux profiles — one command to connect"
  homepage "https://github.com/ainsuotain/sshtie"
  url "https://github.com/ainsuotain/sshtie/archive/refs/tags/v0.1.3.tar.gz"
  sha256 "4b70aea0f2f65af8e3b4d06f1754d7cceafba2beb89c26bc6c32d0988f30eb74"
  license "Apache-2.0"
  head "https://github.com/ainsuotain/sshtie.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "."
  end

  test do
    assert_match "sshtie", shell_output("#{bin}/sshtie --help")
  end
end
