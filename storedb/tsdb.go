package storedb

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"quantbot/utils"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func getDBTable(dbinfo, tableName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath.Join(utils.Conf.StockRoot, dbinfo))
	if nil != err {
		utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
		return nil, err
	}
	dable := "SELECT * FROM sqlite_master WHERE type=\"table\";"
	rows, err := db.Query(dable)
	if nil != err {
		utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
		return nil, err
	}
	find := false
	for rows.Next() {
		info := TableInfo{}
		if err = rows.Scan(&info.Type, &info.Name, &info.Name1, &info.Cnt, &info.Sql); nil != err {
			utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
			return nil, err
		}
		if tableName == info.Name {
			find = true
		}
	}
	if !find {
		//"ID" VARCHAR(32) PRIMARY KEY AUTOINCREMENT,
		sql_table := `
		CREATE TABLE IF NOT EXISTS %s (	 
			 "ID" VARCHAR(32) PRIMARY KEY,
			 "TS" INTEGER,
			 "Date" INTEGER,
			 "Open" DOUBLE,
			 "OpenFr" INTEGER,
			 "Close" DOUBLE,
			 "CloseFr" INTEGER,
			 "Max" DOUBLE,
			 "MaxFr" INTEGER,
			 "Min" DOUBLE,
			 "MinFr" INTEGER,
			 "Volume" DOUBLE,
			 "VolumeFr" INTEGER,
			 "Amount" DOUBLE,
			 "AmountFr" INTEGER,
			 "Code" VARCHAR(10)
		);`

		sql_table = fmt.Sprintf(sql_table, tableName)
		_, err := db.Exec(sql_table)
		if nil != err {
			utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
			return nil, err
		}
		sql_index := fmt.Sprintf("CREATE INDEX Code_index ON %s (Code)", tableName)
		_, err = db.Exec(sql_index)
		if nil != err {
			utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
			return nil, err
		}

		sql_index = fmt.Sprintf("CREATE INDEX ts_index ON %s (TS)", tableName)
		_, err = db.Exec(sql_index)
		if nil != err {
			utils.Logger.Info("getDBTable", zap.String("err", err.Error()))
			return nil, err
		}
	}
	return db, nil
}

func InsertMany(items []*TsItem, dbInfo, tableName string) error {
	db, err := getDBTable(dbInfo, tableName)
	if nil != err {
		utils.Logger.Info("InsertMany", zap.String("err", err.Error()))
		return err
	}
	defer db.Close()

	ts, err := db.Begin()
	if nil != err {
		utils.Logger.Info("InsertMany", zap.String("err", err.Error()))
		return err
	}
	sql := fmt.Sprintf("REPLACE INTO %s(ID, TS, Date, Open, Close, Max, Min, Volume, Amount, Code, OpenFr, CloseFr, MaxFr, MinFr, VolumeFr, AmountFr) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tableName)
	stmt, err := ts.Prepare(sql)
	if nil != err {
		return err
	}
	for _, v := range items {
		stmt.Exec(v.ID, v.TS, v.Date, v.Open, v.Close,
			v.Max, v.Min, v.Volume, v.Amount, v.Code, v.OpenFr,
			v.CloseFr, v.MaxFr, v.MinFr, v.VolumeFr, v.AmountFr)
	}
	err = ts.Commit()
	if nil != err {
		utils.Logger.Info("InsertMany", zap.String("err", err.Error()))
		return err
	}
	return nil
}

func OnGetAll(dbInfo, tableName string) (map[string]map[int64]*TsItem, error) {
	db, err := getDBTable(dbInfo, tableName)
	result := map[string]map[int64]*TsItem{}
	if nil != err {
		utils.Logger.Info("OnGetAll", zap.String("err", err.Error()))
		return nil, err
	}
	sql := fmt.Sprintf("SELECT ID, TS, Date, Open, Close, Max, Min, Volume, Amount, Code FROM %s", tableName)
	rows, err := db.Query(sql)
	if nil != err {
		return nil, err
	}
	for rows.Next() {
		item := TsItem{}
		err = rows.Scan(&item.ID, &item.TS, &item.Date, &item.Open,
			&item.Close, &item.Max, &item.Min, &item.Volume,
			&item.Amount, &item.Code)
		if nil != err {
			fmt.Println(err)
			continue
		}
		if smap, ok := result[item.Code]; ok {
			smap[item.TS] = &item
		} else {
			smap := map[int64]*TsItem{}
			smap[item.TS] = &item
			result[item.Code] = smap
		}
	}
	return result, nil
}
