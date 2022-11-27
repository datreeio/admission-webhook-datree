
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
#   BUILD CODE  #
#################
_builder:
	go build -o ${BUILD_ARGS_OUTPUT} -tags ${BUILD_ARGS_ENV} -ldflags=$(LD_FLAGS) $(BUILD_ARGS_DIR)
	
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



#################
#      TEST     #
#################

test:
	DATREE_ENFORCE="true" go test ./...


##################
#   BUILD IMAGE  #
##################
_image_builder:
	docker build -t ${BUILD_ARGS_OUTPUT} -f $(BUILD_ARGS_DIR)/Dockerfile . --build-arg BUILD_ENVIRONMENT=${BUILD_ARGS_ENV}
	
build-image-cert-generator-%:
	 $(MAKE) _image_builder \
  		-e BUILD_ARGS_DIR=$(CERT_GENERATOR_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="cert-generator"

build-image-init-webhook-%:
	 $(MAKE) _image_builder \
  		-e BUILD_ARGS_DIR=$(INIT_WEBHOOK_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="init-webhook"

build-image-webhook-server-%:
	 $(MAKE) _image_builder \
  		-e BUILD_ARGS_DIR=$(WEBHOOK_SERVER_DIR) \
  		-e BUILD_ARGS_ENV="$*"	 \
		-e BUILD_ARGS_OUTPUT="webhook-server"		

#################
#      DEPLOY   #
#################

.PHONY: deploy-in-minikube
deploy-in-minikube:
	bash ./scripts/deploy-in-minikube.sh

.PHONY: run-in-minikube
run-in-minikube:
	bash ./scripts/run-in-minikube.sh

.PHONY: test-in-minikube
test-in-minikube:
	bash ./scripts/test-in-minikube.sh

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

# to be continued...

deploy-datree-awsmp-rc:
# helm lint chart 
	helm lint ./charts/datree-admission-webhook-awsmp

# docker build binaries images
	$(MAKE) build-image-webhook-server-awsmp
	${MAKE} build-image-init-webhook-awsmp
	${MAKE} build-image-cert-generator-awsmp

# docker login
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com

# docker push image to ECR
	IMAGE_VERSION=$(yq '.initContainer.tag' ./charts/datree-admission-webhook-awsmp/values.yaml)
	IMAGE_VERSION=$(shell echo $${IMAGE_VERSION} | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)
	yq eval '.initContainer.tag = "${IMAGE_VERSION}"' -i ./charts/datree-admission-webhook-awsmp/values.yaml
#	docker tag webhook-server:latest 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-cert-generator:${IMAGE_VERSION}
	docker tag cert-generator:latest localhost:5000/cert-generator:${IMAGE_VERSION}
#	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-cert-generator:${IMAGE_VERSION}

	IMAGE_VERSION=$(yq '.imageWebhook.tag' ./charts/datree-admission-webhook-awsmp/values.yaml)
	IMAGE_VERSION=$(shell echo $${IMAGE_VERSION} | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)
	yq eval '.imageWebhook.tag = "${IMAGE_VERSION}"' -i ./charts/datree-admission-webhook-awsmp/values.yaml

	docker tag webhook-server:latest localhost:5000/webhook-server:${IMAGE_VERSION}
#	docker tag init-webhook:latest 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-webhook-init:${IMAGE_VERSION}
#	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-webhook-init:${IMAGE_VERSION}

	IMAGE_VERSION=$(yq '.image.tag' ./charts/datree-admission-webhook-awsmp/values.yaml)
	IMAGE_VERSION=$(shell echo $${IMAGE_VERSION} | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)
	yq eval '.image.tag = "${IMAGE_VERSION}"' -i ./charts/datree-admission-webhook-awsmp/values.yaml
	
	docker tag init-webhook:latest localhost:5000/init-webhook:${IMAGE_VERSION}
#	docker tag webhook-server:latest 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-admission-webhook:${IMAGE_VERSION}
#	docker push 709825985650.dkr.ecr.us-east-1.amazonaws.com/datree/datree-admission-webhook:${IMAGE_VERSION}

# bump version 
	HELM_CHART_VERSION=$(helm show chart ./charts/datree-admission-webhook-awsmp | grep version: | awk 'NR==2{print $2}')
	HELM_CHART_VERSION=$(shell echo $${HELM_CHART_VERSION} | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)

# helm package chart
	sed -i '' -e 's/version: .*/version: $(HELM_CHART_VERSION)/' ./charts/datree-admission-webhook-awsmp/Chart.yaml
	helm package ./charts/datree-admission-webhook-awsmp --version ${HELM_CHART_VERSION}
# helm push chart to ECR
	aws ecr get-login-password --region us-east-1 | helm login --username AWS --password-stdin 000000000000.dkr.ecr.us-east-1.amazonaws.com
#	helm push datree-admission-webhook-awsmp-${HELM_CHART_VERSION}.tgz 000000000000.dkr.ecr.us-east-1.amazonaws.com