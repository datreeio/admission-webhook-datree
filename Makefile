
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
