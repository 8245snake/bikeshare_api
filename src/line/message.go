package main

import (
	"fmt"

	bikeshareapi "github.com/8245snake/bikeshare-client"
	"github.com/8245snake/bikeshare_api/src/lib/static"
	"github.com/line/line-bot-sdk-go/linebot"
)

//MakeConfirmMessage 確認ダイアログ
func MakeConfirmMessage() *linebot.TemplateMessage {
	leftBtn := linebot.NewMessageAction("left", "left clicked")
	rightBtn := linebot.NewMessageAction("right", "right clicked")
	template := linebot.NewConfirmTemplate("Hello World", leftBtn, rightBtn)
	return linebot.NewTemplateMessage("Sorry :(, please update your app.", template)
}

//CreateQuickReplyItems クイックリプライを作成
func CreateQuickReplyItems() *linebot.QuickReplyItems {
	items := linebot.NewQuickReplyItems()
	// items.Items = append(items.Items, linebot.NewQuickReplyButton("https://i.imgur.com/UdEkcB7.png", linebot.NewPostbackAction("お気に入り", GetPostbackDataFavoriteList(), "", "")))
	// items.Items = append(items.Items, linebot.NewQuickReplyButton("https://i.imgur.com/A5au5SF.png", linebot.NewPostbackAction("履歴", GetPostbackDataForHistory(), "", "")))
	// items.Items = append(items.Items, linebot.NewQuickReplyButton("https://i.imgur.com/UdEkcB7.png", linebot.NewPostbackAction("コマンド", GetPostbackDataForCommands(), "", "")))
	items.Items = append(items.Items, linebot.NewQuickReplyButton("", linebot.NewLocationAction("位置情報で検索")))
	return items
}

//MakeServiceStatusMessage テンプレートメッセージ
func MakeServiceStatusMessage() linebot.SendingMessage {
	status, err := BikeshareAPI.GetStatus()
	if err != nil {
		return linebot.NewTextMessage("APIとの通信に失敗しています")
	}

	message := linebot.NewTextMessage("システムは正常に稼働しています")

	if status.Status == static.StatusOK {
		message = linebot.NewTextMessage("システムは正常に稼働しています")
	} else {
		if status.Connection != static.StatusOK {
			message = linebot.NewTextMessage("DBとの接続が切れています")
		}
		if status.Scraping != static.StatusOK {
			message = linebot.NewTextMessage("台数データの取得に失敗しています")
		}
	}
	return message
}

//MakeSpotListMessageForLocation 位置情報への返信
func MakeSpotListMessageForLocation(lat, lon float64) linebot.SendingMessage {
	distances, err := BikeshareAPI.GetDistances(bikeshareapi.SearchDistanceOption{Lat: lat, Lon: lon})
	if err != nil {
		return linebot.NewTextMessage("検索に失敗しました")
	}
	var spotinfos []bikeshareapi.SpotInfo
	for _, place := range distances.Spots {
		info := bikeshareapi.SpotInfo{
			Area:   place.SpotInfo.Area,
			Spot:   place.SpotInfo.Spot,
			Name:   place.SpotInfo.Name + "\n" + place.Distance,
			Counts: place.SpotInfo.Counts,
		}
		spotinfos = append(spotinfos, info)
	}
	title := "位置情報検索結果"
	container := CreateSpotListBubbleContainer(title, "近いスポットを10件表示します", spotinfos)
	return linebot.NewFlexMessage(title, &container)
}

//MakeSpotListMessage テンプレートメッセージ
func MakeSpotListMessage(query string) linebot.SendingMessage {
	var reply linebot.SendingMessage
	//件数によってテンプレートを振り分ける
	spotinfos, err := BikeshareAPI.GetPlaces(bikeshareapi.SearchPlacesOption{Query: query})
	if err != nil {
		reply = linebot.NewTextMessage("駐輪場の検索に失敗しました")
		return reply
	}
	count := len(spotinfos)
	title := fmt.Sprintf("「%s」を含むスポットが%d件見つかりました", query, count)

	if count == 0 {
		reply = linebot.NewTextMessage(fmt.Sprintf("「%s」を含むスポットが見つかりませんでした", query))
	} else if count < 20 {
		container := CreateSpotListBubbleContainer(title, "検索結果を表示します", spotinfos)
		reply = linebot.NewFlexMessage(title, &container)
	} else if count < 100 {
		container := CreateSpotListCarouselContainer(title, "検索結果を表示します", spotinfos)
		reply = linebot.NewFlexMessage(title, &container)
	} else {
		reply = linebot.NewTextMessage(fmt.Sprintf("「%s」を含むスポットが多すぎて表示できませんでした(%d件)\n検索クエリを変えてください", query, count))
	}
	return reply
}

//MakeFavriteListMessage テンプレートメッセージ
func MakeFavriteListMessage(userID string) linebot.SendingMessage {
	user := GetUserConfigFromCache(userID)
	if user == nil {
		return linebot.NewTextMessage("ユーザ設定が読み込まれませんでした")
	}
	if len(user.Favorites) < 1 {
		return linebot.NewTextMessage("お気に入りがまだ登録されていません")
	}
	spotinfos, err := BikeshareAPI.GetPlaces(bikeshareapi.SearchPlacesOption{Places: user.Favorites})
	if err != nil {
		return linebot.NewTextMessage("検索に失敗しました")
	}
	if len(spotinfos) < 1 {
		return linebot.NewTextMessage("お気に入り登録したスポットがありません。")
	}
	title := "お気に入り登録されたスポットを表示します"
	var reply linebot.SendingMessage
	container := CreateSpotListBubbleContainer(title, "検索結果を表示します", spotinfos)
	reply = linebot.NewFlexMessage(title, &container)
	return reply
}

//MakeRankingMessage ランキング
func MakeRankingMessage(limit int) linebot.SendingMessage {
	spotinfos, err := BikeshareAPI.GetPlaces(bikeshareapi.SearchPlacesOption{Sort: "countd", Limit: limit})
	if err != nil {
		return linebot.NewTextMessage("検索に失敗しました")
	}
	count := len(spotinfos)
	title := fmt.Sprintf("台数が多いスポットTop %d を表示します", count)
	var reply linebot.SendingMessage
	if count == 0 {
		reply = linebot.NewTextMessage("検索結果が0件でした")
	} else if count < 20 {
		container := CreateSpotListBubbleContainer(title, "検索結果を表示します", spotinfos)
		reply = linebot.NewFlexMessage(title, &container)
	} else if count < 100 {
		container := CreateSpotListCarouselContainer(title, "検索結果を表示します", spotinfos)
		reply = linebot.NewFlexMessage(title, &container)
	} else {
		reply = linebot.NewTextMessage("検索結果が多すぎます")
	}
	return reply
}

//MakeAnalysisMessage グラフ表示メッセージの作成
func MakeAnalysisMessage(area string, spot string, span int, userID string) linebot.SendingMessage {
	option := bikeshareapi.SearchGraphOption{
		Area:        area,
		Spot:        spot,
		Property:    "500,380",
		UploadImgur: false,
	}
	graph, err := BikeshareAPI.GetGraph(option)
	if err != nil {
		return linebot.NewTextMessage("グラフの作成に失敗しました")
	}

	//お気に入り登録/解除の判定
	user := GetUserConfigFromCache(userID)
	if user == nil {
		return linebot.NewTextMessage("user設定が読み込まれませんでした")
	}
	param := TemplateMessageParameter{
		Area:             area,
		Spot:             spot,
		Title:            graph.Title,
		URL:              graph.URL,
		Description:      graph.SpotInfo.Description,
		LastUpdate:       getLastUpdateTime(graph.SpotInfo),
		RegButtonVisible: !contains(user.Favorites, area+"-"+spot),
	}
	container := CreateAnalysisBubbleContainer(param)
	reply := linebot.NewFlexMessage(param.Title, &container)
	return reply
}

//MakeCommandListMessage  コマンド一覧表示メッセージの作成
func MakeCommandListMessage() linebot.SendingMessage {
	list := []CommandListItem{
		{ActionType: linebot.ActionTypePostback, Label: "台数ランキング", Data: GetPostbackDataRanking(), Text: "台数が多い順にスポットを表示します"},
		{ActionType: linebot.ActionTypePostback, Label: "設定", Data: GetPostbackDataConfigOpen(), Text: "設定画面を開きます"},
		// {ActionType: linebot.ActionTypePostback, Label: "Slack連携", Data: "slack"},
		{ActionType: linebot.ActionTypePostback, Label: "システム障害状況", Data: GetPostbackDataServiceStatus(), Text: "稼働状況の確認中です..."},
	}
	container := CreateCommandListBubbleContainer("コマンド一覧です", list)
	reply := linebot.NewFlexMessage("コマンド一覧を表示します", &container)
	return reply
}

//MakeHistryListMessage  履歴一覧表示メッセージの作成
func MakeHistryListMessage(userID string) linebot.SendingMessage {
	var list []CommandListItem
	user := GetUserConfigFromCache(userID)
	if user == nil {
		reply := linebot.NewTextMessage("履歴がありません")
		return reply
	}
	for _, history := range user.Histories {
		list = append(list, CommandListItem{ActionType: linebot.ActionTypeMessage, Label: history, Data: history})
	}
	container := CreateCommandListBubbleContainer("履歴の一覧を表示します", list)
	reply := linebot.NewFlexMessage("検索履歴を10件まで表示します", &container)
	return reply
}

//MakeDateAnalysisMessage 任意の日付のグラフ表示メッセージの作成
func MakeDateAnalysisMessage(area string, spot string, userID string, days ...string) linebot.SendingMessage {
	option := bikeshareapi.SearchGraphOption{
		Area:        area,
		Spot:        spot,
		Property:    "500,380",
		UploadImgur: false,
		Days:        days,
	}
	graph, err := BikeshareAPI.GetGraph(option)
	if err != nil {
		return linebot.NewTextMessage("グラフの作成に失敗しました")
	}

	//お気に入り登録/解除の判定
	user := GetUserConfigFromCache(userID)
	if user == nil {
		return linebot.NewTextMessage("user設定が読み込まれませんでした")
	}
	param := TemplateMessageParameter{
		Area:             area,
		Spot:             spot,
		Title:            graph.Title,
		URL:              graph.URL,
		RegButtonVisible: !contains(user.Favorites, area+"-"+spot),
	}
	container := CreateAnalysisBubbleContainer(param)
	reply := linebot.NewFlexMessage(param.Title, &container)
	return reply
}

//MakeDateConfigWindowMessage 設定画面メッセージ作成
func MakeDateConfigWindowMessage(userID string) linebot.SendingMessage {
	var reply linebot.SendingMessage
	user := GetUserConfigFromCache(userID)
	if user == nil {
		reply = linebot.NewTextMessage("ユーザー設定の読み込みに失敗しました")
		return reply
	}
	container := CreateConfigBubbleContainer(user)
	reply = linebot.NewFlexMessage("設定画面", &container)
	return reply
}
