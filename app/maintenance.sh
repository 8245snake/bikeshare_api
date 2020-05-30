#!/bin/sh

cd ./bin

./archiver &
./stationfiller &
./notify &

tail -f /dev/null	