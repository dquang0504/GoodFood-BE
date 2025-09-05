package handlers

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

type CartDetailResponse struct{
	models.CartDetail
	Product *models.Product `json:"product"`
}

func GetCartDetail(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID");
	}

	//creating redis key after accountID
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID);
	//fetching redis key
	cachedCart,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result()
	if err == nil{
		return c.JSON(json.RawMessage(cachedCart))
	}

	cartDetails, err := models.CartDetails(
		qm.Where("\"accountID\" = ?", accountID),
		qm.Load(models.CartDetailRels.ProductIDProduct), // Load quan hệ Product
	).All(c.Context(), boil.GetContextDB())

	if err != nil {
		return service.SendError(c, 500, "Cart detail not found")
	}

	// Converting data into custom struct that has both cart details and products
	response := make([]CartDetailResponse, len(cartDetails))
	for i, cart := range cartDetails {
		response[i] = CartDetailResponse{
			CartDetail: *cart,
			Product: cart.R.ProductIDProduct,
		}
	}

	resp := fiber.Map{
		"status":  "Success",
		"data":    response,
		"message": "Successfully fetched cart detail of user",
	}

	savingKeyJson, _ := json.Marshal(resp)
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,savingKeyJson, 10*time.Minute).Err()
	if rdsErr != nil{
		fmt.Println("Failed to cache cart data: ",rdsErr)
	}

	return c.JSON(resp)
}

func Cart_ModifyQuantity(c *fiber.Ctx) error{
	var cartDetail models.CartDetail
	cartID := c.QueryInt("cartID",0);
	if (cartID == 0){
		return service.SendError(c,401,"Did not receive cartID");
	}
	accountID := c.QueryInt("accountID",0);
	if (accountID == 0){
		return service.SendError(c,401,"Did not receive accountID");
	}

	//Parsing request body into struct
	if err := c.BodyParser(&cartDetail); err != nil{
		return service.SendError(c,400,"Invalid request body");
	}

	update,err := models.FindCartDetail(c.Context(),boil.GetContextDB(),cartID)
	if err != nil{
		return service.SendError(c,500,"Cart detail not found");
	}
	update.Quantity = cartDetail.Quantity;
	_, err = update.Update(c.Context(),boil.GetContextDB(),boil.Infer())
	if err != nil {
		return service.SendError(c, 500, "Failed to update cart quantity")
	}

	//deleting cart cache after modfication
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID);
	cachedCart,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		rdsErr := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if rdsErr != nil {
			fmt.Println("Failed to clear cache:", rdsErr)
			fmt.Println("Cached cart:", cachedCart)
		}
	}

	resp := fiber.Map{
		"status":"Success",
		"data": update,
		"message": "Successfully modified quantity",
	}
	return c.JSON(resp)
}

func DeleteCartItem(c *fiber.Ctx) error{
	cartID := c.QueryInt("cartID",0);
	if cartID == 0{
		return service.SendError(c,401,"Did not receive cartID");
	}
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,401,"Did not receive accountID");
	}

	cartItem,err := models.CartDetails(qm.Where("\"cartID\" = ?",cartID)).One(c.Context(),boil.GetContextDB()) 
	if err != nil {
		return service.SendError(c,500,"No cart item found");
	}

	_,err = cartItem.Delete(c.Context(),boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, "Failed to delete cart item")
	}

	//deleting cart cache after deletion
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID);
	fmt.Println("redis key: ",redisKey);
	cachedCart,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		rdsErr := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if rdsErr != nil {
			fmt.Println("Failed to clear cache:", rdsErr)
			fmt.Println("Cached cart:", cachedCart)
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": cartItem,
		"message": "Successfully deleted a cart item",
	}
	return c.JSON(resp);
}

func AddToCart(c *fiber.Ctx) error{
	var cartDetail models.CartDetail
	if err := c.BodyParser(&cartDetail); err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	//deleting cart cache after insertion
	redisKey := fmt.Sprintf("cart:accountID=%d:",cartDetail.AccountID);
	_, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result()
	if err == nil{
		rdsErr := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if rdsErr != nil {
			fmt.Println("Failed to clear cache:", rdsErr)
		}
	}

	//checking if a product has already existed in the cart
	check,err := models.CartDetails(
		qm.Where("\"productID\" = ? AND \"accountID\" = ?",cartDetail.ProductID,cartDetail.AccountID),
		qm.Load(models.CartDetailRels.ProductIDProduct),
	).One(c.Context(),boil.GetContextDB())
	if err != nil{
		if errors.Is(err,sql.ErrNoRows){
			if err := cartDetail.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
				return service.SendError(c,500,"Couldn't insert new cart detail");
			}

			// Load the inserted row again
			insertedCart, err := models.CartDetails(
				qm.Where("\"productID\" = ? AND \"accountID\" = ?", cartDetail.ProductID, cartDetail.AccountID),
				qm.Load(models.CartDetailRels.ProductIDProduct),
			).One(c.Context(), boil.GetContextDB())
			if err != nil {
				return service.SendError(c, 500, "Couldn't retrieve newly inserted cart detail")
			}

			response := CartDetailResponse{
				CartDetail: *insertedCart,
				Product: insertedCart.R.ProductIDProduct,
			}

			resp := fiber.Map{
				"status": "Success",
				"data": response,
				"message": "Successfully added product to cart",
			}
			return c.JSON(resp);
		}
		return service.SendError(c,500,"Database error");
	}
	check.Quantity += cartDetail.Quantity
	if _,err = check.Update(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,"Couldn't modify item's quantity");
	}
	
	resp := fiber.Map{
		"status": "Success",
		"data": check,
		"message": "Successfully added product to cart",
	}

	return c.JSON(resp);
}

func FetchCart(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,401,"Did not receive accountID");
	}
	cartDetail,err := models.CartDetails(
		qm.Where("\"accountID\" = ?",accountID),
		qm.Load(models.CartDetailRels.ProductIDProduct),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Cart detail not found");
	}

	response := make([]CartDetailResponse,len(cartDetail));
	for i,cart := range cartDetail{
		response[i] = CartDetailResponse{
			CartDetail: *cart,
			Product: cart.R.ProductIDProduct,
		}
	}

	resp := fiber.Map{
		"status":"Success",
		"data": response,
		"message": "Successfully fetched user's cart",
	}
	return c.JSON(resp);
}

func DeleteAllItems(c *fiber.Ctx) error{
	accountID := c.QueryInt("accountID",0);

	cartToDelete,err := models.CartDetails(
		qm.Where("\"accountID\" = ?",accountID),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,401,"Did not receive accountID");
	}

	_,err = cartToDelete.DeleteAll(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Couldn't delete cart items of the provided accountID");
	}

	//deleting cart cache after deletion of all items
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID);
	fmt.Println("redis key: ",redisKey);
	cachedCart,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		rdsErr := redisdatabase.Client.Del(redisdatabase.Ctx,redisKey).Err();
		if rdsErr != nil {
			fmt.Println("Failed to clear cache:", rdsErr)
			fmt.Println("Cached cart:", cachedCart)
		}
	}

	resp := fiber.Map{
		"status":"Success",
		"message": "Successfully deleted all cart items",
	}
	return c.JSON(resp);
}