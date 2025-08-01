package utils

import (
	"GoodFood-BE/internal/auth"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
	"google.golang.org/genai"
	"gopkg.in/gomail.v2"
)

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

func SendMessageCustomerSent(fromEmail string, message string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "williamdang0404@gmail.com")
	mailer.SetHeader("To", "williamdang0404@gmail.com")
	mailer.SetHeader("Subject", "📩 Contact Message from Customer")

	emailBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; color: #333; padding: 20px; max-width: 600px; margin: auto; border: 1px solid #ddd; border-radius: 8px;">
			<h2 style="color: #4CAF50;">📨 Bạn nhận được tin nhắn từ khách hàng!</h2>
			<p><strong>Địa chỉ email khách hàng:</strong> <a href="mailto:%s">%s</a></p>

			<hr style="margin: 20px 0;">

			<p style="white-space: pre-line; line-height: 1.6;">%s</p>

			<hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
			<p style="font-size: 13px; color: #888;">Email này được gửi từ hệ thống website GoodFood24h.</p>
		</div>
	`, fromEmail, fromEmail, message)

	mailer.SetBody("text/html", emailBody)
	
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "williamdang0404@gmail.com", "yhjd uzhk hhvp zfiq")

	err := dialer.DialAndSend(mailer)
	return err
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
	config := &firebase.Config{
		StorageBucket: "fivefood-datn-8a1cf.appspot.com",
	}
	app, err := firebase.NewApp(ctx,config,option.WithCredentialsFile("./config/fivefood-datn-8a1cf-firebase-adminsdk-n0vxi-9ad735160d.json"))
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

func UploadFirebaseImages(images map[string][]byte, ctx context.Context) (map[string]string, error) {
	app := InitializeFirebaseApp(ctx)

	storageClient, err := app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	bucket, err := storageClient.DefaultBucket()
	if err != nil {
		return nil, err
	}

	// Kết quả trả về: map tên ảnh -> URL công khai
	uploadedURLs := make(map[string]string)

	for imageName, imageData := range images {
		obj := bucket.Object("AnhDanhGia/" + imageName) // optional: prefix folder "reviews/"
		writer := obj.NewWriter(ctx)
		writer.ContentType = "image/jpeg"

		if _, err := writer.Write(imageData); err != nil {
			writer.Close()
			return nil, fmt.Errorf("failed to upload %s: %w", imageName, err)
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close writer for %s: %w", imageName, err)
		}
		// Cho phép truy cập public
		err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader)
		if err != nil{
			return nil, fmt.Errorf("failed to set files public %s: %w", imageName, err)
		}

		objectPath := "AnhDanhGia/" + imageName
		encodedPath := url.QueryEscape(objectPath)
		publicURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", bucket.BucketName(), encodedPath)
		uploadedURLs[imageName] = publicURL
	}

	return uploadedURLs, nil
}

func CreateTokenForUser(ctx *fiber.Ctx,username string) (accessToken string, error error){
	//provide user with a token
	accessToken,refreshToken,_, err := auth.CreateToken(username)
	if err != nil{
		return "",err;
	}

	//set refreshToken as HTTP-only Cookie
	ctx.Cookie(&fiber.Cookie{
		Name: "refreshToken",
		Value: refreshToken,
		Path: "/",
		MaxAge: 7*24*60*60, //7 days
		HTTPOnly: true,
		Secure: false, //Switch to `true` if running HTTPS
		SameSite: "None",
	})

	return accessToken,nil;
}

type FacebookUserStruct struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Picture struct{
		Data struct{
			URL string `json:"url"`
		} `json:"data"`
	}`json:"picture"`
}

func GetFacebookUserInfo(accessToken string)(*FacebookUserStruct,error){
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)&access_token=" + url.QueryEscape(accessToken))
	if err != nil{
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200{
		return nil,fmt.Errorf("Facebook API eror: %v",resp.Status)
	}

	var fbUser FacebookUserStruct
	if err := json.NewDecoder(resp.Body).Decode(&fbUser); err != nil{
		return nil, err
	}

	return &fbUser, nil
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
		return nil,service.SendError(c, 500, "Failed to create client: "+err.Error())
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
		return nil,service.SendError(c, 500, "Failed to generate content: "+err.Error())
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