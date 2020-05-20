package main

import (
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/carlescere/scheduler"
	_ "github.com/mattn/go-sqlite3"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  概要：postgres→SQLiteへの変換
//
//　機能：1. postgresのデータを日毎にSQLiteに変換する
//　　　　2. postgresの古いデータを削除する
//
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//ini_section iniのセクション
const ini_section = "ARCHIVE"

//sql_key SQLで日付検索するキー部分
const sql_key = "to_char(date(time),'YYYY-MM-DD')"

//max_insert バルクインサートの最大件数
var max_insert int

//delete_interval Spotinfoから削除する間隔
var delete_interval int

//archive_time アーカイブ実行時刻
var archive_time string

//RunArchive postgresから検索してSQLiteに保存しpostgresから削除する
func RunArchive() {
	logger.Debugf("RunArchive_start")
	db_psql, err := rdb.GetConnectionPsql()
	if err != nil {
		return
	}
	defer db_psql.Close()
	//対象日を取得
	sql := "select " +
		sql_key +
		" from public.analyze" +
		" where date(time) <= date(CURRENT_TIMESTAMP - INTERVAL '2 DAY') group by date(time)"
	rows, err := db_psql.Query(sql)
	if err != nil {
		return
	}

	for rows.Next() {
		var value string
		_ = rows.Scan(&value)
		logger.Debugf("RunArchive アーカイブ対象日=%s", value)
		targetdate, err := time.Parse(filer.ModTimeLayout("yyyy-mm-dd"), value)
		if err != nil {
			continue
		}
		err = insert(db_psql, targetdate)
		if err != nil {
			logger.Debugf("RunArchive insert失敗 : %v", err)
			continue
		}
		logger.Debugf("RunArchive insert成功")
		err = delete(db_psql, targetdate)
		if err != nil {
			logger.Debugf("RunArchive delete失敗 : %v", err)
			continue
		}
		logger.Debugf("RunArchive delete成功")
	}
	logger.Debugf("RunArchive_end")
}

//insert SQLiteに保存
func insert(db *sql.DB, targetdate time.Time) error {
	//postgresから検索
	qry := fmt.Sprintf("SELECT time, trim(area), trim(spot), trim(count) FROM public.analyze where %s = '%s'",
		sql_key, targetdate.Format(filer.ModTimeLayout("yyyy-mm-dd")))
	rows, err := db.Query(qry)
	if err != nil {
		return err
	}
	//SQLiteに接続
	sqlite, err := rdb.GetConnectionSQLite(targetdate, true)
	if err != nil {
		return err
	}
	defer sqlite.Close()
	//SQLiteにInsert
	var rowTried int64    //Insertしようとした件数
	var rowAffected int64 //実際にInsertされた件数
	var result int64      //一時変数
	var rows_sqlite []rdb.Spotinfo
	for rows.Next() {
		var e rdb.Spotinfo
		err := rows.Scan(&e.Time, &e.Area, &e.Spot, &e.Count)
		if err != nil {
			continue
		}
		rows_sqlite = append(rows_sqlite, e)
		//インサート
		if len(rows_sqlite) >= max_insert {
			result, err = rdb.BulkInsertSpotinfo(sqlite, rows_sqlite)
			if err != nil {
				logger.Debugf("BulkInsertSpotinfoでエラー %v \n", err)
			}
			rowAffected += result
			rowTried += int64(len(rows_sqlite))
			rows_sqlite = []rdb.Spotinfo{}
		}
	}
	//インサート
	if len(rows_sqlite) > 0 {
		result, err = rdb.BulkInsertSpotinfo(sqlite, rows_sqlite)
		if err != nil {
			logger.Debugf("BulkInsertSpotinfoでエラー %v \n", err)
		} else {
			rowAffected += result
			rowTried += int64(len(rows_sqlite))
		}
	}

	logger.Debugf("%d件のInsertを試行しました", rowTried)
	logger.Debugf("%d件Insertされました", rowAffected)
	return err
}

//delete 指定日のデータを削除
func delete(db *sql.DB, targetdate time.Time) error {
	addWhere := fmt.Sprintf(" %s = '%s'", sql_key, targetdate.Format(filer.ModTimeLayout("yyyy-mm-dd")))
	option := rdb.SearchOptions{AddWhere: addWhere}
	RowsAffected, err := rdb.Delete(db, "public.analyze", option)
	logger.Debugf("analyzeから%d件削除されました。", RowsAffected)
	return err
}

//RunDeleteOld Spotinfoから古いデータを削除するメイン関数
func RunDeleteOld() {
	logger.Debugf("RunDeleteOld_start")
	db_psql, err := rdb.GetConnectionPsql()
	if err != nil {
		logger.Debugf("RunDeleteOld GetConnectionPsqlでエラー : %v", err)
	}
	defer db_psql.Close()
	//開始
	err = deleteOldRecords(db_psql)
	if err != nil {
		logger.Debugf("RunDeleteOld deleteOldRecordsでエラー : %v", err)
	}
	logger.Debugf("RunDeleteOld_end")
}

//deleteOldRecords spotinfoから古いデータを削除
func deleteOldRecords(db *sql.DB) error {
	sqlwhere := fmt.Sprintf("time < (CURRENT_TIMESTAMP - INTERVAL '%d MINUTE')", delete_interval)
	option := rdb.SearchOptions{AddWhere: sqlwhere}
	RowsAffected, err := rdb.Delete(db, "spotinfo", option)
	logger.Debugf("spotinfoから%d件削除されました。", RowsAffected)
	return err
}

//loadConfig 設定を読み込む
func loadConfig() {
	max_insert = filer.GetIniDataInt(ini_section, "MAXROWS", 5000)
	delete_interval = filer.GetIniDataInt(ini_section, "INTERVAL", 30)
	archive_time = filer.GetIniData(ini_section, "START", "00:00")
	logger.Info("設定を読み込みました")
	logger.Infof("MAXROWS=%d", max_insert)
	logger.Infof("INTERVAL=%d", delete_interval)
	logger.Infof("START=%s", archive_time)
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

	//設定ロード
	loadConfig()

	//開始
	_, _ = scheduler.Every().Day().At(archive_time).Run(RunArchive)
	_, _ = scheduler.Every(delete_interval).Minutes().Run(RunDeleteOld)

	//終了させない
	runtime.Goexit()
}
