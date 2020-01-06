#!/usr/bin/env bash

source ./helpers.sh

set -e

echo "########## Using docker image repo ${IMAGE_REPO}"
echo "########## Set \$IMAGE_REPO to change"

echo "########## Initializing test"
trap 'cleanup && kill $(jobs -p)' EXIT

source ./assert.sh

echo "########## note: set \$LOCAL to scale operator pods to 0"

echo "######### Cleaning up previous Canary"
k delete -f ${DIR}/canary_example.yaml --ignore-not-found || echo cleanup failed, skipping

echo "########## Init Canary project"
# ap init canary --skip-gomod
mkdir canary
pushd canary

ap init . --kind Canary --group examples.io --version v1  --module canary.examples.io

echo "########## Building canary project"
cp ${DIR}/autopilot.yaml.txt autopilot.yaml

ap generate

echo "########## Writing spec.go && generating zz_deepcopy..."
cp ${DIR}/spec.go.txt pkg/apis/canaries/v1/spec.go && ap generate

echo "########## Writing initializing worker..."

cp ${DIR}/initializing_worker.go.txt pkg/workers/initializing/worker.go

echo "########## Writing Waiting worker..."

cp ${DIR}/waiting_worker.go.txt pkg/workers/waiting/worker.go

echo "########## Writing Evaluating worker..."

cp ${DIR}/evaluating_worker.go.txt pkg/workers/evaluating/worker.go

echo "########## Writing Promoting worker..."

cp ${DIR}/promoting_worker.go.txt pkg/workers/promoting/worker.go

echo "########## Writing Rollback worker..."

cp ${DIR}/rollback_worker.go.txt pkg/workers/rollback/worker.go

echo "########## Writing shared code..."
mkdir -p pkg/weights
cp ${DIR}/virtual_service_weights.go.txt pkg/weights/virtual_service_weights.go

ap build ${IMAGE_REPO}/canary
ap deploy ${IMAGE_REPO}/canary -p -d

if [[ -n ${LOCAL} ]]; then
    kubectl -n canary-operator scale deployment canary-operator --replicas=0
fi

sleep 1
k create ns e2etest || echo Namespace exists

k label namespace e2etest istio-injection=enabled --overwrite
k apply -f ${DIR}/canary_example.yaml

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
eventually_eq $(k get canaries.canary.examples.io example -ojson | jq ".status.history[0].promotionSucceeded") "true"

echo "########## Modifying the target deployment"
k set image deployment/example podinfod=stefanprodan/podinfo:3.1.0

sleep 1
echo "########## Expect Evaluating state after second change"
eventually_eq $(phase) Evaluating

generate_traffic 500 &

sleep 10
echo "########## Expect Waiting state"
eventually_eq $(phase) Waiting
eventually_eq $(k get canaries.canary.examples.io example -ojson | jq ".status.history[1].promotionSucceeded") "false"


