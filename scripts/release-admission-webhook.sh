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
    cecho "RED" "yq doesn't exist, please install yq"
    exit 1
fi

if ! command -v gh &> /dev/null; then
    cecho "RED" "yq doesn't exist, please install yq"
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
current_version=$(yq e '.version' ../charts/datree-admission-webhook/Chart.yaml)
cecho "CYAN" "current version: $current_version"
new_version=$(echo $current_version | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')
cecho "CYAN" "new version: $new_version"
# update Chart.yaml
yq e -i ".version = \"$new_version\"" ../charts/datree-admission-webhook/Chart.yaml


#helm
helm dependency build ../charts/datree-admission-webhook/
helm package ../charts/datree-admission-webhook/ -d /tmp/
cecho "GREEN" "helm package done"
cecho "CYAN" "switch to temp branch to create PR"
git stash
git checkout gh-pages
git pull
git checkout -b "release-chart-$new_version"
mv "/tmp/datree-admission-webhook-$new_version.tgz" ../
helm repo index --url https://datreeio.github.io/admission-webhook-datree/ ../ --merge ../index.yaml
git add ../index.yaml
git add ../datree-admission-webhook-$new_version.tgz
git commit -m "release chart $new_version"
git push --set-upstream origin "release-chart-$new_version"
cecho "CYAN" "open PR"
gh pr create --title "release chart $new_version" --body "release chart $new_version" --base gh-pages --head release-chart-$new_version
gco -
cecho "CYAN" "switch back"
git stash pop
cecho "GREEN" "done"

