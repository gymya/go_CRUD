package http

import (
	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Auth service.AuthService
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// Login godoc
// @Summary      Login
// @Description  Login with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginRequest  true  "Login credentials"
// @Success      200          {object}  loginResponse
// @Failure      401          {object}  domain.AppError
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}

	token, expiresIn, err := h.Auth.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:     token,
		ExpiresIn: int64(expiresIn.Seconds()),
	})
}

// Register godoc
// @Summary      Register
// @Description  Create a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      registerRequest  true  "User data"
// @Success      201   {object}  domain.User
// @Failure      409   {object}  domain.AppError
// @Failure      422   {object}  domain.AppError
// @Router       /users [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}

	user, err := h.Auth.Register(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, user)
}
