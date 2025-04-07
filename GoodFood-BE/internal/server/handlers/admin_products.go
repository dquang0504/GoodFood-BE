package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Cards struct{
	TotalProduct int `boil:"totalproduct"`
	TotalInactive int `boil:"totalinactive"`
}
type ProductResponse struct{
	models.Product
	ProductType models.ProductType `json:"productType"`
	ProductImages []models.ProductImage `json:"productImages"`
}
func GetAdminProducts(c *fiber.Ctx) error{
	page := c.QueryInt("page",1);
	sort := c.Query("sort","");
	search := c.Query("search","");
	offset := (page - 1) * 6;

	queryMods := []qm.QueryMod{}

	//fetching cards
	var cards Cards
	err := queries.Raw(`
		SELECT COALESCE(COUNT(*),0) AS totalproduct,
		COUNT(CASE WHEN status = false THEN 1 END) AS totalinactive
		FROM product
	`).Bind(c.Context(),boil.GetContextDB(),&cards);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//fetching product types
	listLoaiSP, err := models.ProductTypes().All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//search and sort logic
	if search != ""{
		fmt.Println("Hello mister")
		switch sort{
			case "Type":
				fetchType, err := models.ProductTypes(qm.Where("\"typeName\" ILIKE ?","%"+search+"%")).One(c.Context(),boil.GetContextDB());
				if err != nil{
					return service.SendError(c,500,err.Error());
				}
				queryMods = append(queryMods, qm.Where("\"productTypeID\" = ?",fetchType.ProductTypeID))
			case "Product Name":
				queryMods = append(queryMods, qm.Where("\"productName\" ILIKE ?","%"+search+"%"))
			case "Weight":
				queryMods = append(queryMods, qm.Where("CAST(weight AS TEXT) ILIKE ?","%"+search+"%"))
			case "Low to high price":
				queryMods = append(queryMods, qm.OrderBy("price ASC"))
				fmt.Println("Hello bro")
			case "High to low price":
				queryMods = append(queryMods, qm.OrderBy("price DESC"))
			case "Active Status":
				queryMods = append(queryMods, qm.Where("status = ?",true))
			case "Inactive Status":
				queryMods = append(queryMods, qm.Where("status = ?",false))
		}
	}
	//fetching total products and calculating total page
	switch sort{
		case "Active Status":
			queryMods = append(queryMods, qm.Where("status = ?",true))
		case "Inactive Status":
			queryMods = append(queryMods, qm.Where("status = ?",false))
		}
	totalProducts, err := models.Products(queryMods...).Count(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	totalPage := int(math.Ceil(float64(totalProducts)/6));

	//appending to query mods and fetching product list
	switch sort{
		case "Low to high price":
			queryMods = append(queryMods, qm.OrderBy("price ASC"))
		case "High to low price":
			queryMods = append(queryMods, qm.OrderBy("price DESC"))
		default: 
			fmt.Println("wrong case");

	}
	queryMods = append(queryMods, qm.OrderBy("\"productID\" DESC"), qm.Limit(6), qm.Offset(offset), qm.Load(models.ProductRels.ProductTypeIDProductType));
	products, err := models.Products(queryMods...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := make([]ProductResponse,len(products))
	for i,p := range products{
		response[i] = ProductResponse{
			Product: *p,
			ProductType: *p.R.ProductTypeIDProductType,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"cards": cards,
		"listLoaiSP": listLoaiSP,
		"totalPage": totalPage,
		"message": "Successfully fetched products values",
	}

	return c.JSON(resp);
}

func GetAdminProductDetail(c *fiber.Ctx) error{
	productID := c.QueryInt("productID",0);
	if productID == 0{
		return service.SendError(c,400,"Did not receive productID!");
	}

	product, err := models.Products(qm.Where("\"productID\" = ?",productID), qm.Load(models.ProductRels.ProductTypeIDProductType),qm.Load(models.ProductRels.ProductIDProductImages)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	listProductImgs, err := models.ProductImages(qm.Where("\"productID\" = ?",productID)).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := ProductResponse{
		Product:      *product,
		ProductType:  *product.R.ProductTypeIDProductType,
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"listHinhSP": listProductImgs,
		"message": "Successfully fetched products detail",
	}

	return c.JSON(resp);

}

type ProductError struct{
	ErrProductName string `json:"errProductName"`
	ErrPrice string `json:"errPrice"`
	ErrWeight string `json:"errWeight"`
	ErrType string `json:"errType"`
	ErrImages string `json:"errImages"`
}

func AdminProductCreate(c *fiber.Ctx) error{
	var insert ProductResponse
	if err := c.BodyParser(&insert); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	if valid, errObj := validationProduct(&insert, true); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	err := insert.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	var productImages = insert.ProductImages
	for i := range productImages{
		productImages[i].ProductID = insert.ProductID
		fmt.Println(productImages[i]);
		err = productImages[i].Insert(c.Context(),boil.GetContextDB(),boil.Infer());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": insert,
		"message": "Successfully created new product",
	}

	return c.JSON(resp);
}

func AdminProductUpdate(c *fiber.Ctx) error{
	var update ProductResponse
	productID := c.QueryInt("productID",0);
	if productID == 0{
		return service.SendError(c,400,"Did not receive productID");
	}

	if err := c.BodyParser(&update); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	var images = update.ProductImages;

	if valid, errObj := validationProduct(&update, false); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	_,err := update.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	if len(images) > 0{
		//checking if new images are uploaded then delete every image of the product
		deleteImgs,err := models.ProductImages(qm.Where("\"productID\" = ?",update.ProductID)).All(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		_,err = deleteImgs.DeleteAll(c.Context(),boil.GetContextDB())
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		//iterate through images and start inserting one by one image
		for i := range images{
			images[i].ProductID = update.ProductID;
			if err := images[i].Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
				return service.SendError(c,500,err.Error());
			}
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": update,
		"message": "Successfully updated the product",
	}

	return c.JSON(resp);
}

func validationProduct(product *ProductResponse, diff bool) (bool,ProductError){
	var error ProductError
	isValid := true
	if product.ProductName == ""{
		error.ErrProductName = "Please input product name!"
		isValid = false
	}
	if product.Price < 0{
		error.ErrPrice = "Price can't be lower than 0!"
		isValid = false
	}
	if product.Weight < 0{
		error.ErrWeight = "Weight can't be lower than 0!"
		isValid = false
	}
	if product.ProductType.TypeName == "" {
		error.ErrType = "Please choose product type!"
		isValid = false
	}
	if len(product.ProductImages) == 0 && diff{
		error.ErrImages = "Please upload the product's image!"
		isValid = false
	}
	return isValid,error
	
}