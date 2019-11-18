---
title: "Getting Started Pt 3"
weight: 3
description: Redeploy and test our operator 
---

## Getting Started with Autopilot Part 3 - Deploying the Operator

In part 3 of the Getting Started tutorial for AutoPilot, we'll deploy and test our changes from [part 2](getting_started_2.md).

To see the completed code for this tutorial, check out https://github.com/solo-io/autorouter

## About 



# Tutorial

## Prerequisites

- Completed parts [one]({{< versioned_link_path fromRoot="/tutorial_code/getting_started_1">}}) and [two]({{< versioned_link_path fromRoot="/tutorial_code/getting_started_2">}}) of this tutorial series.

## Build the Operator

Just like we did in part 1, let's build and deploy the operator. Notice we deploy with the `-d` flag which tells 
Kubernetes to delete the existing operator pod (to trigger a fresh image pull)

```bash
ap build <image>
ap deploy <image> -p -d
```

We should see that our pod is running:

```bash
kubectl get pod -n autoroute-operator
```

```bash
NAME                                 READY   STATUS    RESTARTS   AGE
autoroute-operator-ccbd545c6-bb2s7   1/1     Running   0          1m
```

If you've still got the `AutoRoute` we deployed in [part 1](getting_started_1.md#test-the-operator-with-an-autoroute-resource), you'll notice that the 
pod is no longer in a `CrashLoop`. This is a good start. If you don't have the "example" `AutoRoute`, deploy it again:

```bash
kubectl create ns example 
kubectl apply -n example -f deploy/autoroute_example.yaml
```

Let's take a look at our logs again: 

```bash
ap logs
```

```json
{"level":"info","ts":1573835381.23586,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":":9091"}
{"level":"info","ts":1573835381.2361934,"msg":"Registering watch for primary resource AutoRoute"}
{"level":"info","ts":1573835381.2362459,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573835381.2364664,"msg":"Registering watch for output resource Services"}
{"level":"info","ts":1573835381.236519,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573835381.2366421,"msg":"Registering watch for output resource VirtualServices"}
{"level":"info","ts":1573835381.2366784,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573835381.2367713,"msg":"Registering watch for output resource Gateways"}
{"level":"info","ts":1573835381.2368004,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573835381.2370074,"logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
{"level":"info","ts":1573835397.8720093,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"autoRoute-controller"}
{"level":"info","ts":1573835397.9722834,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"autoRoute-controller","worker count":1}
{"level":"info","ts":1573835743.5527482,"msg":"Syncing AutoRoute in phase Initializing","autoRoute":"example.example","phase":"","name":"example"}
{"level":"info","ts":1573835743.55888,"msg":"Syncing AutoRoute in phase Syncing","autoRoute":"example.example","phase":"Syncing","name":"example"}
{"level":"info","ts":1573835743.6597307,"msg":"cycle through each deployment and check that the labels match our selector","autoRoute":"example.example","phase":"Syncing"}
{"level":"info","ts":1573835743.6598568,"msg":"ensuring the gateway, services and virtual service outputs are created","autoRoute":"example.example","phase":"Syncing","status":{"syncedDeployments":null},"gateway":"example","virtual services":0,"kube services":0}
{"level":"info","ts":1573835743.6726131,"msg":"Updating status of primary resource","autoRoute":"example.example","phase":"Syncing"}
{"level":"info","ts":1573835743.6782513,"msg":"Syncing AutoRoute in phase Ready","autoRoute":"example.example","phase":"Ready","name":"example"}
{"level":"info","ts":1573835743.67874,"msg":"cycle through each deployment and check that the labels match our selector","autoRoute":"example.example","phase":"Ready"}

```

Now we see that our operator processed the `AutoRoute`. The `AutoRoute` was synced and is now in state ready:

```bash
kubectl get autoroute -n example -oyaml example
```

```yaml
apiVersion: examples.io/v1
kind: AutoRoute
metadata:
  creationTimestamp: "2019-11-15T16:35:43Z"
  generation: 1
  labels:
    app: autoroute-operator
    app.kubernetes.io/name: autoroute-operator
  name: example
  namespace: example
  resourceVersion: "394326"
  selfLink: /apis/examples.io/v1/namespaces/example/autoroutes/example
  uid: f7cba80e-07c5-11ea-beb2-42010a8e0142
status:
  observedGeneration: 1
  phase: Ready
  syncedDeployments: null
```

We should also see that our operator created a Gateway:

```bash
kubectl get gateways.networking.istio.io -n example
```

```
NAME      AGE
example   11m
```

We expect no Virtual Services or Kube Services to be created yet. Let's create a few deployments to see them get spun up:

```bash
for i in 1 2 3; do 
cat <<EOF | kubectl apply -n example -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin-${i}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin-${i}
  template:
    metadata:
      labels:
        app: httpbin-${i}
    spec:
      containers:
        - image: docker.io/kennethreitz/httpbin
          imagePullPolicy: IfNotPresent
          name: httpbin
          ports:
            - containerPort: 80
EOF
done
```

```
deployment.apps/httpbin-1 created
deployment.apps/httpbin-2 created
deployment.apps/httpbin-3 created
```


If everything worked, we should immediately be able  to hit these deployments with traffic! Let's try it out.

First, we need the IP:PORT of the Istio Ingress. The way we reach this depends on your environment setup

* In environments using an External LoadBalancer:

    ```bash
    export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
    ```

* In Environments using a NodePort:

    ```bash
    export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
    ```
    
    Getting the Ingress IP depends on the cluster provider:       

    * GKE:
    
    ```bash
    export INGRESS_HOST=<workerNodeAddress>
    ```
    
      You need to create firewall rules to allow the TCP traffic to the ingressgateway service's ports. Run the following commands to allow the traffic for the HTTP port, the secure port (HTTPS) or both:
    
    ```bash
    gcloud compute firewall-rules create allow-gateway-http --allow tcp:$INGRESS_PORT $ gcloud compute firewall-rules create allow-gateway-https --allow tcp:$SECURE_INGRESS_PORT 
    ```
    
    * Minikube:
    
    ```bash
    export INGRESS_HOST=$(minikube ip) 
    ```
    
    * Docker For Desktop:
        
    ```bash
    export INGRESS_HOST=127.0.0.1 
    ```
    
    * Other environments (e.g., IBM Cloud Private etc):
    
    ```bash
    export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}') 
    ```

Finally, send some requests using `curl`:


```bash
curl -I -HHost:httpbin-1.example http://$INGRESS_HOST:$INGRESS_PORT/status/200
curl -I -HHost:httpbin-2.example http://$INGRESS_HOST:$INGRESS_PORT/status/200
curl -I -HHost:httpbin-3.example http://$INGRESS_HOST:$INGRESS_PORT/status/200
```

All 3 of our services should reply with `HTTP/1.1 200 OK`:

```
HTTP/1.1 200 OK
server: istio-envoy
date: Fri, 15 Nov 2019 17:34:05 GMT
content-type: text/html; charset=utf-8
access-control-allow-origin: *
access-control-allow-credentials: true
content-length: 0
x-envoy-upstream-service-time: 1
```

Awesome! We've just automated our Istio ingress with **Autopilot**!

We can see that the `AutoRoute` automatically created some Istio Virtual Serivces:

```bash
kubectl get virtualservices.networking.istio.io -n example
```

```
NAME                GATEWAYS    HOSTS                 AGE
httpbin-1.example   [example]   [httpbin-1.example]   2m
httpbin-2.example   [example]   [httpbin-2.example]   2m
httpbin-3.example   [example]   [httpbin-3.example]   2m
```

As well as standard Kube services:

```bash
kubectl get svc -n example
```

```
NAME        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
httpbin-1   ClusterIP   10.3.252.230   <none>        80/TCP    4m51s
httpbin-2   ClusterIP   10.3.243.217   <none>        80/TCP    4m51s
httpbin-3   ClusterIP   10.3.253.81    <none>        80/TCP    4m51s
```

## Tearing down

To tear down the example environment:

```bash
kubectl delete ns example
kubectl delete ns autorouter-operator
kubectl delete crd autoroutes.examples.io
```

# Summary

This guide provides a demonstration on how Autopilot can be used to automate Mesh features. For a more robust, real-world example,
see the [Autopilot Canary Project](https://github.com/solo-io/autopilot/tree/master/test/e2e), which is used as an automated end-to-end test for Autopilot. 

Please submit questions and feedback to [the solo.io slack channel](https://slack.solo.io/), or [open an issue on GitHub](https://github.com/solo-io/autopilot).
