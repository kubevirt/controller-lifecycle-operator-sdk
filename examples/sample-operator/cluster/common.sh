#!/usr/bin/env bash

set -e

function debug {
    ./cluster/kubectl.sh get events
    ./cluster/kubectl.sh get all -n $1
    ./cluster/kubectl.sh get pod -n $1 | awk 'NR>1 {print $1}' | xargs ./cluster/kubectl.sh logs -n $1 --tail=50
}