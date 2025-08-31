package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"database/sql"
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//GetAdminProductTypes fetches product types with pagination, searching and number of products in each type.
func GetAdminProductTypes(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	search := c.Query("search", "")
	offset := (page - 1) * utils.PageSize

	//Build query
	queryMods := []qm.QueryMod{}
	if search != "" {
		queryMods = append(queryMods, qm.Where("\"typeName\" ILIKE ?", "%"+search+"%"))
	}
	//Count total types and calc total pages
	totalTypes, err := models.ProductTypes(queryMods...).Count(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	totalPage := int(math.Ceil(float64(totalTypes) / float64(6)))

	//get paginated product types
	queryMods = append(queryMods, qm.OrderBy("\"productTypeID\" DESC"), qm.Limit(6), qm.Offset(offset))
	productTypes, err := models.ProductTypes(queryMods...).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Instead of loading all products, run count query for each type (cheaper memory)
	response := make([]dto.ProductTypesResponse, len(productTypes))
	for i, pt := range productTypes {
		count, err := models.Products(qm.Where("\"productTypeID\" = ?",pt.ProductTypeID)).Count(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500, err.Error())
		}
		response[i] = dto.ProductTypesResponse{
			ProductType:  *pt,
			TotalProduct: int(count),
		}
	}

	resp := fiber.Map{
		"status":    "Success",
		"data":      response,
		"totalPage": totalPage,
		"message":   "Successfully fetched product types values",
	}

	return c.JSON(resp)
}

//GetAdminProductTypeDetail fetches a specific product type along with its detailed information
func GetAdminProductTypeDetail(c *fiber.Ctx) error {
	typeID := c.QueryInt("typeID", 0)
	if typeID == 0 {
		return service.SendError(c, 400, "Did not receive typeID")
	}

	pt, err := models.ProductTypes(qm.Where("\"productTypeID\" = ?", typeID)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return service.SendError(c,404,"Product type not found!");
		}
		return service.SendError(c, 500, err.Error())
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    pt,
		"message": "Successfully fetched product types detail",
	}

	return c.JSON(resp)
}
//AdminProductTypeCreate inserts a new record into ProductType table, also checks for duplicates
func AdminProductTypeCreate(c *fiber.Ctx,) error {
	var pt models.ProductType
	if err := c.BodyParser(&pt); err != nil {
		return service.SendError(c, 400, "Invalid body")
	}

	//Field validation
	if valid, errObj := validationProductType(c,&pt); !valid {
		return service.SendErrorStruct(c, 400, errObj)
	}

	//Insert
	if err := pt.Insert(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
		return service.SendError(c, 500, err.Error())
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    pt,
		"message": "Successfully created new product type",
	}

	return c.JSON(resp)
}

func AdminProductTypeUpdate(c *fiber.Ctx) error {
	var pt models.ProductType
	if err := c.BodyParser(&pt); err != nil {
		return service.SendError(c, 400, "Invalid body")
	}

	//Field validation
	if valid, errObj := validationProductType(c,&pt); !valid {
		return service.SendErrorStruct(c, 500, errObj)
	}
	//Update
	if _, err := pt.Update(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
		return service.SendError(c, 500, err.Error())
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    pt,
		"message": "Successfully updated product type",
	}

	return c.JSON(resp)
}

func validationProductType(c *fiber.Ctx,pt *models.ProductType) (bool, dto.ProductTypeError) {
	var errObj dto.ProductTypeError
	if pt.TypeName == "" {
		errObj.ErrTypeName = "Please input product type!"
		return false, errObj
	}
	//Check duplicate name
	exists, err := models.ProductTypes(
		qm.Where("\"typeName\" = ? AND \"productTypeID\" <> ?",pt.TypeName,pt.ProductTypeID),
	).Exists(c.Context(),boil.GetContextDB());
	if err != nil{
		errObj.ErrTypeName = "Validation failed, please try again"
		return false, errObj
	}
	if exists{
		errObj.ErrTypeName = "Product type already exists"
		return false, errObj
	}
	return true, errObj
}
