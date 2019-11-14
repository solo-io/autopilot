# HelloWorld with Autopilot - AutoRouter

This guide will walk you through building a minimalistic "HelloWorld" Operator with Autopilot. 
This operator will automatically generate ingress routes for discovered deployments.

The Operator will process CRDs of kind `AutoRoute` in API group `example.io/v1`.

The `AutoRoute` CRD will have three phases: `Initializing`, `Syncing`, and `Ready`.

In the `Initializing` phase, the `AutoRoute` will be marked as `Syncing` to confirm processing has begun.

In the `Syncing` phase, the `AutoRoute` ensure Istio provides ingress routes to any deployments matching the `spec.selector` on the `AutoRoute`. 
 The `AutoRoute` will ensure that Istio has an ingress route for each selected deployment. Once the routes are created, the 
 state will change to the `Ready` phase to mark that the `AutoRoute` is ready to serve traffic.

In the `Ready` phase, the Operator watch for new deployments. If the list of deployments falls out of sync with the configured routes,
the `AutoRoute` will go back to the `Syncing` phase.

The purpose of this application is to demonstrate basic functionality in Autopilot, as well as provide an example of how to get
started.

# Tutorial

## Install the Autopilot CLI

Download the latest release of the Autopilot CLI from the GitHub releases page: https://github.com/solo-io/autopilot/releases

or install from the terminal:

```bash
curl -sL https://run.solo.io/autopilot/install | sh
export PATH=$HOME/.autopilot/bin:$PATH
```

## Initialize the `autorouter` operator

Run the following to create the operator directory in your workspace. Can be inside or outside `GOPATH` (uses Go modules):

```bash
ap init autorouter --kind AutoRoute --group examples.io --version v1 
```

> Note: if running outside your $GOPATH, you will need to add the flag
> `--module autorouter.examples.io` 

This should produce the output:

```
INFO[0000] Creating Project Config: kind:"AutoRoute" apiVersion:"examples.io/v1" operatorName:"autoroute-operator" phases:<name:"Initializing" description:"AutoRoute has begun initializing" initial:true outputs:"virtualservices" > phases:<name:"Processing" description:"AutoRoute has begun processing" inputs:"metrics" outputs:"virtualservices" > phases:<name:"Finished" description:"AutoRoute has finished" final:true >
go: creating new go.mod: module autorouter.examples.io
```

If you run `tree` on the newly created `autorouter` directory, you'll see the following files were created:

```
.
└── autorouter
    ├── autopilot-operator.yaml
    ├── autopilot.yaml
    └── go.mod
``` 

* `autopilot.yaml`: the project configuration file. Use this to generate the operator code, build, and deployment files.
* `autopilot-operator.yaml`: the runtime configuration for the Operator.
* `go.mod` - a generated `go.mod` file with some `replace` statements necessary to import Autopilot and its dependencies.

## Update the `autopilot.yaml` file

Let's take a look at the generated `autopilot.yaml` file:

```yaml
apiVersion: examples.io/v1
kind: AutoRoute
operatorName: autoroute-operator
phases:
- description: AutoRoute has begun initializing
  initial: true
  name: Initializing
  outputs:
  - virtualservices
- description: AutoRoute has begun processing
  inputs:
  - metrics
  name: Processing
  outputs:
  - virtualservices
- description: AutoRoute has finished
  final: true
  name: Finished
```

This is a default template and meant to be edited. Let's replace the contents with the following:

```yaml
apiVersion: examples.io/v1
kind: AutoRoute
operatorName: autoroute-operator
phases:

- description: AutoRoute is pending processing
  # initial: true indicates this state is the initial state of the CRD
  # initial can only be true for one phase
  initial: true
  name: Initializing
  
- description: AutoRoute is syncing with deployments in the cluster
  name: Syncing
  inputs:
  - deployments
  outputs:
  - services
  - virtualservices
  - gateways

- description: AutoRoute is in-sync and ready to serve traffic
  inputs:
  - deployments
  name: Ready
```

Our `autopilot.yaml` specifies each of the *phases* the `AutoRoute` can be in, as well as the 
inputs and outputs used by the system during the given phase.

We can now generate our project structure by issuing the following command:

```bash
cd autorouter
ap generate
```

Next we'll configure our `autopilot.yaml` file to generate 

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
