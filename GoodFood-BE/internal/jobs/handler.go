package jobs

import (
	"GoodFood-BE/internal/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func HandleResetPasswordEmailTask(ctx context.Context, t *asynq.Task) error{
	var payload ResetPasswordPayload
	if err := json.Unmarshal(t.Payload(),&payload); err != nil{
		return fmt.Errorf("failed to unmarshal payload: %v",err);
	}

	err := utils.SendResetPasswordEmail(payload.ToEmail,payload.ResetLink)
	if err != nil{
		return fmt.Errorf("failed to send email: %v",err);
	}
	return nil
}

func HandleContactCustomerSent(ctx context.Context, t *asynq.Task) error{
	var payload CustomerSentContactPayload
	if err := json.Unmarshal(t.Payload(),&payload); err != nil{
		return fmt.Errorf("failed to unmarshal payload: %v",err);
	}

	err := utils.SendMessageCustomerSent(payload.FromEmail,payload.Message)
	if err != nil{
		return fmt.Errorf("failed to send email: %v",err);
	}
	return nil
}