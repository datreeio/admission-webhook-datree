#!/bin/sh
set -e
set -o pipefail

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

function verify_updated_main_branch(){
    git checkout main
    git pull origin main

    if [ -n "$(git status --porcelain)" ]; then
        cecho "RED" "Please commit your changes before running this script"
        exit 1
    fi
}
function verify_updated_gh_pages_branch(){
    git checkout gh-pages
    git pull origin gh-pages

    if [ -n "$(git status --porcelain)" ]; then
        cecho "RED" "Please commit your changes before running this script"
        exit 1
    fi
}

function clean_up(){
    cecho "CYAN" "Cleaning up"
    if git show-ref --verify --quiet "refs/heads/release-helm-chart"; then
        git branch -D $(git branch | grep "release-helm-chart-")
    fi
}

clean_up
verify_command_exists
verify_updated_gh_pages_branch
verify_updated_main_branch


# bump patch version Chart.yaml
cecho "CYAN" "Bumping patch version..."
current_version=$(yq e '.version' ./charts/datree-admission-webhook/Chart.yaml)
cecho "CYAN" "Current version: $current_version"
new_version=$(echo $current_version | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')
cecho "CYAN" "New version: $new_version"

#helm
yq e -i ".version = \"$new_version\"" ./charts/datree-admission-webhook/Chart.yaml
cecho "GREEN" "Export helm package to /tmp"
helm dependency build ./charts/datree-admission-webhook/
helm package ./charts/datree-admission-webhook/ --version=$new_version -d /tmp/
cecho "GREEN" "PR to main release-helm-chart-$new_version"
git checkout -b "release-helm-chart-$new_version"
git add ./charts/datree-admission-webhook/Chart.yaml
git commit -m "Bump helm chart $new_version"
git push origin "release-helm-chart-$new_version"
cecho "GREEN" "Creating PR"
gh pr create --title "bump version to $new_version" --body "bump version to $new_version" --base main --head "release-helm-chart-$new_version"

cecho "GREEN" "Done!"
cecho "YELLOW" "Prepare to release helm chart to gh-pages..."

git checkout gh-pages
mv "/tmp/datree-admission-webhook-$new_version.tgz" ./
helm repo index --url https://datreeio.github.io/admission-webhook-datree/ ./ --merge ./index.yaml
git checkout -b "release-chart-$new_version"

git add ./index.yaml
git add ./datree-admission-webhook-$new_version.tgz

git commit -m "feat: Release chart datree-admission-webhook-$new_version.tgz"
git push --set-upstream origin "release-chart-$new_version"
cecho "CYAN" "open PR"
gh pr create --title "Release chart datree-admission-webhook-$new_version" --body "release chart $new_version" --base gh-pages --head release-chart-$new_version

# git checkout main
# git pull
# git checkout -b "update-chart-$new_version"
# cecho "CYAN" "switch to main"
# # update Chart.yaml
# yq e -i ".version = \"$new_version\"" ./charts/datree-admission-webhook/Chart.yaml
# git add ./charts/datree-admission-webhook/Chart.yaml
# git commit -m "bump: Update Chart.yaml version"
# git push --set-upstream origin "update-chart-$new_version"
# gh pr create --title "Update Chart.yaml version" --body "update Chart.yaml $new_version" --base main --head update-chart-$new_version
# git checkout main
# cecho "CYAN" "switch to main"
# cecho "GREEN" "done"

