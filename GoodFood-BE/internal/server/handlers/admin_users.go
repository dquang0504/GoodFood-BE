package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"context"
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserCards struct {
	TotalUsers    int `boil:"totalusers"`
	TotalDisabled int `boil:"totaldisabled"`
}
func GetAdminUsers(c *fiber.Ctx) error{

	var cards UserCards

	err := queries.Raw(`
		SELECT COALESCE(COUNT("accountID"),0) AS totalusers,
		COUNT(CASE WHEN status = false THEN 1 END) AS totaldisabled
		FROM account
	`).Bind(c.Context(),boil.GetContextDB(),&cards)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//fetching page number
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	//fetching sort and search
	sort := c.Query("sort","");
	search := c.Query("search","");
	//calculating offset
	offset := (page-1)*6;
	
	//creating query mod
	queryMods := []qm.QueryMod{}

	// handling search and sort filter and filling queryMods
	if search != ""{
		fmt.Println(sort)
		switch sort{
			case "Username":
				queryMods = append(queryMods, qm.Where("username ILIKE ?","%"+search+"%"));
			case "Phone number":
				queryMods = append(queryMods, qm.Where("\"phoneNumber\" ILIKE ?","%"+search+"%"))
			case "Email":
				queryMods = append(queryMods, qm.Where("email ILIKE ?","%"+search+"%"))
			case "Full name":
				queryMods = append(queryMods, qm.Where("\"fullName\" ILIKE ?","%"+search+"%"))
			default:
				//fallback do nothing
		}
	}

	//counting total users matching filter
	totalUser, err := models.Accounts(queryMods...).Count(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	//calculating total pages
	totalPage := int(math.Ceil(float64(totalUser)/float64(6)))

	queryMods = append(queryMods, qm.OrderBy("\"accountID\" DESC"), qm.Limit(6), qm.Offset(offset))

	
	users, err := models.Accounts(queryMods...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	resp := fiber.Map{
		"status": "Success",
		"data": users,
		"cards": cards,
		"totalPage": totalPage,
		"message": "Successfully fetched user values",
	}

	return c.JSON(resp);
}

func GetAdminUserDetail(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID");
	}

	account, err := models.Accounts(qm.Where("\"accountID\" = ?",accountID)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": account,
		"message": "Successfully fetched user values",
	}

	return c.JSON(resp);
}

type UserError struct{
	ErrName string `json:"errName"`
	ErrUsername string `json:"errUsername"`
	ErrPhone string `json:"errPhone"`
	ErrEmail string `json:"errEmail"`
	ErrPassword string `json:"errPassword"`
}

func AdminUserCreate(c *fiber.Ctx) error{
	var user models.Account
	if err := c.BodyParser(&user); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	if valid, errObj := validationUser(&user); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	err := user.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully created new user",
	}

	return c.JSON(resp);
}

func AdminUserUpdate(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID");
	}
	var userBody models.Account
	if err := c.BodyParser(&userBody); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	user, err := models.Accounts(qm.Where("\"accountID\" = ?",accountID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	//update
	user.FullName = userBody.FullName
	user.Role = userBody.Role
	user.Gender = userBody.Gender
	user.Status = userBody.Status
	_,err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully updated user",
	}

	return c.JSON(resp);

}

func validationUser(user *models.Account) (bool,UserError){
	var error UserError
	isValid := true
	if user.FullName == ""{
		error.ErrName = "Please input your full name!"
		isValid = false;
	}
	if user.Email == ""{
		error.ErrEmail = "Please input your email!"
		isValid = false;
	}else if res, err := models.Accounts(qm.Where("email = ?",user.Email)).Exists(context.Background(),boil.GetContextDB()); err == nil{
		if res{
			error.ErrEmail = "Email already exists!"
			isValid = false;
		}
	}
	if user.Username == ""{
		error.ErrUsername = "Please input your username!"
		isValid = false;
	}else if res, err := models.Accounts(qm.Where("username = ?",user.Username)).Exists(context.Background(),boil.GetContextDB()); err == nil{
		if res{
			error.ErrUsername = "Username already exists!"
			isValid = false;
		}
	}
	if user.PhoneNumber.String == ""{
		error.ErrPhone = "Please input your phone number!"
		isValid = false;
	}else if res, err := models.Accounts(qm.Where("\"phoneNumber\" = ?",user.PhoneNumber.String)).Exists(context.Background(),boil.GetContextDB()); err == nil{
		if res{
			error.ErrPhone = "Phone number already exists!"
			isValid = false;
		}
	}
	if user.Password == ""{
		error.ErrPassword = "Please input your password!"
		isValid = false;
	}
	return isValid,error;
}