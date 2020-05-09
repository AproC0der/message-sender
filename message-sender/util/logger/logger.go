package logger

import (
	"time"

	log "github.com/sirupsen/logrus"
	jack "gopkg.in/natefinch/lumberjack.v2"
)

var logger *jack.Logger
var Weblog =log.Fields{"product":"夜莺监控系统","program":"message-sender"}
func init() {
	//
	logger = &jack.Logger{
		Filename:   "log.txt",
		MaxSize:    100, // megabytes
		MaxBackups: 365,
		MaxAge:     180,  //days
		Compress:   true, // disabled by default
		LocalTime:  true,
	}
	log.SetOutput(logger)
	log.SetFormatter(TimeFormatter{&log.JSONFormatter{}})
}

//GetLogger 返回日志记录器
func GetLogger() *jack.Logger {
	return logger
}

//TimeFormatter 时间格式
type TimeFormatter struct {
	log.Formatter
}

//Format 格式化
func (t TimeFormatter) Format(e *log.Entry) ([]byte, error) {
	location, _ := time.LoadLocation("Asia/Chongqing")
	e.Time = e.Time.In(location)
	return t.Formatter.Format(e)
}
