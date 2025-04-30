package handlers

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
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
	ngayFromStr := c.Query("ngayFrom","");
	ngayToStr := c.Query("ngayTo","");
	//parsing dates
	var ngayFrom,ngayTo time.Time
	var errTime error
	if ngayFromStr != ""{
		ngayFrom, errTime = time.Parse("2006-01-02",ngayFromStr);
		if errTime != nil {
			return service.SendError(c, 400, "Invalid format for ngayFrom (expect yyyy-mm-dd)")
		}
	}
	if ngayToStr != ""{
		ngayTo, errTime = time.Parse("2006-01-02",ngayToStr);
		if errTime != nil {
			return service.SendError(c, 400, "Invalid format for ngayTo (expect yyyy-mm-dd)")
		}
	}

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
			default:
				fmt.Println("Do nothing in fallback")
			}	
	}else if sort == "Review Date" && ngayFrom.Before(ngayTo){
		queryMods = append(queryMods, qm.Where("DATE(\"reviewDate\") BETWEEN ? AND ?",ngayFrom,ngayTo));
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

// Define a struct to hold the response data
type ClauseAnalysis struct {
    Clause    string `json:"clause"`
    Sentiment string `json:"sentiment"`
}

type AnalyzeResult struct {
    Review   string           `json:"review"`
    Clauses  []string         `json:"clauses"`
    Analysis []ClauseAnalysis `json:"analysis"`
    Summary  string           `json:"summary"`
}

func GetAdminReviewDetail(c *fiber.Ctx) error{
	
	//establishing connection to python backend
	client := resty.New()

	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}

	//creating redisKey following reviewID
	redisKey := fmt.Sprintf("review:reviewID=%d:",reviewID);
	//fetching redisKey
	cachedReview, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		fmt.Println("Đã lưu vào redis và trả về!");
		return c.JSON(json.RawMessage(cachedReview))
	}

	review, err := models.Reviews(qm.Where("\"reviewID\" = ?",reviewID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	listHinhDG, err := models.ReviewImages(qm.Where("\"reviewID\" = ?",reviewID)).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	reply, err := models.Replies(qm.Where("\"reviewID\" = ?",reviewID)).One(c.Context(),boil.GetContextDB());
	if err != nil && err.Error() != "sql: no rows in result set"{
		return service.SendError(c,500,err.Error())
	}

	result := []AnalyzeResult{}
	_, err = client.R().
		SetHeader("Content-Type","application/json").
		SetBody(map[string]interface{}{"review": []string{review.Comment}}).
		SetResult(&result).
		Post("http://192.168.240.1:5000/analyze")
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := fiber.Map{
		"status": "Success",
		"data": review,
		"listHinhDG": listHinhDG,
		"reply": reply,
		"result": result,
		"message": "Successfully fetched review details!",
	}

	savingKeyJson, _ := json.Marshal(response);
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,savingKeyJson,10*time.Minute);
	if rdsErr != nil{
		fmt.Println("Failed to cache review data: ",rdsErr)
	}

	return c.JSON(response);
}

func InsertReviewReply(c *fiber.Ctx) error{
	var reply models.Reply
	err := c.BodyParser(&reply);
	if err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	if reply.IsReplied{
		return service.SendError(c,500,"Already replied to this review!")
	}

	//setting isReplied to true
	reply.IsReplied = true
	err = reply.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": reply,
		"message": "Successfully replied to review!",
	}

	return c.JSON(resp);
}

func UpdateReviewReply(c *fiber.Ctx) error{
	var reply models.Reply
	err := c.BodyParser(&reply);
	if err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	replyID := c.QueryInt("replyID",0);
	if replyID == 0{
		return service.SendError(c,400,"Did not receive replyID");
	}
 
	_,err = reply.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": reply,
		"message": "Successfully updated reply!",
	}

	return c.JSON(resp);
}