class Sshtie < Formula
  desc "SSH + mosh + tmux profiles — one command to connect"
  homepage "https://github.com/ainsuotain/sshtie"
  url "https://github.com/ainsuotain/sshtie/archive/refs/tags/v0.2.1.tar.gz"
  sha256 "ecf50f0afb1ee359c60d15932c9bfd56009cb83057154a0f98342d341fec040c"
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
