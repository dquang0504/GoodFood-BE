package main

import (
	"GoodFood-BE/internal/jobs"
	"log"

	"github.com/hibiken/asynq"
)

//function main initializes and starts the Asynq worker server.
//Connects to Redis, configures concurrency, and register task handlers.
func main(){
	//Configuring Redis connection for Asynq
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379",Password: "",DB: 0},
		asynq.Config{
			Concurrency: 5, //maximum of 5 jobs handled concurrently
		},
	)

	//Register task handlers using a ServeMux
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeResetPasswordEmail, jobs.HandleResetPasswordEmailTask)
	mux.HandleFunc(jobs.TypeSendContactMessage,jobs.HandleContactCustomerSent)

	//Start the server and log fatal error if failed to run
	if err := srv.Run(mux); err != nil{
		log.Fatalf("Could not run asynq server: %v",err);
	}
}