package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetFour(c *fiber.Ctx) error {
    // Tạo context từ Fiber
	ctx := c.Context()

	// Truy vấn danh sách sản phẩm
	products, err := models.Products(qm.Limit(4)).All(ctx, boil.GetContextDB())
	if err != nil {
		return service.SendError(c,500,"Faield to fetch products")
	}

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"message": "Successfully fetched featuring items",
	}

	// Trả về danh sách sản phẩm dưới dạng JSON
	return c.JSON(resp)
}

func GetTypes(c *fiber.Ctx) error{
	types,err := models.ProductTypes().All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Failed to fetch product types!")
	}
	resp := fiber.Map{
		"status":"Success",
		"data": types,
		"message": "Successfully fetched product types",
	}
	return c.JSON(resp);
}

func GetProductsByPage(c *fiber.Ctx) error{
	boil.DebugMode = true
	page,err := strconv.Atoi(c.Query("page","1"));
	if err != nil{
		return service.SendError(c,500,"Error converting pageNum");
	}

	typeName := c.Query("type","");

	offset := (page-1)*6;

	queryMods := []qm.QueryMod{
		qm.Limit(6),
		qm.Offset(offset),
		qm.Load(models.ProductRels.ProductTypeIDProductType),
	}

	if typeName != ""{
		productType,err := models.ProductTypes(qm.Where("\"typeName\" = ?",typeName)).One(c.Context(),boil.GetContextDB())
		if err != nil {
			return service.SendError(c, 500, "Product type not found")
		}
		queryMods = append(queryMods, qm.Where( "\"productTypeID\" = ?",productType.ProductTypeID))
	}

	products, err := models.Products(queryMods...).All(c.Context(), boil.GetContextDB())
	
	if err != nil {
		println(err.Error())
		return service.SendError(c, 500, "Failed to fetch products by page")
	}

	totalProduct := len(products);

	totalPage := int(math.Ceil(float64(totalProduct) / float64(6)))

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"totalPage": totalPage,
		"message": "Successfully fetched products by page",
	}
	return c.JSON(resp);
}