#!/usr/bin/env bash
set -ex
export GO111MODULE=on

rm -rf lib/cuckoo/target/*
rm -rf lib/opencl/*
rm -rf libOpenCL.zip libcuckoo.zip

wget -O libcuckoo.zip https://github.com/Qitmeer/cuckoo-lib/releases/download/v0.0.1/libcuckoo.zip
unzip libcuckoo.zip -d lib/cuckoo/target/
wget -O libOpenCL.zip https://github.com/Qitmeer/OpenCL-ICD-Loader/releases/download/v0.0.1/libopencl.zip
unzip libOpenCL.zip -d lib/opencl/

export LD_LIBRARY_PATH=`pwd`/lib/cuckoo/target/x86_64-unknown-linux-musl/release:`pwd`/lib/opencl/linux:$LD_LIBRARY_PATH
echo $LD_LIBRARY_PATH
sudo cp `pwd`/lib/opencl/linux/libOpenCL.a /usr/lib/x86_64-linux-musl/

cd lib/cuda

nvcc -m64 -arch=sm_35 -o libcudacuckoo.so --shared -std=c++11 -Xcompiler -fPIC -DEDGEBITS=24 -DSIPHASH_COMPAT=1 mean.cu ./crypto/blake2b-ref.c
sudo cp `pwd`/libcudacuckoo.so /usr/lib/x86_64-linux-musl/
cd ../../

go mod tidy

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
CGO_ENABLED=1 CGO_ENABLED=1 CC=musl-gcc CXX=g++ GOOS=linux GOARCH=amd64 go build -o linux-miner -tags cuda main.go
echo -e "\n Success!"


