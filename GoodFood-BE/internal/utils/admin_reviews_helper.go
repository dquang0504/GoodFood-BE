package utils

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/models"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// BuildReviewFilters constructs query mods based on search & sort params
func BuildReviewFilters(c *fiber.Ctx,sort,search string, from, to time.Time) ([]qm.QueryMod, error){
	queryMods := []qm.QueryMod{}

	switch sort {
	case "Product Name":
		products, err := models.Products(qm.Where("\"productName\" ILIKE ?", "%"+search+"%")).All(c.Context(),boil.GetContextDB())
		if err != nil {
			return nil, err
		}
		if len(products) > 0 {
			ids := make([]int, len(products))
			for i, p := range products {
				ids[i] = p.ProductID
			}
			queryMods = append(queryMods, qm.WhereIn("\"productID\" in ?", convertIntSliceToInterface(ids)...))
		}

	case "Stars":
		stars, err := strconv.Atoi(search)
		if err != nil {
			return nil, fmt.Errorf("invalid stars value: %v", err)
		}
		queryMods = append(queryMods, qm.Where("stars = ?", stars))

	case "Comment":
		queryMods = append(queryMods, qm.Where("\"comment\" ILIKE ?", "%"+search+"%"))
	
	case "Review Date":
		if(search == "" && !from.IsZero() && !to.IsZero() && from.Before(to)){
			queryMods = append(queryMods, qm.Where("DATE(\"reviewDate\") BETWEEN ? AND ?", from, to))
		}
	}

	return queryMods, nil
	
}

func convertIntSliceToInterface(s []int) []interface{}{
	result := make([]interface{},len(s))
	for i, v := range s{
		result[i] = v
	}
	return result
}

// fetchReviewCards gets overall review stats
func FetchReviewCards(c *fiber.Ctx) (dto.ReviewCards, error) {
	var cards dto.ReviewCards
	err := queries.Raw(`
		SELECT COALESCE(COUNT(*),0) AS total_review,
			   COUNT(CASE WHEN stars = 5 THEN 1 END) AS total_5s
		FROM review
	`).Bind(c.Context(), boil.GetContextDB(), &cards)
	return cards, err
}

// analyzeReviews sends comments to Python service for sentiment analysis
func AnalyzeReviews(comments []string, reviewIDs []int) ([]dto.AnalyzeResult, error) {
	client := resty.New()
	var result []dto.AnalyzeResult
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"review": comments, "reviewID": reviewIDs}).
		SetResult(&result).
		Post("http://192.168.1.10:5000/analyze")
	return result, err
}

//AppendSortingResult appends all analysis result into a filtered list and return to the caller
func AppendSortingResult(result []dto.AnalyzeResult, sort string) []dto.AnalyzeResult {
	sortingResult := []dto.AnalyzeResult{}
	keywords := []string{"Khen", "Chê", "Ý kiến trung lập"}

	for _, s := range result {
		count := 0
		for _, keyword := range keywords {
			if strings.Contains(s.Summary, keyword) {
				count++
			}
		}

		switch sort {
			case "Positive Sentiment":
				if count == 1 && strings.Contains(s.Summary, "Khen") {
					sortingResult = append(sortingResult, s)
				}
			case "Negative Sentiment":
				if count == 1 && strings.Contains(s.Summary, "Chê") {
					sortingResult = append(sortingResult, s)
				}
			case "Neutral Sentiment":
				if count == 1 && strings.Contains(s.Summary, "Ý kiến trung lập") {
					sortingResult = append(sortingResult, s)
				}
			case "Mixed Sentiment":
				if count >= 2 {
					sortingResult = append(sortingResult, s)
				}
		}
	}

	return sortingResult
}

// fetchReviewData concurrently fetches review, images, and reply for a given reviewID.
// Returns error if any query fails (except reply not found case).
func FetchReviewData(c *fiber.Ctx,reviewID int) (*models.Review,models.ReviewImageSlice,*models.Reply, error){
	var(
		review *models.Review
		reviewImages models.ReviewImageSlice
		reply *models.Reply
	)

	wg := sync.WaitGroup{}
	errChan := make(chan error,3)

	//Fetch review
	wg.Add(1)
	go func(){
		defer wg.Done()
		r, err := models.Reviews(qm.Where("\"reviewID\" = ?",reviewID)).One(c.Context(),boil.GetContextDB());
		if err != nil{
			errChan <- err
			return
		}
		review = r
	}()

	//Fetch images
	wg.Add(1)
	go func() {
		defer wg.Done()
		imgs, err := models.ReviewImages(qm.Where("\"reviewID\" = ?", reviewID)).All(c.Context(), boil.GetContextDB())
		if err != nil {
			errChan <- err
			return
		}
		reviewImages = imgs
	}()

	//Fetch reply
	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := models.Replies(qm.Where("\"reviewID\" = ?", reviewID)).One(c.Context(), boil.GetContextDB())
		if err != nil && err.Error() != "sql: no rows in result set" {
			errChan <- err
			return
		}
		reply = r
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, nil, nil, err
		}
	}

	if review == nil {
		return nil, nil, nil, fmt.Errorf("review not found")
	}

	return review, reviewImages, reply, nil
}