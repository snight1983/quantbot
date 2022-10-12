package stocks

import (
	"encoding/json"
	"fmt"
	"quantbot/utils"
	"time"
)

type XueQiuItem struct {
	Symbol  string  `json:"symbol"`
	Current float64 `json:"current"`
	Amount  float64 `json:"amount"`
	Volume  int     `json:"volume"`
}

type XueQiuData struct {
	Count int          `json:"count"`
	List  []XueQiuItem `json:"list"`
}

type XueQiuRS struct {
	Data             XueQiuData `json:"data"`
	ErrorCode        int        `json:"error_code"`
	ErrorDescription string     `json:"error_description"`
}

type XueQiuOnTime struct {
	stockListUrl string
	stockMap     map[string]*StockAnalyze
}

// https://xueqiu.com/service/v5/stock/screener/quote/list?page=1&size=5000&order=desc&orderby=percent&order_by=percent&market=CN&type=sh_sz&_=1663758799313
func (obj *XueQiuOnTime) analyzeItem(curTs int64, itemXQ XueQiuItem, hour, nim int,
	collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	stocItem := StockItem{}
	if len(itemXQ.Symbol) == 8 {
		stocItem.LatestPrice = itemXQ.Current
		stocItem.Code = itemXQ.Symbol[2:]
		stocItem.TradeDeal = itemXQ.Volume / 100
		stocItem.TradeAmount = itemXQ.Amount
		onStockItem(StockXueQiu, curTs, &obj.stockMap, &stocItem, collect)
	}
}

func (obj *XueQiuOnTime) onGetPrice(collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	if nil == obj.stockMap {
		obj.stockMap = map[string]*StockAnalyze{}
	}
	obj.stockListUrl = "https://xueqiu.com/service/v5/stock/screener/quote/list?page=1&size=5000&order=desc&orderby=percent&order_by=percent&market=CN&type=sh_sz&_=%d"
	go func() {
		for {
			cur, _, _, hour, min := utils.GetTimeMin5()
			if isLegalCNTS(hour, min) && utils.Conf.IsRealTime {
				url := fmt.Sprintf(obj.stockListUrl, time.Now().UnixMilli())
				buf, err := utils.GetWithJSON(url, GHEADERS, nil)
				beg := time.Now().UnixMilli()
				if nil == err {
					out := XueQiuRS{}
					json.Unmarshal(buf, &out)
					maxCnt := len(out.Data.List)
					if out.Data.List != nil && maxCnt > 0 {
						for _, v := range out.Data.List {
							obj.analyzeItem(cur, v, hour, min, collect)
						}
					}
					fmt.Printf("ts:%d | tUse:%d | mCnt:%d xueq\n", cur, time.Now().UnixMilli()-beg, maxCnt)
				}
			} else {
				obj.stockMap = map[string]*StockAnalyze{}
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
