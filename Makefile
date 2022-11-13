start-watch:
	gow run -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" main.go

start:
	go run -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" main.go
start-dev:
	make datree_build_env=dev start
start-staging:
	make datree_build_env=staging start
start-production:
	make datree_build_env=main start

build:
	go build -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" -o webhook-datree
build-dev:
	make datree_build_env=dev build
build-staging:
	make datree_build_env=staging build
build-production:
	make datree_build_env=main build

test:
	DATREE_ENFORCE="true" go test ./...

deploy-in-minikube:
	bash ./scripts/deploy-in-minikube.sh
run-in-minikube:
	bash ./scripts/run-in-minikube.sh
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