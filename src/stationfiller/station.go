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
	"runtime"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/carlescere/scheduler"

	_ "github.com/lib/pq"
)

var ini_section = "STATION"

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
func (r Response) GetDescriptions() (description string, err error) {
	for _, station := range r.Stations {
		description += station.GetDescription()
	}
	if description == "" {
		return description, fmt.Errorf("descriptionの生成に失敗")
	}
	return
}

//GetStations 駅名をカンマ区切りで返す
func (r Response) GetStations() (result string, err error) {
	if len(r.Stations) > 0 {
		result = r.Stations[0].Name
	}
	for i := 1; i < len(r.Stations)-1; i++ {
		result += "," + r.Stations[i].Name
	}
	if result == "" {
		return result, fmt.Errorf("stationの生成に失敗")
	}
	return
}

//FillStationName 駅名補完
func FillStationName(db *sql.DB) {
	opt := rdb.SearchOptions{AddWhere: "endtime is null and (trim(description) = '' or description is null) "}
	rows, err := rdb.SearchSpotmaster(db, opt)
	if err != nil {
		logger.Debugf("FillStationName SearchSpotmasterでエラー : %v", err)
		return
	}
	logger.Infof("FillStationName %d件処理します", len(rows))
	var station Heartrails
	for _, row := range rows {
		time.Sleep(1 * time.Second)
		station, err = requestStationInfo(row.Lon, row.Lat)
		if err != nil {
			logger.Debugf("FillStationName requestStationInfoでエラー(area=%s, spot=%s) : %v", row.Area, row.Spot, err)
			continue
		}
		row.Description, err = station.Response.GetDescriptions()
		if err != nil {
			logger.Debugf("FillStationName GetDescriptionsでエラー(area=%s, spot=%s) : %v", row.Area, row.Spot, err)
			continue
		}
		row.Station, err = station.Response.GetStations()
		if err != nil {
			logger.Debugf("FillStationName GetStationsでエラー(area=%s, spot=%s) : %v", row.Area, row.Spot, err)
			continue
		}
		err = rdb.UpsertSpotmaster(db, row)
		if err != nil {
			logger.Debugf("FillStationName UpsertSpotmasterでエラー(area=%s, spot=%s) : %v", row.Area, row.Spot, err)
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

//RunFiler 駅名補完メイン関数
func RunFiler() {
	logger.Debugf("RunFiler_start")
	db, err := rdb.GetConnectionPsql()
	if err != nil {
		logger.Debugf("GetConnectionPsqlでエラー : %v", err)
		return
	}
	defer db.Close()

	//補完処理実行
	FillStationName(db)
	logger.Debugf("RunFiler_end")
}

func main() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		fmt.Printf("InitDirSettingでエラー : %v", err)
		return
	}
	exeName := filer.GetExeName()
	logger.Info(exeName, "開始")

	//開始
	scheduledTime := filer.GetIniData(ini_section, "START", "00:00")
	_, _ = scheduler.Every().Day().At(scheduledTime).Run(RunFiler)
	logger.Infof("%sに実行します", scheduledTime)

	//終了させない
	runtime.Goexit()
}
