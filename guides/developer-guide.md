# Developer Guide

This guide explains how to set up your environment for developing on admission-webhook-datree.  
This guide was written for macOS and Linux machines.

## For developing on a local server with thunder client (faster build)

### Prerequisites

- Go version 1.18
- Git
- **optional**: [gow](https://github.com/mitranim/gow#installation) (go file watcher)
- **optional**: VS Code + Thunder Client

### Running admission-webhook-datree as a local server
```
make start
```

### Alternatively, use watch mode for hot reload 🤩
```
make start-watch
```

### Then make requests using Thunder Client
- GET /health
- POST /validate (webhook-demo.yaml)

## For developing on minikube (slower build)

### Prerequisites

- Go version 1.18
- Git
- [Docker](https://docs.docker.com/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [minikube](https://kubernetes.io/docs/tasks/tools/#minikube)

### Background processes that need to run:
- Run `minikube start --extra-config=apiserver.enable-admission-plugins=ValidatingAdmissionWebhook`
- Run Docker daemon by opening Docker desktop

### Deploy to your local minikube
- Run `make run deploy-in-minikube` - this will build a docker image and deploy it to minikube
- check the webhook is deployed: `kubectl get pods -n datree`
- try to apply a demo file to the deployment: `kubectl apply -f ./scripts/webhook-demo.yaml`

### Remove from local minikube
```
./scripts/uninstall.sh
```

### Deploy and apply webhook-demo.yaml to minikube
```
make run-in-minikube
```

### Run a basic E2E test (Not currently in CI)
this will apply the webhook-demo.yaml file to 
minikube and compare the output to ./internal/fixtures/webhook-demo-expected-output.txt
```
make test-in-minikube
```

### Build the docker image locally
```
./scripts/build-docker-image.sh
```

### Deployment
**Important things to keep in mind when releasing to production**:

When uploading a new version you should run the github action and wait until a new build is uploaded to dockerhub.

After that when a new build was published successfully you should test it and open a new PR to update admission-webhook-datree.yaml image property.


**The release will fail in the following scenarios:**
* Release is performed and then immediately a new commit is pushed to main - will release a production tag with commit hash (instead of semantic version)
* Release is performed right after merging to main - will release a staging tag with semantic version (instead of commit hash)
* Release is performed twice in a row (without pushing a new commit) - will try to release the same docker tag again (and fail)

**When releasing a new version to production notice if cloudfront invalidation failed - if so re-run the failed release workflow**
