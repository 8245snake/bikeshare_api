#!/bin/bash

# イメージがない場合はビルド
ARRAY=(`docker images | grep go-build`)
if [ ${#ARRAY[@]} = 0 ]; then
    docker build --network=host -t go-build:latest .
fi

# ビルドして移動
mkdir -p ./app/bin
docker-compose up --build
mv -f ./app/bin/* ../app/bin/
rm ./app/bin/.gitkeep

# 正しく移動できていればフォルダ削除できるはず
rmdir ./app/bin
rmdir ./app
