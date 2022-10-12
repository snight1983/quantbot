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

// https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=sh688258&scale=5&ma=5&datalen=1000

type SinaHistory struct {
}

type SinaHtItem struct {
	Day       string  `json:"day"`
	Open      string  `json:"open"`
	High      string  `json:"high"`
	Low       string  `json:"low"`
	Close     string  `json:"close"`
	Volume    string  `json:"volume"`
	MaPrice5  float64 `json:"ma_price5"`
	MaVolume5 int     `json:"ma_volume5"`
}

func (obj *SinaHistory) onGetHestoryItem(code string, item SinaHtItem) (*storedb.TsItem, error) {

	stockItem := &storedb.TsItem{}
	stockItem.Code = code
	item.Day = strings.ReplaceAll(item.Day, "-", "")
	item.Day = strings.ReplaceAll(item.Day, ":", "")
	item.Day = strings.ReplaceAll(item.Day, " ", "")
	dTime, err := strconv.ParseInt(item.Day, 10, 64)
	if nil != err {
		utils.Logger.Info("onGetHestoryItem",
			zap.String("err", err.Error()),
			zap.Any("err", item.Day))
		return nil, err
	}
	stockItem.Code = code
	stockItem.TS = dTime / 100
	stockItem.Date = dTime / 1000000

	stockItem.Open, err = strconv.ParseFloat(item.Open, 64)
	if nil != err {
		utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
		return nil, err
	}
	stockItem.Close, err = strconv.ParseFloat(item.Close, 64)
	if nil != err {
		utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
		return nil, err
	}
	stockItem.Max, err = strconv.ParseFloat(item.High, 64)
	if nil != err {
		utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
		return nil, err
	}
	stockItem.Min, err = strconv.ParseFloat(item.Low, 64)
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
	stockItem.Volume, err = strconv.ParseFloat(item.Volume, 64)
	if nil != err {
		utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
		return nil, err
	}
	stockItem.Amount = stockItem.Volume * item.MaPrice5
	if nil != err {
		utils.Logger.Warn("onGetHestoryItem", zap.String("err", err.Error()))
		return nil, err
	}
	if stockItem.Volume <= 0 || stockItem.Amount <= 0 {
		return nil, err
	}
	stockItem.OpenFr = StockSina
	stockItem.CloseFr = StockSina
	stockItem.MaxFr = StockSina
	stockItem.MaxFr = StockSina
	stockItem.VolumeFr = StockSina
	stockItem.AmountFr = StockSina
	stockItem.ID = fmt.Sprintf("%s_%d", stockItem.Code, stockItem.TS)
	return stockItem, nil
}

func (obj *SinaHistory) hestory(url, db string, stocks []string) {
	for _, v := range stocks {
		time.Sleep(time.Duration(2) * time.Second)
		tsItemList := []*storedb.TsItem{}
		url := fmt.Sprintf(url, v)
		buf, err := utils.GetWithJSON(url, GHEADERS, nil)
		if nil != err {
			utils.Logger.Error("hestory", zap.String("err", err.Error()))
			continue
		}
		items := []SinaHtItem{}
		json.Unmarshal(buf, &items)
		code := strings.ReplaceAll(v, "sh", "")
		code = strings.ReplaceAll(code, "sz", "")
		for _, tsItem := range items {
			tsItem, err := obj.onGetHestoryItem(code, tsItem)
			if nil == err && nil != tsItem {
				tsItemList = append(tsItemList, tsItem)
			}
		}
		if len(tsItemList) > 0 {
			storedb.InsertMany(tsItemList, db, "STOCKSTS")
			utils.Logger.Info("hestory sina", zap.Int("cnt", len(tsItemList)),
				zap.String("code", code))
		}
	}
}

func (obj *SinaHistory) onGetHestory() {
	bfqurl := "https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=5&ma=5&datalen=480"
	go func() {
		for {
			if utils.Conf.IsSinaHestory {
				_, _, _, hour, _ := utils.GetTimeMin5()
				if hour >= 18 {
					listStocks := []string{}
					shMapLock.RLock()
					for k := range shMap {
						listStocks = append(listStocks, "sh"+k)
					}
					shMapLock.RUnlock()
					szMapLock.RLock()
					for k := range szMap {
						listStocks = append(listStocks, "sz"+k)
					}
					szMapLock.RUnlock()
					if len(listStocks) > 0 {
						obj.hestory(bfqurl, "BFQ-SINA-HESTORY.DB", listStocks)
					}
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}
