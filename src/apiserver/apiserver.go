package main

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  概要：REST APIのサーバ
//
//　機能：1. スクレーパーからPOSTされた台数情報をDBに書き込む
//　　　　2. WEB APIを提供する
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
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"

	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/mattn/go-sqlite3"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  変数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//Db DB接続オブジェクト
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

	rows, err := rdb.SearchCountsByDay(Db, area, spot, day)
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

func init() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		return
	}
	logger.Info(filer.GetExeName(), "開始")
	//DB接続
	Db, err = rdb.GetConnectionPsql()
	if err != nil {
		log.Fatal(err)
	}
	//起動時にキャッシュ
	GetCacheSpotMaster()
}

func main() {
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
		rest.Get("/private/users", GetUser),
		rest.Post("/private/counts", SetSpotinfo),
		rest.Post("/private/places", SetSpotMaster),
		rest.Post("/private/user", UpdateUser),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer Db.Close()

	//サーバ開始
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":5001", api.MakeHandler()))
}
