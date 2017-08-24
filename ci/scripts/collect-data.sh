#!/bin/bash
set -e

dstate_json=$(cat dstate-event/event.json | jq -r .text)
CELL_NAME=$(echo $dstate_json | jq -r '.cell_name')
CELL_ID=$(echo $dstate_json | jq -r '.cell_id')
TAR_PATH=$(echo $dstate_json | jq -r '.logs_path')

echo "$BOSH_CA_CERT" > ca_cert.crt

echo "Running performance tests..."
bosh -e $BOSH_TARGET --ca-cert ca_cert.crt alias-env bosh-director
bosh -e bosh-director --client $BOSH_CLIENT --client-secret $BOSH_CLIENT_SECRET login

bosh -e bosh-director -d $BOSH_DEPLOYMENT scp $CELL_NAME/$CELL_ID:$TAR_PATH diagnostics
