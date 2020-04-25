#!/bin/bash

# リポジトリ最新化
git checkout .
git pull origin master
chmod -R 777 .
# ビルド
cd build
./build.sh
# コンテナ開始
cd ../
docker-compose up -d --build
