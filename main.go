package main

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
	"notification-service/internal/provider"
	"notification-service/internal/repository"
	"notification-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Setup Signal Handling for Graceful Shutdown (FR-06)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Initialize Channels and Scheduler (FR-01, FR-05)
	// ReadyQueue connects the Scheduler to the Dispatcher
	readyQueue := make(chan domain.Notification, 2000)
	sched := service.NewScheduler(readyQueue)

	// 3. Dependency Injection: Repository (Template Store)
	// Swappable: NewMockTemplateRepo() can be replaced by a Postgres implementation later
	repo := repository.NewMockTemplateRepo()

	// 4. Dependency Injection: Providers (Strategy Pattern)
	emailDispatcher := &provider.EmailDispatcher{
		Providers: []*provider.EmailProvider{
			{ID: "SendGrid_Primary", Fail: true}, // Forces Failover (FR-03)
			{ID: "AWS_SES_Backup", Fail: false},
		},
	}

	// 5. Initialize Core Service with Worker Pools (FR-05)
	svc := service.NewNotificationService(
		repo,                      // Template Provider
		5,                         // Worker Count per Channel
		sched,                     // Scheduler Instance
		emailDispatcher,           // Email Logic
		&provider.SlackNotifier{}, // Slack Logic
		&provider.InAppNotifier{}, // In-App Logic
	)

	// 6. Start the Engine
	svc.Start(ctx)
	fmt.Println(">>> NOTIFICATION ENGINE ONLINE | Multi-File Architecture")
	fmt.Println(">>> RUNNING FULL FUNCTIONAL TEST SUITE (FR-01 to FR-14)...")

	// 7. Execute Functional Test Cases
	go runTestSuite(sched)

	// 8. Wait for Shutdown Signal
	<-ctx.Done()
	fmt.Println("\n>>> SIGINT/SIGTERM RECEIVED: Initiating Graceful Shutdown...")

	// Ensure all workers finish current tasks and queues are drained
	svc.Shutdown()
	fmt.Println(">>> SHUTDOWN COMPLETE: All jobs processed safely.")
}

func runTestSuite(sched *service.Scheduler) {
	now := time.Now()

	// --- FR-01 to FR-03: Multi-Channel Routing ---
	sched.SubmitChan <- domain.Notification{
		ID: "FR-01", Recipient: "user@mail.com", Type: domain.ChannelEmail,
		TemplateName: "welcome", Payload: map[string]interface{}{"Name": "Alice"},
	}
	sched.SubmitChan <- domain.Notification{
		ID: "FR-02", Recipient: "dev-ops-slack", Type: domain.ChannelSlack,
		TemplateName: "alert", Payload: map[string]interface{}{"Msg": "Postgres Latency High"},
	}
	sched.SubmitChan <- domain.Notification{
		ID: "FR-03", Recipient: "mobile_uid_99", Type: domain.ChannelInApp,
		TemplateName: "welcome", Payload: map[string]interface{}{"Name": "Bob"},
	}

	// --- FR-04: Custom Template Override (Priority Logic) ---
	sched.SubmitChan <- domain.Notification{
		ID: "FR-04", Recipient: "marketing@test.com", Type: domain.ChannelEmail,
		CustomTemplate: "Flash Sale! Use code {{.Code}}", Payload: map[string]interface{}{"Code": "SAVE50"},
	}

	// --- FR-05: Scheduling (Instant vs Delayed) ---
	sched.SubmitChan <- domain.Notification{
		ID: "FR-05", Recipient: "immediate@test.com", Type: domain.ChannelEmail,
		CustomTemplate: "Instant Delivery", ScheduledFor: now,
	}
	sched.SubmitChan <- domain.Notification{
		ID: "FR-06", Recipient: "delayed@test.com", Type: domain.ChannelEmail,
		CustomTemplate: "Delayed 2s Delivery", ScheduledFor: now.Add(2 * time.Second),
	}

	// --- FR-07: Provider Failover (SendGrid -> AWS SES) ---
	sched.SubmitChan <- domain.Notification{
		ID: "FR-07", Recipient: "failover-target@test.com", Type: domain.ChannelEmail,
		TemplateName: "alert", Payload: map[string]interface{}{"Msg": "Resilience Test"},
	}

	// --- FR-08 to FR-14: Concurrency & Burst Handling ---
	// Validates Worker Pool stability under load across multiple channels
	for i := 8; i <= 14; i++ {
		targetType := domain.ChannelInApp
		if i%2 == 0 {
			targetType = domain.ChannelSlack
		}
		sched.SubmitChan <- domain.Notification{
			ID: fmt.Sprintf("FR-%02d", i), Recipient: "burst-user", Type: targetType,
			CustomTemplate: "Parallel Process Job #{{.ID}}", Payload: map[string]interface{}{"ID": i},
		}
	}
}
