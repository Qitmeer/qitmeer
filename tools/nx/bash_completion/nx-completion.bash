#
#  Command-line completion for nx.
#
_nx()
{
    local current="${COMP_WORDS[COMP_CWORD]}"
    
    # Generated from XML data source.
    local commands="
        base58-decode
        base58-encode
        base58check-decode
        base58check-encode
        blake2b256
        blake256
        sha256
        ripemd160
        bitcoin160
        hash160
        seed
        hd-new
        hd-to-public
        mnemonic-new
        mnemonic-to-seed
    "

    if [[ $COMP_CWORD == 1 ]]; then
        COMPREPLY=( `compgen -W "$commands" -- $current` )
        return
    fi

    local command=COMP_WORDS[1]
    local options="--help"

    # TODO: Generate per-command options here

    COMPREPLY=( `compgen -W "$options" -- $current` )
}
complete -F _nx nx
