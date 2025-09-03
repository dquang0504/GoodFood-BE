package jobs

import (
	"GoodFood-BE/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

//This function handles the execution of the "reset password" email job.
//It unmarshals the task payload into ResetPasswordPayload, then sends an email with the reset link
func HandleResetPasswordEmailTask(ctx context.Context, t *asynq.Task) error{
	var payload ResetPasswordPayload
	if err := json.Unmarshal(t.Payload(),&payload); err != nil{
		return fmt.Errorf("failed to unmarshal payload: %v",err);
	}

	//Attempt to send reset password email
	if err := utils.SendResetPasswordEmail(payload.ToEmail,payload.ResetLink);err != nil{
		return fmt.Errorf("failed to send email: %v",err);
	}
	return nil
}

//This function handles the execution of a "customer contact message" job.
//Unmarshals the task payload into CustomerContactPayload, then sends the message to customer support (my email: williamdang0404@gmail.com)
func HandleContactCustomerSent(ctx context.Context, t *asynq.Task) error{
	var payload CustomerSentContactPayload
	if err := json.Unmarshal(t.Payload(),&payload); err != nil{
		return fmt.Errorf("failed to unmarshal payload: %v",err);
	}

	//Attempt to send the customer message
	if err := utils.SendMessageCustomerSent(payload.FromEmail,payload.Message);err != nil{
		return fmt.Errorf("failed to send email: %v",err);
	}
	return nil
}