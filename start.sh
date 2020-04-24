#!/bin/bash
# ビルド用コンテナと実行用コンテナで共通利用する

HOME=${GOPATH}/src/github.com/bikeshare_api
NAMES=(${BINARY_NAMES//,/ })
SIZE=${#NAMES[@]}

# gitフォルダをチェックしあればpull、なければcloneする
SourcePrepare(){
    if [ -e $HOME ]; then
        cd $HOME
        git checkout .
        git pull origin master
    else
        cd ${GOPATH}/src/github.com
        git clone https://github.com/8245snake/bikeshare_api.git
    fi
    chmod -R 777 $HOME
}

# 全部ビルドする
FullBuild(){
    for BINARY_NAME in ${NAMES[@]}; do
        cd ${HOME}/src/${BINARY_NAME}
        godep restore
        go build -o ${BINARY_NAME}
        mv ${BINARY_NAME} ../../app/bin
    done
    # ビルド成果物を作業フォルダにコピー
    chmod -R 777 $HOME
    cp -R -f  ${HOME}/ /usr/
}

# 実行
Execute(){
    cd /usr/bikeshare_api/app/bin
    exec ./${NAMES[0]}
}

# メイン処理
if [ $SIZE = 1 ]; then
    # 実行
    Execute
else
    # BINARY_NAMESが複数あるときはビルドのみ
    SourcePrepare
    FullBuild
fi