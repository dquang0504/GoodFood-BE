package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var userConnections = make(map[int]*websocket.Conn) // for user
var adminConnections = make(map[int]*websocket.Conn)

type ChatMessage struct{
	FromID int `json:"from_id"`
	ToID int `json:"to_id"`
	Sender string `json:"sender"` //user or admin
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

// Handler WebSocket for user
func HandleUserWebsocket(c *websocket.Conn){
	accountID, err := strconv.Atoi(c.Params("accountID"));
	if err != nil || accountID <= 0{
		log.Println("Invalid accountID from user WebSocket")
		_ = c.WriteMessage(websocket.TextMessage,[]byte("Invalid accountID"))
		_ = c.Close()
		return
	}
	userConnections[accountID] = c
	defer func(){
		delete(userConnections, accountID)
		c.Close()
	}()

	log.Printf("User %d connected via WebSocket\n",accountID)
	
	for{
		_, msg, err := c.ReadMessage()
		if err != nil{
			log.Printf("User %d disconnected: %v\n",accountID, err)
			break;
		}

		//Send message to admins
		for _, adminConn := range adminConnections{
			fmt.Println(websocket.TextMessage);
			_ = adminConn.WriteMessage(websocket.TextMessage, msg)
		}
	}

}

func HandleAdminWebSocket(c *websocket.Conn){
	adminID, err := strconv.Atoi(c.Params("adminID"))
	if err != nil || adminID <= 0{
		log.Println("Invalid adminID from admin websocket")
		_ = c.WriteMessage(websocket.TextMessage, []byte("Invalid adminID"))
		_ = c.Close()
		return
	}

	adminConnections[adminID] = c
	defer func(){
		delete(adminConnections, adminID)
		c.Close()
	}()

	log.Printf("Admin %d connected via WebSocket\n",adminID)

	for{
		_, msg, err := c.ReadMessage()
		if err != nil{
			log.Printf("Admin %d disconnected: %v\n",adminID,err)
			break;
		}

		payload := ChatMessage{}
		if err := json.Unmarshal(msg,&payload); err != nil{
			log.Println("Invalid message from admin: ",err)
			continue
		}
		fmt.Println(payload.Message);
		if userConn, ok := userConnections[payload.ToID]; ok{
			_ = userConn.WriteJSON(fiber.Map{
				"fromAdmin": adminID,
				"message": payload.Message,
			})
			fmt.Println("Not okay");
		}
	}

	// for{
	// 	_, msg, err := c.ReadMessage()
	// 	if err != nil{
	// 		log.Printf("User %d disconnected: %v\n",accountID, err)
	// 		break;
	// 	}

	// 	//Send message to admins
	// 	for _, adminConn := range adminConnections{
	// 		fmt.Println(websocket.TextMessage);
	// 		_ = adminConn.WriteMessage(websocket.TextMessage, msg)
	// 	}
	// }
}

