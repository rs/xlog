package xlog

import "testing"

func TestNopLogger(t *testing.T) {
	// cheap cover score upper
	NopLogger.SetField("name", "value")
	NopLogger.OutputF(LevelInfo, 0, "", nil)
	NopLogger.Debug()
	NopLogger.Debugf("format")
	NopLogger.Info()
	NopLogger.Infof("format")
	NopLogger.Warn()
	NopLogger.Warnf("format")
	NopLogger.Error()
	NopLogger.Errorf("format")
	exit1 = func() {}
	NopLogger.Fatal()
	NopLogger.Fatalf("format")
	NopLogger.Write([]byte{})
	NopLogger.Output(0, "")
}
