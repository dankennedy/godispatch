package godispatch

import (
	"os"
	"testing"
)

func TestStandardLoggerNoPrefix(t *testing.T) {
	log := NewStandardLogger(os.Stdout)
	for _, prefix := range []string{"", "[context]"} {
		log.SetPrefix(prefix)
		log.Debug("some test text")
		log.Debugf("some %s text %d", "numeric", 5)
		log.Info("some test text")
		log.Infof("some %s text %d", "numeric", 5)
		log.Warn("some test text")
		log.Warnf("some %s text %d", "numeric", 5)
		log.Error("some test text")
		log.Errorf("some %s text %d", "numeric", 5)
	}
}
