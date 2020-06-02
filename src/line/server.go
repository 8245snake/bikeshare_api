package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	bikeshareapi "github.com/8245snake/bikeshare-client"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	//LineOAuthEndpoint アクセストークンの取得に使用
	LineOAuthEndpoint = "https://api.line.me/v2/oauth/accessToken"
	//MaxHistory 履歴の保存件数
	MaxHistory = 10
	//MaxFavorite お気に入りの登録件数
	MaxFavorite = 5
	//MaxNotifyTimes 通知時刻の設定可能件数
	MaxNotifyTimes = 2
)

var (
	//Client デフォルトのHTTPクライアント
	Client http.Client
	//ClientID LINEのクライアントID
	ClientID string
	//ClientSecret LINEのクライアントシークレットキー
	ClientSecret string
	//AccessToken LINEのアクセストークン
	AccessToken string
	//LineBotAPI LINEのAPIクライアント
	LineBotAPI *linebot.Client
	//BikeshareAPI BikeshareのAPIクライアント
	BikeshareAPI bikeshareapi.ApiClient
	//SpotNamesDictionary スポット名の辞書
	SpotNamesDictionary = make(map[string]string)
)

//getAccessToken アクセストークン取得
func getAccessToken() string {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Add("client_id", ClientID)
	values.Add("client_secret", ClientSecret)

	req, err := http.NewRequest(
		"POST",
		LineOAuthEndpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := Client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	type Body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var data Body
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	return data.AccessToken
}

//CallbackHandler コールバック処理
func CallbackHandler(w http.ResponseWriter, req *http.Request) {
	events, err := LineBotAPI.ParseRequest(req)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				//普通のテキストメッセージ
				ReplyToTextMessage(event, message)
			case *linebot.StickerMessage:
				//スタンプ
				ReplyToStickerMessage(event, message)
			case *linebot.LocationMessage:
				//位置情報
				ReplyToLocationMessage(event, message)
			}
		case linebot.EventTypeFollow:
			ReplyToFollowEvent(event)
		case linebot.EventTypeUnfollow:
			fmt.Printf("%v\n", event)
		case linebot.EventTypePostback:
			// Postbackのコマンド振り分け
			switch command := ParsePostbackData(event.Postback.Data); command.Type {
			case PostBackCommandTypeAnalyze:
				ReplyToPostbackAnalyze(event, &command)
			case PostBackCommandTypeHistory:
				ReplyToPostbackHistory(event, &command)
			case PostBackCommandTypeCommands:
				ReplyToPostbackCommand(event, &command)
			case PostBackCommandTypeFavoriteList:
				ReplyToPostbackFavList(event, &command)
			case PostBackCommandTypeFavorite:
				ReplyToPostbackFav(event, &command)
			case PostBackCommandTypeDatePicker:
				ReplyToPostbackDatePicker(event, &command)
			case PostBackCommandTypeConfigOpen:
				ReplyToPostbackConfigOpen(event, &command)
			case PostBackCommandTypeNotify:
				ReplyToPostbackNotifyConfig(event, &command)
			case PostBackCommandTypeStatus:
				ReplyToPostbackServiceStatus(event, &command)
			case PostBackCommandTypeRanking:
				ReplyToPostbackRanking(event, &command)
			}

		case linebot.EventTypeJoin:
		case linebot.EventTypeLeave:
		case linebot.EventTypeMemberJoined:
		case linebot.EventTypeMemberLeft:
		case linebot.EventTypeBeacon:
		case linebot.EventTypeAccountLink:
		case linebot.EventTypeThings:
		default:
			w.WriteHeader(400)
		}
	}
}

//NotifyHandler 通知指示
func NotifyHandler(w http.ResponseWriter, req *http.Request) {
	//パース
	req.ParseForm()
	params := req.Form
	userID := params.Get("user")
	SendScheduledNotify(userID)
}

//GetPlaceNameByCode コードから名前を返す
//ない場合は空文字を返す
func GetPlaceNameByCode(code string) (name string) {
	if val, ok := SpotNamesDictionary[code]; ok {
		name = val
	}
	return name
}

//SplitAreaSpot area-spotを切り離す
func SplitAreaSpot(code string) (area string, spot string) {
	arr := strings.Split(code, "-")
	if len(arr) >= 2 {
		area = arr[0]
		spot = arr[1]
	}
	return
}

func init() {
	ClientID = os.Getenv("LINE_CLIENT_ID")
	ClientSecret = os.Getenv("LINE_CLIENT_SECRET")
	AccessToken = getAccessToken()
	if bot, err := linebot.New(ClientSecret, AccessToken, linebot.WithHTTPClient(&Client)); err == nil {
		LineBotAPI = bot
	} else {
		panic(err)
	}
	BikeshareAPI = bikeshareapi.NewApiClient()
	BikeshareAPI.SetCertKey(os.Getenv("API_CERT"))
	if os.Getenv("MODE") == "DEBUG" {
		//デバッグ用
		BikeshareAPI.SetEndpoint("http://localhost:5001/")
	} else if os.Getenv("MODE") == "LOCAL" {
		//APIサーバと同じサーバにあるとき
		BikeshareAPI.SetEndpoint("http://apiserver:5001/")
	}

	//ユーザー設定を取得
	if err := CacheUsrConfigs(); err != nil {
		panic(err)
	}
	//スポット名の辞書を初期化
	places, err := BikeshareAPI.GetAllSpotNames()
	if err != nil {
		panic(err)
	}
	for _, place := range places {
		SpotNamesDictionary[place.Area+"-"+place.Spot] = place.Name
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	http.HandleFunc("/callback", CallbackHandler)
	http.HandleFunc("/notify", NotifyHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
