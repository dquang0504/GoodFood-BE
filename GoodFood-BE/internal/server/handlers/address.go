package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func FetchAddress(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID");
	}
	//Receive page num
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,400,"Did not receive pageNum");
	}
	//Calculate offset
	offset := (page-1)*6;

	addresses,err := models.Addresses(
		qm.Where("\"accountID\" = ?",accountID),
		qm.Limit(6),
		qm.Offset(offset),
		qm.OrderBy("\"addressID\" DESC"),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Error fetching addresses");
	}

	//counting the total of address
	totalAddress,err := models.Addresses().Count(c.Context(),boil.GetContextDB());
	if err != nil {
		return service.SendError(c, 500, "Total address not found")
	}

	totalPage := int(math.Ceil(float64(totalAddress) / float64(6)))

	resp := fiber.Map{
		"status": "Success",
		"data": addresses,
		"totalPage": totalPage,
		"message": "Successfully fetched addresses",
	}

	return c.JSON(resp);
}

func AddressInsert(c *fiber.Ctx) error{
	body := c.Body()
    fmt.Println("Request Body:", string(body)) // Log request body để kiểm tra
	var addressDetails models.Address
	if err := c.BodyParser(&addressDetails); err != nil{
		return service.SendError(c,400,"Invalid request body");
	}

	_,err := models.Accounts(qm.Where("\"accountID\" = ?",addressDetails.AccountID)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,"AccountID not found!");
	}

	//insert
	if err = addressDetails.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		fmt.Printf("Insert error: %+v\n", err) // In lỗi chi tiết
		return service.SendError(c,500,"Couldnt insert new address");
	}

	resp := fiber.Map{
		"status": "Success",
		"data": addressDetails,
		"message": "Successfully inserted new address",
	}
	return c.JSON(resp);
}

func AddressDetail(c *fiber.Ctx) error{
	addressID := c.QueryInt("addressID",0)
	if addressID == 0{
		return service.SendError(c,400,"Did not receive addressID");
	}
	accountID := c.QueryInt("accountID",0)
	if accountID == 0{
		return service.SendError(c,400,"Did not receive addressID");
	}

	addressDetail,err := models.Addresses(qm.Where("\"addressID\" = ? AND \"accountID\" = ?",addressID,accountID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Error fetching address details");
	}

	resp := fiber.Map{
		"status": "Success",
		"data": addressDetail,
		"message": "Successfully fetched address details",
	}

	return c.JSON(resp);
}

func AddressUpdate(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID")
	}
	addressID := c.QueryInt("addressID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive addressID")
	}

	var updateDetails models.Address
	if err := c.BodyParser(&updateDetails);err != nil{
		return service.SendError(c,400,"Invalid body request");
	}

	toBeUpdated,err := models.Addresses(
		qm.Where("\"accountID\" = ? AND \"addressID\" = ?",accountID,addressID),
	).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Cannot find the specified address");
	}

	//update details
	toBeUpdated.Address = updateDetails.Address
	toBeUpdated.DistrictID = updateDetails.DistrictID
	toBeUpdated.FullName = updateDetails.FullName
	toBeUpdated.PhoneNumber = updateDetails.PhoneNumber
	toBeUpdated.SpecificAddress = updateDetails.SpecificAddress
	toBeUpdated.Status = updateDetails.Status
	toBeUpdated.WardCode = updateDetails.WardCode

	_,err = toBeUpdated.Update(c.Context(),boil.GetContextDB(),boil.Infer())
	if err != nil{
		return service.SendError(c,500,"Cannot update the specified address")
	}

	resp := fiber.Map{
		"status": "Success",
		"data": toBeUpdated,
		"message": "Successfully updated the address",
	}

	return c.JSON(resp);
}

func AddressDelete(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID")
	}
	addressID := c.QueryInt("addressID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive addressID")
	}

	toBeDeleted,err := models.Addresses(
		qm.Where("\"accountID\" = ? AND \"addressID\" = ?",accountID,addressID),
	).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Cannot find the specified address");
	}

	_,err = toBeDeleted.Delete(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,"Cannot delete the address")
	}

	resp := fiber.Map{
		"status": "Success",
		"data": toBeDeleted,
		"message": "Successfully updated the address",
	}

	return c.JSON(resp);
}