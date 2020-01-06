#!/usr/bin/env bash

## Variables
DIR="${GOPATH}/src/github.com/solo-io/autopilot/test/e2e"
IMAGE_REPO=${IMAGE_REPO:-"docker.io/ilackarms"}

function k {
    kubectl --namespace e2etest $@
}

function phase {
    k get canaries.canary.examples.io example -ojson | jq ".status.phase" -r
}

function eventually_eq {
    TIMEOUT=30 # hard-coded 30s timeout
    until [[ ${TIMEOUT} -lt 1 ]]
    do
      echo "Trying '$1' == '$2'"
      local res=$(assert_eq $1 $2; echo $?)
      if [[ ${res} -eq 0 ]]; then
        echo "Passed '$@'"
        return 0
      fi
      echo "failed: $@"
      ((TIMEOUT=TIMEOUT-1))
      sleep 1
    done
    return 1
}

function generate_traffic {
    echo "########## Generating Traffic with status code $1"
    while true;  do
        k get canary -o yaml
        k exec $(k get pod | grep Running | grep hey | awk '{print $1}') -c hey -- hey -z 1s -c 1 http://example:9898/status/$1 || \
            echo container not ready
        sleep 1
    done
}

function cleanup {
    set -x
    kubectl delete ns canary-operator
    kubectl delete ns e2etest
    kubectl delete crd canaries.canary.examples.io
    set +x
    echo cleaned up!
}