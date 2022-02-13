#!/usr/bin/dumb-init /bin/bash

# Environment Variables:
# * GH_HOST
# * GH_TOKEN
# * RUNNER_ENTERPRISE
# * RUNNER_ORG
# * RUNNER_REPO
# * RUNNER_WORKDIR
# * RUNNER_NAME
# * RUNNER_LABELS
# * RUNNER_GROUP

github_host=${GH_HOST:="github.com"}
github_url="https://${github_host}"

set_registration_token() {
  registration_token="$(curl -XPOST -fsSL \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Authorization: token ${GH_TOKEN}" \
    "${registration_endpoint}" |
    jq -r '.token')"
}

deregister_runner() {
  echo "deregistering runner ${RUNNER_NAME}"
  set_registration_token
  ./config.sh remove --token "${registration_token}"
  exit
}

if [[ ${github_host} = "github.com" ]]; then
  github_api_url="https://api.${github_host}"
else
  github_api_url="https://${github_host}/api/v3"
fi

if [[ -n ${RUNNER_ORG} ]] && [[ -n ${RUNNER_REPO} ]]; then
  registration_token_path="repos/${RUNNER_ORG}/${RUNNER_REPO}"
elif [ -n "${RUNNER_ORG}" ]; then
  registration_token_path="orgs/${RUNNER_ORG}"
elif [ -n "${RUNNER_ENTERPRISE}" ]; then
  registration_token_path="enterprises/${RUNNER_ENTERPRISE}"
else
  echo "Requires at least one of environment variables RUNNER_ENTERPRISE, RUNNER_ORG or RUNNER_REPO to be set."
  exit 1
fi

registration_endpoint="${github_api_url}/${registration_token_path}/actions/runners/registration-token"
echo "requesting registration token, endpoint: ${registration_endpoint}"

if [[ -n ${RUNNER_ORG} ]] && [[ -n ${RUNNER_REPO} ]]; then
  configure_path="${RUNNER_ORG}/${RUNNER_REPO}"
elif [ -n "${RUNNER_ORG}" ]; then
  configure_path="${RUNNER_ORG}"
elif [ -n "${RUNNER_ENTERPRISE}" ]; then
  configure_path="enterprises/${RUNNER_ENTERPRISE}"
else
  echo "Requires at least one of environment variables RUNNER_ENTERPRISE, RUNNER_ORG or RUNNER_REPO to be set."
  exit 1
fi

set_registration_token

configure_endpoint="${github_url}/${configure_path}"
echo "configuring runner, endpoint: ${configure_endpoint}"
./config.sh \
  --url "${configure_endpoint}" \
  --token "${registration_token}" \
  --name "${RUNNER_NAME:-awesome-ec2-runner}" \
  --work "${RUNNER_WORKDIR}" \
  --labels "${RUNNER_LABELS:-default}" \
  --runnergroup "${RUNNER_GROUP:-default}" \
  --unattended \
  --replace \
  --ephemeral

#trap deregister_runner SIGINT SIGQUIT SIGTERM INT TERM QUIT
trap deregister_runner SIGTERM SIGINT SIGQUIT

while (! docker ps >/dev/null 2>&1); do
  echo "waiting for docker daemon..."
  sleep 1
done

"$@"
