package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/mariomac/manana"
)

func TimeConsumingSynchronousFunction() int {
	<-time.After(3 * time.Second)
	return 42
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	future := manana.Do(func() (interface{}, error) {
		return TimeConsumingSynchronousFunction(), nil
	})
	future.OnSuccess(func(result interface{}) {
		fmt.Println("The result is", result)
		wg.Done()
	})
	future2 := manana.Do(func() (interface{}, error) {
		return TimeConsumingSynchronousFunction(), nil
	})
	future2.OnSuccess(func(result interface{}) {
		fmt.Println("The result of the second invocation is", result)
		wg.Done()
	})
	wg.Wait()
}
