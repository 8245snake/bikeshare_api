package main

import (
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

//ParseComamnd パース
func ParseComamnd(data string) (postback PostBackCommand) {
	command := strings.Replace(strings.TrimSpace(data), "/", "", 1)
	postback.Type = PostBackCommandType(command)
	return
}

//CommandHandler コマンドを処理
func CommandHandler(event *linebot.Event, message *linebot.TextMessage) {
	command := ParseComamnd(message.Text)
	switch command.Type {
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
	case PostBackCommandTypeLacation:
		reply := linebot.NewTextMessage("現在メニューから位置情報検索ができません。\n↓にある「位置情報で検索」をタップしてください").WithQuickReplies(CreateQuickReplyItems())
		ReplyMessage(event.ReplyToken, reply)
	}
}
