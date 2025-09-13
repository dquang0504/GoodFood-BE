package handlers

import (
	"GoodFood-BE/internal/dto"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

//GetAdminReview fetches paginated list of reviews from DB, also sending the list over to microservice flask for sentiment analysis then cache into redis.
//Returns the paginated list along with sentiment analysis results.
func GetAdminReview(c *fiber.Ctx) error{

	//Fetch & parse query params
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	sort := c.Query("sort","Product Name");
	search := c.Query("search","");
	offset := (page-1) * utils.PageSize;
	dateFrom, dateTo, err := utils.ParseDateRange(c.Query("dateFrom", ""), c.Query("dateTo", ""));
	if err != nil{
		return service.SendError(c,400,err.Error());
	}

	//Redis cache
	redisKey := fmt.Sprintf("review:list:page=%d:sort=%s:search=%s:ngayFrom=%s:ngayTo=%s",page,sort,search,dateFrom,dateTo)
	//Fetch redis key
	cachedReview, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		return c.JSON(json.RawMessage(cachedReview))
	}

	// Fetch cards
	query := `SELECT COALESCE(COUNT(*),0) AS total_review,
			   COUNT(CASE WHEN stars = 5 THEN 1 END) AS total_5s
		FROM review
	`
	cards, err := utils.FetchCards(c,query,&dto.ReviewCards{})
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Build filters
	queryMods, err := utils.BuildReviewFilters(c, sort, search, dateFrom, dateTo)
	if err != nil {
		return service.SendError(c, 400, err.Error())
	}

	//Count & paginate
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

	//Analyze reviews
	comments := make([]string, len(reviews))
	reviewIDs := make([]int, len(reviews))
	for i, r := range reviews {
		comments[i] = r.Comment
		reviewIDs[i] = r.ReviewID
	}
	analysisResult, err := utils.AnalyzeReviews(comments, reviewIDs)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Build response
	response := make([]dto.ReviewResponse,len(reviews));
	for i, r := range reviews{
		response[i] = dto.ReviewResponse{
			Review: *r,
			ReviewAccount: *r.R.AccountIDAccount,
			ReviewProduct: *r.R.ProductIDProduct,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"result": analysisResult,
		"cards": cards,
		"totalPage": totalPage, 
		"message": "Successfully fetched review values!",
	}

	//Cache response
	savingKeyJson, _ := json.Marshal(resp);
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,savingKeyJson, 10 * 24 * time.Hour)
	if rdsErr != nil{
		fmt.Println("Failed to cache review data: ",rdsErr)
	}

	return c.JSON(resp);
}

func GetAdminReviewAnalysis(c *fiber.Ctx) error{
	//Fetch query params
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	sort := c.Query("sort","Positive Sentiment");

	//Redis cache
	redisKey := fmt.Sprintf("reviewAnalysis:list:page=%d:sort=%s",page,sort)
	//Fetch redis key
	cachedReviewAnalysis, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		return c.JSON(json.RawMessage(cachedReviewAnalysis))
	}

	//Fetch all reviews to send to python to analyze
	reviews, err := models.Reviews().All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	//Send review list to python backend for sentiment analysis
	comments := make([]string, len(reviews))
	reviewIDs := make([]int, len(reviews))
	for i, r := range reviews {
		comments[i] = r.Comment
		reviewIDs[i] = r.ReviewID
	}
	analysisResult, err := utils.AnalyzeReviews(comments, reviewIDs)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	sortingResult := []dto.AnalyzeResult{}

	//Sentiment filter
	switch sort{
		case "Positive Sentiment":
			sortingResult = utils.AppendSortingResult(analysisResult,sort);
		case "Negative Sentiment":
			sortingResult = utils.AppendSortingResult(analysisResult,sort);
		case "Neutral Sentiment":
			sortingResult = utils.AppendSortingResult(analysisResult,sort);
		case "Mixed Sentiment":
			sortingResult = utils.AppendSortingResult(analysisResult,sort);
		default:
			fmt.Println("Do nothing in fallback")
	}
	
	//pagination
	total := len(sortingResult);
	totalPage := int(math.Ceil(float64(total)/6))

	startIndex := (page - 1) * 6
	endIndex := startIndex + 6
	if endIndex > total{
		endIndex = total
	}
	pagedResult := sortingResult[startIndex:endIndex]

	//Build response
	resp := fiber.Map{
		"status": "Success",
		"result": pagedResult,
		"page": page,
		"totalPage": totalPage,
		"message": "Successfully fetched review values!",
	}

	//Cache response
	savingKeyJson, _ := json.Marshal(resp);
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,savingKeyJson, 10 * 24 * time.Hour)
	if rdsErr != nil{
		fmt.Println("Failed to cache cart data: ",rdsErr)
	}

	return c.JSON(resp);
}

//GetAdminReviewDetail fetches the specified review and send it to microservice flask to analyze
//Returns the review along with the sentiment analysis.
func GetAdminReviewDetail(c *fiber.Ctx) error{
	//Fetch query param
	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}

	//Redis cache
	redisKey := fmt.Sprintf("review:reviewID=%d:",reviewID);
	//Fetch redisKey
	cachedReview, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		return c.JSON(json.RawMessage(cachedReview))
	}

	//Fetch data concurrently
	review,reviewImages,reply,err := utils.FetchReviewData(c,reviewID);
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Send review details to python backend for sentiment analysis
	comments := []string{review.Comment}
	reviewIDs := []int{review.ReviewID}

	analysisResult, err := utils.AnalyzeReviews(comments, reviewIDs)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Build response
	response := fiber.Map{
		"status": "Success",
		"data": review,
		"listHinhDG": reviewImages,
		"reply": reply,
		"result": analysisResult,
		"message": "Successfully fetched review details!",
	}

	//Cache response
	savingKeyJson, _ := json.Marshal(response);
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,savingKeyJson,10*24*time.Hour);
	if rdsErr != nil{
		fmt.Println("Failed to cache review data: ",rdsErr)
	}

	return c.JSON(response);
}

//InsertReviewReply inserts a reply to a specified review.
//Returns the inserted reply
func InsertReviewReply(c *fiber.Ctx) error{
	var reply models.Reply
	if err := c.BodyParser(&reply); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	if reply.IsReplied{
		return service.SendError(c,500,"Already replied to this review!")
	}

	//setting isReplied to true
	reply.IsReplied = true
	if err := reply.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": reply,
		"message": "Successfully replied to review!",
	}

	return c.JSON(resp);
}

//UpdateReviewReply updates an existing reply
//Returns the updated reply
func UpdateReviewReply(c *fiber.Ctx) error{
	var reply models.Reply
	if err := c.BodyParser(&reply); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	replyID := c.QueryInt("replyID",0);
	if replyID == 0{
		return service.SendError(c,400,"Did not receive replyID");
	}
 
	if _,err := reply.Update(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": reply,
		"message": "Successfully updated reply!",
	}

	return c.JSON(resp);
}