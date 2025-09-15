package utils

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

//HandleOAuthLogin is a shared function for Google/Facebook login.
//It checks if OAuth account exists, links account if needed, or creates new account.
func HandleOAuthLogin(c *fiber.Ctx, provider, providerUserID, email, fullName, avatarURL string) error{
	//Check if oauth user already exists
	existingOAuth, err := models.OauthAccounts(qm.Where("\"providerUserID\" = ?",providerUserID),qm.Load(models.OauthAccountRels.AccountIDAccount)).One(c.Context(),boil.GetContextDB())
	if err == nil{
		//OAuth exists => login user
		return LoginWithAccount(c,existingOAuth.R.AccountIDAccount);
	}

	//Check if account exists with same email
	existingAccount, err := models.Accounts(qm.Where("email = ?",email)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		if existingAccount.EmailVerified{
			//Link OAuth to this account
			oauth := models.OauthAccount{
				AccountID: existingAccount.AccountID,
				Provider: provider,
				ProviderUserID: providerUserID,
			}
			if err := oauth.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
				return service.SendError(c,500,"Failed to link OAuth account")
			}
		}
		return service.SendError(c,403,"This email is associated with an unverified account");
	}

	//Otherwise, create new account and link to oauth
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(providerUserID),bcrypt.DefaultCost)
	newUser := models.Account{
		Username: providerUserID,
		Password: string(hashedPass),
		Email: email,
		FullName: fullName,
		Avatar: null.StringFrom(avatarURL),
		Status: true,
		Role: false,
		PhoneNumber: null.String{},
		Gender: true,
	}
	if err := newUser.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500, "Failed to create account");
	}

	oauthUser := models.OauthAccount{
		AccountID: newUser.AccountID,
		Provider: provider,
		ProviderUserID: providerUserID,
	}
	if err := oauthUser.Insert(c.Context(), boil.GetContextDB(), boil.Infer()); err != nil {
        return service.SendError(c, fiber.StatusInternalServerError, "Failed to create OAuth account")
    }

	return LoginWithAccount(c, &newUser)
}

func LoginWithAccount(c *fiber.Ctx, acc *models.Account) error{
	//provide user with a token
	accessToken,refreshToken,sessionID, err := auth.CreateToken(acc.Username,"")
	if err != nil{
		return service.SendError(c,500,"No username found!")
	}

	//set refreshToken and httpOnly cookie
	SaveCookie(refreshToken,c);

	response := fiber.Map{
		"status": "Success",
		"data": fiber.Map{
			"user": acc,
			"accessToken": accessToken,
			"refreshToken": refreshToken,
			"sessionID": sessionID,
		},
		"message": "Login successful!",
	}
	return c.JSON(response);
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

func SaveCookie(refreshToken string,c *fiber.Ctx){
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
