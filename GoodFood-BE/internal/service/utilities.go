package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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