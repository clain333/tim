package pkg

import (
	"log"
	"time"
)

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("Error parsing time \"Asia/Shanghai\": %v", err)
	}
}
func GetLocalTime() time.Time {
	return time.Now().In(loc)
}

func TimeTrunStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
func GetLocalTimeUnix(expireSeconds int64) int64 {
	sTime := time.Now().In(loc)
	expireTimestamp := sTime.Unix() + expireSeconds
	return expireTimestamp
}
