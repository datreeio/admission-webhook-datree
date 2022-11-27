
#################
#    DEFAULTS   #
#################

CMD_DIR        := ./cmd
INIT_WEBHOOK_DIR    := $(CMD_DIR)/init-webhook
CERT_GENERATOR_DIR := $(CMD_DIR)/cert-generator
WEBHOOK_SERVER_DIR := $(CMD_DIR)/webhook-server
WEBHOOK_VERSION := 0.0.1
LD_FLAGS := "-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$(WEBHOOK_VERSION)"
BUILD_ARGS_ENV ?= staging
BUILD_ARGS_DIR ?= $(WEBHOOK_SERVER_DIR)
BUILD_ARGS_OUTPUT ?= webhook-server
		

#################
#      RUN      #
#################

_runner: 
	go run -tags ${BUILD_ARGS_ENV} -ldflags=$(LD_FLAGS) $(BUILD_ARGS_DIR)

run-cert-generator-%:
	 $(MAKE) _runner \
		-e BUILD_ARGS_DIR=$(CERT_GENERATOR_DIR) \
		-e BUILD_ARGS_ENV="$*"

run-init-webhook-%:
	 $(MAKE) _runner \
  		-e BUILD_ARGS_DIR=$(INIT_WEBHOOK_DIR) \
  		-e BUILD_ARGS_ENV="$*"

run-webhook-server-%:
	 $(MAKE) _runner \
  		-e BUILD_ARGS_DIR=$(WEBHOOK_SERVER_DIR) \
  		-e BUILD_ARGS_ENV="$*"

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
	
build-cert-generator-%:
	 $(MAKE) _builder \
  		-e BUILD_ARGS_DIR=$(CERT_GENERATOR_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="cert-generator"

build-init-webhook-%:
	 $(MAKE) _builder \
  		-e BUILD_ARGS_DIR=$(INIT_WEBHOOK_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="init-webhook"

build-webhook-server-%:
	 $(MAKE) _builder \
  		-e BUILD_ARGS_DIR=$(WEBHOOK_SERVER_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="webhook-server"		

#################
#    DEPLOY     #
#################

.PHONY: deploy-in-minikube
deploy-in-minikube:
	bash ./scripts/deploy-in-minikube.sh


deploy-webhook-server: 
	$(eval IMAGE_TAG := $(shell yq '.image.tag' ./charts/datree-admission-webhook-awsmp/values.yaml | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	$(MAKE) build-webhook-server-awsmp
	docker tag webhook-server 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-admission-webhook:$(IMAGE_TAG)
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-admission-webhook:${IMAGE_TAG}
	tag=${IMAGE_TAG} yq e --inplace '.image."tag" |= strenv(tag)' ./charts/datree-admission-webhook-awsmp/values.yaml


deploy-init-webhook:
	$(eval IMAGE_TAG := $(shell yq '.imageWebhook.tag' ./charts/datree-admission-webhook-awsmp/values.yaml | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	$(MAKE) build-init-webhook-awsmp
	docker tag init-webhook 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/init-webhook:$(IMAGE_TAG)
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/init-webhook:${IMAGE_TAG}
	tag=${IMAGE_TAG} yq e --inplace '.imageWebhook."tag" |= strenv(tag)' ./charts/datree-admission-webhook-awsmp/values.yaml

deploy-cert-generator:
	$(eval IMAGE_TAG := $(shell yq '.initContainer.tag' ./charts/datree-admission-webhook-awsmp/values.yaml | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	$(MAKE) build-cert-generator-awsmp
	docker tag cert-generator 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/cert-generator:$(IMAGE_TAG)
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/cert-generator:${IMAGE_TAG}
	tag=${IMAGE_TAG} yq e --inplace '.initContainer."tag" |= strenv(tag)' ./charts/datree-admission-webhook-awsmp/values.yaml

deploy-datree-awsmp:
	$(MAKE) deploy-cert-generator
	$(MAKE) deploy-init-webhook
	$(MAKE) deploy-webhook-server
	
# verify that the chart is valid
	helm lint ./charts/datree-admission-webhook-awsmp

# bump Helm Chart version 
	$(eval HELM_CHART_VERSION=$(shell yq '.version' ./charts/datree-admission-webhook-awsmp/Chart.yaml | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	version=${HELM_CHART_VERSION} yq e --inplace '."version" |= strenv(version)' ./charts/datree-admission-webhook-awsmp/Chart.yaml

# helm push chart to ECR
	helm package ./charts/datree-admission-webhook-awsmp
	aws ecr get-login-password --region us-east-1 | helm login --username AWS --password-stdin 000000000000.dkr.ecr.us-east-1.amazonaws.com
	helm push datree-admission-webhook-awsmp-${HELM_CHART_VERSION}.tgz 000000000000.dkr.ecr.us-east-1.amazonaws.com


#################
#    HELM     #
#################

helm-install-local-in-minikube:
	eval $(minikube docker-env) && \
	./scripts/build-docker-image.sh && \
	helm install -n datree datree-webhook ./charts/datree-admission-webhook --set datree.token="${DATREE_TOKEN}"

helm-upgrade-local:
	helm upgrade -n datree datree-webhook ./charts/datree-admission-webhook --reuse-values --set datree.enforce="true"

helm-uninstall:
	helm uninstall -n datree datree-webhook

helm-install-staging:
	helm install -n datree datree-webhook ./charts/datree-admission-webhook --set datree.token="${DATREE_TOKEN}" --set scan_job.image.repository="datree/scan-job-staging" \
	--set scan_job.image.tag="latest" --set image.repository="datree/webhook-staging" --set image.tag="latest"