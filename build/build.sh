#!/bin/bash

# イメージがない場合はビルド
ARRAY=(`docker images | grep go-build`)
if [ ${#ARRAY[@]} = 0 ]; then
    docker build --network=host -t go-build:latest .
fi
# ビルドコンテナ開始
mkdir -p ./app/bin
docker-compose up --build
# 運用中のコンテナ停止
cd ../
docker-compose down
# 移動
cd build
mv -f ./app/bin/* ../app/bin/
rm -f ./app/bin/.gitkeep
chmod -R 777 ../app/bin/
# 後片付け（正しく移動できていればフォルダ削除できるはず）
rmdir ./app/bin
rmdir ./app
