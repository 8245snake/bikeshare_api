package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"
	"github.com/ant0ine/go-json-rest/rest"
)

//GetConfig 設定を返す
func GetConfig(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}
	//レスポンス用
	var jBody static.JConfig

	//パース
	r.ParseForm()
	params := r.Form
	hostid := params.Get("hostid")
	//検索
	option := rdb.SearchOptions{AddWhere: "trim(hostid) in ('', '" + hostid + "')", OrderBy: "hostid"}
	configs, err := rdb.SearchConfig(Db, option)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("設定の検索に失敗しました")
		return
	}
	for _, config := range configs {
		if config.Key == "imgur_id" {
			jBody.ImgurID = config.Value
		} else if config.Key == "twitter_access_token" {
			jBody.TwitterAccessToken = config.Value
		} else if config.Key == "twitter_access_token_secret" {
			jBody.TwitterAccessTokenSecret = config.Value
		} else if config.Key == "twitter_consumer_key" {
			jBody.TwitterConsumerKey = config.Value
		} else if config.Key == "twitter_consumer_key_secret" {
			jBody.TwitterConsumerKeySecret = config.Value
		} else if config.Key == "client_id" {
			jBody.ClientID = config.Value
		} else if config.Key == "client_secret" {
			jBody.ClientSecret = config.Value
		} else if config.Key == "channel_secret" {
			jBody.ChannelSecret = config.Value
		}
	}
	//返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(jBody)
}

//SetSpotinfo スクレイパーからのPOSTに対応
func SetSpotinfo(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}
	body := static.JSpotinfo{}
	if err := r.DecodeJsonPayload(&body); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//body.Spotinfo をconn.Spotinfoの配列にする
	var rows []rdb.Spotinfo
	for _, row := range body.Spotinfo {
		time, _ := time.Parse(rdb.TimeLayout, row.Time)
		temp := rdb.Spotinfo{Area: row.Area, Spot: row.Spot, Time: time, Count: row.Count}
		rows = append(rows, temp)
	}

	if _, err := rdb.BulkInsertSpotinfo(Db, rows); err != nil {
		Db.Close()
		Db, err = rdb.GetConnectionPsql()
	}
}

//SetSpotMaster スクレイパーからのPOSTに対応
func SetSpotMaster(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}
	body := static.JSpotmaster{}
	if err := r.DecodeJsonPayload(&body); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//body.Spotmaster をrdb.Spotmasterの配列にする
	var rows []rdb.Spotmaster
	for _, row := range body.Spotmaster {
		rows = append(rows, rdb.Spotmaster{Area: strings.TrimSpace(row.Area),
			Spot: strings.TrimSpace(row.Spot),
			Name: strings.TrimSpace(row.Name),
			Lat:  strings.TrimSpace(row.Lat),
			Lon:  strings.TrimSpace(row.Lon)})
	}

	//更新があるかチェック
	var updateList []rdb.Spotmaster
	now := time.Now()
	//ミリ秒はいらない
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local)
	for _, row := range rows {
		olds, err := GetSpotmasterFromCache(row.Area, row.Spot)
		if err != nil {
			//見つからなかったら無条件で追加
			row.Starttime = now
			updateList = append(updateList, row)
		} else {
			//見つかったら名前を比較し変わっていたら更新
			if row.Name != olds.Name {
				row.Starttime = now
				updateList = append(updateList, row)
				//旧データは－1秒して更新
				olds.Endtime = now.Add(-1 * time.Second)
				updateList = append(updateList, olds)
			}
		}
	}
	//Upsert
	for _, item := range updateList {
		err := rdb.UpsertSpotmaster(Db, item)
		if err != nil {
			logger.Debugf("UpsertSpotmaster_Error %v \n", err)
		}
	}
	//キャッシュ最新化
	GetCacheSpotMaster()
}

//GetUser ユーザー設定を返す
func GetUser(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}
	users, err := rdb.GetAllUsers(Db)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type Users struct {
		Users []rdb.User `json:"users"`
	}

	//返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(Users{users})
}

//UpdateUser ユーザー設定を更新する
func UpdateUser(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}

	body := rdb.User{}
	if err := r.DecodeJsonPayload(&body); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := rdb.UpsertUser(Db, &body); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//最新情報を返す
	users, err := rdb.GetAllUsers(Db)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type Users struct {
		Users []rdb.User `json:"users"`
	}

	//返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(Users{users})
}
