package gazprom

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type Handler struct {
	DB *sqlx.DB
}

func (h *Handler) AddRouter(r gin.IRoutes) {
	r.POST("auth", h.Auth)
}

func (h *Handler) validateError(c *gin.Context, err error) {
	ginError := c.Error(err)
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, ginError.JSON())
}

func (h *Handler) responseError(c *gin.Context, err error) {
	ginError := c.Error(err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, ginError.JSON())
}


func (h *Handler) Auth(c *gin.Context) {
	var auth User

	if err := c.ShouldBindJSON(&auth); err != nil {
		h.validateError(c, err)
		return
	}

	err := h.DB.Get(&auth, `SELECT * FROM auth WHERE login = ? AND password = ?`, auth.Login, auth.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusUnauthorized
		}

		ginError := c.Error(err)
		c.AbortWithStatusJSON(status, ginError.JSON())
	}

	c.JSON(http.StatusOK, gin.H{
		"id": auth.ID,
		"name": auth.Login,
		"role": auth.Role,
	})
}


