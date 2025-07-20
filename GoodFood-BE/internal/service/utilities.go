package service

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
	"google.golang.org/genai"
	"gopkg.in/gomail.v2"
)

func SendError(c *fiber.Ctx, statusCode int, message string) error{
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "error",
		"message": message,
	})
}

func SendErrorStruct(c *fiber.Ctx, statusCode int, err interface{}) error{
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "error",
		"err": err,
	})
}

func SendJSON(c *fiber.Ctx,status string, data interface{}, extras map[string]interface{},message string) error{
	
	//creating base response
	resp := fiber.Map{
		"status": status,
		"data": data,
		"message": message,
	}

	//adding extra variables
	for key, value := range extras{
		resp[key] = value
	}

	return c.JSON(resp)
}

func SendResetPasswordEmail(toEmail string, resetLink string) error{
	mailer := gomail.NewMessage();
	mailer.SetHeader("From","williamdang0404@gmail.com")
	mailer.SetHeader("To",toEmail)
	mailer.SetHeader("Subject","Reset Your Password")

	emailBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; color: #333; padding: 20px; max-width: 600px; margin: auto; border: 1px solid #ddd; border-radius: 8px;">
			<div style="text-align: center;">
				<img src="https://firebasestorage.googleapis.com/v0/b/fivefood-datn-8a1cf.appspot.com/o/test%%2Fcomga.png?alt=media&token=0367b2f7-2129-49c1-be47-76e936603dd8" alt="GoodFood24h Logo" style="width: 150px; margin-bottom: 20px;">
			</div>
			<h2 style="color: #ff5722;">Xin chào,</h2>
			<p>Chúng tôi nhận được yêu cầu <strong>đặt lại mật khẩu</strong> cho tài khoản của bạn tại <strong>GoodFood24h</strong>.</p>
			<p>Nếu bạn không yêu cầu điều này, bạn có thể <em>bỏ qua email này</em>.</p>

			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #ff5722; color: white; padding: 12px 24px; border-radius: 5px; text-decoration: none; font-weight: bold;">Đặt lại mật khẩu</a>
			</div>

			<p>Hoặc bạn có thể sao chép và dán đường dẫn sau vào trình duyệt:</p>
			<p style="word-break: break-all;"><a href="%s">%s</a></p>

			<hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">

			<p style="font-size: 14px; color: #888;">Email này được gửi từ hệ thống của GoodFood24h. Vui lòng không trả lời lại email này.</p>

			<p style="margin-top: 30px;">Thân mến,<br><strong>Đội ngũ GoodFood24h</strong></p>
		</div>
	`, resetLink, resetLink, resetLink)

	mailer.SetBody("text/html", emailBody)
	dialer := gomail.NewDialer("smtp.gmail.com",587,"williamdang0404@gmail.com","yhjd uzhk hhvp zfiq")
	err := dialer.DialAndSend(mailer);
	return err;
}

func ClearProductCache(productID int) error{
	redisSetKey := fmt.Sprintf("product:detail:%d:keys",productID)

	//Fetch all saved cached keys
	keys, err := redisdatabase.Client.SMembers(redisdatabase.Ctx,redisSetKey).Result()
	if err != nil{
		return fmt.Errorf("failed to get cache keys: %v", err)
	}
	//deleting key one by one
	if len(keys) > 0{
		err = redisdatabase.Client.Del(redisdatabase.Ctx,keys...).Err()
		if err != nil{
			return fmt.Errorf("failed to delete cached keys: %v",err)
		}
	}
	return nil
}

func InitializeFirebaseApp(ctx context.Context) *firebase.App{
	//firebase app initialization
	app, err := firebase.NewApp(ctx,nil,option.WithCredentialsFile("./config/fivefood-datn-8a1cf-firebase-adminsdk-n0vxi-9ad735160d.json"))
	if err != nil{
		log.Fatalf("error initializing app: %v\n", err)
	}
	return app
}

func DeleteFirebaseImage(imgPath string, ctx context.Context) error{
	app := InitializeFirebaseApp(ctx)

	storageClient, err := app.Storage(ctx)
	if err != nil{
		return err
	}

	bucket, err := storageClient.DefaultBucket()
	if err != nil{
		return err
	}

	obj := bucket.Object(imgPath)
	if err := obj.Delete(ctx); err != nil{
		return err
	}

	return nil
}

func FunctionDeclaration() []*genai.FunctionDeclaration {
	// ✅ Function declaration
	functions := []*genai.FunctionDeclaration{
		{
			Name: "get_order_status",
			Description: "Retrieve the status of an order in GoodFood24h",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"order_id": {Type: genai.TypeString, Description: "The ID of the order"},
				},
				Required: []string{"order_id"},
			},
		},
		{
			Name: "get_top_product",
			Description: "Retrieve the information of the most sold product in GoodFood24h",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
			},
		},
		{
			Name: "place_order",
			Description: "Proceed to place an order of the given product along with its quantity",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"products":{
						Type: genai.TypeArray,
						Description: "The name of the product",
						Items: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"product_name":{
									Type: genai.TypeString,
									Description: "Name of the product",
								},
								"quantity":{
									Type: genai.TypeInteger,
									Description: "Number of units to order",
								},
							},
							Required: []string{"product_name","quantity"},
						},
					},
					"payment_method": {
						Type: genai.TypeString,
						Description: "Payment method: 'COD' means pay when receiving (true), 'ONLINE' means pay online (false).",
						Enum: []string{"COD", "ONLINE"},
					},
				},
				Required: []string{"products","payment_method"},
			},
		},
	}
	return functions
}

func CallVertexAI(prompt string,c *fiber.Ctx, withFunction bool) (*genai.GenerateContentResponse, error){
	client, err := genai.NewClient(c.Context(), &genai.ClientConfig{
		Project:  "322745191572",
		Location: "us-central1",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil,SendError(c, 500, "Failed to create client: "+err.Error())
	}

	// ✅ Generation config with function calling
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 1024,
		Temperature:     Float32Ptr(0.7),
		TopP:            Float32Ptr(0.9),
	}

	if withFunction {
		config.Tools = []*genai.Tool{
			{FunctionDeclarations: FunctionDeclaration()},
		}
	}

	// ✅ Prompt
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "You are a concise assistant specialized in answering questions about GoodFood24h - an e-commerce website for ordering food online. " +
					"Answer this question or call a function if needed: " + prompt},
			},
		},
	}

	// ✅ Generate content
	res, err := client.Models.GenerateContent(c.Context(),
		"projects/322745191572/locations/us-central1/endpoints/5530821664155107328",
		contents,
		config,
	)
	if err != nil {
		return nil,SendError(c, 500, "Failed to generate content: "+err.Error())
	}

	return res,nil
}

func GiveStructuredAnswer(question string,prompt string, c *fiber.Ctx) (string, error) {
	instructionNdPrompt := fmt.Sprintf(
		"The question is about %s in GoodFood24h. Write a concise, natural answer with key details: product name, total quantity sold (exclude product id): %s.",
		question,prompt,
	)
	res, err := CallVertexAI(instructionNdPrompt, c, false)
	if err != nil {
		return "", err
	}

	// 🟢 DEBUG: In toàn bộ response
	fmt.Printf("Full response: %+v\n", len(res.Candidates))

	result := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				result += part.Text
			}
		}
	}
	return result, nil
}

func GiveAnswerForUnreachableData(question string,c *fiber.Ctx) (string, error) {
	instructionNdPrompt := fmt.Sprintf(
		"The question is about %s in GoodFood24h. Write a short answer explaining to the user why the data is restricted from them and try to make it easy to understand, don't use any coding terminology.",
		question,
	)
	res, err := CallVertexAI(instructionNdPrompt, c, false)
	if err != nil {
		return "", err
	}

	result := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				result += part.Text
			}
		}
	}
	return result, nil
}

func Float32Ptr(f float32) *float32 {
	return &f
}