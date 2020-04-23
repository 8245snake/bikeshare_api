#!/bin/sh

HOME=${GOPATH}/src/github.com/bikeshare_api

# ブランチをチェック
if [ -e $HOME ]; then
    cd $HOME
    git checkout .
    git pull origin master
else
    cd ${GOPATH}/src/github.com
    git clone https://github.com/8245snake/bikeshare_api.git
fi
chmod -R 777 $HOME

# ビルドする
cd ${HOME}/src/${BINARY_NAME}
make

# ビルド成果物を移動
chmod -R 777 $HOME
cp -R -f  ${HOME}/ /usr/

# 実行
cd /usr/bikeshare_api/app/bin
exec ./${BINARY_NAME}
