package log

import (
	"fmt"
	"os"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/sirupsen/logrus"
)

func NewDiscardLogger() iface.Logger {
	return &DiscardLogger{}
}

// DiscardLogger just discards every log statement
type DiscardLogger struct{}

func (d *DiscardLogger) Fatal(args ...interface{}) {
	d.Error(args...)
	os.Exit(1)
}

func (d *DiscardLogger) Fatalf(format string, args ...interface{}) {
	d.Errorf(format, args...)
	os.Exit(1)
}

func (d *DiscardLogger) Panic(args ...interface{}) {
	panic(args)
}

func (d *DiscardLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (d *DiscardLogger) Done(args ...interface{}) {
	return
}

func (d *DiscardLogger) Donef(format string, args ...interface{}) {
	return
}

func (d *DiscardLogger) Fail(args ...interface{}) {
	return
}

func (d *DiscardLogger) Failf(format string, args ...interface{}) {
	return
}

func (d *DiscardLogger) Print(level logrus.Level, args ...interface{}) {
	return
}

func (d *DiscardLogger) Printf(level logrus.Level, format string, args ...interface{}) {
	return
}

func (d *DiscardLogger) Write(message []byte) (int, error) {
	return len(message), nil
}

func (d *DiscardLogger) WriteString(message string) {
	return
}

func (d *DiscardLogger) SetLevel(level logrus.Level) {
	return
}

func (d *DiscardLogger) GetLevel() logrus.Level {
	return logrus.DebugLevel
}

// Debug implements logger interface
func (d *DiscardLogger) Debug(args ...interface{}) {}

// Debugf implements logger interface
func (d *DiscardLogger) Debugf(format string, args ...interface{}) {}

// Info implements logger interface
func (d *DiscardLogger) Info(args ...interface{}) {}

// Infof implements logger interface
func (d *DiscardLogger) Infof(format string, args ...interface{}) {}

// Warn implements logger interface
func (d *DiscardLogger) Warn(args ...interface{}) {}

// Warnf implements logger interface
func (d *DiscardLogger) Warnf(format string, args ...interface{}) {}

// Error implements logger interface
func (d *DiscardLogger) Error(args ...interface{}) {}

// Errorf implements logger interface
func (d *DiscardLogger) Errorf(format string, args ...interface{}) {}
