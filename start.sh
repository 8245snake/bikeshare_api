#!/bin/sh

HOME=${GOPATH}/src/github.com/bikeshare_api

if [ -e $HOME ]; then
    rm -r $HOME
fi

cd ${GOPATH}/src/github.com
git clone https://github.com/8245snake/bikeshare_api.git
chmod -R 777 $HOME

cd ${$HOME}/src/${BINARY_NAME}
make

cd ${$HOME}/app/bin/${BINARY_NAME}

${BINARY_NAME}