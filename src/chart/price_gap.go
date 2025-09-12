package chart

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
	"trade/src/common"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type PriceGapChart struct {
	path  string
	datas map[string][]*PriceGapChartData
	chart *charts.Line
}

type PriceGapChartData struct {
	Timestamp int64
	Symbol    string
	Ratio     float64
}

func NewPriceGapChart() *PriceGapChart {
	return &PriceGapChart{
		path:  common.ChartDir + "/price_gap/",
		datas: make(map[string][]*PriceGapChartData),
	}
}

func (c *PriceGapChart) Run(ctx context.Context) error {
	common.Logger.Sugar().Info("PriceGapChart Run")

	files, err := common.ListFiles(c.path)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = c.readLogFile(file)
		if err != nil {
			common.Logger.Sugar().Errorf("PriceGapChart readLogFile error: %v", err)
			continue
		}
	}

	err = c.constructGraph()
	if err != nil {
		return err
	}
	chartFile, err := os.Create(c.path + "price_gap.html")
	if err != nil {
		return err
	}
	defer chartFile.Close()
	return c.chart.Render(chartFile)
}

func (c *PriceGapChart) readLogFile(file string) error {
	var (
		rawStr string
		err    error
	)
	if strings.HasSuffix(file, ".log.gz") {
		rawStr, err = common.ReadGzLogFile(c.path + file)
	} else if strings.HasSuffix(file, ".log") {
		rawStr, err = common.ReadLogFile(c.path + file)
	}
	if err != nil {
		return err
	}
	common.Logger.Sugar().Infof("PriceGapChart readLogFile: %s", file)
	for _, line := range strings.Split(rawStr, "\n") {
		if line == "" {
			common.Logger.Sugar().Warnf("PriceGapChart readLogFile empty line")
			continue
		}
		logData := &common.LogData{}
		err = json.Unmarshal([]byte(line), logData)
		if err != nil {
			common.Logger.Sugar().Errorf("PriceGapChart readLogFile json.Unmarshal error: %v", err)
			continue
		}
		params := strings.Split(logData.Message, ",")
		if len(params) != 3 {
			common.Logger.Sugar().Warnf("PriceGapChart readLogFile invalid line format: %s", line)
			continue
		}
		timestamp, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			common.Logger.Sugar().Errorf("PriceGapChart readLogFile strconv.ParseInt error: %v", err)
			continue
		}
		symbol := params[1]
		ratio, err := strconv.ParseFloat(params[2], 64)
		if err != nil {
			common.Logger.Sugar().Errorf("PriceGapChart readLogFile strconv.ParseFloat error: %v", err)
			continue
		}
		if c.datas[symbol] == nil {
			c.datas[symbol] = make([]*PriceGapChartData, 0, 1024*8)
		}
		c.datas[symbol] = append(c.datas[symbol], &PriceGapChartData{
			Timestamp: timestamp,
			Symbol:    symbol,
			Ratio:     ratio,
		})
	}
	return nil
}

func (c *PriceGapChart) constructGraph() error {
	c.chart = charts.NewLine()
	c.chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Price Gap",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Min: -0.5,
			Max: 0.5,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithDataZoomOpts(
			opts.DataZoom{
				Type:  "slider",
				Start: 0,
				End:   100,
			},
			opts.DataZoom{
				Type:  "inside",
				Start: 0,
				End:   100,
			},
		),
	)
	for symbol, data := range c.datas {
		line := make([]opts.LineData, 0, len(data))
		for _, d := range data {
			line = append(line, opts.LineData{Value: []interface{}{time.Unix(d.Timestamp, 0), d.Ratio}})
		}
		c.chart.AddSeries(symbol, line)
	}
	return nil
}
