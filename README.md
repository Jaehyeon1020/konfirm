# konfirm

**konfirm** is a simple wrapper around kubectl that confirms the effective context before executing any `kubectl` command.

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

To enable autocompletion after installing `konfirm`, run the snippet below.

For example, `konfirm kubectl get pods <tab>` autocompletes available Pod names.

```bash
{
  echo ''
  echo '# konfirm setup'
  echo 'autoload -Uz compinit && compinit'
  echo 'source <(konfirm completion zsh)'
} >> ~/.zshrc

source ~/.zshrc
```

### Update (Homebrew)

```bash
brew update
brew upgrade Jaehyeon1020/konfirm/konfirm
```

### Uninstall
```bash
brew uninstall Jaehyeon1020/konfirm/konfirm
```

If you also want to remove the configuration file (including allowed contexts and subcommands):
```bash
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

### Command details and examples

`konfirm kubectl <kubectl args...>`  
Runs kubectl only after confirming the effective context (including `--context`).

```bash
# Confirm the current context before running kubectl.
konfirm kubectl get pods -n kube-system

# Confirm a context override explicitly.
konfirm kubectl --context prod-cluster get deploy
```

`konfirm add <subcommand>`  
Always allow a specific kubectl subcommand for the current context.

```bash
# Allow `kubectl apply` for the current context.
konfirm add apply

# Allow `kubectl delete` for the current context.
konfirm add delete
```

`konfirm add --all`  
Always allow all kubectl subcommands for the current context.

```bash
# Allow all kubectl subcommands for the current context.
konfirm add --all
```

`konfirm remove <subcommand>`  
Remove a previously allowed kubectl subcommand for the current context.

```bash
# Revoke approval for the `kubectl apply` command.
konfirm remove apply
```

`konfirm remove --all`  
Remove all allowances for the current context (back to full confirmation).

```bash
# Revoke all allowances for the current context.
konfirm remove --all
```

`konfirm status`  
Show the effective context and the current allowlist.

```bash
# Check what is allowed for the current context.
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

---
Korean README: [README.ko.md](README.ko.md)