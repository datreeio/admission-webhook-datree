bash ./scripts/uninstall.sh

eval $(minikube docker-env)

bash ./scripts/build-docker-image.sh
build_exit_code=$?
if [ $build_exit_code != 0 ]; then
  exit $build_exit_code
fi

IS_MINIKUBE=true bash ./scripts/development-install.sh

kubectl get pods -n datree
