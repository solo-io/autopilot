---
title: "autopilot deploy"
weight: 5
---
## autopilot deploy

Deploys the Operator with the provided image to Kubernetes

### Synopsis




```
autopilot deploy <image> [flags]
```

### Options

```
  -c, --cluster-scoped             Deploy the operator as a cluster-wide operator. This is required to provide the operator with the ClusterRole required to read and write to other namespaces (default true)
  -d, --deletepods                 Delete existing pods after pushing images (to force Kubernetes to pull the newly pushed image)
  -h, --help                       help for deploy
  -n, --namespace string           Namespace to which to deploy the operator
  -p, --push docker push <image>   Push the operator image before deploying. Use in place of docker push <image>.
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [autopilot](../autopilot)	 - An SDK for building Service Mesh Operators with ease

