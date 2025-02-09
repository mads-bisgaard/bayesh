# Bayesh
A faster and more efficient way to use your shell history.

## Description
Bayesh suggests relevant commands to you in your shell (using fzf). This is done by maintaining a database of your shell history and suggesting you relevant commands based on a statistical model.


## Installation
The following has been tested on Fedora and Ubuntu:
1. Ensure the dependencies are installed ([fzf](https://github.com/junegunn/fzf) and [python3](https://www.python.org/))
2. clone this repository
3. run `bash install.bash` from the root of this repository.

Now check that Bayesh is installed by running `bayesh --help`. To get the most out of Bayesh you should integrate it into your shell. Currently [Bash](#bash-integration) and [Zsh](#zsh-integration) are supported.

### Bash integration
Add the following lines to your `~/.bashrc` to create a keybinding for triggering Bayesh.

```bash
source "<bayesh root dir>/shell/bayesh.bash"
bind -x '"\C-e":"bayesh_infer_cmd"'
```

Now you can generate bash command suggestions by pressing `Ctrl+e` in your bash shell.

### Zsh integration
Add the following lines to your `~/.zshrc` to trigger Bayesh automatically in your zsh shell:

```bash
autoload -Uz add-zle-hook-widget
source "<bayesh root dir>/shell/bayesh.zsh"
add-zle-hook-widget zle-line-init bayesh_infer_cmd
```
To disable the automatic trigger, remove `add-zle-hook-widget zle-line-init bayesh_infer_cmd` from your `~/.zshrc`:

## Remarks
The purpose of Bayesh is to detect your repetitive workflows in your shell and via great UI (=[fzf](https://github.com/junegunn/fzf)) allow you to quickly reuse commands. In that sense its purpose is similar to the auto suggestion/complete feature smartphones offer. A key difference however, is that Bayesh is not "trained" on any external data. That means it will only ever suggest commands you have previously used. In particular it will only start generating useful suggestions after a short learning phase.

## Gotchas
Bayesh uses the Bash `history` builtin to determine when a new command is executed and when not. Hence, to get the most out of Bayesh you are advised to ensure that, when you execute the same command twice in a row it is visible in your bash history that you ran it twice. In Bash you can do that by adding `export HISTCONTROL=` in your `~/.bashrc` (see [here](https://www.gnu.org/software/bash/manual/bash.html#index-HISTCONTROL)). In Zsh you can do that by adding `unsetopt HIST_IGNORE_DUPS` to your `~/.zshrc` (see [here](https://zsh.sourceforge.io/Doc/Release/Options.html)). You can test your setup by running `history -1` twice in your shell and check that the two lines printed are different (either because they display different timestamps or history event numbers).

### Some inspirations for this project
- [autojump](https://github.com/wting/autojump), [z](https://github.com/rupa/z) and [zoxide](https://github.com/ajeetdsouza/zoxide)
- The exceptional [fzf](https://github.com/junegunn/fzf)
- [Peter Norvig](https://norvig.com/)'s excellent [blogpost](https://norvig.com/spell-correct.html) on how to build a spelling corrector.

### Why "Bayesh"? 👀
**Bayes**ian statistics on your **bash** history 🤷‍♂️.