# bayesh
Bash command prediction using fzf. Based on your bash history, bayesh gives you statistical preditions for the next command you are going to type. That way, most of the time, you can let bayesh type your command for you.

This project is still in a very experimental state, but I use it everyday myself with joy. Let me know if you are interested in contributing 😁.

## Installation
I have tested the following on Fedora and Ubuntu
1. Ensure the dependencies are installed ([fzf](https://github.com/junegunn/fzf) and [python3](https://www.python.org/))
2. clone this repository
3. run `bash install.bash` from the root of this repository.
4. Add the following lines to your `~/.bashrc`
```bash
source "<bayesh root dir>/shell/bayesh.bash"
bind -x '"\C-e":"bayesh_infer_cmd"'
```
Now you should be able to generate bash predictions by pressing `Ctrl+e` in your bash terminal. Probably you don't see any predictions at first, but after some time it will have collected enough data to (hopefully) generate great predictions.

### fzf-tab-completion integration
Alternatively (or maybe "preferably") you can setup bayesh as part of bash autocompletion. That way, if you hit `Tab` in your bash shell with an empty prompt you will get bayesh's preditions. If the prompt is non-empty you get bash's own autocompletion. To achieve this, install [fzf-tab-completion](https://github.com/lincheney/fzf-tab-completion) and add
```bash
source <bayesh root dir>/shell/bayesh.bash
source <bayesh root dir>/shell/fzf-tab-completion-plugin.bash
bind -x '"\t": bayesh_autocomplete'
```
to your `~/.bashrc`.


### Some inspirations for this project
- [Peter Norvig](https://norvig.com/)'s excellent [blogpost](https://norvig.com/spell-correct.html) on how to build a spelling corrector.
- The exceptional [fzf](https://github.com/junegunn/fzf) and how to integrate it with bash's own [tab-completion](https://github.com/lincheney/fzf-tab-completion) framework

### Why "bayesh"? 👀
**Bayes**ian statistics on your **bash** history 🤷‍♂️. Currently this repo doesn't actually do any bayesian statistics, but it does compute conditional probabilities 🤓.