package middleware

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

// Función helper para validación en controladores específicos
func(ctx *gin.Context) {
			if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" || ctx.Request.Method == "PATCH" {
				contentType := ctx.GetHeader("Content-Type")
				if !strings.Contains(strings.ToLower(contentType), "application/json") {
					ctx.JSON(http.StatusUnsupportedMediaType, gin.H{
						"error": "Content-Type must be application/json",
						"code":  "INVALID_CONTENT_TYPE",
						"details": map[string]string{
							"received": contentType,
							"expected": "application/json",
						},
					})
					ctx.Abort()
					return
				}
			}
			ctx.Next()
		}