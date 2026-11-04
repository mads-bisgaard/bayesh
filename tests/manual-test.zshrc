# Minimal Zsh configuration used by `make manual-test`
HISTSIZE=1000
unsetopt HIST_IGNORE_DUPS
PROMPT='%% '
autoload -Uz compinit
compinit
source <(bayesh --zsh)
