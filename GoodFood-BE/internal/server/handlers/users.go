package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/jobs"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"fmt"
	"os"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/hibiken/asynq"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

func HandleRegister(c *fiber.Ctx) error{
	hasPassword := true;
	hasUsername := true
	hasEmail := true
	hasPhone := true;
	var user models.Account
	if err := c.BodyParser(&user); err != nil{
		return service.SendError(c,400,err.Error());
	}

	if valid, errObj := validationUser(&user,hasPassword,hasUsername,hasEmail,hasPhone); !valid{
		return service.SendErrorStruct(c,500,errObj)
	}

	//encrypting password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password),bcrypt.DefaultCost)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	user.Password = string(hashedPass);
	err = user.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully registered!",
	}
	return c.JSON(response);
}

func HandleLogin(c *fiber.Ctx) error{
	//fetching parameters from url
	username := c.Query("username")
	password := c.Query("password")
	if(username == "" || password == "" ){
		return service.SendError(c,400,"Either username or password wasn't valid")
	}

	//comparing login details with users db
	user, err := models.Accounts(qm.Where("username = ?", username)).One(c.Context(), boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,"User not found!")
	}

	//comparing 2 hashes
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(password)); err != nil{
		return service.SendError(c,401, "Username or password does not match!");
	}

	//provide user with a token
	accessToken,refreshToken,sessionID, err := auth.CreateToken(username,"")
	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	//set refreshToken and httpOnly cookie
	saveCookie(refreshToken,c);

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": user,
			"accessToken": accessToken,
			"refreshToken": refreshToken,
			"sessionID": sessionID,
		},
		"message": "Successfully fetched login details!",
	}
	return c.JSON(response);
}

type OAuthLoginStruct struct{
	AccessToken string `json:"accessToken"`
}

func HandleLoginGoogle(c *fiber.Ctx) error{
	body := OAuthLoginStruct{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error())
	}

	idToken := body.AccessToken
	audience := os.Getenv("GOOGLE_AUDIENCE")

	//authorize token with google
	payload, err := idtoken.Validate(c.Context(),idToken,audience);
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	//parsing payload
	sub := fmt.Sprintf("%v",payload.Claims["sub"]) //unique google id
	name := fmt.Sprintf("%v",payload.Claims["name"])
	email := fmt.Sprintf("%v",payload.Claims["email"])
	picture := fmt.Sprintf("%v", payload.Claims["picture"])

	//checking if the oauth user exists
	existingOAuth, err := models.OauthAccounts(qm.Where("\"providerUserID\" = ?",sub),qm.Load(models.OauthAccountRels.AccountIDAccount)).One(c.Context(),boil.GetContextDB());
	if err == nil{

		//provide user with a token
		accessToken,refreshToken, sessionID, err := auth.CreateToken(existingOAuth.R.AccountIDAccount.Username, "")
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		
		//set refreshToken and httpOnly cookie
		saveCookie(refreshToken,c);

		return c.JSON(fiber.Map{
			"status": "Success",
			"data": fiber.Map{
				"user": existingOAuth.R.AccountIDAccount,
				"accessToken": accessToken,
				"refreshToken": refreshToken,
				"sessionID": sessionID,
			},
			"message": "Login successfully!",
		})
	}

	//checking if there is already an account with the same email
	existingAccount, err := models.Accounts(qm.Where("email = ?",email)).One(c.Context(),boil.GetContextDB());
	if err == nil{
		//if exists, bind OAuth account to this account
		if existingAccount.EmailVerified{
			oauth := models.OauthAccount{
				AccountID: existingAccount.AccountID,
				Provider: "google",
				ProviderUserID: sub,
			}
			if err := oauth.Insert(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
				return service.SendError(c, 500, "Failed to link OAuth: "+err.Error())
			}

			//provide user with a token
			accessToken,refreshToken, sessionID, err := auth.CreateToken(existingAccount.Username, "")
			if err != nil{
				return service.SendError(c,500,"No username found!")
			}

			//set refreshToken and httpOnly cookie
			saveCookie(refreshToken,c);

			return c.JSON(fiber.Map{
				"status": "Success",
				"data": fiber.Map{
					"user": existingAccount,
					"accessToken": accessToken,
					"refreshToken": refreshToken,
					"sessionID": sessionID,
				},
				"message": "OAuth linked & login successfully!",
			})
		}else{
			return service.SendError(c,403,"This email is already associated with an unverified account. Please verify your email or contact support.")
		}
	}

	//if oauth account does not exist
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(sub),bcrypt.DefaultCost)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert into account table
	newUser := models.Account{
		Username: sub,
		Password: string(hashedPass),
		PhoneNumber: null.String{},
		Email: email,
		Gender: true,
		FullName: name,
		Avatar: null.StringFrom(picture),
		Status: true,
		Role: false,
	}
	if err := newUser.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert into oauth_account table
	oauthUser := models.OauthAccount{
		AccountID: newUser.AccountID,
		Provider: "google",
		ProviderUserID: sub,
	}
	if err := oauthUser.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	//provide user with a token
	accessToken,refreshToken, sessionID, err := auth.CreateToken(newUser.Username,"")
	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	//set refreshToken and httpOnly cookie
	saveCookie(refreshToken,c);

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": newUser,
			"accessToken": accessToken,
			"refreshToken": refreshToken,
			"sessionID": sessionID,
		},
		"message": "Successfully fetched login details!",
	}
	return c.JSON(response);
}

func HandleLoginFacebook(c *fiber.Ctx) error{
	body := OAuthLoginStruct{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error())
	}

	idToken := body.AccessToken
	fbUser, err := utils.GetFacebookUserInfo(idToken);
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	//checking if the oauth user exists
	existingOAuth, err := models.OauthAccounts(qm.Where("\"providerUserID\" = ?",fbUser.ID),qm.Load(models.OauthAccountRels.AccountIDAccount)).One(c.Context(),boil.GetContextDB());
	if err == nil{

		//provide user with a token
		accessToken,refreshToken,sessionID, err := auth.CreateToken(existingOAuth.R.AccountIDAccount.Username, "")
		if err != nil{
			return service.SendError(c,500,"No username found!")
		}

		//set refreshToken and httpOnly cookie
		saveCookie(refreshToken,c);

		return c.JSON(fiber.Map{
			"status": "Success",
			"data": fiber.Map{
				"user": existingOAuth.R.AccountIDAccount,
				"accessToken": accessToken,
				"refreshToken": refreshToken,
				"sessionID": sessionID,
			},
			"message": "Login successfully!",
		})
	}

	//checking if there is already an account with the same email
	existingAccount, err := models.Accounts(qm.Where("email = ?",fbUser.Email)).One(c.Context(),boil.GetContextDB());
	if err == nil{
		//if exists, bind OAuth account to this account
		if existingAccount.EmailVerified{
			oauth := models.OauthAccount{
				AccountID: existingAccount.AccountID,
				Provider: "facebook",
				ProviderUserID: fbUser.ID,
			}
			if err := oauth.Insert(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
				return service.SendError(c, 500, "Failed to link OAuth: "+err.Error())
			}

			//provide user with a token
			accessToken,refreshToken,sessionID, err := auth.CreateToken(existingAccount.Username, "")
			if err != nil{
				return service.SendError(c,500,"No username found!")
			}

			//set refreshToken and httpOnly cookie
			saveCookie(refreshToken,c);

			return c.JSON(fiber.Map{
				"status": "Success",
				"data": fiber.Map{
					"user": existingAccount,
					"accessToken": accessToken,
					"refreshToken": refreshToken,
					"sessionID": sessionID,
				},
				"message": "OAuth linked & login successfully!",
			})
		}else{
			return service.SendError(c,403,"This email is already associated with an unverified account. Please verify your email or contact support.")
		}
	}

	//if oauth account does not exist
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(fbUser.ID),bcrypt.DefaultCost)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert into account table
	newUser := models.Account{
		Username: fbUser.ID,
		Password: string(hashedPass),
		PhoneNumber: null.String{},
		Email: fbUser.Email,
		Gender: true,
		FullName: fbUser.Name,
		Avatar: null.StringFrom(fbUser.Picture.Data.URL),
		Status: true,
		Role: false,
	}
	fmt.Println(newUser.Avatar);
	if err := newUser.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert into oauth_account table
	oauthUser := models.OauthAccount{
		AccountID: newUser.AccountID,
		Provider: "facebook",
		ProviderUserID: fbUser.ID,
	}
	if err := oauthUser.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	//provide user with a token
	accessToken,refreshToken,sessionID, err := auth.CreateToken(newUser.Username, "")
	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	//set refreshToken and httpOnly cookie
	saveCookie(refreshToken,c);

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": newUser,
			"accessToken": accessToken,
			"refreshToken": refreshToken,
			"sessionID": sessionID,

		},
		"message": "Successfully fetched login details!",
	}
	return c.JSON(response);
}

func HandleUpdateAccount(c *fiber.Ctx) error{
	hasPassword := false;
	hasUsername := false;
	hasEmail := false;
	hasPhone := false;
	var body models.Account
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID!");
	}

	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error());
	}
	account, err := models.Accounts(qm.Where("\"accountID\" = ?",accountID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//check logic for hasEmail and hasPhone
	if body.Email != account.Email{
		hasEmail = true;
	}
	if body.PhoneNumber.String != account.PhoneNumber.String{
		hasPhone = true;
	}

	if ok, errObj := validationUser(&body,hasPassword,hasUsername,hasEmail,hasPhone); !ok{
		return service.SendErrorStruct(c,400,errObj);
	}

	//start updating account info
	account.FullName = body.FullName
	if account.PhoneNumber != null.StringFrom(body.PhoneNumber.String){
		account.PhoneNumber = null.StringFrom(body.PhoneNumber.String)
	}
	if account.Email != body.Email{
		account.Email = body.Email
	}
	//start here
	if body.Avatar.String != ""{
		account.Avatar = null.StringFrom(body.Avatar.String)
	}
	fmt.Println("Account: ",account.Avatar)
	_, err = account.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": account,
		"message": "Successfully updated account information!",
	}

	return c.JSON(resp);
}

type ForgotPassStruct struct{
	Email string `json:"email"`
	CodeOTP string `json:"codeOTP"`
}
var asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379",Password: "",DB: 0})
func HandleForgotPassword(c *fiber.Ctx) error{
	body := ForgotPassStruct{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error())
	}
	if body.Email == ""{
		return service.SendError(c,400,"Please input your email!");
	}
	
	token, err := uuid.NewV4();
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s",token.String());
	//only send mmail if the user exists in database
	_, err = models.Accounts(qm.Where("email = ?",body.Email)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		resp := fiber.Map{
			"status": "Succes",
			"message": "If the email is registered, a password reset link will be sent to your inbox.",
		}
		return c.JSON(resp);
	}

	//caching token key if user exists
	redisKey := fmt.Sprintf("resetPass:token=%s",token.String())
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,body.Email, 10*time.Minute).Err()
	if rdsErr != nil{
		fmt.Println("Failed to cache cart data: ",rdsErr)
	}

	//creating email sending task
	task, err := jobs.NewResetPasswordEmailTask(body.Email,resetLink);
	if err != nil{
		return service.SendError(c,500, err.Error())
	}
	//enqueuing
	_,err = asynqClient.Enqueue(task)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	// err = service.SendResetPasswordEmail(body.Email,resetLink)
	// if err != nil{
	// 	return service.SendError(c,500,err.Error());
	// }

	resp := fiber.Map{
		"status": "Succes",
		"data": token,
		"message": "If the email is registered, a password reset link will be sent to your inbox.",
	}
	return c.JSON(resp);
}

func ValidateResetToken(c *fiber.Ctx) error{
	paramToken := c.Query("token","");
	if paramToken == ""{
		return service.SendError(c,400,"Did not receive token!");
	}
	redisKey := fmt.Sprintf("resetPass:token=%s",paramToken);
	cachedToken, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		fmt.Println("Hello teacher");
		resp := fiber.Map{
			"status": "Success",
			"isValid": true,
			"email": cachedToken,
			"message": "Successfully validated reset password token!",
		}
		return c.JSON(resp);
	}
	return nil
}

type ResetPass struct{
	NewPass string `json:"newPass"`
	ConfirmPass string `json:"confirmPass"`
	Email string `json:"email"`
}
func HandleResetPassword(c *fiber.Ctx) error{
	body := ResetPass{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error())
	}
	
	if(body.NewPass == ""){
		return service.SendError(c,400,"Please input your new password!");
	}

	//encrypting password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(body.NewPass),bcrypt.DefaultCost);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//finding user with email from body
	user, err := models.Accounts(qm.Where("email = ?",body.Email)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//updating password
	user.Password = string(hashedPass);
	_, err = user.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	//deleting token in redis
	paramToken := c.Query("token","");
	if paramToken != ""{
		redisKey := fmt.Sprintf("resetPass:token=%s",paramToken);
		err := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if err != nil{
			fmt.Println("Failed to delete reset token from Redis:", err)
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": user,
		"message": "Successfully reset password!",
	}
	return c.JSON(resp);
}

func RefreshToken(c *fiber.Ctx) error{
	//Fetch sessionID from response
	var req struct{
		SessionID string `json:"sessionID"`
	}
	if err := c.BodyParser(&req); err != nil{
		return service.SendError(c,401,"Mising sessionID");
	}

	cookie := c.Cookies("refreshToken");
	if cookie == ""{
		return service.SendError(c,401,"No refresh token");
	}

	//Verify refreshToken
	claims, err := auth.VerifyToken(cookie)
	if err != nil{
		return service.SendError(c, 401, "Invalid refresh token")
	}

	if claims.SessionID != req.SessionID{ //check for matching sessionID
		fmt.Println("claims ID: ",claims.SessionID)
		fmt.Println("req session ID: ",req.SessionID)
		return service.SendError(c,401,"Session mismatch")
	}

	fmt.Println("claims ID current: ",claims.SessionID)

	//Generate new accessToken
	accessToken, _,_, err := auth.CreateToken(claims.Username, claims.SessionID)
	if err != nil{
		return service.SendError(c, 500, "Cound not generate token")
	}

	response := fiber.Map{
		"status": "Success",
		"accessToken": accessToken,
	}
	return c.JSON(response)
}

type ContactResponse struct{
	Fullname string `json:"name"`
	Email string `json:"fromEmail"`
	Message string `json:"content"`
}
func HandleContact(c *fiber.Ctx) error{
	body := ContactResponse{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error())
	}

	//creating email sending task
	task, err := jobs.NewCustomerSentContactTask(body.Fullname,body.Email,body.Message);
	if err != nil{
		return service.SendError(c,500, err.Error())
	}
	//enqueuing
	_,err = asynqClient.Enqueue(task)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	response := fiber.Map{
		"status": "Success",
		"data": "",
		"message": "Successfully sent a message. We will contact you back shortly!",
	}

	return c.JSON(response);
}

func saveCookie(refreshToken string,c *fiber.Ctx){
	c.Cookie(&fiber.Cookie{
        Name:     "refreshToken",
        Value:    refreshToken,
        HTTPOnly: true,
        Secure:   false,
        SameSite: "Lax",
        Path:     "/",
        Expires:  time.Now().Add(7 * 24 * time.Hour),
    })
}