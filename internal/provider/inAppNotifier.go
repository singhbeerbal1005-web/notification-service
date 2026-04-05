package provider

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
)

type InAppNotifier struct{}

func (i *InAppNotifier) Channel() domain.Channel { return domain.ChannelInApp }
func (i *InAppNotifier) Send(_ context.Context, r, c string) error {
	fmt.Printf("      [OUTPUT] IN-APP     -> @%-14s | %q\n", r, c)
	return nil
}
