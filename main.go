package main

import (
	"context"
	"encoding/json"
	"github.com/TOmorrowarc1/ClassSelectionSystem/account"
	"github.com/TOmorrowarc1/ClassSelectionSystem/course"
	"github.com/TOmorrowarc1/ClassSelectionSystem/priviledge"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	system_logger *logger.Logger
)

func main() {
	system_logger = logger.GetLogger()
	system_logger.SetLogFile("system.log")
	system_logger.SetLogLevel(logger.DEBUG)

	system_logger.Log(logger.INFO, "System starting...")
	account.InitAccountSystem()
	course.InitCourseSystem()
	priviledge.InitPriviledgeSystem()
	system_logger.Log(logger.INFO, "All systems initialized.")

	mux := http.NewServeMux()
	mux.HandleFunc("/api", RequestRoute)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		system_logger.Log(logger.INFO, "Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil {
			system_logger.Log(logger.FATAL, "Failed to start server: %v", err)
		}
		system_logger.Log(logger.INFO, "Server stopped.")
	}()

	quit_channel := make(chan os.Signal, 1)
	signal.Notify(quit_channel, syscall.SIGINT, syscall.SIGTERM)

	<-quit_channel
	system_logger.Log(logger.WARN, "Shutdown signal received. Starting graceful shutdown...")
	account.StoreAccountData()
	course.StoreCourseData()
	system_logger.Log(logger.INFO, "All systems closed.")

	shutdown_ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdown_ctx)
	system_logger.Log(logger.INFO, "Server gracefully stopped.")
	system_logger.Close()
}

func RequestRoute(w http.ResponseWriter, r *http.Request) {

	type Request struct {
		Token      string          `json:"token"`
		Action     string          `json:"action"`
		Parameters json.RawMessage `json:"parameters"`
		Meta       json.RawMessage `json:"meta"`
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	var priviledge_level int
	if req.Action != "LogIn" {
		priviledge_level, err = priviledge.UserAccess(req.Token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	}

	switch req.Action {
	case "Register":
		HandleRegister(w, req.Parameters, priviledge_level)
	case "Remove":
		HandleRemove(w, req.Parameters, priviledge_level)
	case "LogIn":
		HandleLogIn(w, req.Parameters)
	case "LogOut":
		HandleLogOut(w, req.Parameters)
	case "ModifyPassword":
		HandleModifyPassword(w, req.Parameters)
	case "GetUserInfo":
		HandleGetUserInfo(w, req.Parameters, priviledge_level)
	case "GetAllUsersInfo":
		HandleGetAllUsersInfo(w, req.Parameters, priviledge_level)
	case "GetPartUsersInfo":
		HandleGetPartUsersInfo(w, req.Parameters, priviledge_level)
	case "AddCourse":
		HandleAddCourse(w, req.Parameters, priviledge_level)
	case "ModifyCourse":
		HandleModifyCourse(w, req.Parameters, priviledge_level)
	case "LaunchCourse":
		HandleLaunchCourse(w, req.Parameters, priviledge_level)
	case "GetAllCoursesInfo":
		HandleGetAllCoursesInfo(w, req.Parameters)
	case "SelectCourse":
		HandleSelectCourse(w, req.Parameters, priviledge_level)
	case "DropCourse":
		HandleDropCourse(w, req.Parameters, priviledge_level)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func HandleRegister(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleRemove(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleLogIn(w http.ResponseWriter, parameters json.RawMessage) {}

func HandleLogOut(w http.ResponseWriter, parameters json.RawMessage) {}

func HandleModifyPassword(w http.ResponseWriter, parameters json.RawMessage) {}

func HandleGetUserInfo(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleGetAllUsersInfo(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleGetPartUsersInfo(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleAddCourse(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleModifyCourse(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleLaunchCourse(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleGetAllCoursesInfo(w http.ResponseWriter, parameters json.RawMessage) {}

func HandleSelectCourse(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}

func HandleDropCourse(w http.ResponseWriter, parameters json.RawMessage, priviledge int) {}
