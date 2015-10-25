package xlog

import "testing"

func TestNopLogger(t *testing.T) {
	// cheap cover score upper
	nopLogger.SetField("name", "value")
	nopLogger.Debug()
	nopLogger.Debugf("format")
	nopLogger.Info()
	nopLogger.Infof("format")
	nopLogger.Warn()
	nopLogger.Warnf("format")
	nopLogger.Error()
	nopLogger.Errorf("format")
}
