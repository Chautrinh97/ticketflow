package http

import (
	"github.com/gin-gonic/gin"

	"ticketflow/pkg/httpauth"
)

func NewRouter(jwtSecret []byte, bookings *BookingHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	v1.Use(httpauth.RequireAuth(jwtSecret))
	{
		v1.POST("/bookings", bookings.Create)
		v1.GET("/bookings/:id", httpauth.RequireOwnership(bookings.OwnerLookup), bookings.GetByID)
		v1.GET("/users/me/bookings", bookings.ListMine)
	}
	return r
}
