package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

//ChangePasswordSubmit changes the password for users and update corresponding tokens in the process.
func ChangePasswordSubmit(c *fiber.Ctx) error{
	body := dto.ChangePassRequest{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	//Fetch user
	user, err := models.FindAccount(c.Context(),boil.GetContextDB(),body.AccountID)
	if err != nil{
		return service.SendError(c,404,err.Error());
	}

	//Validate input
	if errObj, ok := validateChangePass(&body,*user); !ok{
		return service.SendErrorStruct(c,400,errObj);
	}

	//Encrypt new password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword),bcrypt.DefaultCost);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//Update
	user.Password = string(hashedPass);
	_, err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Clear refresh cookie if exists
	refreshToken := c.Cookies("refreshToken","");
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

	resp := fiber.Map{
		"status": "Success",
		"data": body,
		"message": "Successfully changed password!",
	}

	return c.JSON(resp);
}

//validateChangePass validates input fields when changing password
func validateChangePass(body *dto.ChangePassRequest, user models.Account) (dto.ChangePassErr, bool){
	valid := true;
	error := dto.ChangePassErr{}
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