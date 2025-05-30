package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeResetPasswordEmail = "email:reset_password"

type ResetPasswordPayload struct{
	ToEmail string
	ResetLink string
}

func NewResetPasswordEmailTask(toEmail, resetLink string) (*asynq.Task, error){
	payload, err := json.Marshal(ResetPasswordPayload{
		ToEmail: toEmail,
		ResetLink: resetLink,
	})
	if err != nil{
		return nil, err
	}
	return asynq.NewTask(TypeResetPasswordEmail, payload),nil
}