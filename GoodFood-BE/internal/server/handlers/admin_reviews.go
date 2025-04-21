package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ReviewCards struct{
	TotalReview int `boil:"total_review"`
	Total5S int `boil:"total_5s"`
}
type ReviewResponse struct{
	models.Review
	ReviewAccount models.Account `json:"reviewAccount"`
	ReviewProduct models.Product `json:"reviewProduct"`
}
func GetAdminReview(c *fiber.Ctx) error{
	var cards ReviewCards
	err := queries.Raw(`
		SELECT COALESCE(COUNT(*),0) AS total_review,
		COUNT(CASE WHEN stars = 5 THEN 1 END) AS total_5s
		FROM review
	`).Bind(c.Context(),boil.GetContextDB(),&cards);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	sort := c.Query("sort","Tên sản phẩm");
	search := c.Query("search","");
	offset := (page-1)*6;
	// ngayFromStr := c.Query("ngayFrom","");
	// ngayToStr := c.Query("ngayTo","");
	//parsing dates
	// var ngayFrom,ngayTo time.Time
	// var errTime error
	// if ngayFromStr != ""{
	// 	ngayFrom, errTime := time.Parse("2006-01-02",ngayFromStr);
	// }
	// if ngayToStr != ""{
	// 	ngayTo, errTime := time.Parse("2006-01-02",ngayToStr);
	// }

	queryMods := []qm.QueryMod{}

	if search != ""{
		switch sort{
			case "Product Name":
				//could return multiple products with a similar name
				product, err := models.Products(qm.Where("\"productName\" ILIKE ?","%"+search+"%")).All(c.Context(),boil.GetContextDB());
				if err != nil{
					return service.SendError(c,500,err.Error());
				}
				//appending the product ids into an int slice
				ids := []int{}
				for _, p := range product{
					ids = append(ids, p.ProductID)
				}
				if len(ids) > 0 {
					fmt.Println(ids)
					queryMods = append(queryMods, qm.WhereIn("\"productID\" in ?", convertIntSliceToInterface(ids)...))
				}
			case "Stars":
				intSearch, err := strconv.Atoi(search)
				if err != nil{
					return service.SendError(c,400, err.Error())
				}
				queryMods = append(queryMods, qm.Where("stars = ?",intSearch))
			case "Comment":
				queryMods = append(queryMods, qm.Where("\"comment\" ILIKE ?","%"+search+"%"))
			}	
				
	}

	totalReviews, err := models.Reviews(queryMods...).Count(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	totalPage := int(math.Ceil(float64(totalReviews)/6))

	queryMods = append(queryMods, qm.Limit(6), qm.Offset(offset), qm.OrderBy("\"reviewID\" DESC"),qm.Load(models.ReviewRels.AccountIDAccount), qm.Load(models.ReviewRels.ProductIDProduct))
	reviews, err := models.Reviews(queryMods...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	response := make([]ReviewResponse,len(reviews));
	for i, r := range reviews{
		response[i] = ReviewResponse{
			Review: *r,
			ReviewAccount: *r.R.AccountIDAccount,
			ReviewProduct: *r.R.ProductIDProduct,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"cards": cards,
		"totalPage": totalPage, 
		"message": "Successfully fetched review values!",
	}

	return c.JSON(resp);
}

func convertIntSliceToInterface(s []int) []interface{}{
	result := make([]interface{},len(s))
	for i, v := range s{
		result[i] = v
	}
	return result
}

func GetAdminReviewDetail(c *fiber.Ctx) error{

	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}
	review, err := models.Reviews(qm.Where("\"reviewID\" = ?",reviewID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": review,
		"message": "Successfully fetched review details!",
	}

	return c.JSON(resp);
}