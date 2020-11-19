#!/usr/bin/env bash
set -ex

export GO111MODULE=on
#go mod init qitmeer
go mod tidy

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
golangci-lint run -v --deadline=2m --disable-all --enable=govet --tests=false ./...

linter_targets=$(go list ./...) && \
go test $linter_targets

if [[ $TRAVIS_PULL_REQUEST != 'false' || $TRAVIS_REPO_SLUG != 'Qitmeer/qitmeer' || $TRAVIS_BRANCH != 'master' ]];
then
    cd ./cmd/qitmeerd && \
    go build && \
    ./qitmeerd --version && \
    exit 0
fi

project_name="qitmeer"
image_name="qitmeer/qitmeerd"
image_label="" #future
baseimage_name="qitmeer/golang:1.12.5-alpine3.9"

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
net="--testnet"
rpclisten="--rpclisten=0.0.0.0:18131"
rpcuser="--rpcuser=test"
rpcpass="--rpcpass=test"
debuglevel="--debuglevel=info"

if [[ "\$@" =~ "miningaddr" ]]
then
  miningaddr=""
else
  miningaddr="--miningaddr=Tmgb3CyW7rGgn89MWEoAoMP47CwASc4KG4N"
fi

cd /qitmeer/

if [[ "\$1" == "cli" ]]
then
  ./\$*
  exit
fi

./qitmeerd \$A \$net \$rpclisten \$rpcuser \$rpcpass \$txindex \$miningaddr \$debuglevel "\$@"
EOF

chmod u+x ./bin/build/launch
# create cli

cp ./script/cli.sh ./bin/build/cli
sed -i 's/127.0.0.1/172.17.0.1/g' ./bin/build/cli
sed -i 's/port=1234/port=18131/g' ./bin/build/cli

# build image
cp ./container/docker/alpine/Dockerfile ./bin/ && \
cd ./bin/ && \
docker build -t $image_name ./ && \
docker push $image_name && \
echo -e "\n Success!"


