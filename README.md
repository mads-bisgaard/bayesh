# bayesh
A faster and more efficient way to use your shell history.

## Description
Bayesh suggests commands to you in your Bash shell. This is done by maintaining a database of your shell history and suggesting you relevant commands based on a [statistical model](https://github.com/mads-bisgaard/bayesh/wiki/The-statistical-model).

## Usage
The recommended way to use bayesh is by setting up a keybinding in bash (see below). When you hit the bound key(s) bayesh will recommend you the shell commands relevant to the context you are in. Bayesh will need to learn a bit about your workflows, so suggestions will improve after a learning phase (especially if you have repetitive workflows).

## Installation
The following has been tested on Fedora and Ubuntu:
1. Ensure the dependencies are installed ([fzf](https://github.com/junegunn/fzf) and [python3](https://www.python.org/))
2. clone this repository
3. run `bash install.bash` from the root of this repository.
4. Add the following lines to your `~/.bashrc`
```bash
source "<bayesh root dir>/shell/bayesh.bash"
bind -x '"\C-e":"bayesh_infer_cmd"'
```
Now you should be able to generate bash predictions by pressing `Ctrl+e` in your bash terminal. Probably you don't see any predictions at first, but after some time it will have collected enough data to (hopefully) generate great predictions.

### Some inspirations for this project
- [autojump](https://github.com/wting/autojump), [z](https://github.com/rupa/z) and [zoxide](https://github.com/ajeetdsouza/zoxide)
- The exceptional [fzf](https://github.com/junegunn/fzf)
- [Peter Norvig](https://norvig.com/)'s excellent [blogpost](https://norvig.com/spell-correct.html) on how to build a spelling corrector.

### Why "bayesh"? 👀
**Bayes**ian statistics on your **bash** history 🤷‍♂️.