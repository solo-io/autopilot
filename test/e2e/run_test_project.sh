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
cat > pkg/apis/tests/v1/spec.go <<EOF
package v1

import "time"

// EDIT THIS FILE!  This file should contain the definitions for your API Spec and Status!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "autopilot generate" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

// TestSpec defines the desired state of Test
// +k8s:openapi-gen=true
type TestSpec struct {
	// Service must stay above this number in order to pass
	SuccessThreshold int64 `json:"successThreshold"`

	// Duration of the test
	Duration time.Duration `json:"duration"`
}

// TestStatusInfo defines an observed condition of Test
// +k8s:openapi-gen=true
type TestStatusInfo struct {
	// used to record the time processing started
	TimeStarted time.Time `json:"timeStarted"`
}

EOF

ap generate

echo "Writing initializing worker..."
