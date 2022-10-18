package storedb

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"quantbot/utils"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"
)

type TableInfo struct {
	Type  string
	Name  string
	Name1 string
	Cnt   int64
	Sql   string
}

var (
	cntSum int32

	//stocksFullDBQFQ *sql.DB
	//stocksFullDBHFQ *sql.DB
	//stocksFullDBBFQ *sql.DB
)

type TsItem struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	TS       int64   `json:"ts"`
	Date     int64   `json:"date"`
	Open     float64 `json:"open"`
	OpenFr   int32   `json:"openfr"`
	Close    float64 `json:"close"`
	CloseFr  int32   `json:"closefr"`
	Max      float64 `json:"max"`
	MaxFr    int32   `json:"maxfr"`
	Min      float64 `json:"min"`
	MinFr    int32   `json:"minfr"`
	Volume   float64 `json:"volume"`
	VolumeFr int32   `json:"volumefr"`
	Amount   float64 `json:"amount"`
	AmountFr int32   `json:"amountfr"`
}

func getStockDB(dbinfo, tableName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbinfo)
	if nil != err {
		return nil, err
	}
	sql := "SELECT * FROM sqlite_master WHERE type=\"table\";"
	rows, err := db.Query(sql)
	if nil != err {
		utils.Logger.Info("InitStock", zap.String("err", err.Error()))
		return nil, err
	}
	find := false
	for rows.Next() {
		info := TableInfo{}
		if err = rows.Scan(&info.Type, &info.Name, &info.Name1, &info.Cnt, &info.Sql); nil != err {
			utils.Logger.Info("InitStock", zap.String("err", err.Error()))
			return nil, err
		}
		if info.Name == tableName {
			find = true
		}
	}
	if !find {
		sql_table := `
		CREATE TABLE IF NOT EXISTS t_%s (
			 "TS" INTEGER PRIMARY KEY,
			 "Date" INTEGER,
			 "Open" DOUBLE,
			 "Close" DOUBLE,
			 "Max" DOUBLE,
			 "Min" DOUBLE,
			 "Volume" DOUBLE,
			 "Amount" DOUBLE,
			 "Code" VARCHAR(10),
			 "Statistics" BLOB
		);`
		sqltable := fmt.Sprintf(sql_table, tableName)
		_, err := db.Exec(sqltable)
		if nil != err {
			utils.Logger.Info("InitStock", zap.String("err", err.Error()))
			return nil, err
		}
	}
	return db, nil
}

func InsertManyLocal(dbinfo, tableName string, items []*TsItem) error {
	db, err := getStockDB(dbinfo, tableName)
	if nil != err {
		return err
	}
	defer db.Close()
	ts, err := db.Begin()
	if nil != err {
		return err
	}
	stmt, err := ts.Prepare(fmt.Sprintf("REPLACE INTO t_%s(TS, Date, Open, Close, Max, Min, Volume, Amount, Code) values(?,?,?,?,?,?,?,?,?)", tableName))
	if nil != err {
		return err
	}
	for _, v := range items {
		stmt.Exec(v.TS, v.Date, v.Open, v.Close, v.Max, v.Min, v.Volume, v.Amount, v.Code)
	}
	ts.Commit()
	return nil
}

func LoadStocksCsv(dbinfo, path, code string) error {
	f, err := os.Open(path)
	if err != nil {
		utils.Logger.Info("LoadStocksCsv", zap.String("err", err.Error()))
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	items := []*TsItem{}
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			utils.Logger.Error("LoadStocksCsv", zap.String("err", err.Error()))
			return err
		}
		if len(line) == 10 {
			dTime, err := strconv.ParseInt(line[1], 10, 64)
			if nil != err {
				continue
			}
			stockItem := &TsItem{}
			sList := strings.Split(line[2], ".")
			if len(sList) != 2 {
				utils.Logger.Warn("LoadStocksCsv code", zap.String("line[2]", line[2]))
				continue
			}
			if sList[1] != code {
				utils.Logger.Warn("LoadStocksCsv code", zap.String(sList[1]+"-"+code, sList[1]+"-"+code))
				continue
			}
			stockItem.Code = sList[1]
			if stockItem.Code != code {
				utils.Logger.Warn("LoadStocksCsv code", zap.String("code err", code))
				continue
			}
			stockItem.TS = dTime / 100000
			stockItem.Date = stockItem.TS / 10000
			stockItem.Open, err = strconv.ParseFloat(line[3], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}
			stockItem.Close, err = strconv.ParseFloat(line[6], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}
			stockItem.Max, err = strconv.ParseFloat(line[4], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}
			stockItem.Min, err = strconv.ParseFloat(line[5], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}

			if stockItem.Max < stockItem.Min {
				utils.Logger.Warn("LoadStocksCsv error max min",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min))
				continue
			}

			if stockItem.Open > stockItem.Max || stockItem.Open < stockItem.Min {
				utils.Logger.Warn("LoadStocksCsv  error open",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min),
					zap.Float64("Open", stockItem.Open))
				continue
			}

			if stockItem.Close > stockItem.Max || stockItem.Close < stockItem.Min {
				utils.Logger.Warn("LoadStocksCsv  error Close",
					zap.Float64("Max", stockItem.Max),
					zap.Float64("Min", stockItem.Min),
					zap.Float64("Close", stockItem.Close))
				continue
			}
			stockItem.Volume, err = strconv.ParseFloat(line[7], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}
			stockItem.Volume /= 100
			stockItem.Amount, err = strconv.ParseFloat(line[8], 64)
			if nil != err {
				utils.Logger.Warn("LoadStocksCsv", zap.String("err", err.Error()))
				continue
			}
			if stockItem.Volume <= 0 {
				if stockItem.Open == stockItem.Close &&
					stockItem.Open == stockItem.Max &&
					stockItem.Open == stockItem.Min {
					continue
				} else {
					utils.Logger.Info("LoadStocksCsv no Volume")
				}
			}
			atomic.AddInt32(&cntSum, 1)
			if cntSum%10000 == 0 {
				fmt.Println("total cnt:", dbinfo, cntSum)
			}
			items = append(items, stockItem)
		}
	}
	if len(items) > 0 {
		return InsertManyLocal(dbinfo, code, items)
	}
	return nil
}

func LocalCsvStore(dbinfo, csvFolder string) {
	filepath.Walk(csvFolder, func(fname string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			name := fi.Name()
			nList := strings.Split(name, ".")
			if len(nList) == 3 {
				nList = strings.Split(nList[1], "_")
				if len(nList) == 2 {
					LoadStocksCsv(dbinfo, fname, nList[0])
					runtime.GC()
				}
			}
		}
		return nil
	})
}
