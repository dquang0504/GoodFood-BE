package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var(
	userConnections = make(map[int]*websocket.Conn)
	adminConnections = make(map[int]*websocket.Conn)

	//Separate mutexes for users and admins to increase performance when reading/writing
	userConnMutex sync.RWMutex
	adminConnMutex sync.RWMutex
)

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

	//Mutex lock to avoid race condition because userConnections is not thread-safe.
	//Not thread-safe means if 2 users or more goroutines shouldn't try to write to the map concurrently
	userConnMutex.Lock()
	userConnections[accountID] = c
	userConnMutex.Unlock()
	defer func(){
		userConnMutex.Lock()
		delete(userConnections, accountID)
		userConnMutex.Unlock()
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
		adminConnMutex.RLock()
		for _, adminConn := range adminConnections{
			fmt.Println(websocket.TextMessage);
			_ = adminConn.WriteMessage(websocket.TextMessage, msg)
		}
		adminConnMutex.RUnlock()
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

	//Mutex lock to avoid race condition because adminConnections is not thread-safe.
	//Not thread-safe means if 2 users or more goroutines shouldn't try to write to the map concurrently
	adminConnMutex.Lock()
	adminConnections[adminID] = c
	adminConnMutex.Unlock()
	defer func(){
		adminConnMutex.Lock()
		delete(adminConnections, adminID)
		adminConnMutex.Unlock()
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
		
		//Send to the correct user
		userConnMutex.RLock()
		if userConn, ok := userConnections[payload.ToID]; ok{
			_ = userConn.WriteJSON(fiber.Map{
				"fromAdmin": adminID,
				"message": payload.Message,
			})
			fmt.Println("Not okay");
		}
		userConnMutex.RUnlock()
	}

}

