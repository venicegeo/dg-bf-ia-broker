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

go get -v github.com/venicegeo/dg-bf-ia-broker

# Planet package
cd $GOPATH/src/github.com/venicegeo/dg-bf-ia-broker/planet

# lint
#gometalinter \
#--deadline=60s \
#--concurrency=6 \
#--vendor \
#--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
#--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
#--exclude="Api.* should be .*API" \
#--exclude="Http.* should be .*HTTP" \
#--exclude="Id.* should be .*ID" \
#--exclude="Json.* should be .*JSON" \
#--exclude="Url.* should be .*URL" \
#--exclude="[iI][dD] can be fmt\.Stringer" \
#--exclude=" duplicate of [A-Za-z\._0-9]*" \
#./... | tee $root/planet-lint.txt
#wc -l $root/planet-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/planet.cov github.com/venicegeo/dg-bf-ia-broker/planet
go tool cover -func=$root/planet.cov -o $root/planet.cov.txt

# Tides package
cd $GOPATH/src/github.com/venicegeo/dg-bf-ia-broker/tides

# lint
#gometalinter \
#--deadline=60s \
#--concurrency=6 \
#--vendor \
#--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
#--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
#--exclude="Api.* should be .*API" \
#--exclude="Http.* should be .*HTTP" \
#--exclude="Id.* should be .*ID" \
#--exclude="Json.* should be .*JSON" \
#--exclude="Url.* should be .*URL" \
#--exclude="[iI][dD] can be fmt\.Stringer" \
#--exclude=" duplicate of [A-Za-z\._0-9]*" \
#./... | tee $root/tides-lint.txt
#wc -l $root/tides-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/tides.cov github.com/venicegeo/dg-bf-ia-broker/tides
go tool cover -func=$root/tides.cov -o $root/tides.cov.txt

# Util package
cd $GOPATH/src/github.com/venicegeo/dg-bf-ia-broker/util

# lint
#gometalinter \
#--deadline=60s \
#--concurrency=6 \
#--vendor \
#--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
#--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
#--exclude="Api.* should be .*API" \
#--exclude="Http.* should be .*HTTP" \
#--exclude="Id.* should be .*ID" \
#--exclude="Json.* should be .*JSON" \
#--exclude="Url.* should be .*URL" \
#--exclude="[iI][dD] can be fmt\.Stringer" \
#--exclude=" duplicate of [A-Za-z\._0-9]*" \
#./... | tee $root/util-lint.txt
#wc -l $root/util-lint.txt

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/util.cov github.com/venicegeo/dg-bf-ia-broker/util
go tool cover -func=$root/util.cov -o $root/util.cov.txt

# gather some data about the repo

cd $root
cp $GOPATH/bin/$APP ./$APP.bin
tar cvzf $APP.$EXT \
    $APP.bin \
    planet.cov \
    planet.cov.txt \
    tides.cov \
    tides.cov.txt \
    util.cov \
    util.cov.txt
#    tides-lint.txt \
#    util-lint.txt \
#    planet-lint.txt \
tar tzf $APP.$EXT
