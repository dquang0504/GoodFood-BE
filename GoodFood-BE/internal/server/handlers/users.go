package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

func HandleRegister(c *fiber.Ctx) error{
	var user models.Account
	if err := c.BodyParser(&user); err != nil{
		return service.SendError(c,400,err.Error());
	}

	if valid, errObj := validationUser(&user); !valid{
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
		return service.SendError(c,401, err.Error());
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