package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PingDB возвращает middleware для проверки доступности PostgreSQL (GET /ping).
// При отсутствии пула соединений отвечает 500.
func PingDB(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if pool == nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}
