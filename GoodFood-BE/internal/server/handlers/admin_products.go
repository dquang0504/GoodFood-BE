package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

//GetAdminProductsData fetches products with pagination, sorting, searching.
func GetAdminProducts(c *fiber.Ctx) error{
	page := c.QueryInt("page",1);
	sort := c.Query("sort","");
	search := c.Query("search","");

	query := `SELECT COALESCE(COUNT(*),0) AS totalproduct,
		COUNT(CASE WHEN status = false THEN 1 END) AS totalinactive
		FROM product
	`
	cards, err := utils.FetchCards(c,query,&dto.ProductCards{})
	products, listLoaiSP, totalPage, err := utils.GetAdminProductsUtil(c,page,search,sort);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"cards": cards,
		"listLoaiSP": listLoaiSP,
		"totalPage": totalPage,
		"message": "Successfully fetched products values",
	}

	return c.JSON(resp);
}

//GetAdminProductDetail returns detailed info of a product including type and images.
func GetAdminProductDetail(c *fiber.Ctx) error{
	productID := c.QueryInt("productID",0);
	if productID == 0{
		return service.SendError(c,400,"Did not receive productID!");
	}

	//Fetch product with relations (type + images)
	product, err := models.Products(qm.Where("\"productID\" = ?",productID), qm.Load(models.ProductRels.ProductTypeIDProductType),qm.Load(models.ProductRels.ProductIDProductImages)).One(c.Context(),boil.GetContextDB());
	if err != nil && err != sql.ErrNoRows{
		return service.SendError(c,500,err.Error());
	}

	//Build response
	response := dto.ProductResponse{
		Product:      *product,
		ProductType:  *product.R.ProductTypeIDProductType,
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"listHinhSP": product.R.ProductIDProductImages,
		"message": "Successfully fetched products detail",
	}

	return c.JSON(resp);
}

//AdminProductCreate creates a new product record in table Product
//Returns insert information
func AdminProductCreate(c *fiber.Ctx) error{
	var insert dto.ProductResponse
	if err := c.BodyParser(&insert); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	//Validate input before inserting
	if valid, errObj := validationProduct(&insert, true); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	//Use transaction for safety
	db, ok := boil.GetContextDB().(*sql.DB);
	if !ok {
		return service.SendError(c, 500, "Invalid DB instance")
	}
	tx, err := db.BeginTx(c.Context(),nil);
	if err != nil{
		return service.SendError(c,500,"Failed to start transaction")
	}
	defer tx.Rollback() //rollback if anything fails

	//Insert product
	if err := insert.Product.Insert(c.Context(),tx,boil.Infer());err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Insert product images
	for i := range insert.ProductImages{
		insert.ProductImages[i].ProductID = insert.ProductID
		if err = insert.ProductImages[i].Insert(c.Context(),tx,boil.Infer()); err != nil{
			return service.SendError(c,500,err.Error()+"over here productImgs");
		}
	}

	//Commit transaction if everything succeeds
	if err := tx.Commit(); err != nil{
		return service.SendError(c,500,"Failed to commit transaction")
	}

	//Clear all redis keys related to products
	utils.ClearRedisByPattern("products:page=*:type=*:search=*:minPrice=*:maxPrice=*:orderByPrice=*");
	//Clear all redis keys related to product detail
	utils.ClearRedisByPattern(fmt.Sprintf("product:detail:%d:filter=*:page=*",insert.ProductID))

	resp := fiber.Map{
		"status": "Success",
		"data": insert,
		"message": "Successfully created new product",
	}

	return c.JSON(resp);
}

//AdminProductUpdate updates an existing product record in table Product
//Returns insert information
func AdminProductUpdate(c *fiber.Ctx) error{
	//Fetch payload and query params
	update := dto.ProductResponse{}
	productID := c.QueryInt("productID",0);
	if productID == 0{
		return service.SendError(c,400,"Did not receive productID");
	}
	if err := c.BodyParser(&update); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	//Validate input
	if valid, errObj := validationProduct(&update, false); !valid{
		return service.SendErrorStruct(c,400,errObj);
	}

	//Fetch DB connection
	db, ok := boil.GetContextDB().(*sql.DB);
	if !ok {
		return service.SendError(c,500,"Invalid DB instance")
	}

	//Commence the transaction
	tx, err := db.BeginTx(c.Context(),nil)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	defer tx.Rollback();

	//Check if product previously had a coverImg
	product, err := models.Products(qm.Where("\"productID\" = ?",productID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	if product.CoverImage != ""{
		update.CoverImage = product.CoverImage
	}

	//Update product
	_,err = update.Update(c.Context(),tx,boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	//If there are new images uploaded, replace old images with new ones
	if len(update.ProductImages) > 0{
		//Delete old images of the product
		oldImgs,err := models.ProductImages(qm.Where("\"productID\" = ?",update.ProductID)).All(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		if _,err = oldImgs.DeleteAll(c.Context(),tx); err != nil{
			return service.SendError(c,500,err.Error());
		}
		//Iterate through new images and start inserting
		for i := range update.ProductImages{
			update.ProductImages[i].ProductID = update.ProductID;
			if err := update.ProductImages[i].Insert(c.Context(),tx,boil.Infer()); err != nil{
				return service.SendError(c,500,err.Error());
			}
		}
	}

	//Commit transaction
	if err := tx.Commit(); err != nil{
		return service.SendError(c,500, err.Error())
	}

	//Clear all related redis cache keys related to products
	utils.ClearRedisByPattern("products:page=*:type=*:search=*:minPrice=*:maxPrice=*:orderByPrice=*")
	utils.ClearRedisByPattern(fmt.Sprintf("product:detail:%d:filter=*:page=*",update.ProductID))

	resp := fiber.Map{
		"status": "Success",
		"data": update,
		"message": "Successfully updated the product",
	}

	return c.JSON(resp);
}

func validationProduct(product *dto.ProductResponse, diff bool) (bool,dto.ProductError){
	var errResp dto.ProductError
	isValid := true
	if product.ProductName == ""{
		errResp.ErrProductName = "Please input product name!"
		isValid = false
	}
	if product.Price < 0{
		errResp.ErrPrice = "Price can't be lower than 0!"
		isValid = false
	}
	if product.Weight < 0{
		errResp.ErrWeight = "Weight can't be lower than 0!"
		isValid = false
	}
	if product.ProductType.TypeName == "" {
		errResp.ErrType = "Please choose product type!"
		isValid = false
	}
	if len(product.ProductImages) == 0 && diff{
		errResp.ErrImages = "Please upload the product's image!"
		isValid = false
	}
	return isValid,errResp
}