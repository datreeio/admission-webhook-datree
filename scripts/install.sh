#!/bin/sh

# Sets up the environment for the admission controller webhook in the active cluster.
# check that user have kubectl installed and openssl
# generate TLS keys
generate_keys () {
  printf "🔑 Generating TLS keys...\n"

  chmod 0700 "${keydir}"
  cd "${keydir}"

  cat >server.conf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
prompt = no
[req_distinguished_name]
CN = webhook-server.datree.svc
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = webhook-server.datree.svc
EOF

# Generate the CA cert and private key that is valid for 5 years
openssl req -nodes -new -x509 -days 1827 -keyout ca.key -out ca.crt -subj "/CN=Admission Controller Webhook Demo CA"
# Generate the private key for the webhook server
openssl genrsa -out webhook-server-tls.key 2048
# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key webhook-server-tls.key -subj "/CN=webhook-server.datree.svc" -config server.conf \
    | openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out webhook-server-tls.crt -extensions v3_req -extfile server.conf

cd -
}

verify_prerequisites () {
  local openssl_path
  local kubectl_path

  openssl_path="$(whereis openssl)"
  kubectl_path="$(which kubectl)"

  if [ -z "${openssl_path}" ] ;
  then
    printf '%s\n' "openssl doesn't exist, please install openssl"
    exit 1
  fi

  if [ -z "${kubectl_path}" ] ;
  then
    printf '%s\n' "kubectl doesn't exist, please install kubectl"
    exit 1
  fi
}

verify_datree_namespace_not_existing () {
  local namespace_exists
  namespace_exists="$(kubectl get namespace/datree --ignore-not-found)"

  if [ -n "${namespace_exists}" ] ;
    then
      printf '%s\n' "datree namespace already exists"
      exit 1
    fi
}

verify_webhook_resources_not_existing () {
  local validating_webhook_exists
  validating_webhook_exists="$(kubectl get validatingwebhookconfiguration.admissionregistration.k8s.io/webhook-datree --ignore-not-found)"

  if [ -n "${validating_webhook_exists}" ] ;
    then
      printf '%s\n' "datree validating webhook already exists"
      exit 1
    fi
}

verify_datree_namespace_not_existing

verify_webhook_resources_not_existing

verify_prerequisites

set -eo pipefail

# Create Temporary directory for TLS keys
keydir="$(mktemp -d)"

# Generate keys into a temporary directory.
generate_keys

basedir="$(pwd)/deployment"

# Create the `datree` namespace. This cannot be part of the YAML file as we first need to create the TLS secret,
# which would fail otherwise.
printf "\n🏠 Creating datree namespace...\n"
kubectl create namespace datree

# Label datree namespace to avoid deadlocks in self hosted webhooks
#  https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#avoiding-deadlocks-in-self-hosted-webhooks
kubectl label namespaces datree admission.datree/validate=skip

# label kube-system namespace to avoid operating on the kube-system namespace
# https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#avoiding-operating-on-the-kube-system-namespace
kubectl label namespaces kube-system admission.datree/validate=skip

# Override DATREE_TOKEN env
if [ -z "$DATREE_TOKEN" ] ;
then
    echo
    echo =====================================
    echo === Finish setting up the webhook ===
    echo =====================================
    echo "👉 Insert token (available at https://app.datree.io/settings/token-management)"
    echo "ℹ️  The token is used to connect the webhook with your account."
    read datree_token
else
    datree_token=$DATREE_TOKEN
fi


# Create the TLS secret for the generated keys.
kubectl -n datree create secret tls webhook-server-tls \
    --cert "${keydir}/webhook-server-tls.crt" \
    --key "${keydir}/webhook-server-tls.key"

printf "\n🔗 Creating webhook resources...\n"

# Read the PEM-encoded CA certificate, base64 encode it, and replace the `${CA_PEM_B64}` placeholder in the YAML
# template with it. Then, create the Kubernetes resources.
ca_pem_b64="$(openssl base64 -A <"${keydir}/ca.crt")"
curl "https://raw.githubusercontent.com/datreeio/webhook-datree/main/deployment/webhook-datree.yaml" |  sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' \
    | sed 's@${DATREE_TOKEN}@'"$datree_token"'@g' \
    | kubectl create -f -

# Delete the key directory to prevent abuse (DO NOT USE THESE KEYS ANYWHERE ELSE).
rm -rf "${keydir}"

deploymentRevision=$(kubectl get deployment webhook-server -n datree -o "jsonpath={.metadata.annotations['deployment\.kubernetes\.io/revision']}")
kubectl rollout status deployment webhook-server -n datree --revision="$deploymentRevision"

printf "\n🎉 DONE! The webhook server is now deployed and configured\n"
