#!/usr/bin/env bash
set -ex

export GO111MODULE=on
#go mod init qitmeer
go mod tidy
#export PATH=$PATH:$(pwd)/build/bin

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
golangci-lint run -v --deadline=2m --disable-all --enable=govet --tests=false ./...

exit 0
# After the account and password are set, we can open it

project_name="qitmeer"
image_name="qitmeer/qitmeerd"
image_label="" #future
baseimage_name="qitmeer/golang:1.14.12-alpine3.12"

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin && \
docker pull $baseimage_name

cur_dir=$(pwd)

if [ -d "$cur_dir/bin" ];then
   rm -rf ./bin
fi

mkdir -p bin/build

docker run --rm -v $cur_dir/:/go/src/$project_name -w /go/src/$project_name/cmd/qitmeerd -e GO111MODULE=on $baseimage_name go build -o ./../../bin/build/qitmeerd && \

# create launch

cat>./bin/build/launch<<EOF
#!/usr/bin/env bash
A="-A=./"
net="--mixnet"
rpcuser="--rpcuser=test"
rpcpass="--rpcpass=test"

cd /qitmeer/

if [[ "\$1" == "cli" ]]
then
  ./\$*
  exit
fi

./qitmeerd \$A \$net \$rpcuser \$rpcpass "\$@"
EOF

chmod u+x ./bin/build/launch && \

#---------------------------------------------------------------------------------------------
# create cli

cp ./src/$project_name/script/cli.sh ./bin/build/cli && \
sed -i '' 's/127.0.0.1/172.17.0.1/g' ./bin/build/cli && \
sed -i '' 's/port=1234/port=28131/g' ./bin/build/cli && \

# build image
cp ./container/docker/alpine/Dockerfile ./bin/ && \
cd ./bin/ && \
docker build -t $image_name ./ && \
docker push $image_name && \
echo -e "\n Success!"


