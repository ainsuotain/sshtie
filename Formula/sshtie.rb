class Sshtie < Formula
  desc "SSH + mosh + tmux profiles — one command to connect"
  homepage "https://github.com/ainsuotain/sshtie"
  url "https://github.com/ainsuotain/sshtie/archive/refs/tags/v0.1.1.tar.gz"
  sha256 "616495240470e71610d8e1fa603784446ab775bf44632b0bb2a16f134640581a"
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
