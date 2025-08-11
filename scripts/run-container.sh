#!/bin/bash

# ACCESS_KEY_NAME=private/obp-pnr/build-reports/http-token
ACCESS_KEY_NAME=obp-pnr/git-repo/accesskeys/http/access-token

podman run --rm -p 18080:8080 \
    -e BBFSSRV_HOST=bitbucket.belastingdienst.nl \
    -e BBFSSRV_PROJECT_KEY=OBDMO \
    -e BBFSSRV_REPOSITORY_SLUG=obp-pnr-build-reports \
    -e BBFSSRV_LOG_FORMAT=text \
    -e BBFSSRV_ACCESS_KEY=$(gopass --password $ACCESS_KEY_NAME) \
    -e BBFSSRV_TITLE="OBP PNR-Parkeren en Routeren" \
    cir-cn-devops.chp.belastingdienst.nl/obp-pnr/bbfsserver:v0.0.12
