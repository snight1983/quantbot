package stocks

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"quantbot/utils"
	"time"
)

type EastMoneyOnTime struct {
	stockListUrl string
	stockMap     map[string]*StockAnalyze
}

// http://22.push2.eastmoney.com/api/qt/clist/get?cb=jQuery11240042884791376247566
// _1662445691924&pn=1&pz=20&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&in
// vt=2&wbp2u=|0|0|0|web&fid=f3&fs=m:0+t:6,m:0+t:80&fields=f1,f2,f3,f4,f5,f6,f7,
// f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,
// f136,f115,f152&_=1662445691925

type DfRS struct {
	Data struct {
		Total int          `json:"total"`
		Diff  []*StockItem `json:"diff"`
	} `json:"data"`
}

func (obj *EastMoneyOnTime) analyzeItem(curTs int64, item *StockItem, hour, nim int,
	collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	onStockItem(StockEM, curTs, &obj.stockMap, item, collect)
}

func (obj *EastMoneyOnTime) onGetPrice(collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	if nil == obj.stockMap {
		obj.stockMap = map[string]*StockAnalyze{}
	}
	obj.stockListUrl = "http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=10000&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,f136,f115,f152&_=%d"
	go func() {
		for {
			cur, _, _, hour, min := utils.GetTimeMin5()
			fmt.Println("++++ CUR ++++:", hour, min)
			if isLegalCNTS(hour, min) && utils.Conf.IsRealTime {
				url := fmt.Sprintf(obj.stockListUrl, rand.Intn(50), time.Now().UnixMilli())
				buf, err := utils.GetWithJSON(url, GHEADERS, nil)
				beg := time.Now().UnixMilli()
				if nil == err {
					out := DfRS{}
					json.Unmarshal(buf, &out)
					maxCnt := len(out.Data.Diff)
					if out.Data.Diff != nil && maxCnt > 0 {
						for _, v := range out.Data.Diff {
							obj.analyzeItem(cur, v, hour, min, collect)
						}
					}
					fmt.Printf("ts:%d | tUse:%d | mCnt:%d emmy\n", cur, time.Now().UnixMilli()-beg, maxCnt)
				}
			} else {
				obj.stockMap = map[string]*StockAnalyze{}
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
