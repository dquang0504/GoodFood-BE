package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

// GetAdminUsers returns paginated users with filter and summary cards.
func GetAdminUsers(c *fiber.Ctx) error{
	//Fetch user cards
	query := `
		SELECT COALESCE(COUNT("accountID"),0) AS totalusers,
		COUNT(CASE WHEN status = false THEN 1 END) AS totaldisabled
		FROM account
	`
	cards, err := utils.FetchCards(c,query,&dto.UserCards{});
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	//Fetch query params
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	sort := c.Query("sort","");
	search := c.Query("search","");

	// Handle search and sort filter and filling queryMods
	queryMods := []qm.QueryMod{}
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

	//Count total users with queryMods
	totalUser, err := models.Accounts(queryMods...).Count(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	//Calculate total pages and offset
	offset, totalPage := utils.Paginate(page,utils.PageSize,int(totalUser))

	queryMods = append(queryMods, qm.OrderBy("\"accountID\" DESC"), qm.Limit(utils.PageSize), qm.Offset(offset))

	
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

// GetAdminUserDetail returns detail of a single user by accountID.
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

// AdminUserCreate creates a new user after validating input.
func AdminUserCreate(c *fiber.Ctx) error{
	var(
		hasPassword = false
		hasUsername = true
		hasEmail = false
		hasPhone = false
	)
	var user models.Account
	if err := c.BodyParser(&user); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	//Fields validation
	if valid, errObj := validationUser(&user,hasPassword,hasUsername,hasEmail,hasPhone); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	//Insert
	if err := user.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully created new user",
	}

	return c.JSON(resp);
}

// AdminUserUpdate updates selected fields of a user.
func AdminUserUpdate(c *fiber.Ctx) error{
	//Fetch query param
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

	// Only update specific fields
	user.FullName = userBody.FullName
	user.Role = userBody.Role
	user.Gender = userBody.Gender
	user.Status = userBody.Status
	if _,err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully updated user",
	}

	return c.JSON(resp);

}

// validateUser checks input fields depending on required flags.
func validationUser(user *models.Account, hasPassword bool,hasUsername bool, hasEmail bool,hasPhone bool) (bool,dto.UserError){
	var error dto.UserError
	isValid := true
	if user.FullName == ""{
		error.ErrName = "Please input your full name!"
		isValid = false;
	}
	if(hasEmail){
		if user.Email == ""{
			error.ErrEmail = "Please input your email!"
			isValid = false;
		}else if res, err := models.Accounts(qm.Where("email = ?",user.Email)).Exists(context.Background(),boil.GetContextDB()); err == nil{
			if res{
				error.ErrEmail = "Email already exists!"
				isValid = false;
			}
		}
	}
	if (hasUsername){
		if user.Username == ""{
			error.ErrUsername = "Please input your username!"
			isValid = false;
		}else if res, err := models.Accounts(qm.Where("username = ?",user.Username)).Exists(context.Background(),boil.GetContextDB()); err == nil{
			if res{
				error.ErrUsername = "Username already exists!"
				isValid = false;
			}
		}
	}
	if(hasPhone){
		if user.PhoneNumber.String == ""{
			error.ErrPhone = "Please input your phone number!"
			isValid = false;
		}else if res, err := models.Accounts(qm.Where("\"phoneNumber\" = ?",user.PhoneNumber.String)).Exists(context.Background(),boil.GetContextDB()); err == nil{
			if res{
				error.ErrPhone = "Phone number already exists!"
				isValid = false;
			}
		}
	}
	if(hasPassword){
		if user.Password == ""{
			error.ErrPassword = "Please input your password!"
			isValid = false;
		}
	}
	return isValid,error;
}