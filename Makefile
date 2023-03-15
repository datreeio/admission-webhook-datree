#################
#    DEFAULTS   #
#################

WEBHOOK_SERVER_DIR := ./
LD_FLAGS := "-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1"
BUILD_ARGS_ENV ?= staging # default environment
BUILD_ARGS_DIR ?= $(WEBHOOK_SERVER_DIR)
BUILD_ARGS_OUTPUT ?= webhook-datree # default output
		

####################
#       RUN        #
####################

.PHONY: run-in-minikube
run-in-minikube:
	bash ./scripts/run-in-minikube.sh

#################
#      TEST     #
#################

test:
	DATREE_ENFORCE="true" go test ./...

.PHONY: test-in-minikube
test-in-minikube:
	bash ./scripts/test-in-minikube.sh


##################
#   BUILD  #
##################
_builder:
	docker build -t ${BUILD_ARGS_OUTPUT} -f $(BUILD_ARGS_DIR)/Dockerfile . --build-arg BUILD_ENVIRONMENT=${BUILD_ARGS_ENV}

# Build the webhook server image, using the environment specified in BUILD_ARGS_ENV (default: staging)
build-webhook-server-staging:
	 $(MAKE) _builder \
  		-e BUILD_ARGS_DIR=$(WEBHOOK_SERVER_DIR) \
  		-e BUILD_ARGS_ENV="staging"	 \
		-e BUILD_ARGS_OUTPUT="$(BUILD_ARGS_OUTPUT)"	

build-webhook-server-production:
	 $(MAKE) _builder \
  		-e BUILD_ARGS_DIR=$(WEBHOOK_SERVER_DIR) \
  		-e BUILD_ARGS_ENV="main"	 \
		-e BUILD_ARGS_OUTPUT="$(BUILD_ARGS_OUTPUT)"	

#################
#    DEPLOY     #
#################

.PHONY: deploy-in-minikube
deploy-in-minikube:
	bash ./scripts/deploy-in-minikube.sh


#################
#    DEBUG      #
#################

build:
	go build -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" -o webhook-datree
build-dev:
	make datree_build_env=dev build
build-staging:
	make datree_build_env=staging build
build-production:
	make datree_build_env=main build

start:
	go run -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" main.go

start-watch:
	go run -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" main.go

start-dev:
	make datree_build_env=dev start
start-staging:
	make datree_build_env=staging start
start-production:
	make datree_build_env=main start


#################
#    	HELM    #
#################

helm-install-local-in-minikube:
	eval $(minikube docker-env) && \
	./scripts/build-docker-image.sh && \
	helm install -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	--set datree.token="${DATREE_TOKEN}" \
	--set datree.clusterName="minikube" \
	--set scanJob.image.repository="datree/scan-job-staging" \
	--set scanJob.image.tag="latest" \
	--set image.repository="webhook-server" \
	--set image.pullPolicy="Never" \
	--set image.tag="latest" \
	--set replicaCount=1 \
	--set scanJob.ttlSecondsAfterFinished=100 \
	--debug

helm-upgrade-local:
	helm upgrade -n datree datree-webhook ./charts/datree-admission-webhook --reuse-values --set datree.enforce="true"

helm-uninstall:
	helm uninstall -n datree datree-webhook

helm-install-staging:
	helm install -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	--set datree.token="${DATREE_TOKEN}" \
	--set scanJob.image.repository="datree/scan-job-staging" \
	--set scanJob.image.tag="latest" \
	--set image.repository="datree/webhook-staging" \
	--set image.tag="latest" \
	--debug

helm-template-staging:
	helm template -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	--set datree.token="${DATREE_TOKEN}" \
	--set scanJob.image.repository="datree/scan-job-staging" \
	--set scanJob.image.tag="latest" \
	--set image.repository="datree/webhook-staging" \
	--debug
