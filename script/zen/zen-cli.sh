#!/bin/bash

app_dir=/work/zenprotocol/zenprotocol/src/Node/bin/Debug/
cd $app_dir

mono zen-cli.exe "$@"

cd -
