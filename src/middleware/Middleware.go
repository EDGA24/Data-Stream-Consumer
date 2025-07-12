package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Constantes de configuración
const (
	// Tamaño máximo por defecto: 1MB
	DefaultMaxRequestBodySize = 1 << 20 // 1MB en bytes
	
	// Tamaños alternativos comunes
	MaxRequestBodySize512KB = 512 << 10  // 512KB
	MaxRequestBodySize2MB   = 2 << 20    // 2MB
	MaxRequestBodySize5MB   = 5 << 20    // 5MB
)

// Métodos HTTP que requieren validación de body
var methodsWithBody = []string{"POST", "PUT", "PATCH"}

// =======================
// MIDDLEWARES ATÓMICOS
// =======================

// ValidateContentType - Middleware atómico para validar Content-Type JSON
func ValidateContentType() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Solo validar métodos que envían body
		if !containsMethod(ctx.Request.Method, methodsWithBody) {
			ctx.Next()
			return
		}

		contentType := ctx.GetHeader("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "application/json") {
			ctx.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json",
				"code":  "INVALID_CONTENT_TYPE",
				"details": map[string]interface{}{
					"received": contentType,
					"expected": "application/json",
					"endpoint": ctx.Request.URL.Path,
					"method":   ctx.Request.Method,
				},
			})
			ctx.Abort()
			return
		}
		
		ctx.Next()
	}
}

// ValidateRequestBodySize - Middleware atómico para validar tamaño del request body
func ValidateRequestBodySize(maxSize int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Solo validar métodos que envían body
		if !containsMethod(ctx.Request.Method, methodsWithBody) {
			ctx.Next()
			return
		}

		// Verificar Content-Length header
		if ctx.Request.ContentLength > maxSize {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large. Maximum allowed size is %d bytes (%.1fKB)", 
					maxSize, float64(maxSize)/1024),
				"code": "REQUEST_TOO_LARGE",
				"details": map[string]interface{}{
					"received_size_bytes": ctx.Request.ContentLength,
					"received_size_kb":    float64(ctx.Request.ContentLength) / 1024,
					"max_allowed_bytes":   maxSize,
					"max_allowed_kb":      float64(maxSize) / 1024,
					"endpoint":            ctx.Request.URL.Path,
					"method":              ctx.Request.Method,
				},
			})
			ctx.Abort()
			return
		}

		// Limitar el reader para prevenir lecturas excesivas
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxSize)
		
		ctx.Next()
	}
}

// ValidateRequestBodySizeDefault - Middleware con tamaño por defecto (1MB)
func ValidateRequestBodySizeDefault() gin.HandlerFunc {
	return ValidateRequestBodySize(DefaultMaxRequestBodySize)
}

// =======================
// MIDDLEWARES COMBINADOS
// =======================

// SecurityValidation - Middleware combinado con validaciones de seguridad básicas
func SecurityValidation() gin.HandlerFunc {
	return SecurityValidationWithSize(DefaultMaxRequestBodySize)
}

// SecurityValidationWithSize - Middleware combinado con tamaño personalizable
func SecurityValidationWithSize(maxSize int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Solo aplicar a métodos que envían body
		if !containsMethod(ctx.Request.Method, methodsWithBody) {
			ctx.Next()
			return
		}

		// 1. Validar tamaño del request body
		if ctx.Request.ContentLength > maxSize {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large. Maximum allowed size is %.1fKB", 
					float64(maxSize)/1024),
				"code": "REQUEST_TOO_LARGE",
				"details": map[string]interface{}{
					"received_size_bytes": ctx.Request.ContentLength,
					"received_size_kb":    float64(ctx.Request.ContentLength) / 1024,
					"max_allowed_bytes":   maxSize,
					"max_allowed_kb":      float64(maxSize) / 1024,
					"endpoint":            ctx.Request.URL.Path,
					"method":              ctx.Request.Method,
				},
			})
			ctx.Abort()
			return
		}

		// Limitar el reader
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxSize)

		// 2. Validar Content-Type
		contentType := ctx.GetHeader("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "application/json") {
			ctx.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json",
				"code":  "INVALID_CONTENT_TYPE",
				"details": map[string]interface{}{
					"received": contentType,
					"expected": "application/json",
					"endpoint": ctx.Request.URL.Path,
					"method":   ctx.Request.Method,
				},
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// =======================
// MIDDLEWARES ESPECIALIZADOS
// =======================

// SensorDataValidation - Middleware específico para endpoints de sensores
func SensorDataValidation() gin.HandlerFunc {
	return SecurityValidationWithSize(DefaultMaxRequestBodySize)
}

// LargeDataValidation - Middleware para endpoints que manejan datos grandes
func LargeDataValidation() gin.HandlerFunc {
	return SecurityValidationWithSize(MaxRequestBodySize5MB)
}

// SmallDataValidation - Middleware para endpoints que solo manejan datos pequeños
func SmallDataValidation() gin.HandlerFunc {
	return SecurityValidationWithSize(MaxRequestBodySize512KB)
}

// =======================
// MIDDLEWARE DE MANEJO DE ERRORES
// =======================

// HandleMaxBytesError - Middleware para capturar errores de MaxBytesReader
func HandleMaxBytesError() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// Verificar errores después del procesamiento
		if len(ctx.Errors) > 0 {
			for _, err := range ctx.Errors {
				if strings.Contains(err.Error(), "request body too large") || 
				   strings.Contains(err.Error(), "http: request body too large") {
					if !ctx.Writer.Written() {
						ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
							"error": "Request body exceeded maximum allowed size during processing",
							"code":  "REQUEST_TOO_LARGE_STREAM",
							"details": map[string]interface{}{
								"endpoint":   ctx.Request.URL.Path,
								"method":     ctx.Request.Method,
								"error_type": "stream_limit_exceeded",
							},
						})
					}
					return
				}
			}
		}
	}
}

// =======================
// MIDDLEWARE DE LOGGING (OPCIONAL)
// =======================

// RequestLogger - Middleware para logging de requests con validaciones
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if containsMethod(ctx.Request.Method, methodsWithBody) {
			fmt.Printf("[MIDDLEWARE] %s %s - Content-Type: %s - Content-Length: %d bytes\n", 
				ctx.Request.Method, 
				ctx.Request.URL.Path, 
				ctx.GetHeader("Content-Type"),
				ctx.Request.ContentLength)
		}
		ctx.Next()
	}
}

// =======================
// FUNCIONES HELPER
// =======================

// containsMethod - Verifica si un método está en la lista de métodos
func containsMethod(method string, methods []string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

// BytesToKB - Convierte bytes a KB para mensajes más legibles
func BytesToKB(bytes int64) float64 {
	return float64(bytes) / 1024
}

// BytesToMB - Convierte bytes a MB para mensajes más legibles
func BytesToMB(bytes int64) float64 {
	return float64(bytes) / (1024 * 1024)
}

// =======================
// CONFIGURACIÓN PREDEFINIDA
// =======================

// MiddlewareConfig - Estructura para configuración de middlewares
type MiddlewareConfig struct {
	MaxBodySize      int64
	RequireJSON      bool
	EnableLogging    bool
	EnableErrorHandler bool
}

// DefaultConfig - Configuración por defecto
var DefaultConfig = MiddlewareConfig{
	MaxBodySize:        DefaultMaxRequestBodySize,
	RequireJSON:        true,
	EnableLogging:      false,
	EnableErrorHandler: true,
}

// ApplyMiddlewares - Aplica middlewares basado en configuración
func ApplyMiddlewares(router *gin.Engine, config MiddlewareConfig) {
	if config.EnableLogging {
		router.Use(RequestLogger())
	}
	
	if config.EnableErrorHandler {
		router.Use(HandleMaxBytesError())
	}
	
	if config.RequireJSON && config.MaxBodySize > 0 {
		router.Use(SecurityValidationWithSize(config.MaxBodySize))
	} else if config.RequireJSON {
		router.Use(ValidateContentType())
	} else if config.MaxBodySize > 0 {
		router.Use(ValidateRequestBodySize(config.MaxBodySize))
	}
}

// ApplyMiddlewaresToGroup - Aplica middlewares a un grupo específico
func ApplyMiddlewaresToGroup(group *gin.RouterGroup, config MiddlewareConfig) {
	if config.EnableLogging {
		group.Use(RequestLogger())
	}
	
	if config.EnableErrorHandler {
		group.Use(HandleMaxBytesError())
	}
	
	if config.RequireJSON && config.MaxBodySize > 0 {
		group.Use(SecurityValidationWithSize(config.MaxBodySize))
	} else if config.RequireJSON {
		group.Use(ValidateContentType())
	} else if config.MaxBodySize > 0 {
		group.Use(ValidateRequestBodySize(config.MaxBodySize))
	}
}