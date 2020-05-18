package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"
	"github.com/ant0ine/go-json-rest/rest"
)

//デフォルト値
const (
	defWidth            float64 = 400.0
	defHeight           float64 = 300.0
	defMarginLeft       float64 = 50.0
	defMarginRight      float64 = 50.0
	defMarginTop        float64 = 50.0
	defMarginBottom     float64 = 50.0
	defDaySpan          int     = 2
	FileNameTimeFormat          = "20060102150405"
	NotCreatedImageName         = "ERROR_NOT_CREATED.png"
)

//GraphConfig グラフリクエスト情報
type GraphConfig struct {
	Area         string
	Spot         string
	Days         []string
	Span         int //基準日から何日遡るか
	Width        float64
	Height       float64
	MarginLeft   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	DrawTitle    bool
	UploadImgur  bool
}

//LoadGraphConfig リクエストを解析し設定を取得する
func LoadGraphConfig(params *url.Values) (conf GraphConfig, err error) {
	conf.Area = params.Get("area")
	conf.Spot = params.Get("spot")
	if conf.Area == "" || conf.Spot == "" {
		return conf, fmt.Errorf("パラメータが不正です")
	}
	//日数
	if span, err := strconv.Atoi(params.Get("span")); err == nil {
		conf.Span = span
	} else {
		conf.Span = defDaySpan
	}

	days := params.Get("days")
	// 優先順位：daysがあればdaysのみで決定（spanは無視）
	// daysが空ならdays = 今日 としてspan日分遡って描画
	if days != "" {
		conf.Days = strings.Split(days, ",")
	} else {
		today := time.Now()
		for i := 0; i < conf.Span; i++ {
			conf.Days = append(conf.Days, today.AddDate(0, 0, -i).Format("20060102"))
		}
	}
	// 画像プロパティ設定
	conf.Width = defWidth
	conf.Height = defHeight
	conf.MarginLeft = defMarginLeft
	conf.MarginTop = defMarginTop
	conf.MarginRight = defMarginRight
	conf.MarginBottom = defMarginBottom
	if property := params.Get("property"); property != "" {
		arr := strings.Split(property, ",")
		for i, num := range arr {
			val, err := strconv.ParseFloat(num, 64)
			if err != nil {
				continue
			}
			switch i {
			case 0:
				conf.Width = val
			case 1:
				conf.Height = val
			case 2:
				conf.MarginLeft = val
			case 3:
				conf.MarginTop = val
			case 4:
				conf.MarginRight = val
			case 5:
				conf.MarginBottom = val
			}
		}
	}
	//フラグ設定
	if title := params.Get("title"); title == "yes" {
		conf.DrawTitle = true
	}

	if imgur := params.Get("imgur"); imgur == "yes" {
		conf.UploadImgur = true
	}

	return
}

//createImgName ファイル名を決定する
func createImgName(area, spot string) string {
	return fmt.Sprintf("%s_%s-%s.png", time.Now().Format(FileNameTimeFormat), area, spot)
}

//createTitle グラフタイトルをセットする
func createTitle(area, spot string) string {
	master, err := rdb.SearchSpotmaster(Db, rdb.SearchOptions{Area: area, Spot: spot})
	if err != nil || len(master) < 1 {
		return ""
	}
	name := master[0].Name
	return fmt.Sprintf("[%s-%s] %s", area, spot, name)
}

//drawGraphImage グラフ作成
func drawGraphImage(conf *GraphConfig, fileName string, title string) {
	graph := NewGraph(conf.Width, conf.Height, conf.MarginLeft, conf.MarginRight, conf.MarginTop, conf.MarginBottom)
	for _, day := range conf.Days {
		graph.SetData(conf.Area, conf.Spot, day)
	}
	if conf.DrawTitle {
		graph.Title = title
	}
	graph.Draw(fileName)
}

//GetGraph グラフ作成
func GetGraph(w rest.ResponseWriter, r *rest.Request) {
	//パース
	r.ParseForm()
	param := r.Form
	conf, err := LoadGraphConfig(&param)
	if err != nil {
		w.WriteJson(err.Error())
	}

	//先にファイル名やタイトルを決定しておく
	fileName := createImgName(conf.Area, conf.Spot)
	title := createTitle(conf.Area, conf.Spot)

	//URLを取得
	var link string
	if conf.UploadImgur {
		//imgurにアップロードする（同期）
		drawGraphImage(&conf, fileName, title)
		path := filepath.Join(static.DirImage, fileName)
		link = UploadImgur(path)
		os.Remove(path)
	} else {
		//ローカルのファイルを見せる（非同期）
		go drawGraphImage(&conf, fileName, title)
		link = "https://hanetwi.ddns.net/bikeshare/graph/img/" + fileName
	}

	//URLを返却
	resp := static.JGraphResponse{Title: title,
		Width:  strconv.Itoa(int(conf.Width)),
		Height: strconv.Itoa(int(conf.Height)),
		URL:    link}
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(resp)
	return
}

//handleFile ファイルを返す
func handleFile(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Replace(r.URL.Path, "/graph/img/", "", -1)
	body, err := ioutil.ReadFile(filepath.Join(static.DirImage, fileName))
	if err != nil {
		body, err = serveErrorImage(err.Error(), fileName)
		if err != nil {
			return
		}
	}
	w.Write(body)
}

//serveErrorImage エラー時の画像表示
//エラーの種類によって画像を出し分ける他、画像が作成される前にアクセスされた場合はできるまで待つ
func serveErrorImage(errString string, fileName string) ([]byte, error) {
	//返却用
	var returnFileName string
	switch 1 {
	// breakするためswitch構文に入れる
	default:
		if strings.Index(errString, "The system cannot find") > 0 {
			//ファイルが存在しないエラー
			datetime, err := time.Parse(FileNameTimeFormat, fileName[:len(FileNameTimeFormat)])
			if err != nil {
				//ファイル名が不正のため諦める
				returnFileName = NotCreatedImageName
				break
			}
			var during int = 100
			if now, err := strconv.Atoi(time.Now().Format(FileNameTimeFormat)); err == nil {
				if target, err := strconv.Atoi(datetime.Format(FileNameTimeFormat)); err == nil {
					during = now - target
				}
			}
			if during < 5 {
				// ５秒以内のリクエストなら様子見
				if filer.WaitForFileCreation(filepath.Join(static.DirImage, fileName), 1, 5) {
					returnFileName = fileName
				} else {
					returnFileName = NotCreatedImageName
				}
			} else {
				returnFileName = NotCreatedImageName
			}
		}
	}

	//返却
	body, err := ioutil.ReadFile(filepath.Join(static.DirImage, returnFileName))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func init() {
	//初期化
	err := filer.InitDirSetting()
	if err != nil {
		panic(err)
	}
	ImgurID = getConfig("imgur_id")
	if ImgurID == "" {
		panic("imgur_idが設定されていません")
	}
	Db, err = rdb.GetConnectionPsql()
	if err != nil {
		panic(err)
	}
}

func main() {
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
		rest.Get("/graph", GetGraph),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	//ハンドラ追加
	http.Handle("/", api.MakeHandler())
	http.Handle("/graph/img/", http.HandlerFunc(handleFile))
	//サーバ開始
	log.Fatal(http.ListenAndServe(":5010", nil))
}
