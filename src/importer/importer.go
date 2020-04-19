package main

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  概要：過去にバックアップしたCSVファイルをDBに戻す
//
//　機能：1. postgresへのインポート
//
/////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
)

var _colArea int
var _colSpot int
var _colTime int
var _colCount int
var _readMax int
var _timeFormatCsv string

var Db *sql.DB

//execImport インポート
func execImport(path string) error {
	fmt.Printf("CSVを読み込みます %s\n", path)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("CSV読み込みでエラー error=%v\n", err)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var line []string
	var spotinfos []rdb.Analyze
	fmt.Printf("インサート開始\n")
	//1行ずつ読み込みながら逐次実行する
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		if len(line) < 4 {
			continue
		}
		//インサートするバッファに詰める
		buff, err := ConvertToSpotInfo(line)
		if err != nil {
			return err
		}
		spotinfos = append(spotinfos, buff)
		if len(spotinfos) >= _readMax {
			err := rdb.BulkInsertAnalyze(Db, spotinfos)
			if err != nil {
				return err
			}
			spotinfos = []rdb.Analyze{}
		}
	}
	//ループを抜けたあとに残っていたら
	if len(spotinfos) > 0 {
		err := rdb.BulkInsertAnalyze(Db, spotinfos)
		if err != nil {
			return err
		}
	}
	fmt.Printf("インサート終了\n")
	return nil
}

//ConvertToSpotInfo CSVの1行を解釈し構造体に変換する
func ConvertToSpotInfo(line []string) (rdb.Analyze, error) {
	datetime, err := time.Parse(_timeFormatCsv, line[_colTime])
	if err != nil {
		return rdb.Analyze{}, err
	}
	return rdb.Analyze{Area: line[_colArea], Spot: line[_colSpot], Time: datetime, Count: line[_colCount]}, nil
}

func main() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		return
	}

	//設定読み込み
	section := "IMPORT"
	_readMax = filer.GetIniDataInt(section, "MAXROWS", 5000)
	_colArea = filer.GetIniDataInt(section, "COL_AREA", 1)
	_colSpot = filer.GetIniDataInt(section, "COL_SPOT", 2)
	_colTime = filer.GetIniDataInt(section, "COL_TIME", 0)
	_colCount = filer.GetIniDataInt(section, "COL_COUNT", 3)
	_timeFormatCsv = filer.ModTimeLayout(filer.GetIniData(section, "TIME_FORMAT", "yyyy-mm-dd HH:MM:SS"))

	Db, err = rdb.GetConnectionPsql()
	if err != nil {
		fmt.Printf("DB接続でエラー error=%v\n", err)
		return
	}
	defer Db.Close()

	//ファイル検索
	files, _ := filepath.Glob("../../app/csv/*.csv")
	for _, path := range files {
		err := execImport(path)
		if err != nil {
			fmt.Printf("%v\n", err)
			_ = filer.FileMove(path, "../../app/csv/NG")
		} else {
			_ = filer.FileMove(path, "../../app/csv/OK")
		}
	}

}
