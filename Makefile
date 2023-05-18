start-watch:
	gow run -tags $(or $(datree_build_env),staging) -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=0.0.1" main.go

change-ping-uninstall-url-to-staging: 
	sed -i '' 's|https://gateway\.datree\.io/cli/cluster/uninstall|https://gateway.staging.datree.io/cli/cluster/uninstall|' charts/datree-admission-webhook/templates/namespace-post-delete.yaml

change-ping-uninstall-url-to-production:
	sed -i '' 's|https://gateway\.staging\.datree\.io/cli/cluster/uninstall|https://gateway.datree.io/cli/cluster/uninstall|' charts/datree-admission-webhook/templates/namespace-post-delete.yaml

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


helm-install-local-in-minikube:
	make change-ping-uninstall-url-to-staging && \
	eval $(minikube docker-env) && \
	./scripts/build-docker-image.sh && \
	helm install -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	-f ./internal/fixtures/values.dev.yaml \
	--set datree.token="${DATREE_TOKEN}" \
	--set scanJob.ttlSecondschange-ping-uninstall-url-to-productionFinished=100 \
	--debug && \
	make change-ping-uninstall-url-to-production

helm-upgrade-local:
	make change-ping-uninstall-url-to-staging && \
	eval $(minikube docker-env) && \
	helm upgrade -n datree datree-webhook ./charts/datree-admission-webhook \
	-f ./internal/fixtures/values.dev.yaml \
	--set datree.token="${DATREE_TOKEN}" \
	--set scanJob.ttlSecondschange-ping-uninstall-url-to-productionFinished=100 \
	--debug && \
	make change-ping-uninstall-url-to-production

helm-uninstall:
	helm uninstall -n datree datree-webhook

helm-install-staging:
	make change-ping-uninstall-url-to-staging && \
	helm install -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	--set datree.token="${DATREE_TOKEN}" \
	--set datree.clusterName="minikube" \
	--set datree.policy="Starter" \
	--set clusterScanner.image.repository="datree/cluster-scanner-staging" \
	--set clusterScanner.image.tag="latest" \
	--set image.repository="datree/webhook-staging" \
	--set image.tag="latest" \
	--debug && \
	make change-ping-uninstall-url-to-production

helm-template-staging:
	make change-ping-uninstall-url-to-staging && \
	helm template -n datree datree-webhook ./charts/datree-admission-webhook \
	--create-namespace \
	--set datree.token="${DATREE_TOKEN}" \
	--set datree.clusterName="minikube" \
	--set datree.policy="Starter" \
	--set clusterScanner.image.repository="datree/cluster-scanner-staging" \
	--set clusterScanner.image.tag="latest" \
	--set image.repository="datree/webhook-staging" \
	--set image.tag="latest" \
	--debug && \
	make change-ping-uninstall-url-to-production

# in order to run the command, first install helm-docs by running: "brew install norwoodj/tap/helm-docs"
# https://github.com/norwoodj/helm-docs
generate-helm-docs:
	helm-docs \
	--sort-values-order=file \
	--output-file ./README.md \
	--template-files=./charts/datree-admission-webhook/README.md.gotmpl \
	&& \
	helm-docs \
	--sort-values-order=file \
	--output-file ../../README.md \
	--template-files=./README.md.gotmpl
