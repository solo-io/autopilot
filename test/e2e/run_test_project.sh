#!/usr/bin/env bash

IMAGE_REPO="docker.io/ilackarms"

set -e

echo "Initializing test operator"
ap init --group test --version v1 Test

pushd test


echo "Cleaning up previous Test"
kubectl delete -f ../test_custom_resource.yaml --ignore-not-found


cp ../autopilot.yaml.txt autopilot.yaml

ap generate

echo "Writing spec.go"
cp ../spec.go.txt pkg/apis/tests/v1/spec.go

ap generate

echo "Writing initializing worker..."

cp ../initializing_worker.go.txt pkg/workers/initializing/worker.go

echo "Writing Processing worker..."

cp ../processing_worker.go.txt pkg/workers/processing/worker.go

ap build ${IMAGE_REPO}/test
ap deploy ${IMAGE_REPO}/test -d

kubectl delete ns test-operator

kubectl label namespace e2etest istio-injection=enabled --overwrite
kubectl apply -f ../test_custom_resource.yaml
