#!/usr/bin/env bash

set -e

echo "Initializing test operator"
ap init --group test --version v1 Test

pushd test

cat > autopilot.yaml <<EOF
apiVersion: test/v1
kind: Test
operatorName: test-operator
phases:
  - description: Test has begun initializing
    initial: true
    name: Initializing
    outputs:
      - deployments
      - services
  - description: Test has begun processing
    inputs:
      - metrics
    name: Processing
  - description: Test has passed
    name: Passed
    final: true
  - description: Test has failed
    name: Failed
    final: true
EOF

ap generate

echo "Writing spec.go"
cp ../spec.go.txt pkg/apis/tests/v1/spec.go

ap generate

echo "Writing initializing worker..."

cp ../initializing_worker.go.txt pkg/workers/initializing/worker.go

echo "Writing Processing worker..."

cp ../processing_worker.go.txt pkg/workers/processing/worker.go

