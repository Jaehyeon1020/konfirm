# konfirm

konfirm is a small wrapper around kubectl that confirms the effective context before executing any kubectl command.

![konfirm_v0 3 0_demo_1](https://github.com/user-attachments/assets/e42665da-ddca-4de9-9354-cd1c3b07892f)


## Features
- Prompts for confirmation based on the effective context (including `--context` overrides).
- Lets you permanently allow a context, or allow specific kubectl subcommands per context.

## Prerequisites
- Go
- kubectl installed and available on your PATH

## Installation

### Homebrew

Create a tap and use the formula in `konfirm.rb`:

```bash
# Install Homebrew: https://brew.sh/
brew tap Jaehyeon1020/konfirm https://github.com/Jaehyeon1020/konfirm
brew install Jaehyeon1020/konfirm/konfirm
```

### Update (Homebrew)

```bash
brew update
brew upgrade Jaehyeon1020/konfirm/konfirm
```

### Uninstall
```bash
brew uninstall Jaehyeon1020/konfirm/konfirm
rm -rf ~/Library/Application\ Support/konfirm
```

## Usage

```bash
konfirm kubectl <kubectl args...>
konfirm add <subcommand>
konfirm add --all
konfirm remove <subcommand>
konfirm remove --all
konfirm status
```

### Examples

Confirm the current context before running kubectl:

```bash
konfirm kubectl get pods -n kube-system
```

Confirm a context override:

```bash
konfirm kubectl --context prod-cluster get deploy
```

Allow the current context permanently:

```bash
konfirm add --all
```

Allow a kubectl subcommand (per current context):

```bash
konfirm add apply
```

Check what is allowed for the current context:

```bash
konfirm status
```

## Tips

Add the alias to your `~/.zshrc`, then reload your shell:

```bash
echo 'alias k="konfirm kubectl"' >> ~/.zshrc
source ~/.zshrc
```

Now you can use kubectl as usual while integrating konfirm:

```bash
# From now on, running this command will display a prompt asking for approval.
k get pods
```


## Build a local binary (Not Recommended)

```bash
go build -o konfirm ./cmd/konfirm
mv konfirm /usr/local/bin/
```

### Shell completion
> **If you install via Homebrew, the completion file is installed automatically.**
>
> **You do not need to run the command below.**


Generate and source completion in your shell startup file:

```bash
# zsh
source <(konfirm completion zsh)
```

### Uninstall

Remove the binary:

```bash
rm -f "$(command -v konfirm)"
```

Remove stored config:

MacOS
```bash
rm -rf ~/Library/Application\ Support/konfirm
```
