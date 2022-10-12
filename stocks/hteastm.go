package stocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"quantbot/storedb"
	"quantbot/utils"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// 深圳行情
// https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&ut=7eea3edcaed734bea9cbfc24409ed989&klt=5&fqt=1&secid=0.000001&beg=0&end=20500000&_=1630930917857
// 上海行情
// https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&ut=7eea3edcaed734bea9cbfc24409ed989&klt=5&fqt=1&secid=1.000001&beg=0&end=20500000&_=1630930917857

type EastMoneyHistory struct {
}

type EmHistoryRS struct {
	Rc     int    `json:"rc"`
	Rt     int    `json:"rt"`
	Svr    int    `json:"svr"`
	Lt     int    `json:"lt"`
	Full   int    `json:"full"`
	Dlmkts string `json:"dlmkts"`
	Data   struct {
		Code      string   `json:"code"`
		Market    int      `json:"market"`
		Name      string   `json:"name"`
		Decimal   int      `json:"decimal"`
		Dktotal   int      `json:"dktotal"`
		PreKPrice float64  `json:"preKPrice"`
		Klines    []string `json:"klines"`
	} `json:"data"`
}

func (obj *EastMoneyHistory) onGetHestoryItem(code, item string) (*storedb.TsItem, error) {
	if len(item) > 0 {
		part := strings.Split(item, ",")
		if len(part) > 6 {
			part[0] = strings.Replace(part[0], "-", "", -1)
			part[0] = strings.Replace(part[0], " ", "", -1)
			part[0] = strings.Replace(part[0], ":", "", -1)
			dTime, err := strconv.ParseInt(part[0], 10, 64)
			if nil != err {
				utils.Logger.Info("onGetHestoryItem",
					zap.String("err", err.Error()),
					zap.Any("err", part[0]))
				return nil, err
			}
			stockItem := &storedb.TsItem{}
			stockItem.Code = code
			stockItem.TS = dTime
			stockItem.Date = stockItem.TS / 10000
			stockItem.Open, err = strconv.ParseFloat(part[1], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			stockItem.Close, err = strconv.ParseFloat(part[2], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			stockItem.Max, err = strconv.ParseFloat(part[3], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			stockItem.Min, err = strconv.ParseFloat(part[4], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			if stockItem.Max < stockItem.Min {
				utils.Logger.Warn("onGetHestoryItem error max min",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min))
				return nil, errors.New("error max min")
			}

			if stockItem.Open > stockItem.Max || stockItem.Open < stockItem.Min {
				utils.Logger.Warn("onGetHestoryItem  error open",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min),
					zap.Float64("Open", stockItem.Open))
				return nil, errors.New("error Open")
			}
			if stockItem.Close > stockItem.Max || stockItem.Close < stockItem.Min {
				utils.Logger.Warn("onGetHestoryItem  error Close",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min),
					zap.Float64("Close", stockItem.Close))
				return nil, errors.New("error Close")
			}
			stockItem.Volume, err = strconv.ParseFloat(part[5], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			stockItem.Amount, err = strconv.ParseFloat(part[6], 64)
			if nil != err {
				utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
				return nil, err
			}
			if stockItem.Volume <= 0 || stockItem.Amount <= 0 {
				return nil, err
			}
			stockItem.OpenFr = StockEM
			stockItem.CloseFr = StockEM
			stockItem.MaxFr = StockEM
			stockItem.MaxFr = StockEM
			stockItem.VolumeFr = StockEM
			stockItem.AmountFr = StockEM
			stockItem.ID = fmt.Sprintf("%s_%d", stockItem.Code, stockItem.TS)
			return stockItem, nil
		}
	}
	return nil, errors.New("unknow")
}

func (obj *EastMoneyHistory) hestory(url, db string, stocks []string, beg, end int) {
	for _, v := range stocks {
		time.Sleep(time.Duration(2) * time.Second)
		tsItemList := []*storedb.TsItem{}
		url := fmt.Sprintf(url, v, beg, end, time.Now().UnixMilli())
		buf, err := utils.GetWithJSON(url, GHEADERS, nil)
		if nil != err {
			utils.Logger.Error("hestory", zap.String("err", err.Error()))
			continue
		}
		codeArray := strings.Split(v, ".")
		if len(codeArray) != 2 {
			continue
		}
		item := EmHistoryRS{}
		json.Unmarshal(buf, &item)
		if nil != item.Data.Klines {
			for _, tsItem := range item.Data.Klines {
				tsItem, err := obj.onGetHestoryItem(codeArray[1], tsItem)
				if nil == err && nil != tsItem {
					tsItemList = append(tsItemList, tsItem)
				}
			}
		}
		if len(tsItemList) > 0 {
			storedb.InsertMany(tsItemList, db, "STOCKSTS")
			utils.Logger.Info("hestory em", zap.Int("cnt", len(tsItemList)),
				zap.String("code", item.Data.Code))
		}
	}
}

func (obj *EastMoneyHistory) onGetHestory() {
	qfqurl := "https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&ut=7eea3edcaed734bea9cbfc24409ed989&klt=5&fqt=1&secid=%s&beg=%d&end=%d&_=%d"
	hfqurl := "https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&ut=7eea3edcaed734bea9cbfc24409ed989&klt=5&fqt=2&secid=%s&beg=%d&end=%d&_=%d"
	bfqurl := "https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&ut=7eea3edcaed734bea9cbfc24409ed989&klt=5&fqt=0&secid=%s&beg=%d&end=%d&_=%d"

	go func() {
		for {
			if utils.Conf.IsEMHestory {
				_, _, _, hour, _ := utils.GetTimeMin5()
				if hour >= 18 {
					listStocks := []string{}
					shMapLock.RLock()
					for k := range shMap {
						listStocks = append(listStocks, "1."+k)
					}
					shMapLock.RUnlock()
					szMapLock.RLock()
					for k := range szMap {
						listStocks = append(listStocks, "0."+k)
					}
					szMapLock.RUnlock()
					if len(listStocks) > 0 {
						beg := utils.GetDayBefore(-13)
						end := utils.GetDayBefore(1)
						obj.hestory(bfqurl, "BFQ-EM-HESTORY.DB", listStocks, beg, end)
						obj.hestory(qfqurl, "QFQ-EM-HESTORY.DB", listStocks, beg, end)
						obj.hestory(hfqurl, "HFQ-EM-HESTORY.DB", listStocks, beg, end)
					}
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}
