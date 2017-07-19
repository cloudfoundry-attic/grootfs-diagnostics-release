#!/bin/bash -ex
cd $(dirname $0)/..

set -e

flyrc_target=$1
if [ -z $flyrc_target ]; then
  echo "No target passed, using 'grootfs-ci'"
  flyrc_target="grootfs-ci"
fi

main() {
  check_fly_alias_exists
  sync_and_login

  pipeline_name="grootfs-diagnostics"
  vars_name="gcp"
  pipeline_file="pipeline.yml"

  if [ $flyrc_target == 'lakitu' ]; then
    set_pipeline_lakitu
  else
    set_pipeline_grootfs
  fi
  expose_pipeline
}

set_pipeline_lakitu() {
  fly --target="$flyrc_target" set-pipeline --pipeline=$pipeline_name \
    --config=ci/${pipeline_file} \
    --var slack-alert-url="$(lpass show 'Shared-Garden/grootfs-pivotal-slack-hook' --url)" \
    --var aws-access-key-id="$(lpass show "Shared-Private-GrootFS/grootfs-diagnostics-pws-s3-user" --username)" \
    --var aws-secret-access-key="$(lpass show "Shared-Private-GrootFS/grootfs-diagnostics-pws-s3-user" --password)" \
    --var datadog-application-key="$(lpass show 'Shared-Garden/grootfs-deployments/pws-datadog-app-key' --password)" \
    --var bosh-deployment-name=cf-cfapps-io2-diego \
    --var dstate-job-worker-tag=prod \
    --var dstate-slack-alert-text='<!subteam^S5YHYPBV4|grootfsteam> D-STATE DETECTED -- https://concourse.lakitu.cf-app.com/teams/cf-grootfs/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME :red_circle: :red_circle: :red_circle: :red_circle:' \
    --var dstate-slack-alert-channel=cf-grootfs-internal \
    --load-vars-from $HOME/workspace/secrets-prod/ci/ci_bosh_secrets.yml \
    --load-vars-from $HOME/workspace/secrets-prod/ci/ci_app_specific_configs.yml
}

set_pipeline_grootfs() {

  lpass show 'Shared-Garden/grootfs-deployments\thanos/certificates' --notes > /tmp/ci-certs.yml
  cert="$(bosh2 int --path /certs/ca_cert /tmp/ci-certs.yml)"
  rm -rf /tmp/ci-certs


  fly --target="$flyrc_target" set-pipeline --pipeline=$pipeline_name \
    --config=ci/${pipeline_file} \
    --var slack-alert-url="$(lpass show 'Shared-Garden/grootfs-pivotal-slack-hook' --url)" \
    --var aws-access-key-id="$(lpass show "Shared-Private-GrootFS/grootfs-diagnostics-pws-s3-user" --username)" \
    --var aws-secret-access-key="$(lpass show "Shared-Private-GrootFS/grootfs-diagnostics-pws-s3-user" --password)" \
    --var datadog-application-key="$(lpass show 'Shared-Garden/grootfs-deployments/datadog-api-keys' --password)" \
    --var bosh-deployment-name=cf-cfapps-io2-diego \
    --var dstate-job-worker-tag=thanos \
    --var prod_bosh_client_id="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/bosh-director' --username)" \
    --var prod_bosh_client_secret="$(lpass show 'Shared-Garden/grootfs-deployments\thanos/bosh-director' --password)" \
    --var prod_bosh_ca_cert="$cert" \
    --var prod_bosh_target="https://10.0.0.6" \
    --var PROD_DATADOG_API_KEY="$(lpass show 'Shared-Garden/grootfs-deployments/datadog-api-keys' --username)" \
    --var bosh-deployment-name="cf" \
    --var dstate-slack-alert-channel=cf-grootfs \
    --var dstate-slack-alert-text="Testing! No D-State processes here! Fake news!"
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
