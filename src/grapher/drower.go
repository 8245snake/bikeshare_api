package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strconv"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/rdb"
	"github.com/8245snake/bikeshare_api/src/lib/static"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/image/font"
)

//Db データベースコネクション
var Db *sql.DB

//ErrorImageName エラー画像
const ErrorImageName = "error.png"

//WeekDays 曜日
var WeekDays = [7]string{"日", "月", "火", "水", "木", "金", "土"}

//Graph グラフ本体
type Graph struct {
	DC                                               *gg.Context
	Title                                            string
	Width, Height                                    float64
	MarginLeft, MarginTop, MarginRight, MarginBottom float64
	Plots                                            []Plot
	XAxis, YAxis                                     Axis
}

//Axis 軸
type Axis struct {
	Tick   float64 //値１につき何ピクセルか（pc/value）
	Labels []AxisLabel
}

//AxisLabel 軸ラベルの要素
type AxisLabel struct {
	Caption string
	Value   float64
}

//Plot 点の集合
type Plot struct {
	Area, Spot       string
	Year, Month, Day int
	ColorIndex       int
	Points           []Point
	LegendCaption    string
}

func (p Plot) String() string {
	return fmt.Sprintf("%s-%s", p.Area, p.Spot)
}

//Point 散布図の点
type Point struct {
	xValue time.Time
	yValue float64
}

//NewPoint 点を作成
func NewPoint(xValue time.Time, yValue float64) (p Point) {
	p.xValue = xValue
	p.yValue = yValue
	return p
}

//GetCoordinate 座標を取得（左上が原点）
func (p *Point) GetCoordinate(xTick float64, yTick float64, xOrigin float64, yOrigin float64) (x float64, y float64) {
	xVal := p.xValue.Hour()*60 + p.xValue.Minute()
	x = float64(xVal)*xTick + xOrigin
	y = yOrigin - p.yValue*yTick
	return
}

func (p Point) String() string {
	return fmt.Sprintf("x:%v y:%v", p.xValue, p.yValue)
}

func getFontFace(size float64) font.Face {
	ftBinary, err := ioutil.ReadFile("../../resource/font/Koruri-Semibold.ttf")
	font, err := truetype.Parse(ftBinary)
	if err != nil {
		panic(err)
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: size,
	})
	return face
}

func initContext(width float64, height float64) *gg.Context {
	dc := gg.NewContext(int(width), int(height))
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(getFontFace(12))
	return dc
}

//DrawText テキストを挿入
func (g *Graph) DrawText(text string, x float64, y float64, angle float64) {
	radian := gg.Radians(angle)
	xp := x*math.Cos(-radian) - y*math.Sin(-radian)
	yp := x*math.Sin(-radian) + y*math.Cos(-radian)
	g.DC.Rotate(gg.Radians(angle))
	g.DC.DrawStringAnchored(text, xp, yp, 0, 0.5)
	g.DC.Rotate(gg.Radians(-angle))
}

//NewGraph グラフ初期化
func NewGraph(width, height, marginLeft, marginRight, marginTop, marginBottom float64) (g Graph) {
	g.Height = height
	g.Width = width
	g.DC = initContext(width, height)
	g.MarginLeft = marginLeft
	g.MarginRight = marginRight
	g.MarginTop = marginTop
	g.MarginBottom = marginBottom
	return
}

//SetData データ作成
func (g *Graph) SetData(area, spot, day string) {
	t, err := time.Parse("20060102", day)
	if err != nil {
		return
	}
	var points []Point
	points, err = createPoints(area, spot, day)

	var plot Plot
	plot.Points = points
	plot.Area = area
	plot.Spot = spot
	plot.Year = t.Year()
	plot.Month = int(t.Month())
	plot.Day = t.Day()
	plot.ColorIndex = len(g.Plots)
	plot.LegendCaption = fmt.Sprintf("%s (%s)", t.Format("2006/01/02"), WeekDays[t.Weekday()])
	g.Plots = append(g.Plots, plot)

}

//createPoints 指定日(yyyymmdd)のデータを検索しPoint構造体配列を作成する
func createPoints(area, spot, day string) (points []Point, err error) {

	spotinfos, err := rdb.SearchCountsByDay(Db, area, spot, day)
	if err != nil {
		return points, err
	}
	for _, bikecount := range spotinfos {
		if val, err := strconv.ParseFloat(bikecount.Count, 64); err == nil {
			points = append(points, NewPoint(bikecount.Time, val))
		}
	}
	return points, nil
}

//SetTitle グラフタイトルをセットする
func (g *Graph) SetTitle(area, spot string) {
	master, err := rdb.SearchSpotmaster(Db, rdb.SearchOptions{Area: area, Spot: spot})
	if err != nil || len(master) < 1 {
		return
	}
	name := master[0].Name
	g.Title = fmt.Sprintf("[%s-%s] %s", area, spot, name)
}

//Draw グラフ描画 ファイル名を返す
func (g *Graph) Draw() string {
	start := time.Now()
	if len(g.Plots) < 1 {
		return ErrorImageName
	}
	filePath := fmt.Sprintf("%s_%s.png", g.Plots[0], start.Format("20060102150405"))
	filePath = filepath.Join(static.DirImage, filePath)
	//軸ラベル作成（x軸：時刻）
	var xAxisLabels []AxisLabel
	for i := 0; i <= 24; i++ {
		//24時間分作成
		label := AxisLabel{
			Caption: fmt.Sprintf("%02d:00", i),
			Value:   float64(60 * i),
		}
		xAxisLabels = append(xAxisLabels, label)
	}
	//刻み計算
	xMax := xAxisLabels[len(xAxisLabels)-1].Value
	xMin := xAxisLabels[0].Value
	var plotWidth float64 = g.Width - g.MarginLeft - g.MarginRight
	xTick := plotWidth / (xMax - xMin)
	//軸作成
	g.XAxis = Axis{Tick: xTick, Labels: xAxisLabels}

	//軸ラベル作成（y軸：台数）
	//TODO:スケール調整
	var yAxisLabels []AxisLabel
	yMax := g.Max()
	yMin := 0.0
	step := 1
	if yMax > 100 {
		step = 10
	} else if yMax > 20 {
		step = 5
	}

	for i := 0; i <= int(yMax)+step; i += step {
		label := AxisLabel{
			Caption: strconv.Itoa(i),
			Value:   float64(i),
		}
		yAxisLabels = append(yAxisLabels, label)
	}
	var plotHeight float64 = g.Height - g.MarginTop - g.MarginBottom
	//刻み計算
	yTick := plotHeight / (yMax - yMin + float64(step) - float64(int(yMax)%step))
	//軸作成
	g.YAxis = Axis{Tick: yTick, Labels: yAxisLabels}

	//外枠
	g.DC.DrawRectangle(g.MarginLeft, g.MarginTop, plotWidth, plotHeight)
	// g.DC.Stroke()
	//プロットエリアを描画
	for _, label := range g.XAxis.Labels {
		//縦線
		startX := g.MarginLeft
		startY := g.MarginTop
		g.DC.SetRGB(0.7, 0.7, 0.7)
		g.DC.DrawLine(startX+label.Value*xTick, startY, startX+label.Value*xTick, startY+plotHeight)
		g.DC.Stroke()
		g.DC.SetRGB(0, 0, 0)
		g.DrawText(label.Caption, startX+label.Value*xTick, startY+plotHeight+5.0, 70)
		g.DC.Stroke()
	}
	for _, label := range g.YAxis.Labels {
		//横線
		startX := g.MarginLeft
		startY := g.MarginTop + plotHeight
		g.DC.SetRGB(0.7, 0.7, 0.7)
		g.DC.DrawLine(startX, startY-label.Value*yTick, startX+plotWidth, startY-label.Value*yTick)
		g.DC.Stroke()
		g.DC.SetRGB(0, 0, 0)
		g.DrawText(label.Caption, startX-20, startY-label.Value*yTick, 0)
		g.DC.Stroke()
	}

	//点と線を描画
	cutoff := 60.0 * xTick //60分以上間隔が空いたら線を引かない
	for i, plot := range g.Plots {
		//プロット
		var xSave, ySave float64
		changeColor(g.DC.SetRGB, plot.ColorIndex)
		for _, point := range plot.Points {
			x, y := point.GetCoordinate(xTick, yTick, g.MarginLeft, g.MarginTop+plotHeight)
			g.DC.DrawCircle(x, y, 2)
			if xSave != 0 && math.Abs(xSave-x) < cutoff {
				g.DC.DrawLine(x, y, xSave, ySave)
			}
			xSave = x
			ySave = y
		}

		//凡例
		space := 120.0
		legendTop := g.MarginTop - 10
		legendLeft := 10 + g.MarginLeft + float64(i)*space
		legendLength := 20.0

		g.DC.DrawLine(legendLeft, legendTop, legendLeft+legendLength, legendTop)
		g.DC.Stroke()
		g.DC.SetRGB(0, 0, 0)
		g.DrawText(plot.LegendCaption, legendLeft+legendLength+5, legendTop, 0)
		g.DC.Stroke()
	}
	//タイトル
	if g.Title != "" {
		g.DC.SetFontFace(getFontFace(18))
		g.DC.DrawStringWrapped(g.Title, g.MarginLeft, g.MarginTop-50, 0, 0, plotWidth, 1, gg.AlignLeft)
	}

	//保存
	err := g.DC.SavePNG(filePath)
	if err != nil {
		filePath = ErrorImageName
	}
	end := time.Now()
	fmt.Printf("%f秒\n", (end.Sub(start)).Seconds())
	return filePath
}

//changeColor 描画色を変える
func changeColor(SetRGB func(float64, float64, float64), colorIndex int) {
	switch colorIndex {
	case 0:
		SetRGB(1, 0, 0)
	case 1:
		SetRGB(0, 0, 1)
	case 2:
		SetRGB(0, 1, 0)
	case 3:
		SetRGB(1, 1, 0)
	case 4:
		SetRGB(1, 0, 1)
	case 5:
		SetRGB(0, 1, 1)
	}
}

//Max Y軸のMAX値を取得
func (p *Plot) Max() (max float64) {
	for _, point := range p.Points {
		if val := point.yValue; float64(max) < val {
			max = val
		}
	}
	return max
}

//Max Y軸のMAX値を取得
func (g *Graph) Max() (max float64) {
	for _, plot := range g.Plots {
		if val := plot.Max(); float64(max) < val {
			max = val
		}
	}
	return max
}
