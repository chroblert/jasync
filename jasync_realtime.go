package jasync

import (
	"context"
	"fmt"
	"github.com/chroblert/jlog"
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
	// handler
	HandlerValues []reflect.Value
	// 池
	pool *sync.Pool
}

// New 创建一个新的异步执行对象
//
// verbose: 是否显示进度条,默认显示
func NewAR(count int64, verbose ...bool) *AsyncRealtime {
	if len(verbose) == 0 {
		return &AsyncRealtime{
			mu:      new(sync.RWMutex),
			verbose: false,
			sem:     semaphore.NewWeighted(count),
			wg:      &sync.WaitGroup{},
			pool: &sync.Pool{
				New: func() interface{} {
					return &AsyncRealtimeTask{
						taskName:        "",
						handlerValues:   make([]reflect.Value, 0),
						inParamsValues:  make([][]reflect.Value, 0),
						outParamsValues: make([][]reflect.Kind, 0),
						handlerNum:      0,
						AsyncRealtime:   nil,
					}
				},
			},
		}
	}
	return &AsyncRealtime{
		mu:      new(sync.RWMutex),
		verbose: verbose[0],
		sem:     semaphore.NewWeighted(count),
		wg:      &sync.WaitGroup{},
		pool: &sync.Pool{
			New: func() interface{} {
				return &AsyncRealtimeTask{
					taskName:        "",
					handlerValues:   make([]reflect.Value, 0),
					inParamsValues:  make([][]reflect.Value, 0),
					outParamsValues: make([][]reflect.Kind, 0),
					handlerNum:      0,
					AsyncRealtime:   nil,
				}
			},
		},
	}
}

//	type taskStatusStruct struct {
//		taskStatusStruct  int   // 任务状态 0: init,1:queue,2: doing,3: done
//		taskBegTime int64 // 任务开始时间
//		taskEndTime int64 // 任务结束时间
//	}
func (ar *AsyncRealtime) Wait() {
	ar.wg.Wait()
}

func (ar *AsyncRealtime) Done() <-chan struct{} {
	ar.wg.Wait()
	var ch = make(chan struct{})
	var sendOnly chan<- struct{} = ch
	var revOnly <-chan struct{} = ch
	go func() {
		sendOnly <- struct{}{}
	}()
	return revOnly
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

func (art *AsyncRealtimeTask) CAdd(funcHandler interface{}, params ...interface{}) *AsyncRealtimeTask {
	// 是否为空
	if art == nil {
		return nil
	}
	if art.err != nil {
		return art
	}
	// 是否设置了任务名
	if art.taskName == "" {
		name, err2 := uuid.GenerateUUID()
		if err2 != nil {
			art.err = err2
			return art
		}
		art.taskName = name
	}
	// 添加handler
	// 获取信号量
	// 判断传入的是否是函数
	handlerValue := reflect.ValueOf(funcHandler)
	if handlerValue.Kind() != reflect.Func {
		art.err = fmt.Errorf("传入的不是函数")
		return art
	}
	// 判断是否是添加的第一个函数
	if art.handlerNum == 0 {
		// 判断函数的形参个数，与传入的实参个数是否相同
		if handlerValue.Type().NumIn() != len(params) {
			art.err = fmt.Errorf("形参与实参个数不同")
			return art
		}
		// 判断本函数的输入参数类型 是否等于 本函数形参类型
		for k, param := range params {
			if reflect.ValueOf(param).Kind() != handlerValue.Type().In(k).Kind() {
				art.err = fmt.Errorf("形参与实参类型不同:%d", k)
				return art
			}
		}
		art.handlerNum += 1
	} else {
		// 上一个函数的输出参数
		lastHandlerOutKinds := art.outParamsValues[art.handlerNum-1]
		// 判断上个函数的输出参数个数+本函数的输入参数个数是否等于本函数的形参个数
		if len(lastHandlerOutKinds)+len(params) != handlerValue.Type().NumIn() {
			art.err = fmt.Errorf("上个函数的输出参数个数(%d)+本函数的输入参数个数(%d) != 本函数的形参个数(%d)", len(lastHandlerOutKinds), len(params), handlerValue.Type().NumIn())
			return art
		}
		// 判断上个函数的输出参数的类型+本函数的输入参数类型 是否等于 本函数形参类型
		for k, lastOutKind := range lastHandlerOutKinds {
			if lastOutKind != handlerValue.Type().In(k).Kind() {
				art.err = fmt.Errorf("形参与实参类型不同:%d", k)
				return art
			}
		}
		// 判断本函数的输入参数类型 是否等于 本函数形参类型
		for k, param := range params {
			if reflect.ValueOf(param).Kind() != handlerValue.Type().In(k+len(lastHandlerOutKinds)).Kind() {
				art.err = fmt.Errorf("形参与实参类型不同:%d", k+len(lastHandlerOutKinds))
				return art
			}
		}
		art.handlerNum += 1
	}

	// 函数输出的参数
	outParamNum := handlerValue.Type().NumOut()
	// 保存输出参数
	var outKinds = make([]reflect.Kind, outParamNum)
	for i := 0; i < outParamNum; i++ {
		outKinds[i] = handlerValue.Type().Out(i).Kind()
		//outKinds = append(outKinds, handlerValue.Type().Out(i).Kind())
	}
	art.outParamsValues = append(art.outParamsValues, outKinds)

	// 添加函数
	art.handlerValues = append(art.handlerValues, handlerValue)
	// 添加参数
	if len(params) == 0 {
		art.inParamsValues = append(art.inParamsValues, []reflect.Value{})
	} else {
		paramValues := make([]reflect.Value, len(params))
		for k, v := range params {
			paramValues[k] = reflect.ValueOf(v)
		}
		art.inParamsValues = append(art.inParamsValues, paramValues)
	}
	return art
}

func (art *AsyncRealtimeTask) CDO() (err error) {
	if art == nil {
		return fmt.Errorf("art对象为nil")
	}
	if art.err != nil {
		//jlog.Error(art.err)
		return art.err
	}
	//
	art.wg.Add(1)
	if !art.sem.TryAcquire(1) {
		if art.verbose {
			// 显示信息
			jlog.Info("Block To Acquire Semaphore")
		}
		// 获取信号量
		art.sem.Acquire(context.Background(), 1)
	}

	go func() {
		defer art.wg.Done()
		defer art.sem.Release(1)
		defer art.pool.Put(art)
		defer art.Clean()
		lastOutValues := make([]reflect.Value, 0)
		for k, handlerValue := range art.handlerValues {
			lastOutValues = append(lastOutValues, art.inParamsValues[k]...)
			lastOutValues = handlerValue.Call(lastOutValues)
		}
		//jlog.Info("done")
	}()
	return nil
}

func (art *AsyncRealtimeTask) Clean() {
	art.handlerValues = art.HandlerValues[:0]
	art.inParamsValues = art.inParamsValues[:0]
	art.outParamsValues = art.outParamsValues[:0]
	art.err = nil
	art.handlerNum = 0
}

type AsyncRealtimeTask struct {
	*AsyncRealtime
	taskName        string
	handlerNum      int
	handlerValues   []reflect.Value
	inParamsValues  [][]reflect.Value
	outParamsValues [][]reflect.Kind
	err             error
}

func (ar *AsyncRealtime) Init(taskName string) *AsyncRealtimeTask {
	art := ar.pool.Get().(*AsyncRealtimeTask)
	art.taskName = taskName
	art.AsyncRealtime = ar

	return art
}
