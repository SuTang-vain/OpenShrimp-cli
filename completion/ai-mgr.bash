# ai-mgr bash completion

_ai_mgr() {
    local cur prev words cword
    _init_completion || return

    local commands=(
        "scan"
        "cleanup"
        "switch"
        "link"
        "check"
        "backup"
        "restore"
        "stats"
        "context"
        "scheduler"
        "credentials"
        "daemon"
        "version"
    )

    case "${cword}" in
        1)
            COMPREPLY=($(compgen -W "${commands[*]}" -- "${cur}"))
            ;;
        2)
            case "${prev}" in
                switch)
                    # Get list of models from config
                    local models=$(ai-mgr switch --list 2>/dev/null | grep -oE '^\s*\*{0,1}\s*[a-zA-Z0-9-]+' | sed 's/[*[:space:]]//g')
                    COMPREPLY=($(compgen -W "${models}" -- "${cur}"))
                    ;;
                cleanup)
                    # Get list of tools
                    local tools=$(ai-mgr scan 2>/dev/null | grep -oE '\[.*\]' | tr -d '[]')
                    COMPREPLY=($(compgen -W "${tools}" -- "${cur}"))
                    ;;
                link)
                    COMPREPLY=($(compgen -W "create remove verify init" -- "${cur}"))
                    ;;
                backup)
                    COMPREPLY=($(compgen -W "--include-data --json" -- "${cur}"))
                    ;;
                scheduler)
                    COMPREPLY=($(compgen -W "--list --add --remove --enabled --type --schedule --tool --json" -- "${cur}"))
                    ;;
                credentials)
                    COMPREPLY=($(compgen -W "--list --set --get --delete --model --key --value --env --provider --json" -- "${cur}"))
                    ;;
                *)
                    ;;
            esac
            ;;
        *)
            case "${words[1]}" in
                scheduler)
                    if [[ "${prev}" == "--type" ]]; then
                        COMPREPLY=($(compgen -W "cleanup backup" -- "${cur}"))
                    fi
                    ;;
                credentials)
                    if [[ "${prev}" == "--model" ]]; then
                        COMPREPLY=($(compgen -W "claude-sonnet-4 minimax-m2.1 glm-4.7" -- "${cur}"))
                    fi
                    ;;
                *)
                    ;;
            esac
            ;;
    esac
}

complete -F _ai_mgr ai-mgr
