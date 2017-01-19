#! /bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

export GOPATH=$root/gopath

source $root/ci/vars.sh

mkdir -p $GOPATH $GOPATH/bin $GOPATH/src $GOPATH/pkg

PATH=$PATH:$GOPATH/bin

go version

# install metalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install

go get -v github.com/venicegeo/bf-ia-broker

# Planet package
cd $GOPATH/src/github.com/venicegeo/bf-ia-broker/planet

# lint
gometalinter \
--deadline=60s \
--concurrency=6 \
--vendor \
--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
--exclude="Api.* should be .*API" \
--exclude="Http.* should be .*HTTP" \
--exclude="Id.* should be .*ID" \
--exclude="Json.* should be .*JSON" \
--exclude="Url.* should be .*URL" \
--exclude="[iI][dD] can be fmt\.Stringer" \
--exclude=" duplicate of [A-Za-z\._0-9]*" \
./... | tee $root/planet-lint.txt
wc -l $root/planet-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=planet.cov github.com/venicegeo/bf-ia-broker/planet
cp ./planet.cov $root/planet.cov

# Tides package
cd $GOPATH/src/github.com/venicegeo/bf-ia-broker/tides

# lint
gometalinter \
--deadline=60s \
--concurrency=6 \
--vendor \
--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
--exclude="Api.* should be .*API" \
--exclude="Http.* should be .*HTTP" \
--exclude="Id.* should be .*ID" \
--exclude="Json.* should be .*JSON" \
--exclude="Url.* should be .*URL" \
--exclude="[iI][dD] can be fmt\.Stringer" \
--exclude=" duplicate of [A-Za-z\._0-9]*" \
./... | tee $root/tides-lint.txt
wc -l $root/tides-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=tides.cov github.com/venicegeo/bf-ia-broker/tides
cp ./tides.cov $root/tides.cov

# Util package
cd $GOPATH/src/github.com/venicegeo/bf-ia-broker/util

# lint
gometalinter \
--deadline=60s \
--concurrency=6 \
--vendor \
--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
--exclude="Api.* should be .*API" \
--exclude="Http.* should be .*HTTP" \
--exclude="Id.* should be .*ID" \
--exclude="Json.* should be .*JSON" \
--exclude="Url.* should be .*URL" \
--exclude="[iI][dD] can be fmt\.Stringer" \
--exclude=" duplicate of [A-Za-z\._0-9]*" \
./... | tee $root/util-lint.txt
wc -l $root/util-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=util.cov github.com/venicegeo/bf-ia-broker/util
cp ./util.cov $root/util.cov


# gather some data about the repo

cd $root
cp $GOPATH/bin/$APP ./$APP.bin
tar cvzf $APP.$EXT \
    $APP.bin \
    planet.cov \
    planet-lint.txt \
    tides.cov \
    tides-lint.txt \
    util.cov \
    util-lint.txt
tar tzf $APP.$EXT
