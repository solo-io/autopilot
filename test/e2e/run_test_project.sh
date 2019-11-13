#!/usr/bin/env bash

IMAGE_REPO="docker.io/ilackarms"

set -e

echo "########## Initializing test"
trap 'kill $(jobs -p)' EXIT

source ./assert.sh

echo "########## Init Canary project"
ap init canary && pushd canary
echo "########## note: set \$LOCAL to scale operator pods to 0"

function k {
    kubectl --namespace e2etest $@
}

function phase {
    k get canarydeployments.autopilot.examples.io -oyaml | grep "phase: " | sed 's/phase: //'
}

function eventually_eq {
    TIMEOUT=30 # hard-coded 30s timeout
    until [[ ${TIMEOUT} -lt 1 ]]
    do
      echo "Trying '$@'"
      local res=$(assert_eq $@; echo $?)
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
        k get canarydeployment -o yaml
        k exec $(k get pod | grep Running | grep hey | awk '{print $1}') -c hey -- hey -z 1s -c 1 http://example:9898/status/$1 || \
            echo container not ready
        sleep 1
    done
}
echo "######### Cleaning up previous CanaryDeployment"
k create ns e2etest || echo Namespace exists
k delete -f ../canary_example.yaml --ignore-not-found || echo cleanup failed, skipping

echo "########## Building canary project"
cp ../autopilot.yaml.txt autopilot.yaml

ap generate

echo "########## Writing spec.go && generating zz_deepcopy..."
cp ../spec.go.txt pkg/apis/canarydeployments/v1/spec.go && ap generate

echo "########## Writing initializing worker..."

cp ../initializing_worker.go.txt pkg/workers/initializing/worker.go

echo "########## Writing Waiting worker..."

cp ../waiting_worker.go.txt pkg/workers/waiting/worker.go

echo "########## Writing Evaluating worker..."

cp ../evaluating_worker.go.txt pkg/workers/evaluating/worker.go

echo "########## Writing Promoting worker..."

cp ../promoting_worker.go.txt pkg/workers/promoting/worker.go

echo "########## Writing Rollback worker..."

cp ../rollback_worker.go.txt pkg/workers/rollback/worker.go

echo "########## Writing shared code..."
mkdir -p pkg/weights
cp ../virtual_service_weights.go.txt pkg/weights/virtual_service_weights.go

ap build ${IMAGE_REPO}/canary
ap deploy ${IMAGE_REPO}/canary -d

if [[ -n ${LOCAL} ]]; then
    kubectl -n canary-operator scale deployment canary-operator --replicas=0
fi

sleep 1

k label namespace e2etest istio-injection=enabled --overwrite
k apply -f ../canary_example.yaml

sleep 5

echo "########## Expect Init to become Waiting state"
eventually_eq $(phase) Waiting

echo "########## Modifying the target deployment"
k set image deployment/example podinfod=stefanprodan/podinfo:3.1.1

sleep 1
echo "########## Expect Evaluating state after change"
eventually_eq $(phase) Evaluating

generate_traffic 200 &

sleep 45
echo "########## Expect Waiting state after promotion"
eventually_eq $(phase) Waiting
eventually_eq $(k get canarydeployments.autopilot.examples.io example -ojson | jq ".status.history[0].promotionSucceeded") "true"

echo "########## Modifying the target deployment"
k set image deployment/example podinfod=stefanprodan/podinfo:3.1.0

sleep 1
echo "########## Expect Evaluating state after second change"
eventually_eq $(phase) Evaluating

kill $(jobs -p)
generate_traffic 500 &

sleep 10
echo "########## Expect Waiting state"
eventually_eq $(phase) Waiting
eventually_eq $(k get canarydeployments.autopilot.examples.io example -ojson | jq ".status.history[1].promotionSucceeded") "false"


