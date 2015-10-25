package xlog

import "testing"

func TestNopLogger(t *testing.T) {
	// cheap cover score upper
	NopLogger.SetField("name", "value")
	NopLogger.Debug()
	NopLogger.Debugf("format")
	NopLogger.Info()
	NopLogger.Infof("format")
	NopLogger.Warn()
	NopLogger.Warnf("format")
	NopLogger.Error()
	NopLogger.Errorf("format")
	NopLogger.Write([]byte{})
}
