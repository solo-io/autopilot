# HelloWorld with AutoPilot

This guide will walk you through building a minimalistic "HelloWorld" Operator with AutoPilot.

The Operator will process CRDs of kind `Hello` in API group `example.io/v1`.

The `Hello` CRD will have three phases: `Initializing`, `Replying`, `Resting`.

In the `Initializing` phase, the Operator will create the desired configmaps. The next phase is `Replying`.

In the `Replying` phase, the Operator will set the target confgimaps' data to `hello: world`. The next phase is `Resting`.

In the `Resting` phase, the Operator will monitor the configmaps. If they get out-of-sync, 
the Operator will send us back to the `Initializing` and `Replying` stages. Otherwise, the `Restign` phase will be maintained.

The `Hello` CRD spec specifies the name of one or more kube ConfigMaps. The `Hello` operator will ensure 
these ConfigMaps exist and contain the data `hello: world`.

# Tutorial

## Build the Autopilot CLI

Currently, binary releases of the Autopilot CLi `ap` are not published. To build `ap` locally:

```
git clone https://github.com/solo-io/autopilot
cd autopilot
go get ./...

```
 
> note: tested with `go` 1.13





## Initialize the `hello` operator
## Update the `autopilot.yaml` file
## Generate the code
## Test the deployment
## Update the API Spec
## Re-Generate the code
## Write an example CRD
## Update the Initializing Worker
## Update the Replying Worker
## Update the Resting Worker
## Redeploy
## Try it out!
