package godispatch

import "testing"

func TestStandardLoggerDebug(t *testing.T) {
	log := NewStandardLogger()
	log.Debug("some test text")
	log.Debugf("some %s text %d", "numeric", 5)
	log.Info("some test text")
	log.Infof("some %s text %d", "numeric", 5)
	log.Warn("some test text")
	log.Warnf("some %s text %d", "numeric", 5)
	log.Error("some test text")
	log.Errorf("some %s text %d", "numeric", 5)
}
