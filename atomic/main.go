package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	NodeName string
	Addr     string
	Count    int
}

var DefatultConfig = &Config{}

func loadNewConfig(v atomic.Value) {
	config := Config{
		NodeName: "北京",
		Addr:     "10.77.95.27",
		Count:    rand.Intn(3),
	}
	v.CompareAndSwap(DefatultConfig, config)

}

func main() {
	var v atomic.Value
	loadNewConfig(v)
	var cond = sync.NewCond(&sync.Mutex{})

	// 设置新的config
	go func() {
		for {
			time.Sleep(time.Duration(5+rand.Int63n(5)) * time.Second)
			loadNewConfig(v) // 这可以动态的监听一个config的变化
			cond.Broadcast() // 通知等待着配置已变更
		}
	}()

	go func() {
		for {
			cond.L.Lock()
			cond.Wait()            // 等待变更信号
			c := v.Load().(Config) // 读取新的配置
			fmt.Printf("new config: %+v\n", c)
			cond.L.Unlock()
		}
	}()

	select {}
}
