package grpcmux

import (
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

func initComponents(conf *config.Config) {
	initLogging(conf.GetSubConfig("logging"))
}

func initLogging(conf *config.Config) {
	if conf == nil {
		return
	}

	level := conf.GetString("level", "debug")
	encoding := conf.GetString("encoding", "console")
	outputPath := conf.GetString("output", "")

	logging.Infow("logging initialized:", "level", level, "encoding", encoding)
	logging.Init(level, encoding, outputPath)
}
