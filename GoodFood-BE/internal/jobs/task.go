package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeResetPasswordEmail = "email:reset_password"
const TypeSendContactMessage = "contact:customer_sent"

type ResetPasswordPayload struct{
	ToEmail string
	ResetLink string
}

type CustomerSentContactPayload struct{
	Fullname string
	FromEmail string
	Message string
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

func NewCustomerSentContactTask(fullname,fromEmail, message string) (*asynq.Task, error){
	payload, err := json.Marshal(CustomerSentContactPayload{
		Fullname: fullname,
		FromEmail: fromEmail,
		Message: message,
	})
	if err != nil{
		return nil,err
	}
	return asynq.NewTask(TypeSendContactMessage,payload),nil
}