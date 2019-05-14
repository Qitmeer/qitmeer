#!/usr/bin/env bash
set -ex

export GO111MODULE=on
go mod init qitmeer
go mod tidy

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
golangci-lint run -v --deadline=2m --disable-all --enable=govet  --enable=gosimple ./...

linter_targets=$(go list ./...) && \
go test $linter_targets

if [[ $TRAVIS_PULL_REQUEST != 'false' || $TRAVIS_REPO_SLUG != 'HalalChain/qitmeer' && $TRAVIS_BRANCH != 'master' ]];
then
    exit 0
fi

project_name="qitmeer"
image_name="halalchain/nox-dag"
image_label="" #future
baseimage_name="halalchain/golang:1.12.5-alpine3.9"

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin && \
docker pull $baseimage_name

cur_dir=$(pwd)

if [ -d "$cur_dir/bin" ];then
   rm -rf ./bin
fi

mkdir bin

docker run --rm -v $cur_dir/:/go/src/$project_name -w /go/src/$project_name -e GO111MODULE=on $baseimage_name go build -o ./bin/noxd && \
docker run --rm -v $cur_dir/:/go/src/$project_name -w /go/src/$project_name/tools/nx -e GO111MODULE=on $baseimage_name go build -o ./../../bin/nx && \

# create launch

cat>./bin/launch<<EOF
#!/usr/bin/env bash
A="-A=./"
net="--testnet"
rpclisten="--rpclisten=0.0.0.0:18131"
rpcuser="--rpcuser=test"
rpcpass="--rpcpass=test"
txindex="--txindex"
debuglevel="--debuglevel=info"

if [[ "\$@" =~ "miningaddr" ]]
then
  miningaddr=""
else
  miningaddr="--miningaddr=Tmgb3CyW7rGgn89MWEoAoMP47CwASc4KG4N"
fi

cd /nox/

if [[ "\$1" == "cli" ]]
then
  ./\$*
  exit
fi

if [[ "\$1" == "nx" ]]
then
  ./\$*
  exit
fi

./noxd \$A \$net \$rpclisten \$rpcuser \$rpcpass \$txindex \$miningaddr \$debuglevel "\$@"
EOF

chmod u+x ./bin/launch
# create cli

cp ./script/nox/nox-cli.sh ./bin/cli
sed -i 's/127.0.0.1/172.17.0.1/g' ./bin/cli
sed -i 's/port=1234/port=18131/g' ./bin/cli

# build image
cp ./container/docker/alpine/Dockerfile ./bin/ && \
cd ./bin/ && \
docker build -t $image_name ./ && \
docker push $image_name && \
echo -e "\n Success!"


