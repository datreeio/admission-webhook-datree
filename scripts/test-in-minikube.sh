bash ./scripts/deploy-in-minikube.sh

function removeTokenAndFileNamePostfixFromResult() {
  # For anonymous invocations, the token changes on every run.
  # Therefore we need to remove the token from the result.
  # Same for the file name, which has a changing postfix
  echo "${1}" | 
  sed -e 's/app.staging.datree.io\/login?t=[^ ]\{1,\}/app.staging.datree.io\/login?t=<TOKEN>               /g' |
  sed -e 's/fileToTest-rss-site-Deployment-[^ ]\{1,\}\.yaml/fileToTest-rss-site-Deployment-<FILE_CHANGING_POSTFIX>.yaml/g'
}

# declare colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

expectedResult=$(cat ./internal/fixtures/webhook-demo-expected-output.txt)
expectedResult=$(removeTokenAndFileNamePostfixFromResult "${expectedResult}")
result=$(kubectl apply -f ./internal/fixtures/webhook-demo.yaml 2>&1) # redirect stderr to stdout
result=$(removeTokenAndFileNamePostfixFromResult "${result}")

if [ "${result}" = "${expectedResult}" ]; then
  echo "${GREEN}Test passed${NC}"
  exit 0
else
  echo "${RED}Test failed${NC}"
  echo "Expected result:"
  echo "${expectedResult}"
  echo "Actual result:"
  echo "${result}"
  exit 1
fi
