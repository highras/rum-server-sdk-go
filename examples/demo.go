package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"../src/rum"
	"github.com/highras/fpnn-sdk-go/src/fpnn"
)

type PrintLocker struct {
	mutex sync.Mutex
}

func (locker *PrintLocker) print(proc func()) {
	locker.mutex.Lock()
	defer locker.mutex.Unlock()

	proc()
}

var locker PrintLocker = PrintLocker{}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client := rum.NewRUMServerClient(41000015, "affc562c-8796-4714-b8ae-4b061ca48a6b", "52.83.220.166:13609")

	//client.SetRid("ttttttt-rid")
	//client.SetSid(123456)

	attrs := make(map[string]interface{})
	attrs["aaa"] = "bbb"
	attrs["bbb"] = 123

	errSync := client.SendCustomEvent("error", attrs)
	locker.print(func() {
		if errSync == nil {
			fmt.Printf("SendCustomEvent in sync mode is fine.\n")
		} else {
			fmt.Printf("SendCustomEvent in sync mode error, err: %v\n", errSync)
		}
	})

	errAsync := client.SendCustomEvent("error", attrs, func(errorCode int, errInfo string) {
		locker.print(func() {
			if errorCode == fpnn.FPNN_EC_OK {
				fmt.Printf("SendCustomEvent callback ok\n")
			} else {
				fmt.Printf("SendCustomEvent callback fail, error code: %d, error info:%s\n", errorCode, errInfo)
			}
		})
	})
	locker.print(func() {
		if errAsync == nil {
			fmt.Printf("SendCustomEvent in async mode is fine.\n")
		} else {
			fmt.Printf("SendCustomEvent in async mode error, err: %v\n", errAsync)
		}
	})

	eventMap := make(map[string]interface{})
	eventMap["ev"] = "error"
	eventMap["attrs"] = attrs

	events := make([]map[string]interface{}, 0, 2)
	events = append(events, eventMap)
	events = append(events, eventMap)

	errSync2 := client.SendCustomEvents(events)
	locker.print(func() {
		if errSync2 == nil {
			fmt.Printf("SendCustomEvents in sync mode is fine.\n")
		} else {
			fmt.Printf("SendCustomEvents in sync mode error, err: %v\n", errSync2)
		}
	})

	errAsync2 := client.SendCustomEvents(events, func(errorCode int, errInfo string) {
		locker.print(func() {
			if errorCode == fpnn.FPNN_EC_OK {
				fmt.Printf("SendCustomEvents callback ok\n")
			} else {
				fmt.Printf("SendCustomEvents callback fail, error code: %d, error info:%s\n", errorCode, errInfo)
			}
		})
	})
	locker.print(func() {
		if errAsync2 == nil {
			fmt.Printf("SendCustomEvents in async mode is fine.\n")
		} else {
			fmt.Printf("SendCustomEvents in async mode error, err: %v\n", errAsync2)
		}
	})

	time.Sleep(time.Second) //-- Waiting for the async callback printed.
}
