package main

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  概要：REST APIのサーバ
//
//　機能：1. スクレーパーからPOSTされた台数情報をDBに書き込む
//　　　　2. 公開APIを提供する
//
/////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"

	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/mattn/go-sqlite3"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  変数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//DB接続オブジェクト
var Db *sql.DB

//MasterSave 駐輪場情報構造体のキャッシュ
var MasterSave []rdb.Spotmaster

//JsonTimeLayout 時刻フォーマット
const JsonTimeLayout = "2006/01/02 15:04"

//ini_section セクション
const ini_section = "API"

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  エンドポイント関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//GetCounts 台数を返す公開API
func GetCounts(w rest.ResponseWriter, r *rest.Request) {
	//レスポンス用
	var jBody static.JCountsBody
	var jCounts []static.JCount
	//パース
	r.ParseForm()
	params := r.Form
	area := params.Get("area")
	spot := params.Get("spot")
	day := params.Get("day")

	if area == "" || spot == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("台数検索の際にはareaとspotの両方を指定する必要があります")
		return
	}

	rows, err := SearchCountsByDay(area, spot, day)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson(err.Error())
		return
	}

	//JSON構造体に変換
	for _, s := range rows {
		datetime := s.Time.Format(JsonTimeLayout)
		y := strconv.Itoa(s.Time.Year())
		m := strconv.Itoa(int(s.Time.Month()))
		d := strconv.Itoa(s.Time.Day())
		h := strconv.Itoa(s.Time.Hour())
		mi := strconv.Itoa(s.Time.Minute())
		jCounts = append(jCounts,
			static.JCount{Count: s.Count, Datetime: datetime, Year: y, Month: m, Day: d,
				Hour: h, Minute: mi})
	}
	jBody.Counts = jCounts
	//マスタ検索
	if master, err := GetSpotmasterFromCache(area, spot); err == nil {
		jBody.Area = master.Area
		jBody.Spot = master.Spot
		jBody.Description = master.Description
		jBody.Lat = master.Lat
		jBody.Lon = master.Lon
		jBody.Name = master.Name
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("マスターの検索に失敗しました")
		return
	}
	//返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(jBody)
}

//GetPlaces スポットマスタを返す公開API
func GetPlaces(w rest.ResponseWriter, r *rest.Request) {
	var jItems []static.JPlaces
	var jBody static.JPlacesBody
	//パース
	r.ParseForm()
	params := r.Form
	area := params.Get("area")
	spot := params.Get("spot")
	query := params.Get("q")
	addwhere := ""
	if query != "" {
		addwhere = fmt.Sprintf(" position( '%s' in trim(area) || '-' || trim(spot) || ',' || name || station ) > 0", query)
	}
	option := rdb.SearchOptions{Area: area, Spot: spot,
		AddWhere: addwhere}
	arr, err := rdb.SearchCurrentFull(Db, option)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("マスターの検索に失敗しました")
		return
	}
	//変換
	for _, view := range arr {
		recent := static.Recent{Count: view.Count, Datetime: view.Time.Format(JsonTimeLayout)}
		json := static.JPlaces{Area: view.Area, Spot: view.Spot, Name: view.Name,
			Lat: view.Lat, Lon: view.Lon, Description: view.Description,
			Recent: recent}
		jItems = append(jItems, json)
	}
	//返却
	jBody.Num = len(jItems)
	jBody.Items = jItems
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(jBody)
}

//GetAllPlaces 全てのスポットマスタを返す公開API
func GetAllPlaces(w rest.ResponseWriter, r *rest.Request) {
	var jBody static.JAllPlacesBody
	//マスタ全検索
	jBody.Num = len(MasterSave)

	//型変換
	for _, master := range MasterSave {
		var chiled static.JAllSpotChiled
		chiled.Area = master.Area
		chiled.Spot = master.Spot
		chiled.Name = master.Name
		jBody.Items = append(jBody.Items, chiled)
	}
	//返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(jBody)
}

//GetDistances 距離を返す公開API
func GetDistances(w rest.ResponseWriter, r *rest.Request) {
	var jItems []static.JDistances
	var jBody static.JDistancesBody
	//パース
	r.ParseForm()
	params := r.Form
	lat := params.Get("lat")
	lon := params.Get("lon")
	if lat == "" || lon == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("latとlonの両方を指定する必要があります")
		return
	}

	qry := `select 
	trim(area) as area 
	, trim(spot) as spot 
	, trim(name) as name 
	, trim(count) as count 
	, time  
	, lat 
	, lon 
	, description
	, trunc(  
	  sqrt(  
		pow(TO_NUMBER(lat, '99.999999') - %s, 2) + pow(TO_NUMBER(lon, '999.999999') - %s, 2) 
	  ) * 109133 
	  , 0 
	) as distance 
    from current_full order by distance limit 10`
	qry = fmt.Sprintf(qry, lat, lon)
	rows, err := Db.Query(qry)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("DBの検索に失敗しました")
		return
	}
	for rows.Next() {
		var area, spot, name, count, lat, lon, description string
		var time time.Time
		var distance int
		err := rows.Scan(&area, &spot, &name, &count, &time,
			&lat, &lon, &description, &distance)
		if err != nil {
			continue
		}
		recent := static.Recent{Count: count, Datetime: time.Format(JsonTimeLayout)}
		distanceStr := fmt.Sprintf("%d m", distance)
		jItems = append(jItems, static.JDistances{Area: area, Spot: spot, Name: name,
			Lat: lat, Lon: lon, Description: description, Distance: distanceStr, Recent: recent})
	}
	//返却
	jBody.Num = len(jItems)
	jBody.Items = jItems
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(jBody)
}

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

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  その他関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//checkHeader ヘッダ情報をチェックし秘密文字列の照合を行う
func checkHeader(r *rest.Request) bool {
	cert := r.Header.Get("cert")
	apiCert := os.Getenv("API_CERT")
	return (cert == apiCert)
}

//GetCacheSpotMaster Spotmasterをキャッシュする
func GetCacheSpotMaster() {
	//マスタ検索
	option := rdb.SearchOptions{OrderBy: "area,spot", AddWhere: "endtime is null"}
	if master, err := rdb.SearchSpotmaster(Db, option); err == nil {
		MasterSave = master
		logger.Infof("GetCacheSpotMaster マスタの取得に成功しました(%d件)", len(MasterSave))
	} else {
		MasterSave = []rdb.Spotmaster{}
		logger.Infof("GetCacheSpotMaster マスタの取得に失敗しました")
	}
}

//GetSpotmasterFromCache キャッシュしたデータからSpotmasterを探す
func GetSpotmasterFromCache(area string, spot string) (rdb.Spotmaster, error) {
	for _, s := range MasterSave {
		if s.Area == area && s.Spot == spot {
			return s, nil
		}
	}
	return rdb.Spotmaster{}, fmt.Errorf("area=%s, spot=%s nothing", area, spot)
}

//SearchCountsByDay 指定日(yyyymmdd)のデータを検索する（psql, SQLite振り分け）
func SearchCountsByDay(area, spot, day string) ([]rdb.Spotinfo, error) {
	var spotinfos []rdb.Spotinfo
	//検索条件作成
	option := rdb.SearchOptions{Area: area, Spot: spot, OrderBy: "time desc"}
	date, err := time.Parse("20060102", day)
	if err != nil {
		//ゼロ値で初期化
		date = time.Time{}
	}

	today := time.Now()
	//2日足して今日より未来ならpostgresにデータがある（dateはhhmmssがオール0のため）
	if date.AddDate(0, 0, 2).After(today) || date.IsZero() {
		if date.IsZero() {
			//日付未指定なら最新の1件のみ
			option.Limit = 1
		} else {
			option.AddWhere = fmt.Sprintf("date(time) = '%s'", date.Format("2006-01-02"))
		}
		analyzes, err := rdb.SearchAnalyze(Db, option)
		if err != nil {
			return spotinfos, err
		}
		//変換
		for _, anal := range analyzes {
			spotinfos = append(spotinfos, anal.ToSpotinfo())
		}
	} else {
		db, err := rdb.GetConnectionSQLite(date)
		if err != nil {
			return spotinfos, err
		}
		defer db.Close()
		//SQLiteから検索
		spotinfos = rdb.SearchSpotinfo(db, option)
	}

	return spotinfos, nil
}

func main() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		return
	}
	exeName := filer.GetExeName()
	logger.Info(exeName, "開始")
	defer logger.Info(exeName, "終了")

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
	router, err := rest.MakeRouter(
		rest.Get("/counts", GetCounts),
		rest.Get("/places", GetPlaces),
		rest.Get("/all_places", GetAllPlaces),
		rest.Get("/distances", GetDistances),
		rest.Get("/private/config", GetConfig),
		rest.Post("/private/counts", SetSpotinfo),
		rest.Post("/private/places", SetSpotMaster),
	)
	if err != nil {
		log.Fatal(err)
	}
	//DB接続
	Db, err = rdb.GetConnectionPsql()
	if err != nil {
		log.Fatal(err)
	}
	defer Db.Close()
	//起動時にキャッシュ
	GetCacheSpotMaster()

	//サーバ開始
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":5001", api.MakeHandler()))
}
