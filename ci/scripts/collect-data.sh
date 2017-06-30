#!/bin/bash
set -e -x

event_text=$(cat dstate-event/event.json | jq .text)
dstate_json=$(echo "$event_test" | tr -d '\')

CELL_NAME=$(echo $dstate_json | jq '.cell_name')
CELL_ID=$(echo $dstate_json | jq '.cell_id')
TAR_PATH=$(echo $dstate_json | jq '.logs_path')

echo "$BOSH_CERTIFICATES" > certificates.yml
bosh2 int --path "/certs/ca_cert" certificates.yml > ca_cert.crt

echo "Running performance tests..."
bosh2 -e $BOSH_TARGET --ca-cert ca_cert.crt alias-env bosh-director
bosh2 -e bosh-director --client $BOSH_CLIENT --client-secret $BOSH_CLIENT_SECRET login

bosh2 -e bosh-director -d $BOSH_DEPLOYMENT scp $CELL_NAME/$CELL_ID:$TAR_PATH diagnostics
