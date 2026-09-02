# This file written by AI :)
#
# ttype zsh plugin: an alias and completions.
#   git clone https://github.com/alirezaudev/ttype ~/.oh-my-zsh/custom/plugins/ttype
#   plugins=(... ttype)

alias tt='ttype'

if (( $+commands[ttype] )); then
  # Cache the generated completion so every new shell does not shell out.
  _ttype_comp_dir="${XDG_CACHE_HOME:-$HOME/.cache}/ttype"
  _ttype_comp_file="$_ttype_comp_dir/_ttype"

  if [[ ! -s "$_ttype_comp_file" || "$commands[ttype]" -nt "$_ttype_comp_file" ]]; then
    mkdir -p "$_ttype_comp_dir"
    ttype completion zsh >| "$_ttype_comp_file" 2>/dev/null
  fi

  fpath=("$_ttype_comp_dir" $fpath)
  autoload -U compinit && compinit -u

  unset _ttype_comp_dir _ttype_comp_file
fi
