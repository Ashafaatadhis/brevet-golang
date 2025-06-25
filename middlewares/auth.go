package middlewares

import (
	"brevet-api/config"
	"brevet-api/utils"
	"strings"

	"slices"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth is a middleware to check if the user is authenticated
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Cek apakah Authorization kosong atau tidak diawali dengan Bearer
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Missing or invalid Authorization header", nil)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		// Cek apakah token kosong atau string "undefined"
		if tokenString == "" || tokenString == "undefined" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
		}

		// Cek blacklist token di Redis
		val, err := config.RedisClient.Get(config.Ctx, tokenString).Result()
		if err == nil && val == "blacklisted" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak berlaku", nil)
		}

		jwtSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key")
		user, err := utils.ExtractClaimsFromToken(tokenString, jwtSecret)
		if err != nil || user == nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid atau kadaluarsa", nil)
		}

		// Simpan ke context
		c.Locals("user", user)
		c.Locals("access_token", tokenString)

		return c.Next()
	}
}

// RequireRole is function to check role
func RequireRole(allowedRoles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRaw := c.Locals("user")
		if userRaw == nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: user not found", nil)
		}

		user, ok := userRaw.(*utils.Claims)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: invalid user data", nil)
		}

		if slices.Contains(allowedRoles, user.Role) {
			return c.Next()
		}

		return utils.ErrorResponse(c, fiber.StatusForbidden, "Forbidden: insufficient role access", nil)
	}
}
