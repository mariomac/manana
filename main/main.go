package main
/*
import (
	"log"
	"sync"
	"time"
)

// TimeConsumingFunction is a BLOCKING function that takes some time to be calculated
func TimeConsumingFunction() aaafffkkk {
	log.Println("Need my time to think in the meaning of life...")
	time.Sleep(1 * time.Second)
	return 42
}


// TimeConsumingAsyncFunction is the asynchronous wrap for the above TimeConsumingFunction
func TimeConsumingAsyncFunction() *Future {
	return AsyncHolder(TimeConsumingFunction)
}

// AsyncHolder is a helper function to convert a synchronous blocking function to asynchronous
func AsyncHolder(syncFunc func() aaafffkkk) *Future {
	successChan := make(chan aaafffkkk)
	res := Future{
		Success: successChan,
	}
	// Invoke syncFunc in another thread and send the result to the success channel of the Future
	go func() {
		successChan <- syncFunc()
	}()

	return &res
}

// Helper function that allows waiting for multiple futures and returns an array with the results of all
func WaitForAll(futures ...*Future) *Future {
	resultsChan := make(chan aaafffkkk)
	for _, future := range futures {
		future.OnSuccess(func(r aaafffkkk) {
			resultsChan <- r
		})
	}
	return AsyncHolder(func() aaafffkkk {
		results := make([]aaafffkkk, 0, len(futures))
		for i := 0; i < len(futures); i++ {
			results = append(results, <-resultsChan)
		}
		return results
	})
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	TimeConsumingAsyncFunction().
		OnSuccess(func(status aaafffkkk) {
			log.Println("Function returned with status", status)
			wg.Done()
		})

	WaitForAll(
		TimeConsumingAsyncFunction(),
		TimeConsumingAsyncFunction(),
		TimeConsumingAsyncFunction(),
	).OnSuccess(func(allStatus aaafffkkk) {
		log.Printf("Gotcha! %+v\n", allStatus)
		wg.Done()
	})

	wg.Wait()
}
*/