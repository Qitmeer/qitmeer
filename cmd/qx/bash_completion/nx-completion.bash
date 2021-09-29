#
#  Command-line completion for qx.
#
_qx()
{
    local current="${COMP_WORDS[COMP_CWORD]}"
    
    # Generated from XML data source.
    local commands="
        base58-decode
        base58-encode
        base58check-decode
        base58check-encode
        rlp-decode
        rlp-encode
        blake2b256
        blake2b512
        blake256
        sha256
        sha3-256
        keccak-256
        ripemd160
        bitcoin160
        hash160
        entropy
        hd-new
        hd-to-public
        hd-to-ec
        hd-decode
        hd-derive
        mnemonic-new
        mnemonic-to-entropy
        mnemonic-to-seed
        ec-new
        ec-to-public
        ec-to-wif
        wif-to-ec
        wif-to-public
        ec-to-addr
        tx-decode
        tx-encode
        tx-sign
        msg-sign
        msg-verify
        compact-to-uint64
        uint64-to-compact
        diff-to-gps
        script-encode
        script-decode
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
complete -F _qx qx
