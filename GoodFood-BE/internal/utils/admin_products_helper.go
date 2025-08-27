package utils

import (
	"GoodFood-BE/internal/dto"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/models"
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// These constants define all available sort filters.
const (
	SortType           = "Type"
	SortProductName    = "Product Name"
	SortWeight         = "Weight"
	SortLowToHigh      = "Low to high price"
	SortHighToLow      = "High to low price"
	SortActiveStatus   = "Active Status"
	SortInactiveStatus = "Inactive Status"
	PageSize           = 6
)

// GetAdminProductsUtil fetches paginated product list for admin panel.
func GetAdminProductsUtil(c *fiber.Ctx, page int, search string, sort string) ([]dto.ProductResponse, dto.ProductCards, []*models.ProductType, int, error) {
	offset := (page - 1) * PageSize
	queryMods := []qm.QueryMod{}

	//Fetch product cards (total + inactive)
	cards := dto.ProductCards{}
	err := queries.Raw(`
		SELECT COALESCE(COUNT(*),0) AS totalproduct,
		COUNT(CASE WHEN status = false THEN 1 END) AS totalinactive
		FROM product
	`).Bind(c.Context(), boil.GetContextDB(), &cards)
	if err != nil {
		return nil, dto.ProductCards{}, nil, 0, err
	}

	//Fetch product types
	listLoaiSP, err := models.ProductTypes().All(c.Context(), boil.GetContextDB())
	if err != nil {
		return nil, dto.ProductCards{}, nil, 0, err
	}

	//Build query conditions (search and sort)
	switch sort {
	case SortType:
		if search != "" {
			fetchType, err := models.ProductTypes(qm.Where("\"typeName\" ILIKE ?", search)).One(c.Context(), boil.GetContextDB())
			if err != nil {
				return nil, dto.ProductCards{}, nil, 0, err
			}
			queryMods = append(queryMods, qm.Where("\"productTypeID\" = ?", fetchType.ProductTypeID))
		}
	case SortProductName:
		if search != "" {
			queryMods = append(queryMods, qm.Where("\"productName\" ILIKE ?", "%"+search+"%"))
		}
	case SortWeight:
		if search != "" {
			queryMods = append(queryMods, qm.Where("CAST(weight AS TEXT) ILIKE ?", "%"+search+"%"))
		}
	case SortActiveStatus:
		queryMods = append(queryMods, qm.Where("status = ?", true))
	case SortInactiveStatus:
		queryMods = append(queryMods, qm.Where("status = ?", false))
	}

	//Total products and total pages
	totalProducts, err := models.Products(queryMods...).Count(c.Context(), boil.GetContextDB())
	if err != nil {
		return nil, dto.ProductCards{}, nil, 0, err
	}
	totalPage := int(math.Ceil(float64(totalProducts) / 6))

	//order by price
	switch sort {
	case SortLowToHigh:
		queryMods = append(queryMods, qm.OrderBy("price ASC"))
	case SortHighToLow:
		queryMods = append(queryMods, qm.OrderBy("price DESC"))
	}

	//Fetch products with pagination
	queryMods = append(queryMods, qm.OrderBy("\"productID\" DESC"), qm.Limit(6), qm.Offset(offset), qm.Load(models.ProductRels.ProductTypeIDProductType))
	products, err := models.Products(queryMods...).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return nil, dto.ProductCards{}, nil, 0, err
	}

	//Map to response
	response := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		response[i] = dto.ProductResponse{
			Product:     *p,
			ProductType: *p.R.ProductTypeIDProductType,
		}
	}

	return response, cards, listLoaiSP, totalPage, nil
}

//ClearRedisByPattern deletes Redis keys matching a pattern
func ClearRedisByPattern(pattern string){
	iter := redisdatabase.Client.Scan(redisdatabase.Ctx, 0, pattern, 0).Iterator()
	for iter.Next(redisdatabase.Ctx){
		key := iter.Val()
		if err := redisdatabase.Client.Del(redisdatabase.Ctx,key).Err(); err != nil{
			fmt.Printf("failed to delete keys products %s: %v\n",key,err)
		}
	}
	if err := iter.Err(); err != nil{
		fmt.Printf("Error during Redis scan: %v\n", err);
	}
}
