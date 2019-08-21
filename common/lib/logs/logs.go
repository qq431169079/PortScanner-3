package logs

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// 设定日志级别和文件
func GetLogger(level, file string) *logrus.Entry {
	level = strings.ToUpper(level)

	var (
		levelMap = map[string]logrus.Level{
			"TRACE": logrus.TraceLevel,
			"DEBUG": logrus.DebugLevel,
			"INFO":  logrus.InfoLevel,
			"WARN":  logrus.WarnLevel,
			"ERROR": logrus.ErrorLevel,
			"PANIC": logrus.PanicLevel,
		}
		logger *logrus.Logger
	)
	logger = logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetReportCaller(true)

	// 不记录绝对路径
	//logger.Formatter = &logrus.TextFormatter{
	//	CallerPrettyfier: func(f *runtime.Frame) (string, string) {
	//		path, err := util.GetAbsPath()
	//		if err != nil {
	//			fmt.Println("Get absolute path fail,", err)
	//		}
	//		filename := strings.Replace(f.File, path, "", -1)
	//		return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
	//	},
	//}

	if file != "" {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			logger.SetOutput(os.Stdout)
			logger.Panicf("please check the path and permission of your log file, error info: ", err)
		} else {
			logger.SetOutput(f)
		}
	} else {
		logger.SetOutput(os.Stdout)
	}
	if l, ok := levelMap[level]; ok {
		logger.SetLevel(l)
	} else {
		logger.SetLevel(logrus.ErrorLevel)
		logger.Panicf(`%s log level is not supported, default log level is "ERROR"`, level)
	}
	Log := logrus.NewEntry(logger)
	return Log
}
