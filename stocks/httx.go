package stocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"quantbot/storedb"
	"quantbot/utils"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=sh688258&scale=5&ma=5&datalen=1000

type TXHistory struct {
}

type TXHistoryRS struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func (obj *TXHistory) onGetHestoryItem(code string, item []interface{}) (*storedb.TsItem, error) {

	stockItem := &storedb.TsItem{}
	stockItem.Code = code
	if nil == item || len(item) < 6 {
		return nil, errors.New("error list")
	}

	if item[0] != nil && reflect.TypeOf(item[0]).Kind() == reflect.String {
		dTime, err := strconv.ParseInt(item[0].(string), 10, 64)
		if nil != err {
			return nil, err
		}
		stockItem.TS = dTime / 100
		stockItem.Date = dTime / 1000000
	} else {
		return nil, errors.New("error item value")
	}

	if item[1] != nil && reflect.TypeOf(item[1]).Kind() == reflect.String {
		value, err := strconv.ParseFloat(item[1].(string), 64)
		if nil != err {
			return nil, err
		}
		stockItem.Open = value
	} else {
		return nil, errors.New("error item value")
	}

	if item[2] != nil && reflect.TypeOf(item[2]).Kind() == reflect.String {
		value, err := strconv.ParseFloat(item[2].(string), 64)
		if nil != err {
			return nil, err
		}
		stockItem.Close = value
	} else {
		return nil, errors.New("error item value")
	}

	if item[3] != nil && reflect.TypeOf(item[3]).Kind() == reflect.String {
		value, err := strconv.ParseFloat(item[3].(string), 64)
		if nil != err {
			return nil, err
		}
		stockItem.Max = value
	} else {
		return nil, errors.New("error item value")
	}

	if item[4] != nil && reflect.TypeOf(item[4]).Kind() == reflect.String {
		value, err := strconv.ParseFloat(item[4].(string), 64)
		if nil != err {
			return nil, err
		}
		stockItem.Min = value
	} else {
		return nil, errors.New("error item value")
	}

	if stockItem.Max < stockItem.Min {
		return nil, errors.New("error max min")
	}
	if stockItem.Open > stockItem.Max || stockItem.Open < stockItem.Min {
		return nil, errors.New("error Open")
	}
	if stockItem.Close > stockItem.Max || stockItem.Close < stockItem.Min {
		return nil, errors.New("error Close")
	}

	if item[5] != nil && reflect.TypeOf(item[5]).Kind() == reflect.String {
		value, err := strconv.ParseFloat(item[5].(string), 64)
		if nil != err {
			return nil, err
		}
		stockItem.Volume = value
	} else {
		return nil, errors.New("error item value")
	}

	stockItem.OpenFr = StockTX
	stockItem.CloseFr = StockTX
	stockItem.MaxFr = StockTX
	stockItem.MaxFr = StockTX
	stockItem.VolumeFr = StockTX
	stockItem.AmountFr = StockTX
	stockItem.ID = fmt.Sprintf("%s_%d", stockItem.Code, stockItem.TS)
	return stockItem, nil
}

func (obj *TXHistory) hestory(url, db string, stocks []string) {
	for _, v := range stocks {
		time.Sleep(time.Duration(utils.Conf.WaitSec) * time.Second)
		tsItemList := []*storedb.TsItem{}
		url := fmt.Sprintf(url, v)
		buf, err := utils.GetWithJSON(url, GHEADERS, nil)
		if nil != err {
			utils.Logger.Error("hestory", zap.String("err", err.Error()))
			continue
		}
		items := TXHistoryRS{}
		json.Unmarshal(buf, &items)
		code := strings.ReplaceAll(v, "sh", "")
		code = strings.ReplaceAll(code, "sz", "")
		for _, v := range items.Data {
			if v != nil && reflect.TypeOf(v).Kind() == reflect.Map {
				if ay, ok := v.(map[string]interface{})["m5"]; ok {
					if nil != ay && reflect.TypeOf(ay).Kind() == reflect.Slice {
						for _, p := range ay.([]interface{}) {
							if nil != p && reflect.TypeOf(p).Kind() == reflect.Slice {
								if len(p.([]interface{})) > 5 {
									tsItem, err := obj.onGetHestoryItem(code, p.([]interface{}))
									if nil == err && nil != tsItem {
										tsItemList = append(tsItemList, tsItem)
									}
								}
							}
						}
					}
					break
				}
			}
		}
		if len(tsItemList) > 0 {
			storedb.InsertMany(tsItemList, db, "STOCKSTS")
			utils.Logger.Info("hestory tx", zap.Int("cnt", len(tsItemList)),
				zap.String("code", code))
		}
	}
}

func (obj *TXHistory) onGetHestory() {
	bfqurl := "https://ifzq.gtimg.cn/appstock/app/kline/mkline?param=%s,m5,,480"
	go func() {
		for {
			if utils.Conf.IsTXHestory {
				_, _, _, hour, _ := utils.GetTimeMin5()
				if hour >= utils.Conf.HsStart {
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
						obj.hestory(bfqurl, "BFQ-TX-HESTORY.DB", listStocks)
					}
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}
