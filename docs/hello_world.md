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

This should produce the output:

```
INFO[0000] Creating Project Config: kind:"Hello" apiVersion:"examples.io/v1" operatorName:"hello-operator" phases:<name:"Initializing" description:"Hello has begun initializing" initial:true outputs:"virtualservices" > phases:<name:"Processing" description:"Hello has begun processing" inputs:"metrics" outputs:"virtualservices" > phases:<name:"Finished" description:"Hello has finished" final:true >
```

If we run `ls`

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
