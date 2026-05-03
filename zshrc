export ZSH="$HOME/.oh-my-zsh"
export EDITOR='nano'
ZSH_THEME="robbyrussell"

plugins=(docker git zsh-autosuggestions)

source $ZSH/oh-my-zsh.sh

## make tab-complete
complete -W "`grep -oE '^[a-zA-Z0-9_.-]+:([^=]|$)' Makefile | sed 's/[^a-zA-Z0-9_.-]*$//'`" make

alias r="source ~/.zshrc"
alias gs="git status"
alias gb="go build ./..."
alias gt="go test ./..."

help() {
  echo "Welcome, here are the commands you can run"
  echo " "
  echo "r       - reload current zshrc"
  echo "gb      - go build ./..."
  echo "gt      - go test ./..."
  echo "make    - run Makefile targets (tab for autocomplete)"
}

bindkey '^ ' autosuggest-accept

PROMPT='%F{blue}🐳%f:%F{green}%~%f $(git_prompt_info)%F{blue}%#%f '

ZSH_THEME_GIT_PROMPT_PREFIX="%F{yellow}"
ZSH_THEME_GIT_PROMPT_SUFFIX="%f"
