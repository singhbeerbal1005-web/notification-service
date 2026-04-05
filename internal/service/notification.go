package service

import (
	"bytes"
	"context"
	"notification-service/internal/domain"
	"sync"
	"text/template"
)

type Notifier interface {
	Send(ctx context.Context, recipient, content string) error
	Channel() domain.Channel
}

type TemplateProvider interface {
	GetTemplate(ctx context.Context, name string) (string, error)
}

type NotificationService struct {
	notifiers     map[domain.Channel]Notifier
	queues        map[domain.Channel]chan domain.Notification
	scheduler     *Scheduler
	templates     TemplateProvider
	templateCache sync.Map
	workers       int
	wg            sync.WaitGroup
}

func NewNotificationService(tp TemplateProvider, workers int, sched *Scheduler, ns ...Notifier) *NotificationService {
	svc := &NotificationService{
		templates: tp,
		scheduler: sched,
		notifiers: make(map[domain.Channel]Notifier),
		queues:    make(map[domain.Channel]chan domain.Notification),
		workers:   workers,
	}
	for _, n := range ns {
		svc.notifiers[n.Channel()] = n
		svc.queues[n.Channel()] = make(chan domain.Notification, 2000)
	}
	return svc
}

func (s *NotificationService) Start(ctx context.Context) {
	go s.scheduler.Run(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case n := <-s.scheduler.ReadyQueue:
				if q, ok := s.queues[n.Type]; ok {
					q <- n
				}
			}
		}
	}()

	for ch := range s.notifiers {
		for i := 0; i < s.workers; i++ {
			s.wg.Add(1)
			go s.worker(ctx, ch)
		}
	}
}

func (s *NotificationService) worker(ctx context.Context, ch domain.Channel) {
	defer s.wg.Done()
	for n := range s.queues[ch] {
		s.execute(ctx, n, ch)
	}
}

func (s *NotificationService) execute(ctx context.Context, n domain.Notification, ch domain.Channel) {
	var tmpl *template.Template
	var err error

	if n.CustomTemplate != "" {
		tmpl, err = template.New(n.ID).Parse(n.CustomTemplate)
	} else {
		if val, ok := s.templateCache.Load(n.TemplateName); ok {
			tmpl = val.(*template.Template)
		} else {
			raw, err := s.templates.GetTemplate(ctx, n.TemplateName)
			if err != nil {
				raw = "Default: {{.Name}}"
			}
			tmpl, err = template.New(n.TemplateName).Parse(raw)
			if err == nil {
				s.templateCache.Store(n.TemplateName, tmpl)
			}
		}
	}

	if err != nil || tmpl == nil {
		return
	}
	var body bytes.Buffer
	_ = tmpl.Execute(&body, n.Payload)
	_ = s.notifiers[ch].Send(ctx, n.Recipient, body.String())
}

func (s *NotificationService) Shutdown() {
	for _, q := range s.queues {
		close(q)
	}
	s.wg.Wait()
}
