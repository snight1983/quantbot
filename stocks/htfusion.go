package stocks

import (
	"fmt"
	"quantbot/storedb"
	"quantbot/utils"
	"time"
)

func onBFQFusion() {
	go func() {
		for {
			_, _, _, hour, _ := utils.GetTimeMin5()
			if hour > 18 {
				emMap, err := storedb.OnGetAll("BFQ-EM-HESTORY.DB", "STOCKSTS")
				if nil != err {
					fmt.Println(1)
				}
				sinaMap, err := storedb.OnGetAll("BFQ-SINA-HESTORY.DB", "STOCKSTS")
				if nil != err {
					fmt.Println(1)
				}
				txMap, err := storedb.OnGetAll("BFQ-TX-HESTORY.DB", "STOCKSTS")
				if nil != err {
					fmt.Println(1)
				}

				for key, v := range emMap {
					partSina, okSina := sinaMap[key]
					partTX, oktx := txMap[key]
					for ts, emItem := range v {
						var txItem *storedb.TsItem = nil
						var sinaItem *storedb.TsItem = nil
						if okSina && nil != partSina {
							sinaItem, _ = partSina[ts*100]
						}
						if oktx && nil != partTX {
							txItem, _ = partTX[ts]
						}

						if emItem != nil && sinaItem != nil {

						}
						if emItem != nil && txItem != nil {

						}
						fmt.Println("EM", emItem)
						fmt.Println("SN", sinaItem)
						fmt.Println("TX", txItem)
					}
				}

				fmt.Println(emMap)
				fmt.Println(sinaMap)
				fmt.Println(txMap)
				//obj.hestory(bfqurl, "BFQ-EM-HESTORY.DB", listStocks, beg, end)
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}
