package jobs

import (
	"GoodFood-BE/internal/service"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func HandleResetPasswordEmailTask(ctx context.Context, t *asynq.Task) error{
	var payload ResetPasswordPayload
	if err := json.Unmarshal(t.Payload(),&payload); err != nil{
		return fmt.Errorf("Failed to unmarshal payload: %v",err);
	}

	err := service.SendResetPasswordEmail(payload.ToEmail,payload.ResetLink)
	if err != nil{
		return fmt.Errorf("Failed to send email: %v",err);
	}
	return nil
}