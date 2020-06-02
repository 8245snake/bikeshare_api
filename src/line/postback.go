package main

import (
	"fmt"
	"strconv"
	"strings"
)

//PostBackCommand ポストバックデータ
type PostBackCommand struct {
	Type                      PostBackCommandType
	Mode                      PostBackCommandMode
	Area, Spot, Target, Value string
	Span                      int
}

//PostBackElement ポストバックのDataに含まれるパラメータに種類
type PostBackElement string

const (
	//PostBackElementCommand コマンドのタイプ
	PostBackElementCommand PostBackElement = "command"
	//PostBackElementArea エリアコードに相当
	PostBackElementArea PostBackElement = "area"
	//PostBackElementSpot スポットコードに相当
	PostBackElementSpot PostBackElement = "spot"
	//PostBackElementSpan 表示期間（標準は２日間）
	PostBackElementSpan PostBackElement = "span"
	//PostBackElementMode モード（登録/解除）
	PostBackElementMode PostBackElement = "mode"
	//PostBackElementValue お気に入り登録に使用
	PostBackElementValue PostBackElement = "value"
	//PostBackElementTarget お気に入り削除に使用
	PostBackElementTarget PostBackElement = "targer"
)

//PostBackCommandType コマンドの種類
type PostBackCommandType string

const (
	//PostBackCommandTypeAnalyze 経時変化グラフの表示
	PostBackCommandTypeAnalyze PostBackCommandType = "analysis"
	//PostBackCommandTypeFavorite お気に入り登録or解除
	PostBackCommandTypeFavorite PostBackCommandType = "favorite"
	//PostBackCommandTypeFavoriteList お気に入り一覧の表示
	PostBackCommandTypeFavoriteList PostBackCommandType = "favlist"
	//PostBackCommandTypeDatePicker 日付情報送信
	PostBackCommandTypeDatePicker PostBackCommandType = "date"
	//PostBackCommandTypeTimePicker 時刻情報送信
	PostBackCommandTypeTimePicker PostBackCommandType = "time"
	//PostBackCommandTypeConfigOpen 設定画面表示
	PostBackCommandTypeConfigOpen PostBackCommandType = "config"
	//PostBackCommandTypeConfigMod 設定変更
	PostBackCommandTypeConfigMod PostBackCommandType = "configModify"
	//PostBackCommandTypeNotify 通知時刻編集
	PostBackCommandTypeNotify PostBackCommandType = "notify"
	//PostBackCommandTypeCommands コマンド一覧の表示
	PostBackCommandTypeCommands PostBackCommandType = "commands"
	//PostBackCommandTypeHistory 履歴の表示
	PostBackCommandTypeHistory PostBackCommandType = "history"
	//PostBackCommandTypeRanking 台数ランキング
	PostBackCommandTypeRanking PostBackCommandType = "ranking"
	//PostBackCommandTypeLacation 位置情報で検索
	PostBackCommandTypeLacation PostBackCommandType = "location"
	//PostBackCommandTypeSlack Slack連携
	PostBackCommandTypeSlack PostBackCommandType = "slack"
	//PostBackCommandTypeStatus システム障害状況
	PostBackCommandTypeStatus PostBackCommandType = "system"
)

//PostBackCommandMode モード（登録/解除）お気に入りに使用
type PostBackCommandMode string

const (
	//PostBackCommandModeReg 登録
	PostBackCommandModeReg PostBackCommandMode = "reg"
	//PostBackCommandModeUnreg 解除
	PostBackCommandModeUnreg PostBackCommandMode = "unreg"
)

//ParsePostbackData パース
func ParsePostbackData(data string) (postback PostBackCommand) {
	keyvalArr := strings.Split(data, "_")
	for _, keyval := range keyvalArr {
		key, val := splitKeyVal(keyval)
		switch PostBackElement(key) {
		case PostBackElementCommand:
			postback.Type = PostBackCommandType(val)
		case PostBackElementArea:
			postback.Area = val
		case PostBackElementSpot:
			postback.Spot = val
		case PostBackElementTarget:
			postback.Target = val
		case PostBackElementValue:
			postback.Value = val
		case PostBackElementMode:
			postback.Mode = PostBackCommandMode(val)
		case PostBackElementSpan:
			if span, err := strconv.Atoi(val); err == nil {
				postback.Span = span
			}
		}
	}
	return
}

//Serialize パラメータを直列化
func (pb *PostBackCommand) Serialize() string {
	params := []string{}
	if pb.Type != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementCommand, pb.Type))
	}
	if pb.Area != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementArea, pb.Area))
	}
	if pb.Spot != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementSpot, pb.Spot))
	}
	if pb.Target != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementTarget, pb.Target))
	}
	if pb.Value != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementValue, pb.Value))
	}
	if pb.Mode != "" {
		params = append(params, fmt.Sprintf("%s=%s", PostBackElementMode, pb.Mode))
	}
	if pb.Span != 0 {
		params = append(params, fmt.Sprintf("%s=%d", PostBackElementSpan, pb.Span))
	}
	return strings.Join(params, "_")
}

func splitKeyVal(keyval string) (string, string) {
	keys := strings.Split(keyval, "=")
	if len(keys) != 2 {
		return "", ""
	}
	key := keys[0]
	val := keys[1]
	return key, val
}

//GetPostbackDataForAnalyze グラフ要求用ポストバック文字列
func GetPostbackDataForAnalyze(area string, spot string, span int) string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeAnalyze,
		Area: area,
		Spot: spot,
		Span: span,
	}
	return postback.Serialize()
}

//GetPostbackDataForDateAnalyze グラフ要求用ポストバック文字列
func GetPostbackDataForDateAnalyze(area string, spot string) string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeDatePicker,
		Area: area,
		Spot: spot,
	}
	return postback.Serialize()
}

//GetPostbackDataForCommands コマンド一覧ポストバック文字列
func GetPostbackDataForCommands() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeCommands,
	}
	return postback.Serialize()
}

//GetPostbackDataForHistory 履歴一覧ポストバック文字列
func GetPostbackDataForHistory() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeHistory,
	}
	return postback.Serialize()
}

//GetPostbackDataFavoriteList お気に入り一覧ポストバック文字列
func GetPostbackDataFavoriteList() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeFavoriteList,
	}
	return postback.Serialize()
}

//GetPostbackDataServiceStatus サービス稼働状況ポストバック文字列
func GetPostbackDataServiceStatus() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeStatus,
	}
	return postback.Serialize()
}

//GetPostbackDataRanking 台数ランキング取得ポストバック文字列
func GetPostbackDataRanking() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeRanking,
	}
	return postback.Serialize()
}

//GetPostbackDataConfigOpen 設定画面表示ポストバック文字列
func GetPostbackDataConfigOpen() string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeConfigOpen,
	}
	return postback.Serialize()
}

//GetPostbackDataForFovarite お気に入り登録用ポストバック文字列
func GetPostbackDataForFovarite(area string, spot string, mode PostBackCommandMode) string {
	postback := PostBackCommand{
		Type: PostBackCommandTypeFavorite,
		Area: area,
		Spot: spot,
		Mode: mode,
	}
	return postback.Serialize()
}

//GetPostbackDataForNotify 通知時刻登録用ポストバック文字列
func GetPostbackDataForNotify(mode PostBackCommandMode, targetTime string) string {
	postback := PostBackCommand{
		Type:   PostBackCommandTypeNotify,
		Target: targetTime,
		Mode:   mode,
	}
	return postback.Serialize()
}
