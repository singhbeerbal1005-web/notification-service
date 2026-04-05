package service

import (
	"container/heap"
	"context"
	"notification-service/internal/domain"
	"time"
)

type scheduledJob struct {
	notification domain.Notification
	runAt        time.Time
}

type jobHeap []*scheduledJob

func (h jobHeap) Len() int            { return len(h) }
func (h jobHeap) Less(i, j int) bool  { return h[i].runAt.Before(h[j].runAt) }
func (h jobHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *jobHeap) Push(x interface{}) { *h = append(*h, x.(*scheduledJob)) }
func (h *jobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

type Scheduler struct {
	pq         jobHeap
	SubmitChan chan domain.Notification
	ReadyQueue chan domain.Notification
}

func NewScheduler(readyChan chan domain.Notification) *Scheduler {
	s := &Scheduler{
		pq:         make(jobHeap, 0),
		SubmitChan: make(chan domain.Notification, 1000),
		ReadyQueue: readyChan,
	}
	heap.Init(&s.pq)
	return s
}

func (s *Scheduler) Run(ctx context.Context) {
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	for {
		if s.pq.Len() > 0 {
			wait := time.Until(s.pq[0].runAt)
			if wait <= 0 {
				for s.pq.Len() > 0 && !s.pq[0].runAt.After(time.Now()) {
					job := heap.Pop(&s.pq).(*scheduledJob)
					s.ReadyQueue <- job.notification
				}
				continue
			}
			timer.Reset(wait)
		}

		select {
		case <-ctx.Done():
			return
		case n, ok := <-s.SubmitChan:
			if !ok {
				return
			}
			heap.Push(&s.pq, &scheduledJob{notification: n, runAt: n.ScheduledFor})
		case <-timer.C:
		}
	}
}
