package main

import "time"

type Evaluator struct {
	averageDelay float64

}

func (evaluator Evaluator) init()  {

}

func (evaluator Evaluator) run()  {
	for {
		time.Sleep(50 * time.Millisecond)
	}
}