package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
)

var (
	//client HTTPクライアント
	client *http.Client = &http.Client{}

	endpoint string
)

//SendRequest リクエスト送信
func SendRequest(userID string) {
	URL := strings.Replace(endpoint, "${USER}", userID, -1)
	client.Get(URL)
}

func init() {
	//共通初期化処理
	err := filer.InitDirSetting()
	if err != nil {
		return
	}
	endpoint = filer.GetIniData("NOTIFY", "REQUEST", "")
	if endpoint == "" {
		panic("通知リクエストURLが設定されていません")
	}
	fmt.Printf("endpoint=%s\n", endpoint)
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
