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
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"

	"github.com/ant0ine/go-json-rest/rest"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  変数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//DB接続オブジェクト
var Db *sql.DB

// var lock = sync.RWMutex{}

//JsonTimeLayout 時刻フォーマット
const JsonTimeLayout = "2006/01/02 15:04"

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  エンドポイント関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//GetTest DB接続状態を返す
func GetTest(w rest.ResponseWriter, r *rest.Request) {

}

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
	//台数情報取得
	option := rdb.SearchOptions{Area: area, Spot: spot, OrderBy: "time desc"}
	if day != "" {
		if dttm, err := time.Parse("20060102", day); err == nil {
			option.AddWhere = fmt.Sprintf("date(time) = '%s'", dttm.Format("2006-01-02"))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteJson("dayの形式が不正です")
			return
		}
	} else {
		option.Limit = 1
	}
	rows, err := rdb.SearchAnalyze(Db, option)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("検索に失敗しました")
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
	option = rdb.SearchOptions{Area: area, Spot: spot, AddWhere: "endtime is null", Limit: 1}
	if master, err := rdb.SearchSpotmaster(Db, option); err == nil && len(master) > 0 {
		jBody.Area = master[0].Area
		jBody.Spot = master[0].Spot
		jBody.Description = master[0].Description
		jBody.Lat = master[0].Lat
		jBody.Lon = master[0].Lon
		jBody.Name = master[0].Name
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

}

//GetAllPlaces 全てのスポットマスタを返す公開API
func GetAllPlaces(w rest.ResponseWriter, r *rest.Request) {
	var jBody static.JAllPlacesBody
	//マスタ全検索
	option := rdb.SearchOptions{OrderBy: "area,spot", AddWhere: "endtime is null"}
	masters, err := rdb.SearchSpotmaster(Db, option)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("マスターの検索に失敗しました")
		return
	}
	jBody.Num = len(masters)

	//型変換
	for _, master := range masters {
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

	if err := rdb.BulkInsertSpotinfo(Db, rows); err != nil {
		Db.Close()
		Db, err = rdb.GetConnectionPsql()
	}
}

//SetSpotMaster スクレイパーからのPOSTに対応
func SetSpotMaster(w rest.ResponseWriter, r *rest.Request) {
	if !checkHeader(r) {
		return
	}
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

func main() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		return
	}
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
		rest.Get("/test", GetTest),
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
	Db, err = rdb.GetConnectionPsql()
	defer Db.Close()

	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":5001", api.MakeHandler()))
}
