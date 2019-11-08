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

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

// TestSpec creates a deployment and a load tester for that deployment
// TestSpec defines the desired state of Test
// +k8s:openapi-gen=true
type TestSpec struct {
	PodSpec v1.PodSpec `json:"podSpec"`

	// Service must stay above this number in order to pass
	SuccessThreshold float64 `json:"successThreshold"`

	// Duration of the test
	Duration time.Duration `json:"duration"`
}

// TestStatusInfo defines an observed condition of Test
// +k8s:openapi-gen=true
type TestStatusInfo struct {
	// used to record the time processing started
	TimeStarted time.Time `json:"timeStarted"`

	// each recorded result
	Results []float64 `json:"results"`
}
EOF

ap generate

echo "Writing initializing worker..."

cat > pkg/workers/initializing/worker.go <<EOF

package initializing

import (
	"context"
	"github.com/solo-io/autopilot/test/e2e/test/pkg/parameters"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"

	"github.com/go-logr/logr"
	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/test/e2e/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

func (w *Worker) Sync(ctx context.Context, test *v1.Test) (Outputs, v1.TestPhase, *v1.TestStatusInfo, error) {
	w.Logger.Info("initializing deployment", "name", test.Name)

	labels := map[string]string{"app": "deploy"}
	return Outputs{
		Deployments: parameters.Deployments{
			Items: []appsv1.Deployment{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.Name,
					Namespace: test.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: pointer.Int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:   test.Name,
							Labels: labels,
						},
						Spec: test.Spec.PodSpec,
					}},
			}},
		},
	},
	v1.TestPhaseProcessing,
	&v1.TestStatusInfo{TimeStarted: time.Now()},
	nil
}
EOF

echo "Writing Processing worker..."

cat > pkg/workers/processing/worker.go <<EOF
package processing

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/test/e2e/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

func (w *Worker) Sync(ctx context.Context, test *v1.Test, inputs Inputs) (v1.TestPhase, *v1.TestStatusInfo, error) {
	w.Logger.Info("checking deployment metrics", "deployment", test.Name, "threshold", test.Spec.SuccessThreshold)

	val, err := inputs.Metrics.GetRequestSuccessRate(test.Name, test.Namespace, "5s")
	if err != nil {
		return "", nil, err
	}

	status := test.Status.TestStatusInfo
	status.Results = append(status.Results, val)

	phase := v1.TestPhaseProcessing
	switch {
	case val < test.Spec.SuccessThreshold:
		phase = v1.TestPhaseFailed
	case time.Now().Sub(test.Status.TimeStarted) >= test.Spec.Duration:
		phase = v1.TestPhasePassed
	}

	return phase, &status, nil
}

EOF
