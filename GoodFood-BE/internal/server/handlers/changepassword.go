package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

type ChangePassResponse struct{
	AccountID int `json:"accountID"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}
type ChangePassErr struct{
	ErrOldPassword string `json:"errOldPassword"`
	ErrNewPassword string `json:"errNewPassword"`
	ErrConfirmPassword string `json:"errConfirmPassword"`
	ErrAccount string
}
func ChangePasswordSubmit(c *fiber.Ctx) error{
	body := ChangePassResponse{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	user, err := models.Accounts(qm.Where("\"accountID\" = ?",body.AccountID)).One(context.Background(),boil.GetContextDB()); 
	if err!= nil{
		return service.SendError(c,500,err.Error());
	}

	if errObj, ok := validateChangePass(&body,*user); !ok{
		return service.SendErrorStruct(c,500,errObj);
	}
	//encrypting password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword),bcrypt.DefaultCost);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	user.Password = string(hashedPass);
	_, err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	refreshToken := c.Cookies("refreshToken","");
	fmt.Println(refreshToken)
	if refreshToken != ""{
		c.Cookie(&fiber.Cookie{
			Name: "refreshToken",
			Value: refreshToken,
			Path: "/",
			MaxAge: -1, //7 days
			HTTPOnly: true,
			Secure: false, //Switch to `true` if running HTTPS
			SameSite: "None",
		})
	}
	fmt.Println(refreshToken)

	resp := fiber.Map{
		"status": "Success",
		"data": body,
		"message": "Successfully changed password!",
	}

	return c.JSON(resp);
}

func validateChangePass(body *ChangePassResponse, user models.Account) (ChangePassErr, bool){
	valid := true;
	error := ChangePassErr{}
	//old password err
	if len(body.OldPassword) <= 7{
		error.ErrOldPassword = "Old password needs to be at least 8 characters!";
		valid = false;
	}else if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword)); err!=nil{
		valid = false;
		error.ErrOldPassword = "Wrong password!";
	}
	//new password err
	if len(body.NewPassword) <= 7{
		error.ErrNewPassword = "New password needs to be at least 8 characters!";
		valid = false;
	}else if err := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(body.NewPassword)); err == nil{
		valid = false;
		error.ErrNewPassword = "New password can't be old password!"
	}
	//confirm password err
	if body.ConfirmPassword != body.NewPassword{
		error.ErrConfirmPassword = "Password does not match!"
		valid = false
	}

	return error,valid
}