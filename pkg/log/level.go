package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type fnTypeInformation struct {
	tag      string
	color    string
	logLevel logrus.Level
	stream   io.Writer
}

// Level type
type logFunctionType uint32

const (
	panicFn logFunctionType = iota
	fatalFn
	errorFn
	warnFn
	infoFn
	debugFn
	failFn
	doneFn
)

var fnTypeInformationMap = map[logFunctionType]*fnTypeInformation{
	debugFn: {
		tag:      "[debug]  ",
		color:    "green+b",
		logLevel: logrus.DebugLevel,
		stream:   stdout,
	},
	infoFn: {
		tag:      "[info]   ",
		color:    "cyan+b",
		logLevel: logrus.InfoLevel,
		stream:   stdout,
	},
	warnFn: {
		tag:      "[warn]   ",
		color:    "red+b",
		logLevel: logrus.WarnLevel,
		stream:   stdout,
	},
	errorFn: {
		tag:      "[error]  ",
		color:    "red+b",
		logLevel: logrus.ErrorLevel,
		stream:   stdout,
	},
	fatalFn: {
		tag:      "[fatal]  ",
		color:    "red+b",
		logLevel: logrus.FatalLevel,
		stream:   stdout,
	},
	panicFn: {
		tag:      "[panic]  ",
		color:    "red+b",
		logLevel: logrus.PanicLevel,
		stream:   stderr,
	},
	doneFn: {
		tag:      "[done] √ ",
		color:    "green+b",
		logLevel: logrus.InfoLevel,
		stream:   stdout,
	},
	failFn: {
		tag:      "[fail] X ",
		color:    "red+b",
		logLevel: logrus.ErrorLevel,
		stream:   stdout,
	},
}
