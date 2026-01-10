class Konfirm < Formula
  desc "Confirm kubectl before execution"
  homepage "https://github.com/Jaehyeon1020/konfirm"
  url "https://github.com/Jaehyeon1020/konfirm/archive/refs/tags/v0.2.1.tar.gz"
  sha256 "e0bcbf30b4e71d6dc6e7fddcda209280591a1eddd49d045419c182f96446a134"
  license "MIT"

  depends_on "go" => :build
  depends_on "kubectl"

  def install
    system "go", "build", "-o", "konfirm", "./cmd/konfirm"
    bin.install "konfirm"

    zsh_output = Utils.safe_popen_read(bin/"konfirm", "completion", "zsh")
    (buildpath/"_konfirm").write zsh_output
    zsh_completion.install "_konfirm"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/konfirm version")
  end
end
