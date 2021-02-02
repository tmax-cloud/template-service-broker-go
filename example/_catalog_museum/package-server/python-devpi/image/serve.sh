#!/bin/sh

set -ex

devpi-init
devpi-gen-config --host 0.0.0.0 --port 3141
supervisord -c gen-config/supervisord.conf
#devpi-server --start --host=0.0.0.0
tail -f /dev/null
