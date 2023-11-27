package jasync

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"golang.org/x/sync/semaphore"
	"reflect"
	"sync"
)

// Async 异步执行对象
type AsyncRealtime struct {
	mu *sync.RWMutex
	//是否显示进度
	verbose bool
	// 信号量
	sem *semaphore.Weighted
	//
	wg       *sync.WaitGroup
	taskName string
}

//	type taskStatusStruct struct {
//		taskStatusStruct  int   // 任务状态 0: init,1:queue,2: doing,3: done
//		taskBegTime int64 // 任务开始时间
//		taskEndTime int64 // 任务结束时间
//	}
func (ar *AsyncRealtime) Wait2() {
	ar.wg.Wait()
}
func (ar *AsyncRealtime) AddAndRun(name string, funcHandler interface{}, printHandler interface{}, params ...interface{}) (task_name string, b_success bool, err error) {
	ar.wg.Add(1)
	if name == "" {
		var err2 error
		name, err2 = uuid.GenerateUUID()
		if err2 != nil {
			return "", false, err2
		}
	}
	// 获取信号量
	ar.sem.Acquire(context.Background(), 1)
	// 判断传入的是否是函数
	handlerValue := reflect.ValueOf(funcHandler)
	if handlerValue.Kind() != reflect.Func {
		return "", false, fmt.Errorf("invalid func handler")
	}
	go func(params ...interface{}) {
		defer ar.wg.Done()
		defer ar.sem.Release(1)

		paramNum := len(params)
		//jlog.Info("params:", params)
		params_list := make([]reflect.Value, paramNum)
		if paramNum > 0 {
			for k, v := range params {
				params_list[k] = reflect.ValueOf(v)
			}
		}

		// 运行函数
		values := handlerValue.Call(params_list)
		//
		var printHandlerValue reflect.Value
		if printHandler != nil && reflect.ValueOf(printHandler).Kind() == reflect.Func {
			printHandlerValue = reflect.ValueOf(printHandler)
		}
		// 传入的函数执行的结果保存在values中
		if valuesNum := len(values); valuesNum > 0 {
			resultItems := make([]interface{}, valuesNum)
			// asyncTaskKey:int,asyncTaskVal:value
			for k, v := range values {
				resultItems[k] = v.Interface()
			}
			//taskResult = resultItems
			// 如果传入了printHandler,并且reqHandler函数返回参数的个数与printHandler函数形参个数相同
			if printHandlerValue.IsValid() && handlerValue.Type().NumOut() == printHandlerValue.Type().NumIn() {
				paramsArg := make([]reflect.Value, len(resultItems))
				for k, v := range resultItems {
					if reflect.ValueOf(v).IsValid() {
						paramsArg[k] = reflect.ValueOf(v)
					} else {
						paramsArg[k] = reflect.Zero(printHandlerValue.Type().In(k))
					}
				}
				// 调用printHandler
				printHandlerValue.Call(paramsArg)
			}
		}
	}(params...)
	return task_name, false, fmt.Errorf(handlerValue.String() + " 不符合格式func(参数...)(返回...){}")
}
