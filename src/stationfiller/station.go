package main

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  概要：スポットマスタの駅名を補完するタスク
//
//　機能：1. 駅名補完
//　　　　2. 説明補完
//
/////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"

	_ "github.com/lib/pq"
)

//Station 駅検索APIの駅データ
type Station struct {
	Name       string  `json:"name"`
	Prefecture string  `json:"prefecture"`
	Line       string  `json:"line"`
	Lon        float32 `json:"x"`
	Lat        float32 `json:"y"`
	Postal     string  `json:"postal"`
	Distance   string  `json:"distance"`
	Prev       string  `json:"prev"`
	Next       string  `json:"next"`
}

//GetDescription 説明を出力
func (s Station) GetDescription() string {
	return fmt.Sprintf("%s「%s駅」から%s。", s.Line, s.Name, s.Distance)
}

//Heartrails 駅検索API 親オブジェクト
type Heartrails struct {
	Response Response `json:"response"`
}

//Response 駅検索API 中間オブジェクト
type Response struct {
	Stations []Station `json:"station"`
}

//GetDescriptions 説明を出力
func (r Response) GetDescriptions() string {
	var description string
	for _, station := range r.Stations {
		description += station.GetDescription()
	}
	return description
}

//GetStations 駅名をカンマ区切りで返す
func (r Response) GetStations() (result string) {

	if len(r.Stations) > 0 {
		result = r.Stations[0].Name
	}
	for i := 1; i < len(r.Stations)-1; i++ {
		result += "," + r.Stations[i].Name
	}
	return
}

//FillStationName 駅名補完
func FillStationName(db *sql.DB) {
	opt := rdb.SearchOptions{AddWhere: "endtime is null and trim(description) is null "}
	rows, err := rdb.SearchSpotmaster(db, opt)
	if err != nil {
		return
	}
	var station Heartrails
	for _, row := range rows {
		station, err = requestStationInfo(row.Lon, row.Lat)
		if err != nil {
			continue
		}
		row.Description = station.Response.GetDescriptions()
		row.Station = station.Response.GetStations()
		err = rdb.UpsertSpotmaster(db, row)
		if err != nil {
			continue
		}
	}
}

//requestStationInfo 座標を渡して駅情報を取得する
func requestStationInfo(lon string, lat string) (Heartrails, error) {
	// form values
	values := url.Values{}
	values.Add("x", lon)
	values.Add("y", lat)
	values.Encode()

	var data Heartrails

	res, err := http.PostForm("http://express.heartrails.com/api/json?method=getStations", values)
	if err != nil {
		return data, err
	}

	// リクエスト送信
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}

	// JSONデコード
	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return data, nil
}

//startStationConllector 駅名補完メイン関数 無限ループ
func startStationConllector() {
	//待ち時間取得
	interval := filer.GetIniDataInt("STATION", "INTERVAL", 24)
	for {
		db, err := rdb.GetConnectionPsql()
		if err != nil {
			goto LBL_CONTINUE
		}
		defer db.Close()

		//補完処理実行
		FillStationName(db)

	LBL_CONTINUE: //一回ごとにDB接続を切る
		db.Close()
		time.Sleep(time.Hour * time.Duration(interval))
	}
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

	//タスク開始
	go startStationConllector()
	//プロセスが終了しないように無限ループとする
	for {
		time.Sleep(time.Minute * 60)
	}
}
