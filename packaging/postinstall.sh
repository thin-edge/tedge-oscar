#!/bin/sh
set -e

if command -V zsh >/dev/null 2>&1; then
    tedge-oscar completion zsh > /usr/share/zsh/vendor-completions/_tedge-oscar
fi
