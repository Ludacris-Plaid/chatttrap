#!/bin/bash
set -e

ROOT="/home/dysthemix/projects/ultimate_spoof"
mkdir -p "$ROOT"/{backend/{cmd/{api,webserver},internal/{handlers,services,models,db,utils}},frontend/app,frontend/components,frontend/public,sip/{freeswitch,coturn},docker}

cd "$ROOT"
echo "Scaffold created in $ROOT"
