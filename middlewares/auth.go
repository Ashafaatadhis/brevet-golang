package middlewares

import (
	"brevet-api/config"
	"brevet-api/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth is a middleware to check for valid JWT tokens
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Missing or invalid Authorization header", nil)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Cek blacklist token di Redis
		val, err := config.RedisClient.Get(config.Ctx, tokenString).Result()
		if err == nil && val == "blacklisted" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)

		}

		jwtSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key") // Consider making this required in production
		user, err := utils.ExtractClaimsFromToken(tokenString, jwtSecret)
		if err != nil || user == nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)
		}

		c.Locals("user", user)
		c.Locals("access_token", tokenString)

		return c.Next()
	}
}
