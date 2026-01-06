# konfirm

konfirm is a small wrapper around kubectl that confirms the effective context before executing any kubectl command.

<img width="473" height="220" alt="스크린샷 2026-01-06 오후 10 43 19" src="https://github.com/user-attachments/assets/dc695fcd-8ba2-4aff-a027-9bf73b5e124b" />

<br />

<img width="472" height="190" alt="스크린샷 2026-01-06 오후 10 44 00" src="https://github.com/user-attachments/assets/02d06f26-7638-4e0c-b62f-9ffbbe89ea12" />


## Features
- Prompts for confirmation based on the effective context (including `--context` overrides).
- Lets you permanently allow a context, or allow specific kubectl subcommands per context.

## Installation

### Build a local binary

```bash
go build -o konfirm ./cmd/konfirm
mv konfirm /usr/local/bin/
```

## Uninstall

Remove the binary:

```bash
rm -f "$(command -v konfirm)"
```

Remove stored config:

MacOS
```bash
rm -f "~/Library/Application\ Support/konfirm/config.json"
```

## Usage

```bash
konfirm kubectl <kubectl args...>
konfirm k <kubectl args...>
konfirm allow add
konfirm allow remove
konfirm allow list
konfirm allow kubectl add <subcommand>
konfirm allow kubectl remove <subcommand>
konfirm allow kubectl list
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
konfirm allow add
```

Allow a kubectl subcommand (per current context):

```bash
konfirm allow kubectl add apply
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
k get pods
```
