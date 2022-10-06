#!/bin/sh

cecho(){
    RED="\033[0;31m"
    GREEN="\033[0;32m"  
    YELLOW="\033[1;33m" 
    CYAN="\033[1;36m"
    NC="\033[0m" # No Color
    printf "${!1}${2} ${NC}\n"
}


function verify_command_exists() {
RED="\033[0;31m"
 NC="\033[0m"

if ! command -v helm &> /dev/null; then
    cecho "RED" "helm doesn't exist, please install helm"
    exit 1
fi

if ! command -v gh &> /dev/null; then
    cecho "RED" "gh doesn't exist, please install gh"
    exit 1
fi

if ! command -v yq &> /dev/null; then
    cecho "RED" "yq doesn't exist, please install yq"
    exit 1
fi

}

verify_command_exists

# bump patch version Chart.yaml
cecho "CYAN" "bump patch version"
current_version=$(yq e '.version' ./charts/datree-admission-webhook/Chart.yaml)
cecho "CYAN" "current version: $current_version"
new_version=$(echo $current_version | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')
cecho "CYAN" "new version: $new_version"

#helm
helm dependency build ./charts/datree-admission-webhook/
git checkout gh-pages index.yaml
helm package ./charts/datree-admission-webhook/ -d /tmp/
helm repo index --url https://datreeio.github.io/admission-webhook-datree/ ./ --merge ./index.yaml
mv index.yaml /tmp/
cecho "GREEN" "helm package done"
cecho "CYAN" "switch to temp branch to create PR"
git stash
git checkout gh-pages
git pull
git checkout -b "release-chart-$new_version"
mv "/tmp/datree-admission-webhook-$new_version.tgz" ./
mv "/tmp/index.yaml" ./
git add ./index.yaml
git add ./datree-admission-webhook-$new_version.tgz
git commit -m "feat: Release chart datree-admission-webhook-$new_version.tgz"
git push --set-upstream origin "release-chart-$new_version"
cecho "CYAN" "open PR"
gh pr create --title "Release chart datree-admission-webhook-$new_version" --body "release chart $new_version" --base gh-pages --head release-chart-$new_version
git checkout main
git pull
git checkout -b "update-chart-$new_version"
cecho "CYAN" "switch to main"
# update Chart.yaml
yq e -i ".version = \"$new_version\"" ./charts/datree-admission-webhook/Chart.yaml
git add ./charts/datree-admission-webhook/Chart.yaml
git commit -m "bump: Update Chart.yaml version"
git push --set-upstream origin "update-chart-$new_version"
gh pr create --title "Update Chart.yaml version" --body "update Chart.yaml $new_version" --base main --head update-chart-$new_version
git checkout main
cecho "CYAN" "switch to main"
cecho "GREEN" "done"

