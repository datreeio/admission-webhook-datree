bash ./scripts/uninstall.sh

eval $(minikube docker-env)

bash ./scripts/build-docker-image.sh
build_exit_code=$?
if [ $build_exit_code != 0 ]; then
  exit $build_exit_code
fi

bash ./scripts/installation.sh

sleep 3 # wait for pods to be ready
kubectl get pods -n datree
