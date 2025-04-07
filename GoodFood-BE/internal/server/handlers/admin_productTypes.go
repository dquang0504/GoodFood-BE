package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"context"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ProductTypesResponse struct{
	models.ProductType
	TotalProduct int `boil:"totalproduct"`
}
func GetAdminProductTypes(c *fiber.Ctx) error{
	page := c.QueryInt("page",1);
	search := c.Query("search","");
	offset := (page-1)*6;

	queryMods := []qm.QueryMod{}

	if search != ""{
		queryMods = append(queryMods, qm.Where("\"typeName\" ILIKE ?","%"+search+"%"))
	}
	totalTypes, err := models.ProductTypes(queryMods...).Count(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	totalPage := int(math.Ceil(float64(totalTypes)/float64(6)));

	queryMods = append(queryMods, qm.OrderBy("\"productTypeID\" DESC"), qm.Limit(6), qm.Offset(offset), qm.Load(models.ProductTypeRels.ProductTypeIDProducts));
	productTypes,err := models.ProductTypes(queryMods...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Iterate through the whole list to count the total products of each product types
	response := make([]ProductTypesResponse,len(productTypes))
	for i, pt := range productTypes{
		totalProduct := 0
		if pt.R != nil && pt.R.ProductTypeIDProducts != nil{
			totalProduct = len(pt.R.ProductTypeIDProducts)
		}
		response[i] = ProductTypesResponse{
			ProductType: *pt,
			TotalProduct: totalProduct,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"totalPage": totalPage,
		"message": "Successfully fetched product types values",
	}

	return c.JSON(resp);
}

func GetAdminProductTypeDetail(c *fiber.Ctx) error{
	typeID := c.QueryInt("typeID",0);
	if typeID == 0{
		return service.SendError(c,400,"Did not receive typeID");
	}

	pt, err := models.ProductTypes(qm.Where("\"productTypeID\" = ?",typeID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": pt,
		"message": "Successfully fetched product types detail",
	}

	return c.JSON(resp);
}

type ProductTypeError struct{
	ErrTypeName string `json:"errTypeName"`
}

func AdminProductTypeCreate(c *fiber.Ctx) error{
	var pt models.ProductType
	if err := c.BodyParser(&pt); err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	if valid, errObj := validationProductType(&pt); !valid{
		return service.SendErrorStruct(c,500,errObj);
	}

	err := pt.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": pt,
		"message": "Successfully created new product type",
	}

	return c.JSON(resp);
}

func AdminProductTypeUpdate(c *fiber.Ctx) error{
	var pt models.ProductType
	if err := c.BodyParser(&pt); err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	if valid, errObj := validationProductType(&pt); !valid{
		return service.SendErrorStruct(c,500,errObj);
	}

	_,err := pt.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": pt,
		"message": "Successfully updated product type",
	}

	return c.JSON(resp);
}

func validationProductType(pt *models.ProductType) (bool,ProductTypeError){
	var error ProductTypeError
	isValid := true
	if pt.TypeName == ""{
		error.ErrTypeName = "Please input product type!"
		isValid = false;
	}else if res,err := models.ProductTypes(qm.Where("\"typeName\" = ?",pt.TypeName)).Exists(context.Background(),boil.GetContextDB()); err == nil{
		if(res){
			error.ErrTypeName = "Product type already exists!"
			isValid = false;
		}
	}
	return isValid,error;
}