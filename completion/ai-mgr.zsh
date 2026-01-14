#compdef ai-mgr

_ai-mgr() {
    local -a commands

    commands=(
        'scan:Scan for AI tools on your system'
        'cleanup:Clean up temporary files'
        'switch:Switch between AI models'
        'link:Manage symbolic links for AI tool configurations'
        'check:Health check for AI tools'
        'backup:Backup AI tool configurations'
        'restore:Restore AI tool configurations from backup'
        'stats:Show usage statistics'
        'context:Manage AI conversation context and project memory'
        'scheduler:Manage scheduled tasks'
        'credentials:Manage API credentials securely'
        'daemon:Start HTTP server for Web UI'
        'version:Show version'
    )

    _describe -t commands 'ai-mgr command' commands
}

_ai-mgr
