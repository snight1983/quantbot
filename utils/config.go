package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var (
	Conf ConfigAll
)

type ConfigAll struct {
	IP            string   `json:"ip"`
	Port          int      `json:"port"`
	StockRoot     string   `json:"stockroot"`
	RegServer     string   `json:"regserver"`
	IsEMHestory   bool     `json:"isemhestory"`
	IsSinaHestory bool     `json:"issinahestory"`
	IsTXHestory   bool     `json:"istxhestory"`
	IsRealTime    bool     `json:"isrealtime"`
	CNTS          []int32  `json:"cnts"`
	LocalDBStore  bool     `json:"localdbstore"`
	LocalDB       []string `json:"localdb"`
}

func init() {
	path, err := GetCurrentPath()
	InitLog("", path, "stocks", "1.0.0", "stocks.log")
	if nil != err {
		Logger.Warn(fmt.Sprintf("getCurrentPath failed Err:%s", err.Error()))
		return
	}
	dest := filepath.Join(path, "config.json")
	filePtr, err := os.Open(dest)
	if err != nil {
		Logger.Warn(fmt.Sprintf("Open file failed Err:%s", err.Error()))
		return
	}
	defer filePtr.Close()
	decoder := json.NewDecoder(filePtr)
	decoder.Decode(&Conf)
}
