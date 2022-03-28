#!/usr/bin/env bash

set -e

function debug {
    ./cluster/kubectl.sh get events
    ./cluster/kubectl.sh get all -n $1
    ./cluster/kubectl.sh get pod -n $1 | awk 'NR>1 {print $1}' | xargs ./cluster/kubectl.sh logs -n $1 --tail=50
}

# Install golang
function ensure_golang {
    GOVERSION='go1.17.5.linux-amd64.tar.gz'
    if [[ "$(go version 2>&1)" =~ "not found" ]]; then
        wget -q https://dl.google.com/go/${GOVERSION}
        tar -C /usr/local -xzf ${GOVERSION}
    fi
}
