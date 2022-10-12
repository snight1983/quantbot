package stocks

import (
	"encoding/json"
	"fmt"
	"quantbot/utils"
	"strconv"
	"time"
)

type SinaOnTime struct {
	stockListUrl string
	stockMap     map[string]*StockAnalyze
}

type SinaRS struct {
	Symbol        string  `json:"symbol"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Trade         string  `json:"trade"`
	Pricechange   float64 `json:"pricechange"`
	Changepercent float64 `json:"changepercent"`
	Buy           string  `json:"buy"`
	Sell          string  `json:"sell"`
	Settlement    string  `json:"settlement"`
	Open          string  `json:"open"`
	High          string  `json:"high"`
	Low           string  `json:"low"`
	Volume        int     `json:"volume"`
	Amount        int     `json:"amount"`
	Ticktime      string  `json:"ticktime"`
	Per           float64 `json:"per"`
	Pb            float64 `json:"pb"`
	Mktcap        float64 `json:"mktcap"`
	Nmc           float64 `json:"nmc"`
	Turnoverratio float64 `json:"turnoverratio"`
}

func (obj *SinaOnTime) analyzeItem(curTs int64, itemSina *SinaRS, hour, nim int,
	collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	stocItem := StockItem{}
	stocItem.LatestPrice, _ = strconv.ParseFloat(itemSina.Trade, 64)
	stocItem.Code = itemSina.Code
	stocItem.TradeDeal = itemSina.Volume / 100      // 手
	stocItem.TradeAmount = float64(itemSina.Amount) // 元
	onStockItem(StockSina, curTs, &obj.stockMap, &stocItem, collect)
}

func (obj *SinaOnTime) OnGetPrice(callback func(curTs int64, stock *StockAnalyze, tp int32)) {
	if nil == obj.stockMap {
		obj.stockMap = map[string]*StockAnalyze{}
	}
	obj.stockListUrl = "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=100&sort=symbol&asc=1&node=hs_a&symbol=&_s_r_a=page"
	go func() {
		for {
			maxCnt := 0
			for i := 0; i < 100; i++ {
				cur, _, _, hour, min := utils.GetTimeMin5()
				if isLegalCNTS(hour, min) && utils.Conf.IsRealTime {
					url := fmt.Sprintf(obj.stockListUrl, i)
					buf, err := utils.GetWithJSON(url, GHEADERS, nil)
					beg := time.Now().UnixMilli()
					if nil == err {
						out := []SinaRS{}
						json.Unmarshal(buf, &out)
						if nil != out {
							if len(out) <= 0 {
								break
							}
							for _, v := range out {
								obj.analyzeItem(cur, &v, hour, min, callback)
								maxCnt++
							}
						}
						fmt.Printf("ts:%d | tUse:%d | mCnt:%d sina\n", cur, time.Now().UnixMilli()-beg, maxCnt)
					}
					time.Sleep(time.Duration(1) * time.Second)
				} else {
					obj.stockMap = map[string]*StockAnalyze{}
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}
