package jasync

import (
	"github.com/chroblert/jlog"
	"testing"
)

func TestAsync_AddAndRun(t *testing.T) {
	a := NewAR(100)
	defer jlog.Flush()
	//jlog.SetStoreToFile(false)
	jlog.IsIniCreateNewLog(true)
	jlog.SetUseConsole(true)
	for i := 0; i < 10100; i++ {
		//jlog.Infof("out i:%d\n", i)
		a.AddAndRun("", func(i int) {
			jlog.Infof("info i:%d\n", i)
		}, func() {

		}, i)
	}
	a.Wait2()
	//time.Sleep(10 * time.Second)
}
