package utils

import (
	"GoodFood-BE/internal/dto"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/models"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
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

func GetProductsUtil(c *fiber.Ctx, search, typeName, orderBy string, page, minPrice, maxPrice int) (models.ProductSlice, int,string, error){
	//Initialize queryMods
	queryMods := []qm.QueryMod{
		qm.Where("status = true"),
		qm.Where("price BETWEEN ? AND ?",minPrice,maxPrice),
		qm.Load(models.ProductRels.ProductTypeIDProductType),
	}

	//typeName filter logic
	var totalProduct int64;
	var err error;
	if typeName != ""{
		productType,err := models.ProductTypes(qm.Where("\"typeName\" = ?",typeName)).One(c.Context(),boil.GetContextDB())
		if err != nil {
			return nil, 0, "",fmt.Errorf("product type not found")
		}
		queryMods = append(queryMods, qm.Where( "\"productTypeID\" = ?",productType.ProductTypeID),)
	}

	//search filter logic
	if search != ""{
		queryMods = append(queryMods, qm.Where("LOWER(\"productName\") LIKE LOWER(?)","%"+search+"%"))
	}

	//Calculate totalProduct
	totalProduct,err = models.Products(
		queryMods...
	).Count(c.Context(),boil.GetContextDB());
	if err != nil {
		return nil, 0, "",fmt.Errorf("count products failed: %w", err)
	}
	totalPage := int(math.Ceil(float64(totalProduct) / float64(PageSize)))

	// Redis cache key
	redisKey := fmt.Sprintf(
		"products:page=%d:type=%s:search=%s:minPrice=%d:maxPrice=%d:orderBy=%s",
		page, typeName, search, minPrice, maxPrice, orderBy,
	)
	//Fetch redis cache
	cachedProducts,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result()
	if err == nil{
		products := models.ProductSlice{}
		if json.Unmarshal([]byte(cachedProducts),&products)==nil{
			return products, totalPage, redisKey,nil
		}
	}

	//Pagination logic
	offset := (page-1)*6;
	queryMods = append(queryMods,qm.OrderBy("price "+orderBy), qm.Limit(6), qm.Offset(offset));
	products, err := models.Products(queryMods...).All(c.Context(), boil.GetContextDB());
	if err != nil {
		return nil, 0, "",fmt.Errorf("fetch products failed: %w", err)
	}

	return products,totalPage,redisKey,nil
}

// buildProductDetail fetches product, images, reviews, and star counts from DB.
func BuildProductDetail(c *fiber.Ctx, id int, filter string, page int) (dto.ProductDetailResponse, int, error) {
	const pageSize = 3
	offset := (page - 1) * pageSize;

	//Fetch product with images
	detail, err := models.Products(qm.Where("\"productID\" = ?",id),qm.Load(models.ProductRels.ProductIDProductImages)).One(c.Context(),boil.GetContextDB());
	if err != nil {
		return dto.ProductDetailResponse{},0,err
	}
	// Build base response
	response := dto.ProductDetailResponse{
		Product:       *detail,
		ProductImages: detail.R.ProductIDProductImages,
	}

	// Fetch reviews
	reviews, reviewErr, totalPage := reviewDisplay(c, id, filter, offset)
	if reviewErr != nil {
		return dto.ProductDetailResponse{}, 0, reviewErr
	}
	response.FiveStarsReview = reviews

	// Fetch star distribution
	stars, err := countStars(c, id)
	if err != nil {
		return dto.ProductDetailResponse{}, 0, err
	}
	response.Stars = stars

	return response, totalPage, nil
}

func reviewDisplay(c *fiber.Ctx,id int, filter string, offset int) (reviews []dto.ReviewResponse,error error, totalPage int){
	queries := []qm.QueryMod{}
	queries = append(queries, qm.Where("\"productID\" = ?",id))

	if filter != "All"{
		star, err := strconv.Atoi(filter);
		if err != nil || star < 1 || star > 5{
			return nil,fmt.Errorf("invalid star filter"), 0;
		}
		queries = append(queries, qm.Where("stars = ?",star))
	}

	//calculating total page
	totalReview, err := models.Reviews(queries...).Count(c.Context(),boil.GetContextDB())
	if err != nil{
		return nil,fmt.Errorf(err.Error()),0;
	}
	totalPage = int(math.Ceil(float64(totalReview)/3));

	//loading necessary references
	queries = append(queries, 
		qm.Load(models.ReviewRels.AccountIDAccount),
		qm.Load(models.ReviewRels.ProductIDProduct),
		qm.Load(models.ReviewRels.ReviewIDReviewImages),
		qm.Load(models.ReviewRels.ReviewIDReplies),
		qm.Offset(offset),
		qm.Limit(3),
		qm.OrderBy("\"reviewID\""),
	)

	review, err := models.Reviews(queries...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return nil,err,totalPage
	}
	
	reviewResult := make([]dto.ReviewResponse,len(review));
	for i, r := range review{
		reviewResult[i] = dto.ReviewResponse{
			Review: *r,
			ReviewAccount: *r.R.AccountIDAccount,
			ReviewProduct: *r.R.ProductIDProduct,
			ReviewImages: r.R.ReviewIDReviewImages,
			ReviewReply: r.R.ReviewIDReplies,
		}
	}
	return reviewResult,nil,totalPage;
}

// countStars queries the number of reviews for each star rating (1–5).
func countStars(c *fiber.Ctx, productID int) (dto.Star, error) {
	result := dto.Star{}
	for stars := 1; stars <= 5; stars++ {
		count, err := models.Reviews(
			qm.Where("\"productID\" = ? AND stars = ?", productID, stars),
		).Count(c.Context(), boil.GetContextDB())
		if err != nil {
			return result, err
		}

		switch stars {
		case 1:
			result.OneStars = int(count)
		case 2:
			result.TwoStars = int(count)
		case 3:
			result.ThreeStars = int(count)
		case 4:
			result.FourStars = int(count)
		case 5:
			result.FiveStars = int(count)
		}
	}
	return result, nil
}
