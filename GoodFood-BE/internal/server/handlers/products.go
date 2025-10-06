package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	// "golang.org/x/text/number"
)

// GetFour function fetches 4 product with types to display at front-end
// Returns a slice of 4
func GetFour(c *fiber.Ctx) error {
	// Fetch 4 products to display at Home.tsx
	products, err := models.Products(qm.Where("status = true"), qm.Limit(4), qm.Load(models.ProductRels.ProductTypeIDProductType)).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Build response
	response := make([]dto.GetFourStruct, len(products))
	for i, product := range products {
		response[i] = dto.GetFourStruct{
			Product:     *product,
			ProductType: product.R.ProductTypeIDProductType,
		}
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    response,
		"message": "Successfully fetched featuring items",
	}

	return c.JSON(resp)
}

// GetTypes function fetches all product types to display at Product.tsx
func GetTypes(c *fiber.Ctx) error {
	types, err := models.ProductTypes(qm.Where("\"status\" = true")).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, "Failed to fetch product types!")
	}
	resp := fiber.Map{
		"status":  "Success",
		"data":    types,
		"message": "Successfully fetched product types",
	}
	return c.JSON(resp)
}

func GetProductsByPage(c *fiber.Ctx) error {

	//Fetch query params
	page := c.QueryInt("page", 0)
	if page == 0 {
		return service.SendError(c, 400, "Did not receive pageNum")
	}
	typeName := c.Query("type", "")
	search := c.Query("search", "")
	minPrice := c.QueryInt("minPrice", 0)
	maxPrice := c.QueryInt("maxPrice", 0)
	orderBy := c.Query("orderBy", "ASC")

	// Redis cache key
	redisKey := fmt.Sprintf(
		"products:page=%d:type=%s:search=%s:minPrice=%d:maxPrice=%d:orderBy=%s",
		page, typeName, search, minPrice, maxPrice, orderBy,
	)
	//Fetch redis cache
	cachedProducts := fiber.Map{}
	ok, _ := utils.GetCache(redisKey, &cachedProducts)
	if ok {
		return c.JSON(cachedProducts)
	}

	//Filters and pagination logic
	products, totalPage, err := utils.GetProductsUtil(c, search, typeName, orderBy, page, minPrice, maxPrice)
	if err != nil {
		println(err.Error())
		return service.SendError(c, 500, "Failed to fetch products by page")
	}

	resp := fiber.Map{
		"status":    "Success",
		"data":      products,
		"totalPage": totalPage,
		"message":   "Successfully fetched products by page",
	}

	//saving redis key to redis database for 10 mins
	utils.SetCache(redisKey, resp, 10*time.Minute, "products:keys")

	return c.JSON(resp)
}

func ClassifyImage(c *fiber.Ctx) error {
	//Initialize resty
	client := resty.New()

	//Lấy dữ liệu ảnh từ request
	file, err := c.FormFile("image")
	if err != nil {
		return service.SendError(c, 400, "Invalid request format")
	}

	// Mở file ảnh
	fileContent, err := file.Open()
	if err != nil {
		return service.SendError(c, 500, "Failed to open uploaded image")
	}
	defer fileContent.Close()

	var result dto.PredictResult
	_, err = client.R().
		SetFileReader("file", file.Filename, fileContent).
		SetResult(&result).
		Post("http://192.168.1.10:5000/callModel")
	if err != nil {
		return service.SendError(c, 500, "Python microservice unavailable!")
	}

	//Trả về kết quả
	return c.JSON(fiber.Map{
		"status":  "Success",
		"message": "Image classified successfully",
		"data":    result,
	})

}

// GetDetail handles fetching product details including reviews, images, and star counts.
// It also applies caching using Redis for performance optimization.
func GetDetail(c *fiber.Ctx) error {
	//Fetch query params
	id := c.QueryInt("id", 0)
	if id == 0 {
		return service.SendError(c, 400, "ID not found")
	}
	filter := c.Query("filter", "All")
	page := c.QueryInt("page", 1)

	//Redis key
	redisKey := fmt.Sprintf("product:detail:%d:filter=%s:page=%d", id, filter, page)
	//Fetch redis cache
	cachedDetail := fiber.Map{}
	if ok, _ := utils.GetCache(redisKey, &cachedDetail); ok {
		return c.JSON(cachedDetail)
	}

	// Build product detail response
	detailedResponse, totalPage, err := utils.BuildProductDetail(c, id, filter, page)
	if err != nil {
		return service.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	resp := fiber.Map{
		"status":    "Success",
		"data":      detailedResponse,
		"totalPage": totalPage,
		"message":   "Successfully fetched detailed product!",
	}

	//Saving redis cache for 30 mins
	redisSetKey := fmt.Sprintf("product:detail:%d:keys", id)
	utils.SetCache(redisKey, resp, 30*time.Minute, redisSetKey)

	return c.JSON(resp)
}

// GetSimilar function fetches products that share the same typeID with one specified product.
func GetSimilar(c *fiber.Ctx) error {
	productID := c.QueryInt("id", 0)
	if productID == 0 {
		return service.SendError(c, 400, "Did not receive ID!")
	}

	typeID := c.QueryInt("typeID", 0)
	if typeID == 0 {
		return service.SendError(c, 400, "Did not receive typeID!")
	}

	//Fetching typeName from typeID
	similars, err := models.Products(qm.Where("\"productID\" != ? AND \"productTypeID\" = ?", productID, typeID)).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, "ID not found!")
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    similars,
		"message": "Successfully fetched similar products",
	}

	return c.JSON(resp)
}
