package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/jobs"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/hibiken/asynq"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

func HandleRegister(c *fiber.Ctx) error{
	hasPassword := true;
	hasUsername := true
	var user models.Account
	if err := c.BodyParser(&user); err != nil{
		return service.SendError(c,400,err.Error());
	}

	if valid, errObj := validationUser(&user,hasPassword,hasUsername); !valid{
		return service.SendErrorStruct(c,500,errObj)
	}

	//encrypting password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password),bcrypt.DefaultCost)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	user.Password = string(hashedPass);
	err = user.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully registered!",
	}
	return c.JSON(response);
}

func HandleLogin(c *fiber.Ctx) error{
	//fetching parameters from url
	username := c.Query("username")
	password := c.Query("password")
	if(username == "" || password == "" ){
		return service.SendError(c,400,"Either username or password wasn't valid")
	}

	//comparing login details with users db
	user, err := models.Accounts(qm.Where("username = ?", username)).One(c.Context(), boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,"User not found!")
	}

	//comparing 2 hashes
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(password)); err != nil{
		return service.SendError(c,401, "Username or password does not match!");
	}

	//provide user with a token
	accessToken,refreshToken, err := auth.CreateToken(username)
	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	//set refreshToken as HTTP-only Cookie
	c.Cookie(&fiber.Cookie{
		Name: "refreshToken",
		Value: refreshToken,
		Path: "/",
		MaxAge: 7*24*60*60, //7 days
		HTTPOnly: true,
		Secure: false, //Switch to `true` if running HTTPS
		SameSite: "None",
	})

	fmt.Println("Cookie set:", c.Cookies("refreshToken"))

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": user,
			"accessToken": accessToken,
		},
		"message": "Successfully fetched login details!",
	}
	return c.JSON(response);
}

func HandleUpdateAccount(c *fiber.Ctx) error{
	hasPassword := false;
	hasUsername := false;
	var body models.Account
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID!");
	}

	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error());
	}
	account, err := models.Accounts(qm.Where("\"accountID\" = ?",accountID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	if ok, errObj := validationUser(&body,hasPassword,hasUsername); !ok{
		return service.SendErrorStruct(c,400,errObj);
	}

	//start updating account info
	account.FullName = body.FullName
	account.PhoneNumber = null.StringFrom(body.PhoneNumber.String)
	account.Email = body.Email
	fmt.Println("Body: ",body.Avatar.String)
	//start here
	if body.Avatar.String != ""{
		account.Avatar = null.StringFrom(body.Avatar.String)
	}
	fmt.Println("Account: ",account.Avatar)
	_, err = account.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": account,
		"message": "Successfully updated account information!",
	}

	return c.JSON(resp);
}

type ForgotPassStruct struct{
	Email string `json:"email"`
	CodeOTP string `json:"codeOTP"`
}
var asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379",Password: "",DB: 0})
func HandleForgotPassword(c *fiber.Ctx) error{
	body := ForgotPassStruct{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error())
	}
	if body.Email == ""{
		return service.SendError(c,400,"Please input your email!");
	}
	
	token, err := uuid.NewV4();
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s",token.String());
	//only send mmail if the user exists in database
	_, err = models.Accounts(qm.Where("email = ?",body.Email)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		resp := fiber.Map{
			"status": "Succes",
			"message": "If the email is registered, a password reset link will be sent to your inbox.",
		}
		return c.JSON(resp);
	}

	//caching token key if user exists
	redisKey := fmt.Sprintf("resetPass:token=%s",token.String())
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,body.Email, 10*time.Minute).Err()
	if rdsErr != nil{
		fmt.Println("Failed to cache cart data: ",rdsErr)
	}

	//creating email sending task
	task, err := jobs.NewResetPasswordEmailTask(body.Email,resetLink);
	if err != nil{
		return service.SendError(c,500, err.Error())
	}
	//enqueuing
	_,err = asynqClient.Enqueue(task)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	// err = service.SendResetPasswordEmail(body.Email,resetLink)
	// if err != nil{
	// 	return service.SendError(c,500,err.Error());
	// }

	resp := fiber.Map{
		"status": "Succes",
		"data": token,
		"message": "If the email is registered, a password reset link will be sent to your inbox.",
	}
	return c.JSON(resp);
}

func ValidateResetToken(c *fiber.Ctx) error{
	paramToken := c.Query("token","");
	if paramToken == ""{
		return service.SendError(c,400,"Did not receive token!");
	}
	redisKey := fmt.Sprintf("resetPass:token=%s",paramToken);
	cachedToken, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		fmt.Println("Hello teacher");
		resp := fiber.Map{
			"status": "Success",
			"isValid": true,
			"email": cachedToken,
			"message": "Successfully validated reset password token!",
		}
		return c.JSON(resp);
	}
	return nil
}

type ResetPass struct{
	NewPass string `json:"newPass"`
	ConfirmPass string `json:"confirmPass"`
	Email string `json:"email"`
}
func HandleResetPassword(c *fiber.Ctx) error{
	body := ResetPass{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error())
	}
	
	if(body.NewPass == ""){
		return service.SendError(c,400,"Please input your new password!");
	}

	//encrypting password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(body.NewPass),bcrypt.DefaultCost);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//finding user with email from body
	user, err := models.Accounts(qm.Where("email = ?",body.Email)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//updating password
	user.Password = string(hashedPass);
	_, err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	//deleting token in redis
	paramToken := c.Query("token","");
	if paramToken != ""{
		redisKey := fmt.Sprintf("resetPass:token=%s",paramToken);
		err := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if err != nil{
			fmt.Println("Failed to delete reset token from Redis:", err)
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully reset password!",
	}
	return c.JSON(resp);
}

func RefreshToken(c *fiber.Ctx) error{
	//Fetch refreshToken from HTTP-only Cookie
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == ""{
		return service.SendError(c,401,"Missing refresh token")
	}

	//Verify refreshToken
	claims, err := auth.VerifyToken(refreshToken)
	if err != nil{
		return service.SendError(c, 401, "Invalid refresh token")
	}

	//Generate new accessToken
	accessToken, newRefreshToken, err := auth.CreateToken(claims.Username)
	if err != nil{
		return service.SendError(c, 500, "Cound not generate token")
	}

	//Return new accessToken and update new refresh token into Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    newRefreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 ngày
		HTTPOnly: true,
		Secure:   false, // Đổi thành `true` nếu chạy HTTPS
		SameSite: "None",
	})
	response := fiber.Map{
		"status": "Success",
		"accessToken": accessToken,
	}
	return c.JSON(response)
}

type ContactResponse struct{
	Fullname string `json:"name"`
	Email string `json:"fromEmail"`
	Message string `json:"content"`
}
func HandleContact(c *fiber.Ctx) error{
	body := ContactResponse{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error())
	}

	//creating email sending task
	task, err := jobs.NewCustomerSentContactTask(body.Fullname,body.Email,body.Message);
	if err != nil{
		return service.SendError(c,500, err.Error())
	}
	//enqueuing
	_,err = asynqClient.Enqueue(task)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	response := fiber.Map{
		"status": "Success",
		"data": "",
		"message": "Successfully sent a message. We will contact you back shortly!",
	}

	return c.JSON(response);
}