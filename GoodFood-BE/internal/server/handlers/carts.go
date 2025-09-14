package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

//GetCartDetail returns cart details (fetch from redis cache or db) and save cart details into cache afterward.
func GetCartDetail(c *fiber.Ctx) error{
	//Fetch query param
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID");
	}

	//Create redis key after accountID
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID);
	//Fetch cache
	cachedCart := fiber.Map{}
	if ok, _ := utils.GetCache(redisKey,&cachedCart); ok{
		return c.JSON(cachedCart)
	}

	//Fetch DB
	cartDetails, err := models.CartDetails(
		qm.Where("\"accountID\" = ?", accountID),
		qm.Load(models.CartDetailRels.ProductIDProduct), // Load Products
	).All(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, "Cart detail not found")
	}

	// Converting data into custom struct that has both cart details and products
	response := utils.BuildCartResponse(cartDetails);

	resp := fiber.Map{
		"status":  "Success",
		"data":    response,
		"message": "Successfully fetched cart detail of user",
	}

	//Save into cache
	utils.SetCache(redisKey,resp,10*time.Minute,"");

	return c.JSON(resp)
}

//Cart_ModifyQuantity modifies the quantities of products within Cart Details
func Cart_ModifyQuantity(c *fiber.Ctx) error{
	//Fetch query params
	cartID := c.QueryInt("cartID",0);
	if (cartID == 0){
		return service.SendError(c,401,"Did not receive cartID");
	}
	accountID := c.QueryInt("accountID",0);
	if (accountID == 0){
		return service.SendError(c,401,"Did not receive accountID");
	}

	//Parse request body into struct
	var cartDetail models.CartDetail
	if err := c.BodyParser(&cartDetail); err != nil{
		return service.SendError(c,400,"Invalid request body");
	}

	update,err := models.FindCartDetail(c.Context(),boil.GetContextDB(),cartID)
	if err != nil{
		return service.SendError(c,500,"Cart detail not found");
	}

	//If found, update the quantity
	update.Quantity = cartDetail.Quantity;
	if _, err = update.Update(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil {
		return service.SendError(c, 500, "Failed to update cart quantity")
	}

	//Clear cache after mutation
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID)
	utils.ClearCache(redisKey);
	
	resp := fiber.Map{
		"status":"Success",
		"data": update,
		"message": "Successfully modified quantity",
	}
	return c.JSON(resp)
}

//DeleteCartItem deletes one item in cart
func DeleteCartItem(c *fiber.Ctx) error{
	//Fetch query params
	cartID := c.QueryInt("cartID",0);
	if cartID == 0{
		return service.SendError(c,401,"Did not receive cartID");
	}
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,401,"Did not receive accountID");
	}

	//Fetch db
	cartItem,err := models.FindCartDetail(c.Context(),boil.GetContextDB(),cartID);
	if err != nil {
		return service.SendError(c,500,"No cart item found");
	}

	//If found, delete cart item
	if _,err = cartItem.Delete(c.Context(),boil.GetContextDB()); err != nil {
		return service.SendError(c, 500, "Failed to delete cart item")
	}

	//Clear cache after mutation
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID)
	utils.ClearCache(redisKey);

	resp := fiber.Map{
		"status": "Success",
		"data": cartItem,
		"message": "Successfully deleted a cart item",
	}
	return c.JSON(resp);
}

//AddToCart adds a product into the customer's cart
func AddToCart(c *fiber.Ctx) error{
	var cartDetail models.CartDetail
	if err := c.BodyParser(&cartDetail); err != nil{
		return service.SendError(c,400,"Invalid body");
	}

	//Clear cache after mutation
	redisKey := fmt.Sprintf("cart:accountID=%d:",cartDetail.AccountID)
	utils.ClearCache(redisKey);

	//Check if the product has already existed in the cart
	check,err := models.CartDetails(
		qm.Where("\"productID\" = ? AND \"accountID\" = ?",cartDetail.ProductID,cartDetail.AccountID),
		qm.Load(models.CartDetailRels.ProductIDProduct),
	).One(c.Context(),boil.GetContextDB())
	if err != nil{
		if errors.Is(err,sql.ErrNoRows){
			//If not found, insert the product into cart
			if err := cartDetail.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
				return service.SendError(c,500,"Couldn't insert new cart detail");
			}

			// Load Products with the inserted row again for api response
			insertedCart, err := models.CartDetails(
				qm.Where("\"productID\" = ? AND \"accountID\" = ?", cartDetail.ProductID, cartDetail.AccountID),
				qm.Load(models.CartDetailRels.ProductIDProduct),
			).One(c.Context(), boil.GetContextDB())
			if err != nil {
				return service.SendError(c, 500, "Couldn't retrieve newly inserted cart detail")
			}

			response := dto.CartDetailResponse{
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

	//If found, update the quantity
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

//FetchCart returns all cart items to the requested user.
func FetchCart(c *fiber.Ctx) error{
	//Fetch query param
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,401,"Did not receive accountID");
	}

	//Fetch db
	cartDetail,err := models.CartDetails(
		qm.Where("\"accountID\" = ?",accountID),
		qm.Load(models.CartDetailRels.ProductIDProduct),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Cart detail not found");
	}

	// Convert data into custom struct that has both cart details and products
	response := utils.BuildCartResponse(cartDetail);

	resp := fiber.Map{
		"status":"Success",
		"data": response,
		"message": "Successfully fetched user's cart",
	}
	return c.JSON(resp);
}

//DeleteAllItems deletes the entire cart of the user.
func DeleteAllItems(c *fiber.Ctx) error{
	//Fetch query param
	accountID := c.QueryInt("accountID",0);

	//Fetch all cart details for deletion
	cartToDelete,err := models.CartDetails(
		qm.Where("\"accountID\" = ?",accountID),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,401,"Did not receive accountID");
	}

	//Delete
	if _,err = cartToDelete.DeleteAll(c.Context(),boil.GetContextDB()); err != nil{
		return service.SendError(c,500,"Couldn't delete cart items of the provided accountID");
	}

	//Clear cache after mutation
	redisKey := fmt.Sprintf("cart:accountID=%d:",accountID)
	utils.ClearCache(redisKey);

	resp := fiber.Map{
		"status":"Success",
		"message": "Successfully deleted all cart items",
	}
	return c.JSON(resp);
}