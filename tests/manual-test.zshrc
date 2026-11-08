# Minimal Zsh configuration used by `make manual-test`
HISTSIZE=1000
PROMPT='%% '
autoload -Uz compinit
compinit
source <(bayesh --zsh)
