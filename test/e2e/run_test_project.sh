#!/usr/bin/env bash

IMAGE_REPO="docker.io/ilackarms"

set -e

echo "########## Initializing test"
# source: https://github.com/torokmark/assert.sh/blob/master/assert.sh
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

echo "########## Cleaning up previous CanaryDeployment"
k create ns e2etest || echo Namespace exists
k delete -f ../canary_example.yaml --ignore-not-found || echo cleanup failed, skipping

#echo "########## Building canary project"
#cp ../autopilot.yaml.txt autopilot.yaml
#
#ap generate
#
#echo "########## Writing spec.go && generating zz_deepcopy..."
#cp ../spec.go.txt pkg/apis/canarydeployments/v1/spec.go && ap generate
#
#echo "########## Writing initializing worker..."
#
#cp ../initializing_worker.go.txt pkg/workers/initializing/worker.go
#
#echo "########## Writing Waiting worker..."
#
#cp ../waiting_worker.go.txt pkg/workers/waiting/worker.go
#
#echo "########## Writing Evaluating worker..."
#
#cp ../evaluating_worker.go.txt pkg/workers/evaluating/worker.go
#
#echo "########## Writing Promoting worker..."
#
#cp ../promoting_worker.go.txt pkg/workers/promoting/worker.go
#
#echo "########## Writing Rollback worker..."
#
#cp ../rollback_worker.go.txt pkg/workers/rollback/worker.go
#
#echo "########## Writing shared code..."
#mkdir -p pkg/weights
#cp ../virtual_service_weights.go.txt pkg/weights/virtual_service_weights.go
#
#ap build ${IMAGE_REPO}/canary
#ap deploy ${IMAGE_REPO}/canary -d

if [[ -n ${LOCAL} ]]; then
    echo k scale deployment canary-operator --replicas=0
    echo k scale deployment canary-operator --replicas=0
    echo k scale deployment canary-operator --replicas=0
    k scale deployment canary-operator --replicas=0
fi

sleep 1

k label namespace e2etest istio-injection=enabled --overwrite
k apply -f ../canary_example.yaml

sleep 1


echo "Expect Waiting state"
assert_eq phase Waiting

echo "Modifying the target deployment"
k apply -f deployment-change-ok.yaml

sleep 1
echo "Expect Evaluating state"
assert_eq phase Evaluating

echo "########## generating traffic..."

while true;  do
    k get canarydeployment -o yaml
    k exec -ti $(k get pod | grep hey | awk '{print $1}') -- hey -z 1s -c 1 http://example:9898/status/200
    sleep 1
done

