package dto

import "GoodFood-BE/models"


//ReviewCards represents 2 pieces of info for AdminReview UI
type ReviewCards struct{
	TotalReview int `boil:"total_review"`
	Total5S int `boil:"total_5s"`
}

//ReviewResponse represents the response structure for AdminReview.tsx to enhance readability
type ReviewResponse struct{
	models.Review
	ReviewAccount models.Account `json:"reviewAccount"`
	ReviewProduct models.Product `json:"reviewProduct"`
	ReviewImages models.ReviewImageSlice `json:"reviewImages"`
	ReviewReply models.ReplySlice `json:"reviewReply"`
	ReviewInvoice models.Invoice `json:"reviewInvoice"`
}

// ClauseAnalysis represents the clauses along with its sentiment received from microservice flask.
type ClauseAnalysis struct {
    Clause    string `json:"clause"`
    Sentiment string `json:"sentiment"`
}

//AnalyzeResult represents the sentiment analysis response received from microservice flask.
type AnalyzeResult struct {
	ReviewID int 			  `json:"reviewID"`
    Review   string           `json:"review"`
    Clauses  []string         `json:"clauses"`
    Analysis []ClauseAnalysis `json:"analysis"`
    Summary  string           `json:"summary"`
}

//ReviewSubmitRequest represents the request body received from front-end
type ReviewSubmitRequest struct{
	models.Review
	ReviewImages []models.ReviewImage `json:"reviewImages"`
}

//NSFWScores represents the 3 categories of NSFW-content recognition model
type NSFWScores struct{
	Unsafe float64 `json:"unsafe"`
	Porn float64 `json:"porn"`
	Sexy float64 `json:"sexy"`
}

//ImageDetectionResult represents the response received from flask python microservice - NSFW & violent content recognition
type ImageDetectionResult struct{
	Image string `json:"image"`
	NSFW bool `json:"nsfw"`
	NSFWScores NSFWScores `json:"nsfw_scores"`
	Violent bool `json:"violent"`
	ViolentLabel string `json:"violent_label"`
}

//ReviewContentDetection represents the response received from flask python microservice - toxic speech recognition
type ReviewContentDetection struct{
	Label string `json:"label"`
	Score float64 `json:"score"`
	Images []ImageDetectionResult `json:"images"`
}