package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mariomac/manana"
)

// Some time consuming function that is executed synchronously
func TimeConsumingFunction(max int) int {
	<-time.After(time.Duration(rand.Intn(max)) * time.Millisecond)
	return max * 3
}

func Synchronous() {
	a := TimeConsumingFunction(5000)
	b := TimeConsumingFunction(1000)
	c := TimeConsumingFunction(100)
	d := TimeConsumingFunction((a + b) / c)

	fmt.Printf("The result is %v", d)
}

func Asynchronous() {
	asyncTimeConsumingFunction := func(max int) func() (interface{}, error) {
		return func() (interface{}, error) {
			return TimeConsumingFunction(max), nil
		}
	}

	f := manana.All(
		manana.Do(asyncTimeConsumingFunction(5000)),
		manana.Do(asyncTimeConsumingFunction(1000)),
		manana.Do(asyncTimeConsumingFunction(100)),
	)

	d := make(chan manana.Future)
	f.OnSuccess(func(results interface{}) {
		a := results.([]interface{})[0].(int)
		b := results.([]interface{})[1].(int)
		c := results.([]interface{})[2].(int)
		d <- manana.Do(asyncTimeConsumingFunction((a + b) / c))
	})

	val, _ := (<-d).Get()
	fmt.Printf("The result is %v", val)
}

func main() {
	start := time.Now()
	Synchronous()
	fmt.Println(" (", time.Now().Sub(start), ")")

	start = time.Now()
	Asynchronous()
	fmt.Println(" (", time.Now().Sub(start), ")")
}
