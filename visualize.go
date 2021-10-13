package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/datasets"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/ipplus360/awdb-golang/awdb-golang"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	YAggregation   = 1
	XAggregation   = 2
	SubAggregation = 3
)

type Chart struct {
	lua.Super
	charts []components.Charter
}

// Pie 绘制饼图
func (c *Chart) Pie(s Search, dimension, title string) {
	if s.typ != "agg" {
		logger.Errorf("if want to generate a pie, must be an aggregation search")
		return
	}

	// 一般聚合查询只有一条结果，等待search的结果channel关闭
	for {
		select {
		case res, ok := <-s.buffer:
			if !ok {
				logger.Infof("pie chart generate complete")
				return
			}

			pie := parseData2Pie(res, dimension, title)
			c.charts = append(c.charts, pie)
		}
	}
}

// Bar 绘制条形统计图。s为搜索结果，typ为搜索类型：2，x轴上一个坐标点对应一类数据；3，轴上一个坐标点对应多类数据
// dimension:统计维度，value或count；x，y 坐标轴名称
func (c *Chart) Bar(s Search, typ int, dimension, title, x, y string) {
	if s.typ != "agg" {
		logger.Errorf("if want to generate a bar chart, must be an aggregation search")
		return
	}

	for {
		select {
		case res, ok := <-s.buffer:
			if !ok {
				logger.Infof("bar chart generate complete")
				return
			}
			c.charts = append(c.charts, parseData2Bar(res, typ, dimension, title, x, y))
		}
	}
}

// Line 绘制折线统计图。s为搜索结果，typ为搜索类型：2，x轴上一个坐标点对应一类数据；3，轴上一个坐标点对应多类数据
// dimension:统计维度，value或count；x，y 坐标轴名称
func (c *Chart) Line(s Search, typ int, dimension, title, x, y string) {
	if s.typ != "agg" {
		logger.Errorf("if want to generate a bar chart, must be an aggregation search")
		return
	}

	for {
		select {
		case res, ok := <-s.buffer:
			if !ok {
				logger.Infof("line chart generate complete")
				return
			}

			chart := parseData2Line(res, typ, dimension, title, x, y)
			c.charts = append(c.charts, chart)
		}
	}
}

// Geo 绘制地图
func (c *Chart) Geo(s Search, mapT, dimension, title, dbPath string) {
	if s.typ != "agg" {
		logger.Errorf("if want to generate a geo map, must be an aggregation search")
		return
	}

	// 一般聚合查询只有一条结果，等待search的结果channel关闭
	for {
		select {
		case res, ok := <-s.buffer:
			if !ok {
				logger.Infof("geo chart generate complete")
				return
			}

			pie := parseData2Geo(res, mapT, dimension, title, dbPath)
			c.charts = append(c.charts, pie)
		}
	}
}

// 解析结果生成饼图所需的数据
func parseData2Pie(res []byte, dimension, title string) *charts.Pie {
	var r SingleRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return nil
	}

	var pieData = make([]opts.PieData, len(r.Buckets))
	for i, bucket := range r.Buckets {
		key := convertKey(bucket.Key)
		if bucket.KeyAsString != "" {
			key = bucket.KeyAsString
		}
		if dimension == "value" {
			pieData[i] = opts.PieData{Name: key, Value: bucket.Agg.Value}
			continue
		}
		pieData[i] = opts.PieData{Name: key, Value: bucket.DocCount}
	}

	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: title,
			Theme:     types.ThemeShine,
			Width:     "100%",
			Height:    "100vh",
		}),
		charts.WithTitleOpts(opts.Title{Title: title, Left: "center"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{
			Show:   true,
			Orient: "vertical",
			X:      "left",
			Top:    "center",
		}),
	)

	pie.AddSeries(title, pieData).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {d}%",
			}),
		)

	return pie
}

// 解析结果生成地图对象
func parseData2Geo(res []byte, mapT, dimension, title, dbPath string) *charts.Geo {
	var r SingleRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return nil
	}

	db, err := awdb.Open(dbPath)
	if err != nil {
		logger.Errorf("open geoip2 db error: %v", err)
		return nil
	}
	defer db.Close()

	geoData := generateGeoItem(r, mapT, dimension, db)
	resSort := sortGeoData(geoData)
	l := len(resSort)
	var top5Data []opts.GeoData
	var lastData []opts.GeoData

	if l < 6 {
		top5Data = resSort
	} else {
		top5Data = resSort[0:5]
		lastData = resSort[5 : l-1]
	}

	// geo chart
	geo := charts.NewGeo()
	geo.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle:       title,
			Width:           "100vw",
			Height:          "100vh",
			BackgroundColor: "#044161",
			ChartID:         "",
			AssetsHost:      "",
			Theme:           types.ThemeInfographic,
		}),
		charts.WithTitleOpts(opts.Title{Title: title, Left: "center", Top: "20px", TitleStyle: &opts.TextStyle{
			Color: "#ffffff"}}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: true,
			Min:        float32(resSort[0].Value.([]float64)[2]),
			Max:        float32(resSort[l-1].Value.([]float64)[2]),
			InRange: &opts.VisualMapInRange{
				Symbol: "circle",
				Color:  []string{"#ffffff", "#50a3ba", "#eac736", "#d94e5d", "#ff0000"},
			},
		}),
		charts.WithGeoComponentOpts(opts.GeoComponent{
			Map: mapT,
			ItemStyle: &opts.ItemStyle{
				Color:       "#004883",
				BorderColor: "#046ac8",
				Opacity:     1,
			},
			//Silent:    true,
		}),
	)

	geo.AddSeries("geoTop", types.ChartEffectScatter, top5Data,
		charts.WithRippleEffectOpts(opts.RippleEffect{
			Period:    4,
			Scale:     4,
			BrushType: "stroke",
		}),
	).AddSeries("geoLast", types.ChartScatter, lastData)

	return geo
}

// 根据结果返回 城市和统计数据
func generateGeoItem(r SingleRes, mapT, dimension string, db *awdb.Reader) []opts.GeoData {
	// 以省份为维度
	var err error
	var cityCount = make(map[string][]float64, 0)
	for _, bucket := range r.Buckets {
		key := convertKey(bucket.Key)
		if bucket.KeyAsString != "" {
			key = bucket.KeyAsString
		}
		key = strings.Trim(key, "\u0000")

		// 获取经纬度，默认北京
		var city = "北京"
		var longitude = float64(datasets.Coordinates[city][0])
		var latitude = float64(datasets.Coordinates[city][1])
		var record interface{}

		ip := net.ParseIP(key)
		err = db.Lookup(ip, &record)
		if err != nil {
			logger.Errorf("get geo error for ip %s: %v", ip, err)
		} else {
			var resMap = record.(map[string]interface{})
			city = convertProvinceName(resMap, mapT)
			coordinate, ok := datasets.Coordinates[city]
			if ok {
				// 国内省份
				longitude = float64(coordinate[0])
				latitude = float64(coordinate[1])
			} else {
				// 国外省份
				longitude, _ = strconv.ParseFloat(fmt.Sprintf("%s", resMap["lngwgs"]), 64)
				latitude, _ = strconv.ParseFloat(fmt.Sprintf("%s", resMap["latwgs"]), 64)
			}
		}

		geoV, ok := cityCount[city]
		if !ok {
			// 未获取过数据
			geoV = make([]float64, 3)
		}
		geoV[0] = longitude
		geoV[1] = latitude

		if dimension == "value" {
			geoV[2] += float64(bucket.Agg.Value)
		} else {
			geoV[2] += float64(bucket.DocCount)
		}

		cityCount[city] = geoV
	}

	var geoData = make([]opts.GeoData, 0)
	for city, count := range cityCount {
		geoData = append(geoData, opts.GeoData{
			Name:  city,
			Value: count,
		})
	}

	return geoData
}

// 解析结果，生成条形图。绘制条形图和折线图的数据无差别
func parseData2Bar(res []byte, typ int, dimension, title, x, y string) *charts.Bar {
	switch typ {
	case XAggregation:
		// 基础类型
		return generateXAggBarItems(res, dimension, title, x, y)
	case SubAggregation:
		// 一个x轴数据点，对应多个类型的值
		return generateSubAggBarItems(res, dimension, title, x, y)
	default:
		return nil
	}
}

// 类型1，一个x轴坐标点对应一个类别数据
func generateXAggBarItems(res []byte, dimension, title, x, y string) *charts.Bar {
	var r SingleRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return nil
	}

	xAxis := make([]string, len(r.Buckets))
	yAxis := make([]opts.BarData, len(r.Buckets))
	for i, bucket := range r.Buckets {
		key := convertKey(bucket.Key)
		if bucket.KeyAsString != "" {
			key = bucket.KeyAsString
		}
		xAxis[i] = key
		yAxis[i] = opts.BarData{Value: bucket.DocCount}
		if dimension == "value" {
			yAxis[i] = opts.BarData{Value: bucket.Agg.Value}
			continue
		}
		yAxis[i] = opts.BarData{Value: bucket.DocCount}
	}

	bar := charts.NewBar()
	bar = setBar(bar, title, x, y)
	bar.SetXAxis(xAxis).
		AddSeries("", yAxis)

	return bar
}

// 类型2，一个x坐标点对应多个类别数据
func generateSubAggBarItems(res []byte, dimension string, title, x, y string) *charts.Bar {
	r, xAxis, yAxis, err := generateSubAggItems(res, dimension)
	if err != nil {
		return nil
	}

	yAxisBar, ok := yAxis.(map[string][]int64)
	if !ok {
		logger.Errorf("assert yAxis to []opts.BarData error")
		return nil
	}

	bar := charts.NewBar()
	bar = setBar(bar, title, x, y)
	bar.SetXAxis(xAxis)
	for cate, data := range yAxisBar {
		items := make([]opts.BarData, len(r.Buckets))
		for k, item := range data {
			items[k] = opts.BarData{Value: item}
		}
		bar.AddSeries(cate, items)
	}

	return bar
}

// bar设置参数
func setBar(bar *charts.Bar, title, x, y string) *charts.Bar {
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeShine,
			Width:  "100%",
			Height: "100vh",
		}),
		charts.WithTitleOpts(opts.Title{Title: title, Left: "center"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
		charts.WithXAxisOpts(opts.XAxis{Name: x}),
		charts.WithYAxisOpts(opts.YAxis{Name: y}),
	)
	return bar
}

// 折线图
func parseData2Line(res []byte, typ int, dimension, title, x, y string) *charts.Line {
	switch typ {
	case XAggregation:
		// 基础类型
		return generateXAggLineItems(res, dimension, title, x, y)
	case SubAggregation:
		return generateSubAggLineItems(res, dimension, title, x, y)
	default:
		return nil
	}
}

func generateXAggLineItems(res []byte, dimension, title, x, y string) *charts.Line {
	var r SingleRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return nil
	}

	xAxis := make([]string, len(r.Buckets))
	yAxis := make([]opts.LineData, len(r.Buckets))
	for i, bucket := range r.Buckets {
		key := convertKey(bucket.Key)
		if bucket.KeyAsString != "" {
			key = bucket.KeyAsString
		}
		xAxis[i] = key
		yAxis[i] = opts.LineData{Value: bucket.DocCount}
		if dimension == "value" {
			yAxis[i] = opts.LineData{Value: bucket.Agg.Value}
			continue
		}
		yAxis[i] = opts.LineData{Value: bucket.DocCount}
	}

	line := charts.NewLine()
	line = setLine(line, title, x, y)
	line.SetXAxis(xAxis).
		AddSeries("", yAxis)

	return line
}

func generateSubAggLineItems(res []byte, dimension, title, x, y string) *charts.Line {
	r, xAxis, yAxis, err := generateSubAggItems(res, dimension)
	if err != nil {
		return nil
	}

	yAxisLine, ok := yAxis.(map[string][]int64)
	if !ok {
		logger.Errorf("assert yAxis to map[string][]int64 error")
		return nil
	}

	line := charts.NewLine()
	line = setLine(line, title, x, y)
	line.SetXAxis(xAxis)
	for cate, data := range yAxisLine {
		items := make([]opts.LineData, len(r.Buckets))
		for k, item := range data {
			items[k] = opts.LineData{Value: item}
		}
		line.AddSeries(cate, items).
			SetSeriesOptions(charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:      true,
					Color:     "",
					Position:  "",
					Formatter: "{b}  {c}",
				},
			}))
	}

	return line
}

// 样式修改
func setLine(line *charts.Line, title, x, y string) *charts.Line {
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "100vh",
			Theme:  types.ThemeShine,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
			Left:  "center",
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: x}),
		charts.WithYAxisOpts(opts.YAxis{Name: y}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			//Left:         "10px",
			Orient:       "vertical",
			X:            "right",
			SelectedMode: "multiple",
			Align:        "right",
		}),
	)

	return line
}

func generateSubAggItems(res []byte, dimension string) (MultiRes, []string, interface{}, error) {
	var r MultiRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return r, nil, nil, err
	}

	if len(r.Buckets) == 0 {
		logger.Errorf("no bucket in search result")
		return r, nil, nil, err
	}

	xAxis := make([]string, len(r.Buckets))
	yAxis := make(map[string][]int64)
	for i, bucket := range r.Buckets {
		// x 轴
		key := convertKey(bucket.Key)
		if bucket.KeyAsString != "" {
			key = bucket.KeyAsString
		}
		xAxis[i] = key

		subBucket3 := bucket.SubBucket.SubBuckets3
		if subBucket3 == nil {
			continue
		}
		for _, subBucket := range subBucket3 {
			// category 类别，每一类应该有len(r.Buckets)个数值
			subKey := convertKey(subBucket.Key)
			if subBucket.KeyAsString != "" {
				subKey = bucket.KeyAsString
			}

			_, ok := yAxis[subKey]
			if !ok {
				// 类别数据之前不存在
				yAxis[subKey] = make([]int64, len(r.Buckets))
			}
			if dimension == "value" {
				yAxis[subKey][i] = subBucket.Agg.Value
				continue
			}
			yAxis[subKey][i] = subBucket.DocCount
		}
	}

	return r, xAxis, yAxis, nil
}

// Page 生成最终页面
func (c *Chart) Page(path string) error {
	page := components.NewPage()
	if len(c.charts) == 0 {
		logger.Errorf("no chart generated")
		return nil
	}

	for _, chart := range c.charts {
		if chart == nil {
			continue
		}
		page.AddCharts(chart)
	}
	// 绘制完成，清空slice
	c.charts = make([]components.Charter, 0)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		logger.Errorf("open res file error: %v", err)
		return err
	}
	defer f.Close()

	err = page.Render(f)
	return err
}

// 将key转化为string
func convertKey(key interface{}) string {
	switch key.(type) {
	case int:
		return fmt.Sprintf("%d", key.(int))
	case int64:
		return fmt.Sprintf("%d", key.(int64))
	case float64:
		return fmt.Sprintf("%f", key.(float64))
	case string:
		return key.(string)
	}

	return "format_error"
}

// 根据IP解析结果，返回城市名称
func convertProvinceName(resMap map[string]interface{}, mapT string) string {
	accuracy := fmt.Sprintf("%s", resMap["accuracy"])

	if accuracy == "国家" {
		// 未获取到省份信息
		country := fmt.Sprintf("%s", resMap["country"])
		longitude := fmt.Sprintf("%s", resMap["lngwgs"])
		latitude := fmt.Sprintf("%s", resMap["latwgs"])
		name := country + "_" + longitude + "_" + latitude
		return name
	}

	// 获取省份信息
	province := strings.Trim(fmt.Sprintf("%s", resMap["province"]), "省")
	//if mapT == "world" {
	//	// 如果是世界地图，则所有中国IP定位北京
	//	province = "北京"
	//}

	switch province {
	case "新疆维吾尔自治区":
		return "新疆"
	case "重庆市":
		return "重庆"
	case "北京市":
		return "北京"
	case "天津市":
		return "天津"
	case "上海市":
		return "上海"
	case "广西壮族自治区":
		return "广西"
	case "中国香港":
		return "香港"
	case "内蒙古自治区":
		return "内蒙古"
	case "":
		longitude := fmt.Sprintf("%s", resMap["lngwgs"])
		latitude := fmt.Sprintf("%s", resMap["latwgs"])
		name := "unknown_zone_" + longitude + "_" + latitude
		return name
	default:
		return province
	}
}

// 排序，获取最大值和最小值
//func sortGeoData(data []opts.GeoData) []float64 {
//	var resSort = make([]float64, len(data))
//	for i, d := range data {
//		v := d.Value.([]float64)[2]
//		resSort[i] = v
//	}
//
//	sort.Float64s(resSort)
//	return resSort
//}

type PairList []opts.GeoData

func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int      { return len(p) }
func (p PairList) Less(i, j int) bool {
	return p[i].Value.([]float64)[2] > p[j].Value.([]float64)[2]
}

// A function to turn a map into a PairList, then sort and return it.
func sortGeoData(p PairList) PairList {
	sort.Sort(p)
	return p
}
