package dto

type ChangePassRequest struct{
	AccountID int `json:"accountID"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}
type ChangePassErr struct{
	ErrOldPassword string `json:"errOldPassword"`
	ErrNewPassword string `json:"errNewPassword"`
	ErrConfirmPassword string `json:"errConfirmPassword"`
	ErrAccount string
}