package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
)

var client *http.Client = &http.Client{}

//SendRequest リクエスト送信
func SendRequest(userID string) {
	client.Get("https://abiding-idea-241002.an.r.appspot.com/notify?user=" + userID)
}

func init() {
	//共通初期化処理
	err := filer.InitDirSetting()
	if err != nil {
		return
	}
}

func main() {
	db, err := rdb.GetConnectionPsql()
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
	for {
		users, err := rdb.GetAllUsers(db)
		if err != nil {
			continue
		}
		now := time.Now().Format("15:04")
		for _, user := range users {
			for _, notify := range user.Notifies {
				if notify == now {
					SendRequest(user.LineID)
				}
			}
		}
		time.Sleep(60 * time.Second)
	}
}
