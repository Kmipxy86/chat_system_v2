package agent

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"qrchat/internal/apperr"
	"qrchat/internal/auth"
	"qrchat/internal/platform/httpx"
)

type Handler struct {
	svc        *Service
	issuer     *auth.Issuer
	signupKey  string
	loginLimit *httpx.KeyedLimiter
}

func NewHandler(svc *Service, issuer *auth.Issuer, signupKey string, loginLimit *httpx.KeyedLimiter) *Handler {
	return &Handler{svc: svc, issuer: issuer, signupKey: signupKey, loginLimit: loginLimit}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/agents/register", h.register)
	r.POST("/agents/login", h.login)
	r.GET("/agents/me", h.RequireAgent(), h.me)
}

const agentKey = "agent"

// RequireAgent authenticates the Bearer account JWT and loads the current agent row.
func (h *Handler) RequireAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok {
			httpx.Error(c, apperr.Unauthorized)
			return
		}
		agentID, err := h.issuer.ParseAgent(token)
		if err != nil {
			httpx.Error(c, apperr.Unauthorized)
			return
		}
		a, err := h.svc.Get(c.Request.Context(), agentID)
		if err != nil {
			httpx.Error(c, apperr.Unauthorized)
			return
		}
		c.Set(agentKey, a)
		c.Next()
	}
}

// AgentFrom returns the agent authenticated by RequireAgent.
func AgentFrom(c *gin.Context) Agent {
	a, _ := c.Get(agentKey)
	return a.(Agent)
}

type agentView struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

func toView(a Agent) agentView {
	return agentView{AgentID: a.ID.String(), Name: a.Name, Email: a.Email}
}

func (h *Handler) register(c *gin.Context) {
	// Signup is closed unless an operator configured AGENT_SIGNUP_KEY and the
	// caller presents it, so anyone cannot mint themselves an agent account.
	if h.signupKey == "" {
		httpx.Error(c, apperr.SignupDisabled)
		return
	}
	if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Agent-Signup-Key")), []byte(h.signupKey)) != 1 {
		httpx.Error(c, apperr.SignupDisabled)
		return
	}
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, apperr.Invalid("invalid JSON body"))
		return
	}
	sess, err := h.svc.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, sess)
}

func (h *Handler) login(c *gin.Context) {
	if !h.loginLimit.Allow(c.ClientIP()) {
		httpx.Error(c, apperr.RateLimited)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, apperr.Invalid("invalid JSON body"))
		return
	}
	sess, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (h *Handler) me(c *gin.Context) {
	c.JSON(http.StatusOK, toView(AgentFrom(c)))
}
