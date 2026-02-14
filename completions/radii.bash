_radii_compreply() {
  COMPREPLY=()
  mapfile -t COMPREPLY < <(compgen "$@" || true)
}

_radii() {
  # prev is set by _init_completion; kept local to avoid clobbering globals
  # shellcheck disable=SC2034
  local cur prev words cword
  _init_completion -n : || return

  local global_flags commands
  global_flags="--verbose --quiet --debug --version --help"
  commands="install remove list"

  # Completing the subcommand (first arg after radii)
  if (( cword == 1 )); then
    if [[ $cur == -* ]]; then
      _radii_compreply -W "$global_flags" -- "$cur"
    else
      _radii_compreply -W "$commands $global_flags" -- "$cur"
    fi
    return 0
  fi

  local cmd=${words[1]}

  # Completing options after the command (including empty cur after a space)
  if [[ -z $cur || $cur == -* ]]; then
    case $cmd in
      install|in)
        _radii_compreply -W "$global_flags --auto-detect --batch --dry-run --force" -- "$cur"
        ;;
      remove|rm)
        _radii_compreply -W "$global_flags --all --batch --dry-run" -- "$cur"
        ;;
      list|ls)
        _radii_compreply -W "$global_flags --available --installed" -- "$cur"
        ;;
      *)
        _radii_compreply -W "$global_flags" -- "$cur"
        ;;
    esac
    return 0
  fi

  # Positional completion (driver IDs) could go here later.
  return 0
}

complete -F _radii radii
