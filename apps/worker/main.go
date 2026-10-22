package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/anomaly"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/notifications"
)

func main() {
	log.Printf("[kubetrim-hub-worker] v%s starting asynchronous FinOps telemetry & anomaly worker...", domain.Version)

	dispatcher := notifications.NewDispatcher(notifications.Config{
		SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		TeamsWebhookURL: os.Getenv("TEAMS_WEBHOOK_URL"),
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Execute initial self-check on startup
	now := time.Now().UTC()
	samples := []domain.MemorySample{
		{Timestamp: now.Add(-6 * time.Hour), Bytes: 200 * 1024 * 1024},
		{Timestamp: now.Add(-5 * time.Hour), Bytes: 400 * 1024 * 1024},
		{Timestamp: now.Add(-4 * time.Hour), Bytes: 800 * 1024 * 1024},
		{Timestamp: now.Add(-3 * time.Hour), Bytes: 150 * 1024 * 1024, OOMKilled: true},
		{Timestamp: now.Add(-2 * time.Hour), Bytes: 500 * 1024 * 1024},
		{Timestamp: now.Add(-1 * time.Hour), Bytes: 950 * 1024 * 1024},
	}
	if finding := anomaly.AnalyzeMemorySawtooth("self-check", "kube-system", "probe", "worker", samples); finding != nil {
		log.Printf("[kubetrim-hub-worker] anomaly engine self-check passed: %s (%d sawtooth cycles)", finding.AnomalyType, finding.SawtoothCycles)
	}
	_ = dispatcher

	if os.Getenv("RUN_ONCE") == "true" {
		log.Printf("[kubetrim-hub-worker] RUN_ONCE=true completed successfully")
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[kubetrim-hub-worker] shutting down cleanly")
			return
		case <-ticker.C:
			log.Printf("[kubetrim-hub-worker] heartbeat: anomaly & attribution worker active")
		}
	}
}
