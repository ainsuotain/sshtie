class Sshtie < Formula
  desc "SSH + mosh + tmux profiles — one command to connect"
  homepage "https://github.com/ainsuotain/sshtie"
  url "https://github.com/ainsuotain/sshtie/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "8df4f49b2b9d0605b8e45e4024a5079ccbe0224034a58b9eeca8d9d1b6c5db95"
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
