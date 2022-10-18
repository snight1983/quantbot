package stocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"quantbot/storedb"
	"quantbot/utils"
	"sort"
	"strings"

	"sync"
	"time"

	"go.uber.org/zap"
)

type StockItem struct {
	LatestPrice float64 `json:"f2"`  // 最新价格
	TradeDeal   int     `json:"f5"`  // 成交量
	TradeAmount float64 `json:"f6"`  // 成交额
	Code        string  `json:"f12"` // 代码
}

type StockAnalyze struct {
	BegItem     StockItem
	EndItem     StockItem
	MaxPrice    float64
	MinPrice    float64
	LastPrice   float64
	LastTS      int64
	NextTS      int64
	TradeVolume float64
	TradeAmount float64
	IsChange    bool
}

// 上海
// http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:1+t:2,m:1+t:23&fields=f12&_=%d
// 深圳
// http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80&fields=f12&_=%d
// 北京
// http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m%3A0+t%3A81+s%3A2048&fields=f12&_=%d

const (
	StockEM int32 = iota
	Stock163
	StockSina
	StockXueQiu
	StockTX
)

var (
	lastTs    int64
	o163      S163OnTime
	easyMoney EastMoneyOnTime
	xueQiu    XueQiuOnTime

	easyMoneyhs EastMoneyHistory
	sinahs      SinaHistory
	txhs        TXHistory

	stockMapEsmLockCur    sync.RWMutex
	stockMapEsmCur        map[string]*StockAnalyze
	stockMapSinaLockCur   sync.RWMutex
	stockMapSinaCur       map[string]*StockAnalyze
	stockMapXueQiuLockCur sync.RWMutex
	stockMapXueQiuCur     map[string]*StockAnalyze
	stockMap163LockCur    sync.RWMutex
	stockMap163Cur        map[string]*StockAnalyze

	cacheList     []LastPriceCache
	cacheListLock sync.RWMutex
	LegalTs       []int64

	stockListTS int64
	shMap       map[string]interface{}
	shMapLock   sync.RWMutex
	szMap       map[string]interface{}
	szMapLock   sync.RWMutex

	cnTSMap map[int32]interface{}
)

type StocksCodeEM struct {
	Rc     int    `json:"rc"`
	Rt     int    `json:"rt"`
	Svr    int    `json:"svr"`
	Lt     int    `json:"lt"`
	Full   int    `json:"full"`
	Dlmkts string `json:"dlmkts"`
	Data   struct {
		Total int `json:"total"`
		Diff  []struct {
			F12 string `json:"f12"`
		} `json:"diff"`
	} `json:"data"`
}

var GHEADERS map[string]string

func getStocksInfo() error {

	url := fmt.Sprintf("http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:1+t:2,m:1+t:23&fields=f12&_=%d", rand.Intn(50), time.Now().UnixMilli())
	buf, err := utils.GetWithJSON(url, GHEADERS, nil)
	if nil != err {
		utils.Logger.Error("getStocksInfo", zap.String("err", err.Error()))
		return err
	}
	semSH := StocksCodeEM{}
	json.Unmarshal(buf, &semSH)
	if nil == semSH.Data.Diff {
		utils.Logger.Error("getStocksInfo", zap.String("err", "empty sh list"))
		return errors.New("empty sh list")
	}
	shMapLock.Lock()
	for _, v := range semSH.Data.Diff {
		shMap[v.F12] = nil
	}
	shMapLock.Unlock()

	url = fmt.Sprintf("http://%d.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80&fields=f12&_=%d", rand.Intn(50), time.Now().UnixMilli())
	buf, err = utils.GetWithJSON(url, GHEADERS, nil)
	if nil != err {
		utils.Logger.Error("getStocksInfo", zap.String("err", err.Error()))
		return err
	}
	semSZ := StocksCodeEM{}
	json.Unmarshal(buf, &semSZ)
	if nil == semSZ.Data.Diff {
		utils.Logger.Error("getStocksInfo", zap.String("err", "empty sh list"))
		return errors.New("empty sh list")
	}
	szMapLock.Lock()
	for _, v := range semSZ.Data.Diff {
		szMap[v.F12] = nil
	}
	szMapLock.Unlock()
	if len(szMap) > 2000 && len(shMap) > 2000 {
		stockListTS = time.Now().Unix()
	}
	return nil
}

type LastPriceCache struct {
	Ts      int64             `json:"ts"`
	IsStore bool              `json:"store"`
	Stocks  []*storedb.TsItem `json:"stocks"`
}

func onPriceItem(curTs int64, stockIn *StockAnalyze, tp int32) {
	stockItem := *stockIn
	if isInCNTS(int32(stockIn.LastTS % 10000)) {
		go func() {
			switch tp {
			case StockEM:
				stockMapEsmLockCur.Lock()
				defer stockMapEsmLockCur.Unlock()
				lastTs = time.Now().Unix()
				stockMapEsmCur[stockItem.BegItem.Code] = &stockItem
			case StockSina:
				stockMapSinaLockCur.Lock()
				defer stockMapSinaLockCur.Unlock()
				lastTs = time.Now().Unix()
				stockMapSinaCur[stockItem.BegItem.Code] = &stockItem
			case StockXueQiu:
				stockMapXueQiuLockCur.Lock()
				defer stockMapXueQiuLockCur.Unlock()
				lastTs = time.Now().Unix()
				stockMapXueQiuCur[stockItem.BegItem.Code] = &stockItem
			case Stock163:
				stockMap163LockCur.Lock()
				defer stockMap163LockCur.Unlock()
				lastTs = time.Now().Unix()
				stockMap163Cur[stockItem.BegItem.Code] = &stockItem
			}
		}()
	}
}

func onLocalStore() {
	if utils.Conf.LocalDBStore {
		for _, v := range utils.Conf.LocalDB {
			parts := strings.Split(v, "|")
			if len(parts) == 2 {
				storedb.LocalCsvStore(parts[0], parts[1])
			}
		}
	}
}

func StartCollect() {
	onLocalStore()
	GHEADERS = map[string]string{}
	GHEADERS["Accept-Language"] = "zh-CN,zh;q=0.9"
	GHEADERS["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36"
	GHEADERS["Connection"] = "keep-alive"
	GHEADERS["Accept"] = "*/*"
	cacheList = []LastPriceCache{}
	cnTSMap = map[int32]interface{}{}
	for _, v := range utils.Conf.CNTS {
		cnTSMap[v] = nil
	}

	stockMapEsmCur = map[string]*StockAnalyze{}
	stockMapSinaCur = map[string]*StockAnalyze{}
	stockMap163Cur = map[string]*StockAnalyze{}
	stockMapXueQiuCur = map[string]*StockAnalyze{}

	shMap = map[string]interface{}{}
	szMap = map[string]interface{}{}

	// 实时数据
	easyMoney.onGetPrice(onPriceItem)
	o163.onGetPrice(onPriceItem)
	xueQiu.onGetPrice(onPriceItem)

	// 历史数据
	easyMoneyhs.onGetHestory()
	sinahs.onGetHestory()
	txhs.onGetHestory()

	//onBFQFusion()

	go func() {
		for {
			now := time.Now().Unix()
			if lastTs > 0 {
				dif := now - lastTs
				fmt.Println("dif:", dif)
				if dif > 15 {
					cacheStocksSql()
					lastTs = 0
				}
			}
			if now-stockListTS > 600 {
				getStocksInfo()
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}

type FromeFloat struct {
	Value   float64
	SrcType int32
}

type FromeTool struct {
	array []FromeFloat
}

func (obj *FromeTool) Add(value float64, srcType int32) {
	if nil == obj.array {
		obj.array = []FromeFloat{}
	}
	obj.array = append(obj.array, FromeFloat{
		Value:   value,
		SrcType: srcType,
	})
}

func (obj *FromeTool) GetNearValue() *FromeFloat {
	ln := len(obj.array)
	if ln <= 0 {
		return nil
	}
	sort.Slice(obj.array, func(i, j int) bool {
		return obj.array[i].Value > obj.array[j].Value
	})
	item := &obj.array[ln/2]
	full := ""
	for _, v := range obj.array {
		part := fmt.Sprintf("value:%f,tp:%d | ", v.Value, v.SrcType)
		full += part
	}
	obj.array = []FromeFloat{}
	return item
}

func cacheStocksSql() bool {
	isChange := false

	stockMapEsmLockCur.Lock()
	defer stockMapEsmLockCur.Unlock()
	stockMapSinaLockCur.Lock()
	defer stockMapSinaLockCur.Unlock()
	stockMapXueQiuLockCur.Lock()
	defer stockMapXueQiuLockCur.Unlock()
	stockMap163LockCur.Lock()
	defer stockMap163LockCur.Unlock()

	if len(stockMap163Cur) > 0 || len(stockMapEsmCur) > 0 ||
		len(stockMapSinaCur) > 0 || len(stockMapXueQiuCur) > 0 {
		total := 0.0
		change := 0.0
		cache := LastPriceCache{}
		cache.Stocks = []*storedb.TsItem{}

		for key, esm := range stockMapEsmCur {
			stockSql := &storedb.TsItem{}

			sinaItem, okSina := stockMapSinaCur[key]
			xueQiuItem, okXueQiu := stockMapXueQiuCur[key]
			o163Item, ok163 := stockMap163Cur[key]

			fTool := FromeTool{}
			// [1] Open
			if esm.BegItem.LatestPrice > 0 {
				fTool.Add(esm.BegItem.LatestPrice, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.BegItem.LatestPrice > 0 {
				fTool.Add(sinaItem.BegItem.LatestPrice, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.BegItem.LatestPrice > 0 {
				fTool.Add(xueQiuItem.BegItem.LatestPrice, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.BegItem.LatestPrice > 0 {
				fTool.Add(o163Item.BegItem.LatestPrice, Stock163)
			}
			itemSel := fTool.GetNearValue()
			if nil == itemSel || itemSel.Value <= 0 {
				continue
			}
			stockSql.Open = itemSel.Value
			stockSql.OpenFr = itemSel.SrcType

			// [2] Close
			if esm.EndItem.LatestPrice > 0 {
				fTool.Add(esm.EndItem.LatestPrice, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.EndItem.LatestPrice > 0 {
				fTool.Add(sinaItem.EndItem.LatestPrice, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.EndItem.LatestPrice > 0 {
				fTool.Add(xueQiuItem.EndItem.LatestPrice, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.EndItem.LatestPrice > 0 {
				fTool.Add(o163Item.EndItem.LatestPrice, Stock163)
			}

			itemSel = fTool.GetNearValue()
			if nil == itemSel || itemSel.Value <= 0 {
				continue
			}
			stockSql.Close = itemSel.Value
			stockSql.CloseFr = itemSel.SrcType

			// [3] Max
			if esm.MaxPrice > 0 {
				fTool.Add(esm.MaxPrice, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.MaxPrice > 0 {
				fTool.Add(sinaItem.MaxPrice, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.MaxPrice > 0 {
				fTool.Add(xueQiuItem.MaxPrice, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.MaxPrice > 0 {
				fTool.Add(o163Item.MaxPrice, Stock163)
			}
			itemSel = fTool.GetNearValue()
			if nil == itemSel || itemSel.Value <= 0 {
				continue
			}
			stockSql.Max = itemSel.Value
			stockSql.MaxFr = itemSel.SrcType

			// [4] Min
			if esm.MinPrice > 0 {
				fTool.Add(esm.MinPrice, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.MinPrice > 0 {
				fTool.Add(sinaItem.MinPrice, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.MinPrice > 0 {
				fTool.Add(xueQiuItem.MinPrice, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.MinPrice > 0 {
				fTool.Add(o163Item.MinPrice, Stock163)
			}

			itemSel = fTool.GetNearValue()
			if nil == itemSel || itemSel.Value <= 0 {
				continue
			}
			stockSql.Min = itemSel.Value
			stockSql.MinFr = itemSel.SrcType
			// [5] Volume
			if esm.TradeVolume >= 0 {
				fTool.Add(esm.TradeVolume, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.TradeVolume > 0 {
				fTool.Add(sinaItem.TradeVolume, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.TradeVolume > 0 {
				fTool.Add(xueQiuItem.TradeVolume, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.TradeVolume > 0 {
				fTool.Add(o163Item.TradeVolume, Stock163)
			}
			itemSel = fTool.GetNearValue()
			if nil == itemSel || itemSel.Value < 0 {
				continue
			}
			stockSql.Volume = itemSel.Value
			stockSql.VolumeFr = itemSel.SrcType
			// [6] Amount
			if esm.TradeAmount > 0 {
				fTool.Add(esm.TradeAmount, StockEM)
			}
			if okSina && sinaItem != nil && sinaItem.TradeAmount >= 0 {
				fTool.Add(sinaItem.TradeAmount, StockSina)
			}
			if okXueQiu && xueQiuItem != nil && xueQiuItem.TradeAmount >= 0 {
				fTool.Add(xueQiuItem.TradeAmount, StockXueQiu)
			}
			if ok163 && o163Item != nil && o163Item.TradeAmount >= 0 {
				fTool.Add(o163Item.TradeAmount, Stock163)
			}
			itemSel = fTool.GetNearValue()
			if nil == itemSel || itemSel.Value < 0 {
				continue
			}
			change++
			stockSql.Amount = itemSel.Value
			stockSql.AmountFr = itemSel.SrcType
			stockSql.Code = key
			stockSql.TS = esm.LastTS
			stockSql.Date = esm.LastTS / 10000
			stockSql.ID = fmt.Sprintf("%s_%d", stockSql.Code, stockSql.TS)
			cache.Stocks = append(cache.Stocks, stockSql)
		}
		utils.Logger.Info("Size",
			zap.Any("EastMoney", len(stockMapEsmCur)),
			zap.Any("Sina", len(stockMapSinaCur)),
			zap.Any("XueQiu", len(stockMapXueQiuCur)),
			zap.Any("163", len(stockMap163Cur)))

		stockMapEsmCur = map[string]*StockAnalyze{}
		stockMapSinaCur = map[string]*StockAnalyze{}
		stockMapXueQiuCur = map[string]*StockAnalyze{}
		stockMap163Cur = map[string]*StockAnalyze{}

		if float64(change)/float64(total) > 0.3 && len(cache.Stocks) > 0 {
			isChange = true
			beg := time.Now().Unix()
			storedb.InsertMany(cache.Stocks, "CUR_REALTIME.DB", "STOCKSTS")
			utils.Logger.Info("cacheStocksSql:",
				zap.Any("TimeUse", time.Now().Unix()-beg),
				zap.Any("Cnt", len(cache.Stocks)))
			cacheListLock.Lock()
			defer cacheListLock.Unlock()
			cacheList = append(cacheList, cache)
			if len(cacheList) > 48 {
				cacheList = cacheList[len(cacheList)-48:]
			}
		}
	}
	return isChange
}

func isLegalCNTS(hour, minute int) bool {
	if hour >= 9 && hour < 16 {
		return true
	}
	return false
}

func isInCNTS(minute int32) bool {
	if _, ok := cnTSMap[minute]; ok {
		return true
	}
	return false
}

func onStockItem(tp int32, curTs int64, stockMap *map[string]*StockAnalyze, stocItem *StockItem,
	collect func(curTs int64, stock *StockAnalyze, tp int32)) {

	if v, ok := (*stockMap)[stocItem.Code]; ok {
		if curTs == v.LastTS {
			if v.MaxPrice < stocItem.LatestPrice {
				v.MaxPrice = stocItem.LatestPrice
			}
			if v.MinPrice > stocItem.LatestPrice {
				v.MinPrice = stocItem.LatestPrice
			}
		} else {
			v.LastPrice = stocItem.LatestPrice
			v.EndItem = *stocItem
			v.TradeVolume = float64(v.EndItem.TradeDeal - v.BegItem.TradeDeal)
			v.TradeAmount = v.EndItem.TradeAmount - v.BegItem.TradeAmount
			v.NextTS = curTs
			collect(curTs, v, tp)
			v.LastTS = curTs
			v.MaxPrice = stocItem.LatestPrice
			v.MinPrice = stocItem.LatestPrice
			v.LastPrice = stocItem.LatestPrice
			v.BegItem = *stocItem
		}
	} else {
		// 首次不处理
		netStock := StockAnalyze{}
		netStock.LastTS = curTs
		netStock.MaxPrice = stocItem.LatestPrice
		netStock.MinPrice = stocItem.LatestPrice
		netStock.LastPrice = stocItem.LatestPrice
		netStock.BegItem = *stocItem
		(*stockMap)[stocItem.Code] = &netStock
	}
}
