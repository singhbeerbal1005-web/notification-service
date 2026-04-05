package provider

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
)

type SlackNotifier struct{}

func (s *SlackNotifier) Channel() domain.Channel { return domain.ChannelSlack }
func (s *SlackNotifier) Send(_ context.Context, r, c string) error {
	fmt.Printf("      [OUTPUT] SLACK      -> #%-14s | %q\n", r, c)
	return nil
}
