package auth

import (
	"GoodFood-BE/internal/service"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//AuthMiddleware verifies access token from header request
func AuthMiddleware(c *fiber.Ctx) error{
	//Get token from authorization header
	authHeader := c.Get("Authorization")
	if authHeader == ""{
		return service.SendError(c,401,"Missing Authorization header")
	}

	//Verifying bearer token format
	splitToken := strings.Split(authHeader, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer"{
		return service.SendError(c,401,"Invalid Authorization format")
	}

	tokenString := splitToken[1]

	//Verify the token
	claims,err := VerifyToken(tokenString)
	if err != nil{
		return service.SendError(c,401,"Invalid or expired token");
	}

	//Save the username into context so that other handlers could use it
	c.Locals("username",claims.Username)

	//Contintue to handle the request
	return c.Next()
}

// OptionalAuthMiddleware tries to parse the token if present, else continue without error.
func OptionalAuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// Không có token => coi như anonymous
		return c.Next()
	}

	// Verifying bearer token format
	splitToken := strings.Split(authHeader, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer" {
		// Có Authorization nhưng sai format => vẫn trả lỗi
		return service.SendError(c, 401, "Invalid Authorization format")
	}

	tokenString := splitToken[1]

	// Verify the token
	claims, err := VerifyToken(tokenString)
	if err != nil {
		return service.SendError(c, 401, "Invalid or expired token")
	}

	///Save the username into context so that other handlers could use it
	c.Locals("username",claims.Username)

	return c.Next()
}

//Helper function to fetch the logged in user in handlers
func GetAuthenticatedUser(c *fiber.Ctx) string{
	username, ok := c.Locals("username").(string)
	if !ok {
		return ""
	}
	return username
}