package service

import (
	"SecLayer/conf"
	"encoding/json"

	"github.com/astaxie/beego/logs"
)

func InitLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = conf.SecLayerSysConfig.LogLevel
	config["level"] = transferLogLevel(conf.SecLayerSysConfig.LogLevel)
	var jstr []byte
	jstr, err = json.Marshal(config)
	if err != nil {
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(jstr))
	if err != nil {
		return
	}
	return
}

func transferLogLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	default:
		return logs.LevelDebug
	}
}
