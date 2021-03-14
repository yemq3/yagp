package main

type Scheduler struct {
	evaluator Evaluator
}

func NewScheduler(messageCenter MessageCenter) Scheduler {
	scheduler := Scheduler{}

	scheduler.evaluator = NewEvaluator(messageCenter)

	return scheduler
}

func (scheduler *Scheduler) run() {
	go scheduler.evaluator.run()
	
}