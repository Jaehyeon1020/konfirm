class Konfirm < Formula
  desc "Confirm kubectl before execution"
  homepage "https://github.com/jaehyeonkim/konfirm"
  url "https://github.com/jaehyeonkim/konfirm/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "추가예정" # TODO: release action
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
