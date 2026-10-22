package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/api"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/copilot"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/gitops"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/notifications"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	store := storage.NewMemoryStore()
	if os.Getenv("SEED_DEMO_DATA") != "false" {
		storage.SeedDemoFleet(store)
	}
	bot := gitops.NewBot(gitops.BotConfig{
		GitHubAPIBaseURL: os.Getenv("GITHUB_API_URL"),
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
	})
	engine := copilot.NewEngine(copilot.Config{
		Endpoint:  os.Getenv("LLM_ENDPOINT"),
		ModelID:   os.Getenv("LLM_MODEL_ID"),
		APIKey:    os.Getenv("LLM_API_KEY"),
		EnableLLM: os.Getenv("LLM_ENABLED") == "true",
	})
	dispatcher := notifications.NewDispatcher(notifications.Config{
		SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		TeamsWebhookURL: os.Getenv("TEAMS_WEBHOOK_URL"),
	})

	srv := api.NewServer(api.ServerConfig{
		APIKey:        os.Getenv("HUB_API_KEY"),
		Store:         store,
		GitOpsBot:     bot,
		CopilotEngine: engine,
		Dispatcher:    dispatcher,
	})

	httpSrv := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("[kubetrim-hub-api] v%s listening on :%s", domain.Version, port)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
