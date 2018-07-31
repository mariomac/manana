package main

import (
	"fmt"
	"time"

	"github.com/mariomac/manana"
)

func TimeConsumingDouble(n int, ctxDone <- chan struct{}) (int, error) {
	select {
		case <- ctxDone:
			fmt.Println("Canceled!!!")
			return 0, manana.ErrorCanceled
		case <- time.After(2 * time.Second):
	}

	fmt.Println("Timeconsumingdouble has ended")
	return n * 2, nil
}

func AsyncDouble(n int) manana.Future {
	return manana.DoCtx(func(ctxDone <- chan struct{}) (interface{}, error) {

		return TimeConsumingDouble(n, ctxDone)
	})
}

func main() {

	ab := manana.All(
		AsyncDouble(3),
		AsyncDouble(5),
	)
	w := make(chan struct{})
	var c manana.Future
	ab.OnSuccess(func(results interface{}) {
		a := results.([]interface{})[0].(int)
		b := results.([]interface{})[1].(int)
		c = AsyncDouble(a + b)
		close(w)
	})

	<-w
	c.Cancel()

	time.Sleep(5*time.Second)
	fmt.Println("Bye!")
}
