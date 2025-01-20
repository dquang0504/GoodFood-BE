package handlers

import (
	"GoodFood-BE/models"
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func HandleLogin(c *fiber.Ctx) error{
	//fetching parameters from url
	username := c.Query("username")
	password := c.Query("password")
	if(username == "" || password == "" ){
		return service.SendError(c,400,"Either username or password wasn't valid")
	}

	//comparing login details with users db
	user, err := models.Accounts(qm.Where("username = ? AND password = ?", username, password)).One(c.Context(), boil.GetContextDB())

	if err != nil{
		return service.SendError(c,500,"User not found!")
	}

	//provide user with a token
	token,err := auth.CreateToken(username);

	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": user,
			"token": token,
		},
		"message": "Successfully fetched login details!",
	}
	return c.JSON(response);
}