package main

import (
	"sync"

	bikeshareapi "github.com/8245snake/bikeshare-client"
)

//UserUpdateType ユーザー情報更新タイプ
type UserUpdateType string

const (
	//UserUpdateTypeUserAdd ユーザー追加
	UserUpdateTypeUserAdd UserUpdateType = "a_user"
	//UserUpdateTypeHistory 履歴
	UserUpdateTypeHistory UserUpdateType = "u_history"
	//UserUpdateTypeFavorite お気に入り
	UserUpdateTypeFavorite UserUpdateType = "u_favorite"
	//UserUpdateTypeNotify 通知時刻
	UserUpdateTypeNotify UserUpdateType = "u_notify"
	//UserUpdateTypeHistoryDelete 履歴
	UserUpdateTypeHistoryDelete UserUpdateType = "d_history"
	//UserUpdateTypeFavoriteDelete お気に入り
	UserUpdateTypeFavoriteDelete UserUpdateType = "d_favorite"
	//UserUpdateTypeNotifyDelete 通知時刻
	UserUpdateTypeNotifyDelete UserUpdateType = "d_notify"
)

//UserConfigs ユーザー設定
var UserConfigs []bikeshareapi.Users

//CacheUsrConfigs ユーザー設定を変数に格納
func CacheUsrConfigs() error {
	//ユーザ情報をキャッシュ
	if user, err := BikeshareAPI.GetUsers(); err == nil {
		UserConfigs = user
	} else {
		return err
	}
	return nil
}

//GetUserConfigFromCache キャッシュから設定を取得（nilが返る可能性がある）
func GetUserConfigFromCache(userID string) *bikeshareapi.Users {
	for _, user := range UserConfigs {
		if user.LineID == userID {
			return &user
		}
	}
	return nil
}

//UpdateUserConfig ユーザー情報を更新
func UpdateUserConfig(updateType UserUpdateType, UsaerID string, value string) {
	//排他制御する
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	//ユーザー設定を取得
	user := GetUserConfigFromCache(UsaerID)
	if user == nil {
		user = &bikeshareapi.Users{LineID: UsaerID}
	}
	switch updateType {
	case UserUpdateTypeUserAdd:
		//なにもしない
	case UserUpdateTypeHistory:
		user.Histories = AddList(user.Histories, value, MaxHistory)
	case UserUpdateTypeNotify:
		user.Notifies = AddList(user.Notifies, value, MaxNotifyTimes)
	case UserUpdateTypeFavorite:
		user.Favorites = AddList(user.Favorites, value, MaxFavorite)
	case UserUpdateTypeHistoryDelete:
		user.Histories = RemoveList(user.Histories, value)
	case UserUpdateTypeNotifyDelete:
		user.Notifies = RemoveList(user.Notifies, value)
	case UserUpdateTypeFavoriteDelete:
		user.Favorites = RemoveList(user.Favorites, value)
	}
	//送信したらレスポンスのデータで内部変数を更新
	if users, err := BikeshareAPI.UpdateUser(*user); err == nil {
		UserConfigs = users
	}
}

//AddList 検索履歴を先頭に追加したスライスを返す
func AddList(slice []string, value string, max int) []string {
	if contains(slice, value) {
		//重複するならそのまま帰す
		return slice
	}
	//先頭に入れる
	buff := []string{value}
	for _, item := range slice {
		buff = append(buff, item)
	}
	//max件数を超えたら切り捨てる
	if len(buff) > max {
		buff = buff[:max]
	}
	return buff
}

//RemoveList スライスから指定した要素を削除する
func RemoveList(slice []string, value string) []string {
	if !contains(slice, value) {
		//無いならそのまま帰す
		return slice
	}
	//一致しないものだけ足していく
	buff := []string{}
	for _, item := range slice {
		if item != value {
			buff = append(buff, item)
		}
	}
	return buff
}

//contains 配列に要素が含まれているか判定
func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
