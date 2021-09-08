package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/ipplus360/awdb-golang/awdb-golang"
	"github.com/oschwald/geoip2-golang"
	"github.com/rock-go/rock/logger"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

type Bar struct {
	Buckets []*Bucket `json:"buckets"`
}

type Bucket struct {
	KeyAsString string      `json:"key_as_string"`
	Key         interface{} `json:"key"`
	DocCount    int64       `json:"doc_count"`
}

func TestPie(t *testing.T) {
	res := `{
      "doc_count_error_upper_bound" : 0,
      "sum_other_doc_count" : 0,
      "buckets" : [
        {
          "key" : "x1.domain.com",
          "doc_count" : 231412470
        },
        {
          "key" : "mx1.domain.com",
          "doc_count" : 29886354
        },
        {
          "key" : "x2.domain.com",
          "doc_count" : 3971891
        },
        {
          "key" : "gb.domain.com",
          "doc_count" : 506776
        },
        {
          "key" : "mx1topic.domain.com",
          "doc_count" : 304999
        },
        {
          "key" : "x1-test.domain.com",
          "doc_count" : 10488
        },
        {
          "key" : "hfmgb.domain.com",
          "doc_count" : 11
        }
      ]
    }`

	var bar Bar
	err := json.Unmarshal([]byte(res), &bar)
	if err != nil {
		fmt.Println(err)
		return
	}

	items := make([]opts.PieData, len(bar.Buckets))
	for i, bucket := range bar.Buckets {
		items[i] = opts.PieData{Name: bucket.Key.(string), Value: bucket.DocCount}
	}

	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "各站点请求数比例", Left: "center"}),
	)

	pie.AddSeries("pie", items).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)

	page := components.NewPage()
	page.AddCharts(pie)
	f, err := os.Create("pie.html")
	if err != nil {
		fmt.Println(err)
	}
	page.Render(io.MultiWriter(f))
}

func TestBar(t *testing.T) {
	res := `{"buckets":[{"key_as_string":"2021-08-29T18:00:00.000+08:00","key":1630231200000,"doc_count":4280832},{"key_as_string":"2021-08-29T18:30:00.000+08:00","key":1630233000000,"doc_count":4506608},{"key_as_string":"2021-08-29T19:00:00.000+08:00","key":1630234800000,"doc_count":4763353},{"key_as_string":"2021-08-29T19:30:00.000+08:00","key":1630236600000,"doc_count":4818766},{"key_as_string":"2021-08-29T20:00:00.000+08:00","key":1630238400000,"doc_count":5049341},{"key_as_string":"2021-08-29T20:30:00.000+08:00","key":1630240200000,"doc_count":5176377},{"key_as_string":"2021-08-29T21:00:00.000+08:00","key":1630242000000,"doc_count":5397897},{"key_as_string":"2021-08-29T21:30:00.000+08:00","key":1630243800000,"doc_count":5485547},{"key_as_string":"2021-08-29T22:00:00.000+08:00","key":1630245600000,"doc_count":5166584},{"key_as_string":"2021-08-29T22:30:00.000+08:00","key":1630247400000,"doc_count":4930842},{"key_as_string":"2021-08-29T23:00:00.000+08:00","key":1630249200000,"doc_count":4610990},{"key_as_string":"2021-08-29T23:30:00.000+08:00","key":1630251000000,"doc_count":4076016},{"key_as_string":"2021-08-30T00:00:00.000+08:00","key":1630252800000,"doc_count":3837624},{"key_as_string":"2021-08-30T00:30:00.000+08:00","key":1630254600000,"doc_count":3071271},{"key_as_string":"2021-08-30T01:00:00.000+08:00","key":1630256400000,"doc_count":2920424},{"key_as_string":"2021-08-30T01:30:00.000+08:00","key":1630258200000,"doc_count":2739912},{"key_as_string":"2021-08-30T02:00:00.000+08:00","key":1630260000000,"doc_count":2587755},{"key_as_string":"2021-08-30T02:30:00.000+08:00","key":1630261800000,"doc_count":2371760},{"key_as_string":"2021-08-30T03:00:00.000+08:00","key":1630263600000,"doc_count":2323595},{"key_as_string":"2021-08-30T03:30:00.000+08:00","key":1630265400000,"doc_count":2421757},{"key_as_string":"2021-08-30T04:00:00.000+08:00","key":1630267200000,"doc_count":2357728},{"key_as_string":"2021-08-30T04:30:00.000+08:00","key":1630269000000,"doc_count":2325194},{"key_as_string":"2021-08-30T05:00:00.000+08:00","key":1630270800000,"doc_count":2256770},{"key_as_string":"2021-08-30T05:30:00.000+08:00","key":1630272600000,"doc_count":2361925},{"key_as_string":"2021-08-30T06:00:00.000+08:00","key":1630274400000,"doc_count":2594129},{"key_as_string":"2021-08-30T06:30:00.000+08:00","key":1630276200000,"doc_count":2666246},{"key_as_string":"2021-08-30T07:00:00.000+08:00","key":1630278000000,"doc_count":2690135},{"key_as_string":"2021-08-30T07:30:00.000+08:00","key":1630279800000,"doc_count":3130495},{"key_as_string":"2021-08-30T08:00:00.000+08:00","key":1630281600000,"doc_count":4032123},{"key_as_string":"2021-08-30T08:30:00.000+08:00","key":1630283400000,"doc_count":5544851},{"key_as_string":"2021-08-30T09:00:00.000+08:00","key":1630285200000,"doc_count":7362871},{"key_as_string":"2021-08-30T09:30:00.000+08:00","key":1630287000000,"doc_count":8426755},{"key_as_string":"2021-08-30T10:00:00.000+08:00","key":1630288800000,"doc_count":9624456},{"key_as_string":"2021-08-30T10:30:00.000+08:00","key":1630290600000,"doc_count":10241945},{"key_as_string":"2021-08-30T11:00:00.000+08:00","key":1630292400000,"doc_count":10181008},{"key_as_string":"2021-08-30T11:30:00.000+08:00","key":1630294200000,"doc_count":9387551},{"key_as_string":"2021-08-30T12:00:00.000+08:00","key":1630296000000,"doc_count":8161405},{"key_as_string":"2021-08-30T12:30:00.000+08:00","key":1630297800000,"doc_count":7890285},{"key_as_string":"2021-08-30T13:00:00.000+08:00","key":1630299600000,"doc_count":8238238},{"key_as_string":"2021-08-30T13:30:00.000+08:00","key":1630301400000,"doc_count":9113250},{"key_as_string":"2021-08-30T14:00:00.000+08:00","key":1630303200000,"doc_count":9535774},{"key_as_string":"2021-08-30T14:30:00.000+08:00","key":1630305000000,"doc_count":9897006},{"key_as_string":"2021-08-30T15:00:00.000+08:00","key":1630306800000,"doc_count":10843867},{"key_as_string":"2021-08-30T15:30:00.000+08:00","key":1630308600000,"doc_count":9070993},{"key_as_string":"2021-08-30T16:00:00.000+08:00","key":1630310400000,"doc_count":9011189},{"key_as_string":"2021-08-30T16:30:00.000+08:00","key":1630312200000,"doc_count":8980662},{"key_as_string":"2021-08-30T17:00:00.000+08:00","key":1630314000000,"doc_count":8509123},{"key_as_string":"2021-08-30T17:30:00.000+08:00","key":1630315800000,"doc_count":6382290},{"key_as_string":"2021-08-30T18:00:00.000+08:00","key":1630317600000,"doc_count":5780}]}`
	var bar Bar
	err := json.Unmarshal([]byte(res), &bar)
	if err != nil {
		fmt.Println(err)
		return
	}

	keys := make([]string, len(bar.Buckets))
	items := make([]opts.BarData, len(bar.Buckets))
	itemsLine := make([]opts.LineData, len(bar.Buckets))
	for i, bucket := range bar.Buckets {
		keys[i] = bucket.KeyAsString
		items[i] = opts.BarData{Value: bucket.DocCount}
		itemsLine[i] = opts.LineData{Value: bucket.DocCount}
	}

	barV := charts.NewBar()
	barV.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "过去24h"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
	)
	barV.SetXAxis(keys).
		AddSeries("", items)

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(
			opts.Title{Title: "过去24h请求趋势图", Left: "center"},
		),
	)

	line.SetXAxis(keys).
		AddSeries("请求数", itemsLine).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{Show: true}),
		)

	page := components.NewPage()
	page.AddCharts(barV, line)
	f, err := os.Create("bar.html")
	if err != nil {
		fmt.Println(err)
	}
	page.Render(io.MultiWriter(f))
}

func TestSubAggBar(t *testing.T) {
	res, err := ioutil.ReadFile("visualization/result_for_pie.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	//bar := parseData2Line(res, 2, "count", "各域名24h访问趋势图", "时间", "请求数")
	bar := parseData2Bar(res, 2, "count", "各域名24h访问趋势图", "时间", "请求数")
	//bar := parseData2Pie(res, "count", "各域名24h访问趋势图")
	page := components.NewPage()
	page.AddCharts(bar)

	f, err := os.OpenFile("bar_multi.html", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("open res file error: %v\n", err)
		return
	}
	defer f.Close()

	page.Render(f)
}

func TestGeo(t *testing.T) {
	db, err := geoip2.Open("../resource/GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP("59.77.233.214")
	record, err := db.City(ip)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Portuguese (BR) city name: %v\n", record.City.Names["en"])

	fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
}

func TestAiwenIP(t *testing.T) {
	db, err := awdb.Open("../resource/IP_trial_single_WGS84.awdb")
	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()

	ip := net.ParseIP("59.77.233.214")

	var record = make(map[string]interface{})
	err = db.Lookup(ip, &record)
	if err != nil {
		fmt.Println(err)
	}
	var resMap = record
	fmt.Printf("accuracy:%s", resMap["accuracy"])
	fmt.Println()
	fmt.Printf("%s", record)
}

// 地图描绘测试
func TestChart_Geo(t *testing.T) {
	data, _ := ioutil.ReadFile("./visualization/top_10_ip.json")
	geo := parseData2Geo(data, "china", "count", "股吧过去24h访问分布图", "../resource/IP_trial_single_WGS84.awdb")
	page := components.NewPage()
	page.AddCharts(geo)
	f, _ := os.Create("test.html")
	defer f.Close()
	page.Render(f)
}

type Res struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func TestParse(t *testing.T) {
	res, _ := ioutil.ReadFile("./visualization/top_10_ip.json")
	var r SingleRes
	var err error
	err = json.Unmarshal(res, &r)
	if err != nil {
		logger.Errorf("unmarshal search result error: %v", err)
		return
	}

	db, err := geoip2.Open("../resource/GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	data := make([]Res, 0)
	data2 := make(map[string][]float64)
	for _, bucket := range r.Buckets {
		name := bucket.Key.(string)
		value := bucket.DocCount
		data = append(data, Res{
			Name:  name,
			Value: value,
		})

		ip := net.ParseIP(name)
		record, err := db.City(ip)
		if err != nil {
			fmt.Println(err)
		}
		data2[name] = []float64{record.Location.Latitude, record.Location.Longitude}
	}

	d, _ := json.Marshal(&data)
	d2, _ := json.Marshal(&data2)
	fmt.Printf("%s", d)
	fmt.Printf("%s", d2)
}
