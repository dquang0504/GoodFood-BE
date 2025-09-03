package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

//Task type constants used to identify job categories in Asynq
const TypeResetPasswordEmail = "email:reset_password"
const TypeSendContactMessage = "contact:customer_sent"

//This struct defines the payload for reset password email tasks
type ResetPasswordPayload struct{
	ToEmail string
	ResetLink string
}

//This one defines the payload for customer contact message task
type CustomerSentContactPayload struct{
	Fullname string
	FromEmail string
	Message string
}


//This function creates a new task for sending a reset password email.
//It marshals the payload and returns an *asynq.Task that can be enqueued.
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

//This function creates a new task for sending a customer contact message.
//It marshals the payload and returns an *asynq.Task that can be enqueued
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