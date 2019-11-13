#!/usr/bin/env bash


function k {
    kubectl --namespace e2etest $@
}

k port-forward deployment/hey 15000&

sleep 2
curl 'localhost:15000/logging?level=debug' -XPOST

kill %1

k logs -l app=example -c istio-proxy -f