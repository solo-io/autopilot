# HelloWorld with Autopilot - AutoRouter

This guide will walk you through building a minimalistic "HelloWorld" Operator with Autopilot. 
This operator will automatically generate ingress routes for discovered deployments.

The Operator will process CRDs of kind `AutoRoute` in API group `example.io/v1`.

The `AutoRoute` CRD will have two phases: `Updating`, and `Ready`.

In the `Updating` phase, the `AutoRoute` ensure Istio provides ingress routes to any deployments matching the `spec.selector` on the `AutoRoute`. 
 The `AutoRoute` will ensure that Istio has an ingress route for each selected deployment. Once the routes are created, the 
 state will change to the `Ready` phase to mark that the `AutoRoute` is ready to serve traffic.

In the `Ready` phase, the Operator watch for new deployments. If the list of deployments falls out of sync with the configured routes,
the `AutoRoute` will go back to the `Updating` phase.

The purpose of this application is to demonstrate basic functionality in Autopilot 

# Tutorial

## Install the Autopilot CLI

Download the latest release of the Autopilot CLI from the GitHub releases page: https://github.com/solo-io/autopilot/releases

or install from the terminal:

```bash
curl -sL https://run.solo.io/autopilot/install | sh
export PATH=$HOME/.autopilot/bin:$PATH
```

## Initialize the `hello` operator

Run the following to create the operator directory in your workspace. Can be inside or outside `GOPATH` (uses Go modules):

```bash
ap init hello --kind Hello --group examples.io --version v1
```

> Note: if running outside your $GOPATH, you will need to add the flag
> `--module hello.io` 

This should produce the output:

```
INFO[0000] Creating Project Config: kind:"Hello" apiVersion:"examples.io/v1" operatorName:"hello-operator" phases:<name:"Initializing" description:"Hello has begun initializing" initial:true outputs:"virtualservices" > phases:<name:"Processing" description:"Hello has begun processing" inputs:"metrics" outputs:"virtualservices" > phases:<name:"Finished" description:"Hello has finished" final:true >
go: creating new go.mod: module hello.io
```

If you run `tree` on the newly created `hello` directory, you'll see the following files were created:

```
.
└── hello
    ├── autopilot-operator.yaml
    ├── autopilot.yaml
    └── go.mod
``` 

* `autopilot.yaml`: the project configuration file. Use this to generate the operator code, build, and deployment files.
* `autopilot-operator.yaml`: the runtime configuration for the Operator.
* `go.mod` - a generated `go.mod` file with some `replace` statements necessary to import Autopilot and its dependencies.


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
