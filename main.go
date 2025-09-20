package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TOmorrowarc1/ClassSelectionSystem/account"
	"github.com/TOmorrowarc1/ClassSelectionSystem/course"
	"github.com/TOmorrowarc1/ClassSelectionSystem/privilege"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
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
	privilege.InitPrivilegeSystem()
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

	var privilege_level int
	if req.Action != "LogIn" {
		privilege_level, err = privilege.UserAccess(req.Token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	}

	switch req.Action {
	case "Register":
		HandleRegister(w, req.Parameters, privilege_level)
	case "Remove":
		HandleRemove(w, req.Parameters, privilege_level)
	case "LogIn":
		HandleLogIn(w, req.Parameters)
	case "LogOut":
		HandleLogOut(w, req.Token)
	case "ModifyPassword":
		HandleModifyPassword(w, req.Parameters)
	case "GetUserInfo":
		HandleGetUserInfo(w, req.Parameters, privilege_level)
	case "GetAllUsersInfo":
		HandleGetAllUsersInfo(w, req.Parameters, privilege_level)
	case "GetPartUsersInfo":
		HandleGetPartUsersInfo(w, req.Parameters, privilege_level)
	case "AddCourse":
		HandleAddCourse(w, req.Parameters, privilege_level)
	case "ModifyCourse":
		HandleModifyCourse(w, req.Parameters, privilege_level)
	case "LaunchCourse":
		HandleLaunchCourse(w, req.Parameters, privilege_level)
	case "GetAllCoursesInfo":
		HandleGetAllCoursesInfo(w, req.Parameters)
	case "SelectCourse":
		HandleSelectCourse(w, req.Parameters, privilege_level)
	case "DropCourse":
		HandleDropCourse(w, req.Parameters, privilege_level)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

// Typical logic for my work handler: (check privilege), decode parameters, execute the corresponding function and write back http reponse.
func HandleRegister(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		User_name     string `json:"username"`
		Password      string `json:"password"`
		Identity_info struct {
			Class struct {
				Grade int `json:"grade"`
				Class int `json:"class"`
			}
			Privilege string `json:"privilege"`
		}
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = account.Register(params.User_name, params.Password, params.Identity_info.Class.Grade, params.Identity_info.Class.Class, params.Identity_info.Privilege)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleRemove(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		User_name string `json:"username"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = account.RemoveUser(params.User_name)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleLogIn(w http.ResponseWriter, parameters json.RawMessage) {
	type Parameters struct {
		User_name string `json:"name"`
		Password  string `json:"password"`
	}
	type Response struct {
		Token   string `json:"authToken"`
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	var params Parameters
	err := json.Unmarshal(parameters, &params)
	if err != nil {
		response.Message = "Invalid parameters"
	} else {
		privilege_level, err := account.LogIn(params.User_name, params.Password)
		if err != nil {
			response.Message = err.Error()
		} else {
			response.Token = privilege.UserLogIn(privilege_level)
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleLogOut(w http.ResponseWriter, token string) {
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	err := privilege.UserLogOut(token)
	if err != nil {
		response.Message = err.Error()
	}
	json.NewEncoder(w).Encode(response)
}

func HandleModifyPassword(w http.ResponseWriter, parameters json.RawMessage) {
	type Parameters struct {
		User_name string `json:"name"`
		Password  string `json:"password"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	var params Parameters
	err := json.Unmarshal(parameters, &params)
	if err != nil {
		response.Message = "Invalid parameters"
	} else {
		err = account.ModifyPassword(params.User_name, params.Password)
		if err != nil {
			response.Message = err.Error()
		}
	}
	json.NewEncoder(w).Encode(response)
}

type UserInfo struct {
	User_name     string `json:"name"`
	Password      string `json:"password"`
	Identity_info struct {
		Class struct {
			Grade int `json:"grade"`
			Class int `json:"class"`
		}
		Privilege string `json:"privilege"`
	}
}

func userInfoConstruct(user_info *account.UserInfo) UserInfo {
	var user UserInfo
	user.User_name = user_info.Uid_
	user.Password = user_info.Password_
	user.Identity_info.Class.Grade = user_info.Class_id_.Grade_
	user.Identity_info.Class.Class = user_info.Class_id_.Class_
	user.Identity_info.Privilege = account.PrivilegeToString(user_info.Privilege_)
	return user
}

func HandleGetUserInfo(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		User_name string `json:"name"`
	}
	type Response struct {
		User_info UserInfo
		Message   string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_TEACHER {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			user_info, err := account.GetUserInfo(params.User_name)
			if err != nil {
				response.Message = err.Error()
			} else {
				user := userInfoConstruct(user_info)
				response.User_info = user
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleGetAllUsersInfo(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Response struct {
		Users   []UserInfo `json:"users"`
		Message string     `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		all_users := account.GetAllUsersInfo()
		for _, user_info := range all_users {
			user := userInfoConstruct(user_info)
			response.Users = append(response.Users, user)
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleGetPartUsersInfo(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		Way   int `json:"way"` // 0 for class, 1 for course
		Class struct {
			Grade int `json:"grade"`
			Class int `json:"class"`
		}
		Course_id string `json:"courseName"`
	}

	type Response struct {
		Users   []UserInfo `json:"users"`
		Message string     `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_TEACHER {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			if params.Way == 0 {
				users, err := account.GetClassUsersInfo(params.Class.Grade, params.Class.Class)
				if err != nil {
					response.Message = err.Error()
				} else {
					for _, user_info := range users {
						user := userInfoConstruct(user_info)
						response.Users = append(response.Users, user)
					}
				}
			} else if params.Way == 1 {
				users, err := account.GetCourseUsersInfo(params.Course_id)
				if err != nil {
					response.Message = err.Error()
				} else {
					for _, user_info := range users {
						user := userInfoConstruct(user_info)
						response.Users = append(response.Users, user)
					}
				}
			} else {
				response.Message = "Invalid way"
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

type CourseInfo struct {
	Course_name  string `json:"name"`
	Teacher_name string `json:"teacherName"`
	Max_student  int    `json:"maximum"`
}

func HandleAddCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		Course_Info CourseInfo `json:"courseInfo"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.AddCourse(params.Course_Info.Course_name, params.Course_Info.Teacher_name, params.Course_Info.Max_student)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleModifyCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		Course_Name string     `json:"courseName"`
		Course_Info CourseInfo `json:"courseInfo"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.ModifyCourse(params.Course_Name, params.Course_Info.Course_name, params.Course_Info.Teacher_name, params.Course_Info.Max_student)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleLaunchCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		Course_Name string `json:"courseName"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PRIVILEGE_ADMIN {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.LaunchCourse(params.Course_Name)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

type CourseFullInfo struct {
	Course_name  string `json:"name"`
	Teacher_name string `json:"teacherName"`
	Max_students int    `json:"maximum"`
	Now_students int    `json:"current"`
	Lanuched     bool   `json:"launched"`
}

func courseFullInfoConstruct(course_info *course.CourseInfo) CourseFullInfo {
	var course CourseFullInfo
	course.Course_name = course_info.Course_name
	course.Teacher_name = course_info.Teacher
	course.Max_students = course_info.Max_students
	course.Now_students = course_info.Now_students
	course.Lanuched = course_info.Lanuched
	return course
}
func HandleGetAllCoursesInfo(w http.ResponseWriter, parameters json.RawMessage) {
	type Response struct {
		Courses []CourseFullInfo `json:"courses"`
		Message string           `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	all_courses := course.GetAllCoursesInfo()
	for _, course_info := range all_courses {
		course_full_info := courseFullInfoConstruct(course_info)
		response.Courses = append(response.Courses, course_full_info)
	}
	json.NewEncoder(w).Encode(response)
}

func HandleSelectCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		Course_Name string `json:"courseName"`
		UserName    string `json:"name"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege != account.PRIVILEGE_STUDENT {
		response.Message = "Only students can select courses"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.SelectCourse(params.UserName, params.Course_Name)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleDropCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		UserName string `json:"name"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}
	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege != account.PRIVILEGE_STUDENT {
		response.Message = "Only students can drop courses"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.DropCourse(params.UserName)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}
