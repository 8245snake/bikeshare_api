#!/bin/bash
# ビルド用コンテナと実行用コンテナで共通利用する

HOME=${GOPATH}/src/github.com/bikeshare_api
NAMES=(${BINARY_NAMES//,/ })
SIZE=${#NAMES[@]}

# 全部ビルドする
FullBuild(){
    # 成果物を削除しておく
    for BINARY_NAME in ${NAMES[@]}; do
        if [ -e /usr/bikeshare_api/app/bin/${BINARY_NAME} ]; then
            echo ${BINARY_NAME} exists and removed
            rm /usr/bikeshare_api/app/bin/${BINARY_NAME}
        fi
    done
    # すべてビルド
    for BINARY_NAME in ${NAMES[@]}; do
        echo start build to ${BINARY_NAME}
        cd ${HOME}/src/${BINARY_NAME}
        godep restore
        echo restore end
        go build -o ${BINARY_NAME}
        echo build end
        mv ${BINARY_NAME} ../../app/bin
    done
    # ビルド成果物を作業フォルダにコピー
    chmod -R 777 $HOME
    cp -R -f  ${HOME}/ /usr/
}

# 実行
Execute(){
    # バイナリが見つかるまでループ
    while true
    do
        if [ -e /usr/bikeshare_api/app/bin/${NAMES[0]} ]; then
            break
        fi
        echo "looking for ${NAMES[0]}..."
        sleep 10
    done

    echo "execute ${NAMES[0]}"
    cd /usr/bikeshare_api/app/bin
    exec ./${NAMES[0]}
}

# メイン処理
if [ $SIZE = 1 ]; then
    # 実行
    Execute
else
    # BINARY_NAMESが複数あるときはビルドのみ
    FullBuild
fi