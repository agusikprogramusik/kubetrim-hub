package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/copilot"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/gitops"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/notifications"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerConfig configures the KubeTrim Hub HTTP API server
type ServerConfig struct {
	APIKey        string
	Store         storage.Store
	GitOpsBot     *gitops.Bot
	CopilotEngine *copilot.Engine
	Dispatcher    *notifications.Dispatcher
}

// Server wraps the Gin router and domain services
type Server struct {
	cfg    ServerConfig
	router *gin.Engine
}

// NewServer initializes the Gin v1.12+ HTTP router and registers all FinOps endpoints
func NewServer(cfg ServerConfig) *Server {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Store == nil {
		cfg.Store = storage.NewMemoryStore()
	}
	if cfg.GitOpsBot == nil {
		cfg.GitOpsBot = gitops.NewBot(gitops.BotConfig{})
	}
	if cfg.CopilotEngine == nil {
		cfg.CopilotEngine = copilot.NewEngine(copilot.Config{})
	}
	if cfg.Dispatcher == nil {
		cfg.Dispatcher = notifications.NewDispatcher(notifications.Config{})
	}

	r := gin.New()
	r.Use(gin.Recovery())

	s := &Server{
		cfg:    cfg,
		router: r,
	}
	s.registerRoutes()
	return s
}

// Router returns the underlying http.Handler
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) registerRoutes() {
	s.router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": domain.Version,
		})
	})

	s.router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	v1 := s.router.Group("/api/v1")
	v1.Use(s.authMiddleware())
	{
		v1.POST("/clusters/:cluster_id/snapshots", s.handleIngestSnapshot)
		v1.POST("/ingest", s.handleIngestSnapshot)
		v1.GET("/executive/kpis", s.handleGetExecutiveKPIs)
		v1.GET("/clusters", s.handleListClusters)
		v1.GET("/clusters/:cluster_id/snapshots", s.handleGetClusterSnapshots)
		v1.GET("/clusters/:cluster_id/attribution", s.handleGetAttribution)
		v1.GET("/clusters/:cluster_id/anomalies", s.handleGetAnomalies)
		v1.POST("/gitops/pr", s.handleCreateGitOpsPR)
		v1.GET("/gitops/prs", s.handleListGitOpsPRs)
		v1.POST("/copilot/chat", s.handleCopilotChat)
	}
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(s.cfg.APIKey)
		if expected == "" {
			c.Next()
			return
		}

		token := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if token == "" {
			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			}
		}

		if token != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or missing API key",
			})
			return
		}
		c.Next()
	}
}

func (s *Server) handleIngestSnapshot(c *gin.Context) {
	var report domain.AuditReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clusterID := strings.TrimSpace(c.Param("cluster_id"))
	if clusterID == "" {
		clusterID = strings.TrimSpace(c.GetHeader("X-Cluster-ID"))
	}
	if clusterID == "" {
		clusterID = strings.TrimSpace(report.ClusterName)
	}
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	if report.ReportID == "" {
		report.ReportID = "snap-" + uuid.NewString()
	}
	if report.GeneratedAt.IsZero() {
		report.GeneratedAt = time.Now().UTC()
	}

	if err := s.cfg.Store.SaveSnapshot(clusterID, report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger asynchronous anomaly & regression notification check
	go s.evaluatePostIngestAlerts(clusterID)

	c.JSON(http.StatusAccepted, gin.H{
		"status":      "accepted",
		"cluster_id":  clusterID,
		"snapshot_id": report.ReportID,
	})
}

func (s *Server) evaluatePostIngestAlerts(clusterID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, finding := range s.cfg.Store.GetAnomalies(clusterID) {
		_ = s.cfg.Dispatcher.SendAnomalyAlert(ctx, finding)
	}
	if attrReport, err := s.cfg.Store.GetAttribution(clusterID); err == nil && attrReport != nil {
		for _, attr := range attrReport.Attributions {
			if attr.Status == domain.StatusRegressionAlert {
				_ = s.cfg.Dispatcher.SendRegressionAlert(ctx, clusterID, attr)
			}
		}
	}
}

func (s *Server) handleGetExecutiveKPIs(c *gin.Context) {
	c.JSON(http.StatusOK, s.cfg.Store.GetExecutiveKPIs())
}

func (s *Server) handleListClusters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"clusters": s.cfg.Store.GetClusters(),
	})
}

func (s *Server) handleGetClusterSnapshots(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	c.JSON(http.StatusOK, gin.H{
		"cluster_id": clusterID,
		"snapshots":  s.cfg.Store.GetClusterSnapshots(clusterID),
	})
}

func (s *Server) handleGetAttribution(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	report, err := s.cfg.Store.GetAttribution(clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (s *Server) handleGetAnomalies(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	c.JSON(http.StatusOK, gin.H{
		"cluster_id": clusterID,
		"anomalies":  s.cfg.Store.GetAnomalies(clusterID),
	})
}

func (s *Server) handleCreateGitOpsPR(c *gin.Context) {
	var req domain.PullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := s.cfg.GitOpsBot.CreatePullRequest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.cfg.Store.SavePullRequest(*record)
	c.JSON(http.StatusCreated, record)
}

func (s *Server) handleListGitOpsPRs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pull_requests": s.cfg.Store.ListPullRequests(),
	})
}

func (s *Server) handleCopilotChat(c *gin.Context) {
	var req copilot.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fleet := copilot.FleetContext{
		KPIs:      s.cfg.Store.GetExecutiveKPIs(),
		Clusters:  s.cfg.Store.GetClusters(),
		Snapshots: s.cfg.Store.GetLatestSnapshots(),
		Anomalies: s.cfg.Store.GetAnomalies(req.ClusterID),
	}
	resp, err := s.cfg.CopilotEngine.Ask(c.Request.Context(), req, fleet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
