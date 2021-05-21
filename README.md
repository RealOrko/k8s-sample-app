# k8s Sample App

A sample application for deploying into AWS EKS where the health check can be setup to fail using env vars.

## Assumptions

This repository assumes you are already up and running with k8s, specifically AWS EKS. You already have [kubectl](https://kubernetes.io/docs/tasks/tools/) installed and can connect. 

## Pre-requisites

Please install the following:

  - [direnv](https://direnv.net/docs/installation.html)

Run `direnv allow` in the checkout directory. 

## Cloning the repo

To clone and activate the repo, please run the following:

```
git clone git@github.com:RealOrko/k8s-sample-app.git && cd k8s-debugging && direnv allow
```

## Up and running

  - `kdeploy`: To deploy the app container in your k8s cluster. 
  - `kdestroy`: To destroy the app container in your k8s cluster. 


## SSH into pod

To SSH into the pod for running diagnostics please run: 

  - `kshell`: This will give you a bash shell into the container. 
