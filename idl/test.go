package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	now := time.Now()
	var wg sync.WaitGroup
	success := 0
	fail := 0
	var mu sync.Mutex

	// 模拟200并发请求
	for i := 0; i < 20000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get("http://127.0.0.1:8888/api/resume/login")
			if err != nil {
				mu.Lock()
				fail++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()
			mu.Lock()
			if resp.StatusCode == 200 {
				success++
			} else if resp.StatusCode == 429 {
				fail++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("成功请求：%d，限流拒绝：%d\n", success, fail)

	fmt.Println(time.Since(now).Seconds())
}
