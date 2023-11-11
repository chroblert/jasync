package jasync

import (
	"github.com/chroblert/jlog"
	"strconv"
	"testing"
	"time"
)

func TestMain2(t *testing.T) {
	jlog.SetVerbose(false)
	a := New(true)
	a.Add("t01", func() (string, int) {
		jlog.NDebug("1xx")
		return "test 2", 22
	}, func() {
	})
	for i := 1; i < 5; i++ {
		t := i
		jlog.Debug(a.AddR("t"+strconv.Itoa(i), func() []int {
			//jlog.Debug("t"+strconv.Itoa(t))
			return []int{1, t}
		}, func(s []int) {
			jlog.Debug("start ", s)
			time.Sleep(time.Duration(t) * time.Second)
			jlog.Debug(s)
		}))
	}
	time.Sleep(2 * time.Second)
	a.Run(2)
	a.Wait()

	jlog.Debug(a.GetTaskAllTotal())
	jlog.Debug(a.GetTaskCurAllTotal())
	jlog.Debug(a.Run(1))
	a.Wait()
	jlog.Debug(a.GetTasksResult())
	jlog.Debug(a.GetTaskResult("kk"))
	jlog.Debug(a.GetTaskAllTotal())
	jlog.Debug(a.GetTaskCurAllTotal())
	jlog.Debug(a.GetTaskCurNeedDoCount())
	a.PrintTaskStatus("t12", true)
	jlog.Debug("kkkkkkkjkj")
}
