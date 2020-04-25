#!/bin/bash

HOME=${GOPATH}/src/github.com/bikeshare_api

# gitフォルダをチェックしあればpull、なければcloneする（普通はある）
if [ -e $HOME ]; then
    cd $HOME
    git checkout .
    git pull origin master
else
    cd ${GOPATH}/src/github.com
    git clone https://github.com/8245snake/bikeshare_api.git
fi
chmod -R 777 $HOME

# 最新化したビルドスクリプトを叩く
bash $HOME/build/start.sh