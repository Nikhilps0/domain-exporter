package main

import (
	"log/slog"
	"net/http"
	"os"
	"platform/domain-exporter/internal/config"
	"platform/domain-exporter/internal/middleware"
	"platform/domain-exporter/internal/server"
)

func main() {

	//Variables
	env, err := config.LoadConfig(); 
	if err !=nil{
		slog.Error("Program Exited", "Error", err)
		return
	}
	
	//Logger setup
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	slog.Info("Application Initilization Started")
	// HTTP Handler function
	slog.Info("Applicaion Start Loading", "Port", env.ServerPort)
	http.HandleFunc("/", middleware.AuditLogger(server.DefaultPath))

	// HTTP Listener
	err = http.ListenAndServe(":"+env.ServerPort, nil)

	if err != nil {
		slog.Error("Application Failed", "Error", err)
		return
	}

}
