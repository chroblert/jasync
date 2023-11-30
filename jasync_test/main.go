package main

import (
	"fmt"
	"github.com/chroblert/jasync"
	"github.com/chroblert/jlog"
	"github.com/schollz/progressbar/v3"
	"time"
)

//func TestAsync_AddAndRun(t *testing.T) {
//	a := jasync.NewAR(100)
//	defer jlog.Flush()
//	//jlog.SetStoreToFile(false)
//	jlog.IsIniCreateNewLog(true)
//	jlog.SetUseConsole(true)
//	for i := 0; i < 10100; i++ {
//		//jlog.Infof("out i:%d\n", i)
//		a.AddAndRun("", func(i int) {
//			jlog.Infof("info i:%d\n", i)
//		}, func() {
//
//		}, i)
//	}
//	a.Wait()
//	//time.Sleep(10 * time.Second)
//}

func main() {
	a := jasync.NewAR(100, false)
	defer jlog.Flush()
	//jlog.SetStoreToFile(false)
	jlog.IsIniCreateNewLog(true)
	jlog.SetUseConsole(true)
	bar := progressbar.Default(1000)
	for i := 0; i < 1000; i++ {
		err := a.Init(fmt.Sprintf("task-%d", i)).CAdd(func(i int) string {
			jlog.Infof("func-1:%d\n", i)
			return fmt.Sprintf("2222-%d", i)
		}, i).CAdd(func(s string, i, d int) {
			jlog.Infof("func-2:%s:%d\n", s, i)
			time.Sleep(100 * time.Millisecond)
		}, i, 4).CAdd(func() {
			bar.Add(1)
		}).CDO()
		if err != nil {
			jlog.Error(err)
		}
	}
	select {
	case <-a.Done():
		jlog.Info("done")
	}
}
