package dto

//UserCards struct represents the summary stats for Admin User module
type UserCards struct {
	TotalUsers    int `boil:"totalusers"`
	TotalDisabled int `boil:"totaldisabled"`
}

//UserError struct represents custom errors for api response in Admin User module
type UserError struct{
	ErrName string `json:"errName"`
	ErrUsername string `json:"errUsername"`
	ErrPhone string `json:"errPhone"`
	ErrEmail string `json:"errEmail"`
	ErrPassword string `json:"errPassword"`
}