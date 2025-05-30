package main

import (
	"GoodFood-BE/internal/jobs"
	"log"

	"github.com/hibiken/asynq"
)

func main(){
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379",Password: "",DB: 0},
		asynq.Config{
			Concurrency: 5, //maximum of 5 jobs handled concurrently
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeResetPasswordEmail, jobs.HandleResetPasswordEmailTask)
	if err := srv.Run(mux); err != nil{
		log.Fatalf("Could not run asynq server: %v",err);
	}
}