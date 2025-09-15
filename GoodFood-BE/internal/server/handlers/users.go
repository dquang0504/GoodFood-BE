package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/jobs"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"fmt"
	"os"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

// HandleRegister handles user registration
func HandleRegister(c *fiber.Ctx) error {
	var user models.Account
	if err := c.BodyParser(&user); err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	//Validate user input
	if valid, errObj := validationUser(&user, true, true, true, true); !valid {
		return service.SendErrorStruct(c, 500, errObj)
	}

	//Hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	user.Password = string(hashedPass)

	//Save user
	err = user.Insert(c.Context(), boil.GetContextDB(), boil.Infer())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	response := fiber.Map{
		"status":  "Success",
		"data":    user,
		"message": "Successfully registered!",
	}
	return c.JSON(response)
}

// HandleLogin handles login with username and password
func HandleLogin(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	//comparing login details with users db
	user, err := models.Accounts(qm.Where("username = ?", body.Username)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, "User not found!")
	}

	//comparing 2 hashes
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return service.SendError(c, 401, "Username or password does not match!")
	}

	return utils.LoginWithAccount(c, user)
}

// HandleLoginGoogle handles login with Google OAuth
func HandleLoginGoogle(c *fiber.Ctx) error {
	body := dto.OAuthLoginStruct{}
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	//Validate token
	audience := os.Getenv("GOOGLE_AUDIENCE")
	payload, err := idtoken.Validate(c.Context(), body.AccessToken, audience)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Parse payload
	sub := fmt.Sprintf("%v", payload.Claims["sub"]) //unique google id
	name := fmt.Sprintf("%v", payload.Claims["name"])
	email := fmt.Sprintf("%v", payload.Claims["email"])
	picture := fmt.Sprintf("%v", payload.Claims["picture"])

	return utils.HandleOAuthLogin(c, "google", sub, email, name, picture)
}

// HandleLoginFacebook handles login with Facebook OAuth
func HandleLoginFacebook(c *fiber.Ctx) error {
	body := dto.OAuthLoginStruct{}
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	fbUser, err := utils.GetFacebookUserInfo(body.AccessToken)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	return utils.HandleOAuthLogin(c, "facebook", fbUser.ID, fbUser.Email, fbUser.Name, fbUser.Picture.Data.URL)
}

// HandleUpdateAccount allows user to update their profile (fullname, avatar, phone, gender).
func HandleUpdateAccount(c *fiber.Ctx) error {
	//Fetch query param
	accountID := c.QueryInt("accountID", 0)
	if accountID == 0 {
		return service.SendError(c, 400, "Did not receive accountID!")
	}
	var body models.Account
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	//Fetch account
	account, err := models.FindAccount(c.Context(), boil.GetContextDB(), accountID)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Check logic for hasEmail and hasPhone
	var (
		hasEmail = false
		hasPhone = false
	)
	if body.Email != account.Email {
		hasEmail = true
	}
	if body.PhoneNumber.String != account.PhoneNumber.String {
		hasPhone = true
	}

	//Validate fields before update
	if ok, errObj := validationUser(&body, false, false, hasEmail, hasPhone); !ok {
		return service.SendErrorStruct(c, 400, errObj)
	}

	//start updating account info
	account.FullName = body.FullName
	if account.PhoneNumber != null.StringFrom(body.PhoneNumber.String) {
		account.PhoneNumber = null.StringFrom(body.PhoneNumber.String)
	}
	if account.Email != body.Email {
		account.Email = body.Email
	}
	if body.Avatar.String != "" {
		account.Avatar = null.StringFrom(body.Avatar.String)
	}

	if _, err = account.Update(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
		return service.SendError(c, 500, err.Error())
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    account,
		"message": "Successfully updated account information!",
	}

	return c.JSON(resp)
}

// HandleForgotPassword generates a password reset token and sends email.
var asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379", Password: "", DB: 0})

func HandleForgotPassword(c *fiber.Ctx) error {
	body := dto.ForgotPassStruct{}
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, err.Error())
	}

	//Find account by email
	acc, err := models.Accounts(qm.Where("email = ?", body.Email)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		resp := fiber.Map{
			"status":  "Success",
			"message": "If the email is registered, a password reset link will be sent to your inbox.",
		}
		return c.JSON(resp)
	}

	//Create reset token
	token, err := uuid.NewV4()
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", token.String())

	//Cache token key in redis if user exists
	redisKey := fmt.Sprintf("resetPass:token=%s", token.String())
	utils.SetCache(redisKey, acc.Email, 15*time.Minute, "")

	//Enqueue email task - asynq
	task, err := jobs.NewResetPasswordEmailTask(acc.Email, resetLink)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	//enqueuing
	_, err = asynqClient.Enqueue(task)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	resp := fiber.Map{
		"status":  "Succes",
		"data":    token,
		"message": "If the email is registered, a password reset link will be sent to your inbox.",
	}
	return c.JSON(resp)
}

// ValidateResetToken checks if reset token is still valid.
func ValidateResetToken(c *fiber.Ctx) error {
	token := c.Query("token", "")
	if token == "" {
		return service.SendError(c, 400, "Did not receive token!")
	}

	//Fetch cache
	redisKey := fmt.Sprintf("resetPass:token=%s", token)
	cachedToken := fiber.Map{}
	ok, err := utils.GetCache(redisKey, &cachedToken)

	//Validating cache
	if err != nil {
		return service.SendError(c, 500, "Error validating token")

	}
	if !ok {
		return service.SendError(c, 500, "Invalid or expired token")

	}

	resp := fiber.Map{
		"status":  "Success",
		"isValid": true,
		"email":   cachedToken,
		"message": "Successfully validated reset password token!",
	}
	return c.JSON(resp)
}

//HandleResetPassword updates password if token is valid.
func HandleResetPassword(c *fiber.Ctx) error {
	body := dto.ResetPass{}
	err := c.BodyParser(&body)
	if err != nil {
		return service.SendError(c, 400, "Invalid request body")
	}

	//Find account with email
	user, err := models.Accounts(qm.Where("email = ?", body.Email)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Hash new pass
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(body.NewPass), bcrypt.DefaultCost)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Update db
	user.Password = string(hashedPass)
	if _, err = user.Update(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Delete token from redis - one-time use
	redisKey := fmt.Sprintf("resetPass:token=%s", body.Token)
	utils.ClearCache(redisKey);

	resp := fiber.Map{
		"status":  "Success",
		"message": "Password has been reset successfully",
	}
	return c.JSON(resp)
}

func RefreshToken(c *fiber.Ctx) error {
	//Fetch sessionID from response
	var req struct {
		SessionID string `json:"sessionID"`
	}
	if err := c.BodyParser(&req); err != nil {
		return service.SendError(c, 401, "Mising sessionID")
	}

	cookie := c.Cookies("refreshToken")
	if cookie == "" {
		return service.SendError(c, 401, "No refresh token")
	}

	//Verify refreshToken
	claims, err := auth.VerifyToken(cookie)
	if err != nil {
		return service.SendError(c, 401, "Invalid refresh token")
	}

	if claims.SessionID != req.SessionID { //check for matching sessionID
		fmt.Println("claims ID: ", claims.SessionID)
		fmt.Println("req session ID: ", req.SessionID)
		return service.SendError(c, 401, "Session mismatch")
	}

	fmt.Println("claims ID current: ", claims.SessionID)

	//Generate new accessToken
	accessToken, _, _, err := auth.CreateToken(claims.Username, claims.SessionID)
	if err != nil {
		return service.SendError(c, 500, "Cound not generate token")
	}

	response := fiber.Map{
		"status":      "Success",
		"accessToken": accessToken,
	}
	return c.JSON(response)
}

func HandleContact(c *fiber.Ctx) error {
	body := dto.ContactResponse{}
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, err.Error())
	}

	//creating email sending task
	task, err := jobs.NewCustomerSentContactTask(body.Fullname, body.Email, body.Message)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	//enqueuing
	_, err = asynqClient.Enqueue(task)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	response := fiber.Map{
		"status":  "Success",
		"data":    "",
		"message": "Successfully sent a message. We will contact you back shortly!",
	}

	return c.JSON(response)
}
