package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"os"
)

const (
	YAggregation   = 1
	XAggregation   = 2
	SubAggregation = 3
)

type Chart struct {
	lua.NoReflect
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

			pie := parseDataPie(res, dimension, title)
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
			c.charts = append(c.charts, parseDataBar(res, typ, dimension, title, x, y))
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

			chart := parseDataLine(res, typ, dimension, title, x, y)
			c.charts = append(c.charts, chart)
		}
	}
}

// 解析结果生成饼图所需的数据
func parseDataPie(res []byte, dimension, title string) *charts.Pie {
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
		charts.WithTitleOpts(opts.Title{Title: title, Left: "center"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
	)

	pie.AddSeries(title, pieData).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)

	return pie
}

// 解析结果，生成条形图。绘制条形图和折线图的数据无差别
func parseDataBar(res []byte, typ int, dimension, title, x, y string) *charts.Bar {
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
		charts.WithTitleOpts(opts.Title{Title: title, Left: "center"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
		charts.WithXAxisOpts(opts.XAxis{Name: x}),
		charts.WithYAxisOpts(opts.YAxis{Name: y}),
	)
	return bar
}

// 折线图
func parseDataLine(res []byte, typ int, dimension, title, x, y string) *charts.Line {
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
		line.AddSeries(cate, items)
	}

	return line
}

func setLine(line *charts.Line, title, x, y string) *charts.Line {
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
			Left:  "center",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "shine",
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: x}),
		charts.WithYAxisOpts(opts.YAxis{Name: y}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
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
