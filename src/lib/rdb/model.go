package rdb

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/static"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  定数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//DriverType DBのドライバ識別子
type DriverType string

const (
	//DriverTypePostgres postgres
	DriverTypePostgres DriverType = "postgres"
	//DriverTypeSQLite3 SQLite
	DriverTypeSQLite3 DriverType = "sqlite3"
)

//ItemKeyLine 項目キー（LINE）
type ItemKeyLine string

const (
	//LineHistory 履歴
	LineHistory = "History"
	//LineNotify 通知時刻
	LineNotify = "Notify"
	//LineFavorite お気に入りスポット
	LineFavorite = "Favorite"
)

//ItemKeySlack 項目キー（Slack）
type ItemKeySlack string

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

//ConfigLINE LINEの設定
type ConfigLINE struct {
	ID, Key, Value string
	Seq            int
}

//ConfigSlack Slackの設定
type ConfigSlack struct {
	ID, Key, Value string
	Seq            int
}

//ServiceConfig 設定
type ServiceConfig interface {
	ServiceConfig()
}

//User ユーザ
type User struct {
	LineID    string   `json:"line_id"`
	SlackID   string   `json:"slack_id"`
	Favorites []string `json:"favorites"`
	Notifies  []string `json:"notifies"`
	Histories []string `json:"histories"`
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
func (s Spotinfo) ToAnalyze() Analyze {
	return Analyze{Area: s.Area, Spot: s.Spot, Time: s.Time, Count: s.Count}
}

func (s Analyze) String() string {
	return fmt.Sprintf("%s-%s %s %s台", s.Area, s.Spot, s.Time.Format(TimeLayout), s.Count)
}

//ToSpotinfo Analyze→Spotinfoの変換
func (s Analyze) ToSpotinfo() Spotinfo {
	return Spotinfo{Area: s.Area, Spot: s.Spot, Time: s.Time, Count: s.Count}
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

//ServiceConfig インターフェース用
func (conf ConfigLINE) ServiceConfig() {}

//ServiceConfig インターフェース用
func (conf ConfigSlack) ServiceConfig() {}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//IdentifyDriverType DBのタイプを取得
func IdentifyDriverType(db *sql.DB) DriverType {
	driverName := reflect.ValueOf(db.Driver()).Type().String()
	if driverName == "*sqlite3.SQLiteDriver" {
		return DriverTypeSQLite3
	} else {
		return DriverTypePostgres
	}
}

//JudgeDBTypeByDate 日付を渡してどちらのDBを見るか判定する
func JudgeDBTypeByDate(date time.Time) DriverType {
	today := time.Now()
	if (date.Year() == today.Year() && today.YearDay()-date.YearDay() <= 1) || date.IsZero() {
		return DriverTypePostgres
	} else {
		return DriverTypeSQLite3
	}
}

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
	defer rows.Close()

	var es []Spotinfo
	for rows.Next() {
		var e Spotinfo
		if IdentifyDriverType(db) == DriverTypeSQLite3 {
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
	defer rows.Close()

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

//SearchCountsByDay 指定日(yyyymmdd)のデータを検索する（psql, SQLite振り分け）
func SearchCountsByDay(psql *sql.DB, area, spot, day string) ([]Spotinfo, error) {
	var spotinfos []Spotinfo
	//検索条件作成
	option := SearchOptions{Area: area, Spot: spot, OrderBy: "time desc"}
	date, err := time.Parse("20060102", day)
	if err != nil {
		//ゼロ値で初期化
		date = time.Time{}
	}

	//どちらのDBを見るか
	if JudgeDBTypeByDate(date) == DriverTypePostgres {
		//今日か昨日ならPostgres
		if date.IsZero() {
			//日付未指定なら最新の1件のみ
			option.Limit = 1
		} else {
			option.AddWhere = fmt.Sprintf("date(time) = '%s'", date.Format("2006-01-02"))
		}
		analyzes, err := SearchAnalyze(psql, option)
		if err != nil {
			return spotinfos, err
		}
		//変換
		for _, anal := range analyzes {
			spotinfos = append(spotinfos, anal.ToSpotinfo())
		}
	} else {
		//昨日より過去ならSQLite
		db, err := GetConnectionSQLite(date, false)
		if err != nil {
			return spotinfos, err
		}
		defer db.Close()
		//SQLiteから検索
		spotinfos = SearchSpotinfo(db, option)
	}

	return spotinfos, nil
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
	defer rows.Close()

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
	defer rows.Close()

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

//SearchCurrentFull ビュー検索
func SearchCurrentFull(db *sql.DB, option SearchOptions) ([]CurrentFull, error) {
	qry := `select 
	trim(area),trim(spot),trim(name),
	trim(count),time,lat,lon,description,station 
	from current_full `
	qry += option.GetSqlWhere()
	rows, err := db.Query(qry)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var arr []CurrentFull
	for rows.Next() {
		var s CurrentFull
		err := rows.Scan(&s.Area, &s.Spot, &s.Name, &s.Count, &s.Time,
			&s.Lat, &s.Lon, &s.Description, &s.Station)
		if err != nil {
			continue
		}
		arr = append(arr, s)
	}
	return arr, nil
}

//Delete 汎用的なレコード削除関数
func Delete(db *sql.DB, table string, option SearchOptions) (int64, error) {
	qry := "delete from " + table
	qry += option.GetSqlWhere()
	result, err := db.Exec(qry)
	RowsAffected, _ := result.RowsAffected()
	return RowsAffected, err
}

//GetAllUsers ユーザー設定をすべて取得
func GetAllUsers(db *sql.DB) ([]User, error) {
	var users []User
	qry := `select line_id,slack_id from public.user`
	rows, err := db.Query(qry)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s User
		err := rows.Scan(&s.LineID, &s.SlackID)
		if err != nil {
			continue
		}
		users = append(users, s)
	}

	var rtnUsers []User
	for _, user := range users {
		qry = `select key,value,seq from line where id = $1 order by id,key,seq`
		stmt, err := db.Prepare(qry)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		defer stmt.Close()

		rows, err = stmt.Query(user.LineID)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		var key string
		var val string
		var seq int
		var Favorites []string = []string{}
		var Histories []string = []string{}
		var Notifies []string = []string{}
		for rows.Next() {
			err := rows.Scan(&key, &val, &seq)
			if err != nil {
				continue
			}
			switch ItemKeyLine(key) {
			case LineFavorite:
				Favorites = append(Favorites, val)
			case LineHistory:
				Histories = append(Histories, val)
			case LineNotify:
				Notifies = append(Notifies, val)
			}
		}
		user.Favorites = Favorites
		user.Histories = Histories
		user.Notifies = Notifies
		rtnUsers = append(rtnUsers, user)
	}

	return rtnUsers, nil
}

//UpsertUser ユーザがあればUpdate無ければInsert
func UpsertUser(db *sql.DB, user *User) (err error) {
	//トランザクション開始
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var qry string
	//user削除
	if user.LineID != "" {
		qry = fmt.Sprintf(`delete from public.user where line_id = '%s'`, user.LineID)
	} else if user.SlackID != "" {
		qry = fmt.Sprintf(`delete from public.user where slack_id = '%s'`, user.SlackID)
	} else {
		tx.Rollback()
		return fmt.Errorf("[ERROR]UpsertUser IDが不明です")
	}
	_, err = tx.Query(qry)
	if err != nil {
		tx.Rollback()
		return err
	}
	//userインサート
	qry = fmt.Sprintf(`insert into public.user(line_id, slack_id) values('%s','%s')`, user.LineID, user.SlackID)
	_, err = tx.Query(qry)
	if err != nil {
		tx.Rollback()
		return err
	}
	//line設定削除
	qry = "delete from public.line where id = $1"
	stmt, err := db.Prepare(qry)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = stmt.Query(user.LineID)
	if err != nil {
		tx.Rollback()
		return err
	}

	qry = "insert into public.line(id,key,value,seq) values "
	template := "('%s','%s','%s','%d')"
	var values string
	values = ""
	for i, fav := range user.Favorites {
		if i != 0 {
			values += ","
		}
		values += fmt.Sprintf(template, user.LineID, LineFavorite, fav, i)
	}
	_, err = db.Exec(qry + values)
	if err != nil {
		tx.Rollback()
		return err
	}

	values = ""
	for i, his := range user.Histories {
		if i != 0 {
			values += ","
		}
		values += fmt.Sprintf(template, user.LineID, LineHistory, his, i)
	}
	_, err = db.Exec(qry + values)
	if err != nil {
		tx.Rollback()
		return err
	}

	values = ""
	for i, nitice := range user.Notifies {
		if i != 0 {
			values += ","
		}
		values += fmt.Sprintf(template, user.LineID, LineNotify, nitice, i)
	}
	_, err = db.Exec(qry + values)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

//CheckScrapingStatus スクレイピングが正常に行われているかチェック
func CheckScrapingStatus(db *sql.DB) error {
	qry := "select count(*) from spotinfo"
	row := db.QueryRow(qry)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count < 1000 {
		return fmt.Errorf("%v", static.StatusScrapingError)
	}
	return nil
}
