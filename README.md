# bayesh
Bash command prediction using fzf. Based on your bash history, bayesh gives you statistical preditions for the next command you are going to type. That way, most of the time, you can let bayesh type your command for you.



### Some inspirations for this project
- [Peter Norvig](https://norvig.com/)'s excellent [blogpost](https://norvig.com/spell-correct.html) on how to build a spelling corrector.
- The exceptional [fzf](https://github.com/junegunn/fzf) and how to integrate it with bash's own [tab-completion](https://github.com/lincheney/fzf-tab-completion) framework

### Why "bayesh"? 👀
**Bayes**ian statistics on your **bash** history 🤷‍♂️. Currently this repo doesn't actually do any bayesian statistics, but it does compute conditional probabilities 🤓.