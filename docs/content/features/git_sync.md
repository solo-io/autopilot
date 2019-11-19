---
title: "GitOps with GitSync"
weight: 2
---

The `GitSync` feature of Autopilot allows Operators to submit automated configuration updates as `git` commits, rather than directly updating cluster config. This can be used as part of a manual approval process as well as part of an automated [GitOps](https://www.weave.works/technologies/gitops/) pipeline.

## Motivation

Many CI/CD systems today take advantage of [GitOps](https://www.weave.works/technologies/gitops/), a pattern that uses `git` repositories as authoritative sources of truth. `git` provides workflows for declarative configuration (via branches) as well as approval processes to manage that configuration (via [pull requests](https://www.weave.works/blog/gitops-operations-by-pull-request)).

The `GitSync` feature of Autopilot allows Operators to submit automated configuration updates as `git` commits, to be used as part of a `GitOps` workflow or manual approval process.

## Usage

To enable `GitSync`, provide the following in your `autopilot-operator.yaml`:

```yaml
#...

# add the following gitSync struct: 
gitSync:
  # make sure enabled is set to true 
  enabled: true
  # e.g. solo-io for https://github.com/solo-io/autopilot
  org: org
  # e.g. dev
  branch: branch
  # the name of a secret and its key containing a valid GITHUB_TOKEN
  # e.g. mygitcredentials.github_token
  # where mygitcredentials is a secret in the operator namespace:
  # 
  #  kind: Secret
  #  metadata:
  #    name: mygitcredentials
  #    namespace: <operator namespace>
  #  type: Opaque
  #  data:
  #    github_token: <github token data>
  credentials: secretName.keyName
```

Once `GitSync` is enabled, your *workers* can now set `outputs.SyncToGit = true`, like so:

{{< highlight go "hl_lines=19-21" >}}
func (w *Worker) Sync(ctx context.Context, canary *v1.CanaryDeployment, inputs Inputs) (Outputs, v1.CanaryDeploymentPhase, *v1.CanaryDeploymentStatusInfo, error) {
	
	// construct outputs
	...
	
	outputs := Outputs{
        Deployments: parameters.Deployments{
            Items: []appsv1.Deployment{
                canaryDeployment,
            },
        },
        VirtualServices: parameters.VirtualServices{
            Items: []v1alpha3.VirtualService{
                virtualService,
            },
        },
    }

    // submit the above outputs to a Git branch
    // rather than kubernetes
    outputs.SyncToGit = true

	return outputs, v1.CanaryDeploymentPhaseWaiting, &status, nil
}
{{< /highlight >}}

The next time this worker syncs, a commit containing all the changed resources will be pushed to the configured git branch:

