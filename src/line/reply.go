package main

import (
	"fmt"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

//ReplyMessage 返信用共通関数
func ReplyMessage(replyToken string, message linebot.SendingMessage) error {
	//_, err := LineBotAPI.ReplyMessage(replyToken, message.WithQuickReplies(CreateQuickReplyItems())).Do()
	_, err := LineBotAPI.ReplyMessage(replyToken, message).Do()
	if err != nil {
		//だめかもしれないけどとりあえずエラーメッセージの再送を試みる
		ReplyMessage(replyToken, linebot.NewTextMessage(err.Error()))
	}
	return err
}

//ReplyToFollowEvent フォローされたとき
func ReplyToFollowEvent(event *linebot.Event) {
	//ユーザー登録
	UpdateUserConfig(UserUpdateTypeUserAdd, event.Source.UserID, "")
	//返信
	ReplyMessage(event.ReplyToken, linebot.NewTextMessage("フォローありがとうございます！\n駐輪場の名前を入力してみてください"))
}

//ReplyToTextMessage テキストメッセージへの返信
func ReplyToTextMessage(event *linebot.Event, message *linebot.TextMessage) {
	replyToken := event.ReplyToken
	text := message.Text

	switch 1 {
	default:
		if strings.Index(text, "/") == 0 {
			//スラッシュコマンド
			CommandHandler(event, message)
			break
		}
		//その他のメッセージは駐輪場検索とする
		reply := MakeSpotListMessage(text)
		ReplyMessage(replyToken, reply)

		// 検索履歴は駐輪場検索のみ保存する
		UpdateUserConfig(UserUpdateTypeHistory, event.Source.UserID, text)
	}
}

//ReplyToStickerMessage スタンプへの返信
func ReplyToStickerMessage(event *linebot.Event, message *linebot.StickerMessage) {
	fmt.Printf("StickerID=%s\n", message.StickerID)
	replyToken := event.ReplyToken
	//適当なスタンプを返す
	reply := linebot.NewStickerMessage("11537", "52002734")
	ReplyMessage(replyToken, reply)
}

//ReplyToLocationMessage 位置情報メッセージへの返信
func ReplyToLocationMessage(event *linebot.Event, message *linebot.LocationMessage) {
	replyToken := event.ReplyToken
	reply := MakeSpotListMessageForLocation(message.Latitude, message.Longitude)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackAnalyze グラフ表示
func ReplyToPostbackAnalyze(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeAnalysisMessage(command.Area, command.Spot, command.Span, event.Source.UserID)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackCommand コマンド一覧の表示
func ReplyToPostbackCommand(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeCommandListMessage()
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackHistory 履歴表示
func ReplyToPostbackHistory(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeHistryListMessage(event.Source.UserID)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackServiceStatus サービス稼働状況の表示
func ReplyToPostbackServiceStatus(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeServiceStatusMessage()
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackFavList お気に入り一覧表示
func ReplyToPostbackFavList(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeFavriteListMessage(event.Source.UserID)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackRanking ランキング表示
func ReplyToPostbackRanking(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeRankingMessage(20)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackDatePicker 日付検索
func ReplyToPostbackDatePicker(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	day := strings.Replace(event.Postback.Params.Date, "-", "", -1)
	reply := MakeDateAnalysisMessage(command.Area, command.Spot, event.Source.UserID, day)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackConfigOpen 設定画面呼び出し
func ReplyToPostbackConfigOpen(event *linebot.Event, command *PostBackCommand) {
	replyToken := event.ReplyToken
	reply := MakeDateConfigWindowMessage(event.Source.UserID)
	ReplyMessage(replyToken, reply)
}

//ReplyToPostbackFav お気に入り登録
func ReplyToPostbackFav(event *linebot.Event, command *PostBackCommand) {
	var reply linebot.SendingMessage
	user := GetUserConfigFromCache(event.Source.UserID)
	if user == nil {
		reply = linebot.NewTextMessage("ユーザー設定の読み込みに失敗しました")
		ReplyMessage(event.ReplyToken, reply)
		return
	}
	// 検索履歴登録
	code := command.Area + "-" + command.Spot
	userID := event.Source.UserID
	switch command.Mode {
	case PostBackCommandModeReg:
		if len(user.Favorites) >= MaxFavorite {
			reply = linebot.NewTextMessage("これ以上お気に入りを登録できません")
			break
		}
		//登録
		UpdateUserConfig(UserUpdateTypeFavorite, userID, code)
		reply = MakeDateConfigWindowMessage(userID)
	case PostBackCommandModeUnreg:
		if len(user.Favorites) < 1 {
			reply = linebot.NewTextMessage("お気に入りを削除できません")
			break
		}
		UpdateUserConfig(UserUpdateTypeFavoriteDelete, userID, code)
		reply = MakeDateConfigWindowMessage(userID)
	}
	//返信
	ReplyMessage(event.ReplyToken, reply)
}

//ReplyToPostbackNotifyConfig 通知時刻編集
func ReplyToPostbackNotifyConfig(event *linebot.Event, command *PostBackCommand) {
	var reply linebot.SendingMessage
	user := GetUserConfigFromCache(event.Source.UserID)
	if user == nil {
		reply = linebot.NewTextMessage("ユーザー設定の読み込みに失敗しました")
		ReplyMessage(event.ReplyToken, reply)
		return
	}

	userID := event.Source.UserID
	switch command.Mode {
	case PostBackCommandModeReg:
		if len(user.Notifies) >= MaxNotifyTimes {
			reply = linebot.NewTextMessage("これ以上時刻を登録できません")
			break
		}
		target := event.Postback.Params.Time
		UpdateUserConfig(UserUpdateTypeNotify, userID, target)
		reply = MakeDateConfigWindowMessage(userID)
	case PostBackCommandModeUnreg:
		if len(user.Notifies) < 1 {
			reply = linebot.NewTextMessage("時刻を削除できません")
			break
		}
		target := command.Target
		UpdateUserConfig(UserUpdateTypeNotifyDelete, userID, target)
		reply = MakeDateConfigWindowMessage(userID)
	}
	//返信
	ReplyMessage(event.ReplyToken, reply)
}

//SendScheduledNotify 通知を送信する
func SendScheduledNotify(userID string) {
	message := MakeFavriteListMessage(userID)
	switch message.(type) {
	case *linebot.FlexMessage:
		//_, err := LineBotAPI.PushMessage(userID, message.WithQuickReplies(CreateQuickReplyItems())).Do()
		_, err := LineBotAPI.PushMessage(userID, message).Do()
		fmt.Printf("%v\n", err)
	case *linebot.TextMessage:
		//バブルコンテナの作成に失敗したときなので何もしない
		return
	}
}
