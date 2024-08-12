#!/bin/bash

podman run --rm -p 18080:8080 \
    -e BBFSSRV_HOST=bitbucket.belastingdienst.nl \
    -e BBFSSRV_PROJECT_KEY=essentials \
    -e BBFSSRV_REPOSITORY_SLUG=olo-kor-build-reports \
    -e BBFSSRV_LOG_FORMAT=text \
    -e BBFSSRV_ACCESS_KEY=$(gopass --password private/olo-kor-build-reports/access-token/test-bbfs) \
    cir-cn.chp.belastingdienst.nl/zandp06/bbfsserver-013f3560b18499010383cf0a71c2c23c
