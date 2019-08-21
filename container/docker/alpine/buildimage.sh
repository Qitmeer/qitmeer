#!/usr/bin/env bash
set -e

cd ./../../../

project_name="qitmeer"
image_name="qitmeer/qitmeerd"
baseimage_name="qitmeer/golang:1.12.5-alpine3.9"

docker pull $baseimage_name

cur_dir=$(pwd)

if [ -d "$cur_dir/bin" ];then
   rm -rf $cur_dir/bin
fi

mkdir -p bin/build

docker run --rm -v $cur_dir/:/go/src/$project_name -w /go/src/$project_name -e GO111MODULE=on $baseimage_name go build -o ./bin/build/qitmeerd && \

# build image
cp ./container/docker/alpine/Dockerfile ./bin/ && \
cd ./bin/ && \
docker build -t $image_name ./  && \

docker images
echo -e "\n Success!"

if [ -d "$cur_dir/bin" ];then
   rm -rf $cur_dir/bin
fi