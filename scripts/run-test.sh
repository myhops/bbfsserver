#!/bin/bash

export PORT=8080
export BBFSSRV_HOST=bitbucket.belastingdienst.nl
export BBFSSRV_PROJECT_KEY=essentials
export BBFSSRV_REPOSITORY_SLUG=olo-kor-build-reports
export BBFSSRV_LOG_FORMAT=text
export BBFSSRV_ACCESS_KEY=$(gopass --password private/olo-kor-build-reports/access-token/test-bbfs)

WDIR=$(dirname ${BASH_SOURCE[0]})
go run $WDIR/../cmd/bbfsserver "$@"