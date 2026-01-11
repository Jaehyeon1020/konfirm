# konfirm
영문 README: [README.md](README.md)

**konfirm**은 `kubectl` 명령을 실행하기 전에 **실제 적용될 context를 확인하고 승인하도록 하는 경량 래퍼(wrapper)** 도구입니다.

![konfirm_v0 3 0_demo_1](https://github.com/user-attachments/assets/e42665da-ddca-4de9-9354-cd1c3b07892f)

## 주요 기능
- `--context` 오버라이드를 포함하여 **실제 적용되는 context 기준으로 확인 프롬프트를 표시**합니다.
- context 단위로:
  - 전체 kubectl 명령을 영구적으로 허용하거나
  - 특정 kubectl 서브커맨드만 선택적으로 허용할 수 있습니다.

## 사전 요구 사항
- Go
- `kubectl`이 설치되어 있고 PATH에 등록되어 있어야 합니다.

## 설치 방법

### Homebrew

탭(tap)을 추가한 뒤 `konfirm.rb` 포뮬러를 사용해 설치합니다.

```bash
# Homebrew 설치: https://brew.sh/
brew tap Jaehyeon1020/konfirm https://github.com/Jaehyeon1020/konfirm
brew install Jaehyeon1020/konfirm/konfirm
```

### 업데이트 (Homebrew)

```bash
brew update
brew upgrade Jaehyeon1020/konfirm/konfirm
```

### 제거

```bash
brew uninstall Jaehyeon1020/konfirm/konfirm
rm -rf ~/Library/Application\ Support/konfirm
```

## 사용 방법

```bash
konfirm kubectl <kubectl args...>
konfirm add <subcommand>
konfirm add --all
konfirm remove <subcommand>
konfirm remove --all
konfirm status
```

### 사용 예시

현재 context를 확인한 뒤 kubectl 실행:

```bash
konfirm kubectl get pods -n kube-system
```

context 오버라이드를 사용하는 경우:

```bash
konfirm kubectl --context prod-cluster get deploy
```

현재 context를 **영구적으로 허용**:

```bash
konfirm add --all
```

특정 kubectl 서브커맨드를 현재 context에서 허용:

```bash
konfirm add apply
```

현재 context에서 허용된 항목 확인:

```bash
konfirm status
```

## 팁

`~/.zshrc`에 alias를 추가한 뒤 쉘을 다시 로드하세요.

```bash
echo 'alias k="konfirm kubectl"' >> ~/.zshrc
source ~/.zshrc
```

이제 기존 kubectl 사용 방식 그대로 `konfirm`을 함께 사용할 수 있습니다.

```bash
# 이후 이 명령을 실행하면 승인 프롬프트가 표시됩니다.
k get pods
```

## 로컬 바이너리 빌드 (권장하지 않음)

```bash
go build -o konfirm ./cmd/konfirm
mv konfirm /usr/local/bin/
```

### 셸 자동 완성
> **Homebrew로 설치한 경우 자동 완성 파일은 자동으로 설치됩니다.**
>
> **아래 명령을 실행할 필요가 없습니다.**

쉘 시작 파일에서 자동 완성을 생성하고 로드하려면:

```bash
# zsh
source <(konfirm completion zsh)
```

### 제거

바이너리 삭제:

```bash
rm -f "$(command -v konfirm)"
```

설정 파일 삭제:

macOS:
```bash
rm -rf ~/Library/Application\ Support/konfirm
```
