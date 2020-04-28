package rdb

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

//TimeLayout 時刻フォーマット
const TimeLayout = "2006/01/02 15:04:05"

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  構造体
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//Spotinfo 台数情報テーブル
type Spotinfo struct {
	Time  time.Time `json:"time"`
	Area  string    `json:"area"`
	Spot  string    `json:"spot"`
	Count string    `json:"count"`
}

//Analyze 統計用テーブル
type Analyze struct {
	Time              time.Time
	Area, Spot, Count string
}

//Spotmaster スポットマスタ
type Spotmaster struct {
	Area, Spot, Name, Lat, Lon string
	Starttime, Endtime         time.Time
	Description, Station       string
}

//SearchOptions 検索オプション
type SearchOptions struct {
	Area, Spot, AddWhere, OrderBy string
	Offset, Limit                 int
	Time                          time.Time
}

//CurrentFull 現在の台数
type CurrentFull struct {
	Area, Spot, Name, Count        string
	Time                           time.Time
	Lat, Lon, Description, Station string
}

//ConfigDB 設定
type ConfigDB struct {
	Key, Value, HostID string
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  レシーバ
/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s Spotinfo) String() string {
	return fmt.Sprintf("%s-%s %s %s台", s.Area, s.Spot, s.Time.Format(TimeLayout), s.Count)
}

//TimeString 時刻のフォーマット(yyyy/mm/dd hh:mm:ss)
func (s Spotinfo) TimeString() string {
	return s.Time.Format(TimeLayout)
}

//ToAnalyze Spotinfo→Analyzeの変換
func (data Spotinfo) ToAnalyze() Analyze {
	return Analyze{Area: data.Area, Spot: data.Spot, Time: data.Time, Count: data.Count}
}

func (s Analyze) String() string {
	return fmt.Sprintf("%s-%s %s %s台", s.Area, s.Spot, s.Time.Format(TimeLayout), s.Count)
}

//ToAnalyze Analyze→Spotinfoの変換
func (data Analyze) ToSpotinfo() Spotinfo {
	return Spotinfo{Area: data.Area, Spot: data.Spot, Time: data.Time, Count: data.Count}
}

//ToCsvStr CSV用文字列生成
func (s Analyze) ToCsvStr() []string {
	return []string{s.Time.Format(TimeLayout), s.Area, s.Spot, s.Count}
}

//TimeString 時刻のフォーマット(yyyy/mm/dd hh:mm:ss)
func (s Analyze) TimeString() string {
	return s.Time.Format(TimeLayout)
}

//GetSqlWhere 検索条件作成
func (option SearchOptions) GetSqlWhere() string {
	qry := ""
	if option.Area != "" {
		qry += fmt.Sprintf(" and area='%s' ", option.Area)
	}
	if option.Spot != "" {
		qry += fmt.Sprintf(" and spot='%s' ", option.Spot)
	}
	if !option.Time.IsZero() {
		qry += fmt.Sprintf(" and time='%s' ", option.Time.Format(TimeLayout))
	}
	if option.AddWhere != "" {
		qry += fmt.Sprintf(" and %s ", option.AddWhere)
	}
	if option.OrderBy != "" {
		qry += fmt.Sprintf(" order by %s ", option.OrderBy)
	}
	if option.Offset != 0 {
		qry += fmt.Sprintf(" offset %d ", option.Offset)
	}
	if option.Limit != 0 {
		qry += fmt.Sprintf(" limit %d ", option.Limit)
	}
	if qry != "" {
		qry = " where 1=1 " + qry
	}
	return qry
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//SearchSpotinfoSingle 1件だけ取得
func SearchSpotinfoSingle(db *sql.DB, option SearchOptions) (Spotinfo, error) {
	qry := "SELECT time, trim(area), trim(spot), trim(count) FROM spotinfo "
	qry += option.GetSqlWhere()
	var obj Spotinfo
	err := db.QueryRow(qry).Scan(&obj.Time, &obj.Area, &obj.Spot, &obj.Count)
	return obj, err
}

//SearchSpotinfo Spotinfoテーブルをspotとareaから検索
func SearchSpotinfo(db *sql.DB, option SearchOptions) []Spotinfo {
	qry := "SELECT time, trim(area), trim(spot), trim(count) FROM spotinfo "
	qry += option.GetSqlWhere()

	rows, err := db.Query(qry)

	if err != nil {
		fmt.Println(err)
	}

	driverName := reflect.ValueOf(db.Driver()).Type().String()
	var es []Spotinfo
	for rows.Next() {
		var e Spotinfo
		if driverName == "*sqlite3.SQLiteDriver" {
			//SQLiteは日付型がないので特殊処理
			var timestr string
			err := rows.Scan(&timestr, &e.Area, &e.Spot, &e.Count)
			if err != nil {
				continue
			}
			e.Time, err = time.Parse(TimeLayout, timestr)
			if err != nil {
				continue
			}
		} else {
			err := rows.Scan(&e.Time, &e.Area, &e.Spot, &e.Count)
			if err != nil {
				continue
			}
		}
		es = append(es, e)
	}
	return es
}

//BulkInsertSpotinfo スポット情報をバルクインサートする
func BulkInsertSpotinfo(db *sql.DB, rows []Spotinfo) (int64, error) {
	qry := "insert into spotinfo (time,area,spot,count) values "
	template := "('%s','%s','%s','%s')"
	values := ""

	for i, row := range rows {
		if i != 0 {
			values += ","
		}
		values += fmt.Sprintf(template, row.Time.Format(TimeLayout), row.Area, row.Spot, row.Count)
	}
	qry += values + " on conflict do nothing"
	result, err := db.Exec(qry)
	RowsAffected, _ := result.RowsAffected()
	return RowsAffected, err
}

//SearchAnalyzeSingle 1件だけ取得
func SearchAnalyzeSingle(db *sql.DB, option SearchOptions) (Analyze, error) {
	qry := "SELECT time, trim(area), trim(spot), trim(count) FROM public.analyze "
	qry += option.GetSqlWhere()
	var obj Analyze
	err := db.QueryRow(qry).Scan(&obj.Time, &obj.Area, &obj.Spot, &obj.Count)
	return obj, err
}

//SearchAnalyze Analyzeテーブルをspotとareaから検索
func SearchAnalyze(db *sql.DB, option SearchOptions) ([]Analyze, error) {
	qry := "SELECT time, trim(area), trim(spot), trim(count) FROM public.analyze "
	qry += option.GetSqlWhere()

	rows, err := db.Query(qry)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var es []Analyze
	for rows.Next() {
		var e Analyze
		err := rows.Scan(&e.Time, &e.Area, &e.Spot, &e.Count)
		if err != nil {
			continue
		}
		es = append(es, e)
	}
	return es, nil
}

//BulkInsertAnalyze スポット情報をバルクインサートする
func BulkInsertAnalyze(db *sql.DB, rows []Analyze) error {
	qry := "insert into public.analyze (time,area,spot,count) values "
	template := "('%s','%s','%s','%s')"
	values := ""

	for i, row := range rows {
		if i != 0 {
			values += ","
		}
		values += fmt.Sprintf(template, row.Time.Format(TimeLayout), row.Area, row.Spot, row.Count)
	}
	qry += values + " on conflict do nothing"
	_, err := db.Exec(qry)
	return err
}

//SearchSpotmaster マスタ検索
func SearchSpotmaster(db *sql.DB, option SearchOptions) ([]Spotmaster, error) {
	qry := `select 
	trim(area),
	trim(spot),
	trim(name),
	lat,
	lon,
	COALESCE(description, ''),
	COALESCE(station, ''),
	starttime,
	COALESCE(endtime, '0001/01/01') 
	from spotmaster `
	qry += option.GetSqlWhere()

	rows, err := db.Query(qry)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var es []Spotmaster
	for rows.Next() {
		var e Spotmaster
		err := rows.Scan(&e.Area, &e.Spot, &e.Name, &e.Lat, &e.Lon, &e.Description, &e.Station, &e.Starttime, &e.Endtime)
		if err != nil {
			continue
		}
		es = append(es, e)
	}
	return es, nil
}

//UpsertSpotmaster あればUpdate無ければInsert
func UpsertSpotmaster(db *sql.DB, m Spotmaster) (err error) {

	qry := `insert into spotmaster(area, spot, name, lat, lon, starttime, endtime, description, station) 
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	on conflict on CONSTRAINT spotmaster_pkey do update 
	set (area, spot, name, lat, lon, starttime, endtime, description, station) 
	= ($1, $2, $3, $4, $5, $6, $7, $8, $9)  `

	if m.Endtime.IsZero() {
		_, err = db.Exec(qry, m.Area, m.Spot, m.Name, m.Lat, m.Lon, m.Starttime, nil, m.Description, m.Station)
	} else {
		_, err = db.Exec(qry, m.Area, m.Spot, m.Name, m.Lat, m.Lon, m.Starttime, m.Endtime, m.Description, m.Station)
	}

	return err
}

//SearchConfig 設定テーブル検索
func SearchConfig(db *sql.DB, option SearchOptions) ([]ConfigDB, error) {
	qry := "SELECT trim(key), trim(value), trim(hostid) FROM public.config "
	qry += option.GetSqlWhere()

	rows, err := db.Query(qry)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var es []ConfigDB
	for rows.Next() {
		var e ConfigDB
		err := rows.Scan(&e.Key, &e.Value, &e.HostID)
		if err != nil {
			continue
		}
		es = append(es, e)
	}
	return es, nil
}

//Delete 汎用的なレコード削除関数
func Delete(db *sql.DB, table string, option SearchOptions) (int64, error) {
	qry := "delete from " + table
	qry += option.GetSqlWhere()
	result, err := db.Exec(qry)
	RowsAffected, _ := result.RowsAffected()
	return RowsAffected, err
}
