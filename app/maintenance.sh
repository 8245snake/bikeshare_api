#!/bin/sh

# バイナリが見つかるまでループ
while true
do
    if [ -e /usr/bikeshare_api/app/bin/stationfiller ]; then
        break
    fi
    echo "looking for stationfiller..."
    sleep 10
done


cd /usr/bikeshare_api/app/bin

./archiver &
./stationfiller &
./notify &

tail -f /dev/null	