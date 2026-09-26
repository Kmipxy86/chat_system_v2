package support

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"qrchat/internal/agent"
	"qrchat/internal/apperr"
	"qrchat/internal/platform/httpx"
)

type Handler struct {
	svc          *Service
	agents       *agent.Handler
	startLimiter *httpx.KeyedLimiter
}

func NewHandler(svc *Service, agents *agent.Handler, startLimiter *httpx.KeyedLimiter) *Handler {
	return &Handler{svc: svc, agents: agents, startLimiter: startLimiter}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/support/conversations", h.start)

	dash := r.Group("/support/conversations", h.agents.RequireAgent())
	dash.GET("", h.list)
	dash.POST("/:id/claim", h.claim)
}

func (h *Handler) start(c *gin.Context) {
	if !h.startLimiter.Allow(c.ClientIP()) {
		httpx.Error(c, apperr.RateLimited)
		return
	}
	var req struct {
		CustomerName string `json:"customer_name"`
		Subject      string `json:"subject"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, apperr.Invalid("invalid JSON body"))
		return
	}
	res, err := h.svc.Start(c.Request.Context(), req.CustomerName, req.Subject)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) list(c *gin.Context) {
	a := agent.AgentFrom(c)
	var views []ConversationView
	var err error
	if c.Query("filter") == "mine" {
		views, err = h.svc.ListMine(c.Request.Context(), a.ID)
	} else {
		views, err = h.svc.ListOpen(c.Request.Context())
	}
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": views})
}

func (h *Handler) claim(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, apperr.NotFound)
		return
	}
	a := agent.AgentFrom(c)
	sess, err := h.svc.Claim(c.Request.Context(), roomID, a.ID, a.Name)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, sess)
}
