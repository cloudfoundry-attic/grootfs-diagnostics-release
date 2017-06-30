#!/bin/bash -ex
cd $(dirname $0)/..

flyrc_target=$1
if [ -z $flyrc_target ]; then
  echo "No target passed, using 'grootfs-ci'"
  flyrc_target="grootfs-ci"
fi
[ -z $DEBUG ] && DEBUG=0

main() {
  check_fly_alias_exists
  sync_and_login

  pipeline_name="grootfs-diagnostics"
  vars_name="gcp"
  pipeline_file="pipeline.yml"

  if [ $DEBUG -eq 1 ]; then
    pipeline_name="grootfs-test"
    vars_name="dummy"

    spruce --concourse merge ci/$pipeline_file ci/dummy-pipeline-spruce.yml > ci/dummy_pipeline.yml
    pipeline_file="dummy_pipeline.yml"
  fi

  set_pipeline
  expose_pipeline
}

set_pipeline() {
  fly --target="$flyrc_target" set-pipeline --pipeline=$pipeline_name \
    --config=ci/${pipeline_file} --load-vars-from=$HOME/workspace/grootfs-ci-secrets/vars/$vars_name.yml \
    --var gnome-private-key="$(lpass show 'Shared-Garden/grootfs-deployments/github-garden-gnome' --notes)" \
    --var github-access-token="$(lpass show 'Shared-Garden/Garden-Gnome-Github-Account' --field=api-key)" \
    --var gamora-bosh-username="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/bosh-director' --username)" \
    --var gamora-bosh-password="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/bosh-director' --password)" \
    --var gamora-bosh-certificates="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/certificates' --notes)" \
    --var dockerhub-username="$(lpass show 'Shared-Garden/cf-garden-docker' --username)" \
    --var dockerhub-password="$(lpass show 'Shared-Garden/cf-garden-docker' --password)" \
    --var garden-tracker-token="$(lpass show 'Shared-Garden/Garden-Gnome-Tracker-Account' --notes)" \
    --var aws-access-key-id="$(lpass show "Shared-Garden/grootfs-deployments\thanos/aws-keys" --username)" \
    --var aws-secret-access-key="$(lpass show "Shared-Garden/grootfs-deployments\thanos/aws-keys" --password)" \
    --var datadog-api-key="$(lpass show 'Shared-Garden/grootfs-deployments/datadog-api-keys' --username)" \
    --var datadog-application-key="$(lpass show 'Shared-Garden/grootfs-deployments/datadog-api-keys' --password)" \
    --var cf-username="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-creds' --username)" \
    --var cf-password="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-creds' --password)" \
    --var cf-secrets="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-secrets' --notes)" \
    --var cf-uaa-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-uaa-certs' --notes)" \
    --var cf-etcd-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-etcd-certs' --notes)" \
    --var cf-cc-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-cc-certs' --notes)" \
    --var cf-consul-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-consul-certs' --notes)" \
    --var cf-diego-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-diego-certs' --notes)" \
    --var cf-loggregator-certs="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-loggregator-certs' --notes)" \
    --var cf-networking="$(lpass show 'Shared-Garden/grootfs-deployments\gamora/cf-networking' --notes)" \
    --var github-client-id="$(lpass show 'Shared-Garden/grootfs-deployments/github-garden-gnome' --username)" \
    --var github-client-secret="$(lpass show 'Shared-Garden/grootfs-deployments/github-garden-gnome' --password)" \
    --var grootfs-release-private-yaml="$(lpass show 'Shared-Garden/grootfs-release-private.yml' --notes)" \
    \
    --var thanos-bosh-username="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/bosh-director' --username)" \
    --var thanos-bosh-password="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/bosh-director' --password)" \
    --var thanos-bosh-certificates="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/certificates' --notes)" \
    --var thanos-cf-diego-certs="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/cf-diego-certs' --notes)"
}

expose_pipeline() {
  fly --target="$flyrc_target" expose-pipeline --pipeline="$pipeline_name"
}

check_fly_alias_exists() {
  set +e
  grep $flyrc_target ~/.flyrc > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Please ensure $flyrc_target exists in ~/.flyrc and that you have run fly login"
    exit 1
  fi
  set -e
}

sync_and_login() {
  fly -t $flyrc_target sync

  set +e
  fly -t $flyrc_target containers > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    fly -t $flyrc_target login
  fi
  set -e
}

main
