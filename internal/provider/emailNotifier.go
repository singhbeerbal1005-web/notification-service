package provider

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
)

type EmailProvider struct {
	ID   string
	Fail bool
}

func (e *EmailProvider) Send(_ context.Context, to, body string) error {
	if e.Fail {
		return fmt.Errorf("%s failed", e.ID)
	}
	fmt.Printf("      [OUTPUT] %-10s -> %-15s | %q\n", e.ID, to, body)
	return nil
}

type EmailDispatcher struct {
	Providers []*EmailProvider
}

func (e *EmailDispatcher) Channel() domain.Channel { return domain.ChannelEmail }
func (e *EmailDispatcher) Send(ctx context.Context, r, c string) error {
	for _, p := range e.Providers {
		if err := p.Send(ctx, r, c); err == nil {
			return nil
		}
		fmt.Printf("      [FAILOVER] %s failed, attempting next...\n", p.ID)
	}
	return fmt.Errorf("exhausted all email providers")
}
