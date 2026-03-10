#!/bin/bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -Lo /usr/local/bin/gita \
  "https://github.com/whoisyurii/gita-cli/releases/latest/download/gita-${OS}-${ARCH}"
chmod +x /usr/local/bin/gita