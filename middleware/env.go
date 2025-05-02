package middleware

import (
	"github.com/gin-gonic/gin"
)

// EnvironmentKey is the key used to store the environment in the context
const EnvironmentKey = "environment"

// DefaultEnvironment is the default environment to use if none is specified
const DefaultEnvironment = "production"

// EnvironmentMiddleware is a middleware that extracts the X-ENV header
// and stores it in the context. If the header is not present, it uses
// the default environment.
func EnvironmentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the X-ENV header
		env := c.GetHeader("X-ENV")
		if env == "" {
			env = DefaultEnvironment
		}

		// Store the environment in the context
		c.Set(EnvironmentKey, env)

		// Continue
		c.Next()
	}
}

// GetEnvironment returns the environment from the context
func GetEnvironment(c *gin.Context) string {
	env, exists := c.Get(EnvironmentKey)
	if !exists {
		return DefaultEnvironment
	}
	return env.(string)
}
