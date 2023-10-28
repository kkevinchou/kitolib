package assetslog

import "github.com/kkevinchou/kitolib/log"

var Logger log.KitoLogger

func init() {
	Logger = log.NewStdOutLogger()
}

func SetLogger(logger log.KitoLogger) {
	Logger = logger
}
