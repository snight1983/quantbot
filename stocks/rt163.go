package stocks

import (
	"encoding/json"
	"fmt"
	"quantbot/utils"
	"time"
)

type S163OnTime struct {
	stockListUrl string
	stockMap     map[string]*StockAnalyze
}

func (obj *S163OnTime) OnNewItem(item *StockAnalyze) {
	fmt.Println(item)
}

type S163StockItem struct {
	CODE       string  `json:"CODE"`
	FIVEMINUTE int     `json:"FIVE_MINUTE"`
	HIGH       float64 `json:"HIGH"`
	HS         float64 `json:"HS"`
	LB         float64 `json:"LB"`
	LOW        float64 `json:"LOW"`
	MCAP       float64 `json:"MCAP"`
	MFRATIO    struct {
		MFRATIO2  float64 `json:"MFRATIO2"`
		MFRATIO10 float64 `json:"MFRATIO10"`
	} `json:"MFRATIO"`
	MFSUM     float64 `json:"MFSUM"`
	NAME      string  `json:"NAME"`
	OPEN      float64 `json:"OPEN"`
	PE        float64 `json:"PE"`
	PERCENT   float64 `json:"PERCENT"`
	PRICE     float64 `json:"PRICE"`
	SNAME     string  `json:"SNAME"`
	SYMBOL    string  `json:"SYMBOL"`
	TCAP      float64 `json:"TCAP"`
	TURNOVER  float64 `json:"TURNOVER"`
	UPDOWN    float64 `json:"UPDOWN"`
	VOLUME    int     `json:"VOLUME"`
	WB        float64 `json:"WB"`
	YESTCLOSE float64 `json:"YESTCLOSE"`
	ZF        float64 `json:"ZF"`
	NO        int     `json:"NO"`
}

type S163Stocks struct {
	Page      int             `json:"page"`
	Count     int             `json:"count"`
	Order     string          `json:"order"`
	Total     int             `json:"total"`
	Pagecount int             `json:"pagecount"`
	Time      string          `json:"time"`
	List      []S163StockItem `json:"list"`
}

func (obj *S163OnTime) analyzeItem(curTs int64, item163 *S163StockItem, hour, nim int,
	collect func(curTs int64, stock *StockAnalyze, tp int32)) {

	stocItem := StockItem{}
	stocItem.LatestPrice = item163.PRICE
	stocItem.Code = item163.SYMBOL
	stocItem.TradeDeal = item163.VOLUME / 100 // 手
	stocItem.TradeAmount = item163.TURNOVER   // 元
	onStockItem(Stock163, curTs, &obj.stockMap, &stocItem, collect)
}

func (obj *S163OnTime) onGetPrice(collect func(curTs int64, stock *StockAnalyze, tp int32)) {
	if nil == obj.stockMap {
		obj.stockMap = map[string]*StockAnalyze{}
	}
	obj.stockListUrl = "http://quotes.money.163.com/hs/service/diyrank.php?host=http%3A%2F%2Fquotes.money.163.com%2Fhs%2Fservice%2Fdiyrank.php&page=0&query=STYPE%3AEQA&fields=NO%2CSYMBOL%2CNAME%2CPRICE%2CPERCENT%2CUPDOWN%2CFIVE_MINUTE%2COPEN%2CYESTCLOSE%2CHIGH%2CLOW%2CVOLUME%2CTURNOVER%2CHS%2CLB%2CWB%2CZF%2CPE%2CMCAP%2CTCAP%2CMFSUM%2CMFRATIO.MFRATIO2%2CMFRATIO.MFRATIO10%2CSNAME%2CCODE%2CANNOUNMT%2CUVSNEWS&sort=PERCENT&order=desc&count=10000&type=query"
	headers := map[string]string{}
	headers["Accept-Language"] = "zh-CN,zh;q=0.9"
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36"
	headers["Connection"] = "keep-alive"
	headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
	headers["Upgrade-Insecure-Requests"] = "1"
	go func() {
		for {
			cur, _, _, hour, min := utils.GetTimeMin5()
			if isLegalCNTS(hour, min) && utils.Conf.IsRealTime {
				buf, err := utils.GetWithJSON(obj.stockListUrl, headers, nil)
				beg := time.Now().UnixMilli()
				if nil == err {
					out := S163Stocks{}
					json.Unmarshal(buf, &out)
					al := map[string]interface{}{}
					for _, v := range out.List {
						al[v.CODE] = nil
						obj.analyzeItem(cur, &v, hour, min, collect)
					}
					fmt.Printf("ts:%d | tUse:%d | mCnt:%d 163\n", cur, time.Now().UnixMilli()-beg, len(out.List))
				}
			} else {
				obj.stockMap = map[string]*StockAnalyze{}
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
