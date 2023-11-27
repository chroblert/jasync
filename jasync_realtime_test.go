package jasync

import (
	"fmt"
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
	a.Wait()
	//time.Sleep(10 * time.Second)
}

func TestAsync_ChainAdd(t *testing.T) {
	a := NewAR(1000, false)
	defer jlog.Flush()
	//jlog.SetStoreToFile(false)
	jlog.IsIniCreateNewLog(true)
	jlog.SetUseConsole(true)
	for i := 0; i < 101000; i++ {
		err := a.Init(fmt.Sprintf("task-%d", i)).CAdd(func(i int) string {
			jlog.Infof("func-1:%d\n", i)
			return fmt.Sprintf("2222-%d", i)
		}, i).CAdd(func(s string, i, d int) { jlog.Infof("func-2:%s:%d\n", s, i) }, i, 4).CDO()
		if err != nil {
			jlog.Error(err)
		}
	}
	a.Wait()
}
