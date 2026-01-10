class Konfirm < Formula
  desc "Confirm kubectl before execution"
  homepage "https://github.com/Jaehyeon1020/konfirm"
  url "https://github.com/Jaehyeon1020/konfirm/archive/refs/tags/v0.2.1.tar.gz"
  sha256 "acfc645b5a7959d4965daa3a6497b955f28ae983b8b8373c3a22020814e4898e"
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

  def caveats
    <<~EOS
      \e[33;1mZsh completion requires compinit and loading konfirm's completion script.\e[0m
      \e[33;1mCopy and run this:\e[0m

        {
          echo ''
          echo '# konfirm setup'
          echo 'autoload -Uz compinit && compinit'
          echo 'source <(konfirm completion zsh)'
        } >> ~/.zshrc

        source ~/.zshrc
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/konfirm version")
  end
end
