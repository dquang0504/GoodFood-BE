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