package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/mattn/go-scan"
)

const (
	//endpoint エンドポイント
	endpoint = "https://api.imgur.com/3/image"
	//ErrorImageURL エラー時に見せる画像のURL
	ErrorImageURL = "https://imgur.com/fXmLgph"
)

var (
	//Configs 設定
	Configs []rdb.ConfigDB
	//ImgurID APIのキー
	ImgurID string
)

//getConfig DBから設定を取得
func getConfig(key string) string {
	if Configs == nil {
		db, err := rdb.GetConnectionPsql()
		if err != nil {
			return ""
		}
		defer db.Close()
		Configs, err = rdb.SearchConfig(db, rdb.SearchOptions{})
		if err != nil {
			return ""
		}
	}
	for _, conf := range Configs {
		if conf.Key == key {
			return conf.Value
		}
	}
	return ""
}

//UploadImgur 画像アップロード
func UploadImgur(imgPath string) string {

	b, err := ioutil.ReadFile(imgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err.Error())
		return ErrorImageURL
	}
	params := url.Values{"image": {base64.StdEncoding.EncodeToString(b)}}

	var res *http.Response

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "post:", err.Error())
		return ErrorImageURL
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Client-ID "+ImgurID)

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "post:", err.Error())
		return ErrorImageURL
	}
	if res.StatusCode != 200 {
		var message string
		err = scan.ScanJSON(res.Body, "data/error", &message)
		if err != nil {
			message = res.Status
		}
		fmt.Fprintln(os.Stderr, "post:", message)
		return ErrorImageURL
	}
	defer res.Body.Close()

	var link string
	err = scan.ScanJSON(res.Body, "data/link", &link)
	if err != nil {
		fmt.Fprintln(os.Stderr, "post:", err.Error())
		return ErrorImageURL
	}
	return link
}
