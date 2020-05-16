package main

import (
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
)

//デフォルト値
const (
	defWidth        float64 = 600.0
	defHeight       float64 = 400.0
	defMarginLeft   float64 = 50.0
	defMarginRight  float64 = 50.0
	defMarginTop    float64 = 50.0
	defMarginBottom float64 = 50.0
)

//GetGraph グラフ作成
func GetGraph(w rest.ResponseWriter, r *rest.Request) {
	//パース
	r.ParseForm()
	params := r.Form
	area := params.Get("area")
	spot := params.Get("spot")
	if area == "" || spot == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson("パラメータが不正です")
		return
	}
	days := params.Get("days")
	var daysArr []string
	//デフォルトは今日と昨日のみ
	if days == "" {
		daysArr = []string{time.Now().Format("20060102"), (time.Now().AddDate(0, 0, -1)).Format("20060102")}
	} else {
		daysArr = strings.Split(days, ",")
	}
	// 画像プロパティ設定
	width := defWidth
	height := defHeight
	mleft := defMarginLeft
	mtop := defMarginTop
	mright := defMarginRight
	mbottom := defMarginBottom
	if property := params.Get("property"); property != "" {
		arr := strings.Split(property, ",")
		for i, num := range arr {
			val, err := strconv.ParseFloat(num, 64)
			if err != nil {
				continue
			}
			switch i {
			case 0:
				width = val
			case 1:
				height = val
			case 2:
				mleft = val
			case 3:
				mright = val
			case 4:
				mtop = val
			case 5:
				mbottom = val
			}
		}
	}
	//グラフ画像作成
	graph := NewGraph(width, height, mleft, mright, mtop, mbottom)
	for _, day := range daysArr {
		graph.SetData(area, spot, day)
	}
	if title := params.Get("title"); title == "yes" {
		graph.SetTitle(area, spot)
	}
	filepath := graph.Draw()
	if filepath == ErrorImageName {
		w.Header().Set("Content-Type", "application/json")
		w.WriteJson(ErrorImageURL)
		return
	}
	link := UploadImgur(filepath)
	os.Remove(filepath)

	resp := static.JGraphResponse{Title: graph.Title,
		Width:  strconv.Itoa(int(graph.Width)),
		Height: strconv.Itoa(int(graph.Height)),
		URL:    link}

	//URLを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteJson(resp)
	return
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

	//サーバ開始
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":5010", api.MakeHandler()))
}
