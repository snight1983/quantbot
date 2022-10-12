package utils

import (
	"time"
)

func init() {

}

func GetToday() int32 {

	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	return int32(year*10000 + month*100 + day)
}

func GetDayBefore(dayChange int) int {
	now := time.Now()
	now = now.AddDate(0, 0, dayChange)
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	return year*10000 + month*100 + day
}

func GetTodayMonth() int32 {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	return int32(year*100 + month)
}

func GetTs() int {
	now := time.Now()

	return now.Hour()*3600 + now.Minute()*60 + now.Minute()
}

func GetTimeTSAfter(after int64) (int32, int32, int32, int32) {
	nowts := time.Now().Unix()
	nowts += after
	now := time.Unix(nowts, 0)
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	return int32(year*10000 + month*100 + day),
		int32(year*100 + month),
		int32((now.UnixMilli() - 316800000) / 604800000),
		int32(now.Hour()*3600 + now.Minute()*60 + now.Minute())
}

func GetHour() int {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	return year*1000000 + month*10000 + day*100 + hour
}

func GetTimeMin5() (int64, int, int, int, int) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	min := (now.Minute()/5 + 1) * 5
	if min == 60 {
		min = 0
		hour++
	}
	if hour > 24 {
		hour = 0
	}
	return int64(year)*100000000 + int64(month)*1000000 + int64(day)*10000 + int64(hour)*100 + int64(min), month, day, hour, min
}

func GetPer4Hour() int {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := int(now.Hour() / 4)
	return year*1000000 + month*10000 + day*100 + hour
}

func GetHour12() (int32, int) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	if hour == 23 || hour == 24 || hour == 0 {
		if hour == 23 {
			nowT := now.AddDate(0, 0, 1)
			day = nowT.Day()
		}
		hour = 1
	} else if hour == 1 || hour == 2 {
		hour = 2
	} else if hour == 3 || hour == 4 {
		hour = 3
	} else if hour == 5 || hour == 6 {
		hour = 4
	} else if hour == 7 || hour == 8 {
		hour = 5
	} else if hour == 9 || hour == 10 {
		hour = 6
	} else if hour == 11 || hour == 12 {
		hour = 7
	} else if hour == 13 || hour == 14 {
		hour = 8
	} else if hour == 15 || hour == 16 {
		hour = 9
	} else if hour == 17 || hour == 18 {
		hour = 10
	} else if hour == 19 || hour == 20 {
		hour = 11
	} else {
		hour = 12
	}
	return int32(year*1000000 + month*10000 + day*100 + hour), hour
}
