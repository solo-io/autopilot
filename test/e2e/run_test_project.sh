#!/usr/bin/env bash

IMAGE_REPO="docker.io/ilackarms"

set -e

echo "Initializing canary operator"
ap init canary && pushd canary

echo "Cleaning up previous CanaryDeployment"
kubectl delete -f ../canary_example.yaml --ignore-not-found || echo cleanup failed, skipping

cp ../autopilot.yaml.txt autopilot.yaml

ap generate

echo "Writing spec.go && generating zz_deepcopy..."
cp ../spec.go.txt pkg/apis/canarydeployments/v1/spec.go && ap generate

echo "Writing initializing worker..."

cp ../initializing_worker.go.txt pkg/workers/initializing/worker.go

echo "Writing syncing worker..."

cp ../syncing_worker.go.txt pkg/workers/syncing/worker.go

echo "Writing Waiting worker..."

cp ../waiting_worker.go.txt pkg/workers/waiting/worker.go

echo "Writing Evaluating worker..."

cp ../evaluating_worker.go.txt pkg/workers/evaluating/worker.go

echo "Writing Promoting worker..."

cp ../promoting_worker.go.txt pkg/workers/promoting/worker.go

echo "Writing Rollback worker..."

cp ../rollback_worker.go.txt pkg/workers/rollback/worker.go

echo "Writing shared code..."
mkdir -p pkg/weights
cp ../virtual_service_weights.go.txt pkg/weights/virtual_service_weights.go

ap build ${IMAGE_REPO}/canary
ap deploy ${IMAGE_REPO}/canary -d

kubectl delete ns canary-operator

kubectl label namespace e2etest istio-injection=enabled --overwrite
kubectl apply -f ../canary_example.yaml
