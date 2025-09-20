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
	system_logger.SetLogLevel(logger.Debug)

	system_logger.Log(logger.Info, "System starting...")
	account.InitAccountSystem()
	course.InitCourseSystem()
	privilege.InitPrivilegeSystem()
	system_logger.Log(logger.Info, "All systems initialized.")

	mux := http.NewServeMux()
	mux.HandleFunc("/api", RequestRoute)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		system_logger.Log(logger.Info, "Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil {
			system_logger.Log(logger.Fatal, "Failed to start server: %v", err)
		}
		system_logger.Log(logger.Info, "Server stopped.")
	}()

	quit_channel := make(chan os.Signal, 1)
	signal.Notify(quit_channel, syscall.SIGINT, syscall.SIGTERM)

	<-quit_channel
	system_logger.Log(logger.Warn, "Shutdown signal received. Starting graceful shutdown...")
	account.StoreAccountData()
	course.StoreCourseData()
	system_logger.Log(logger.Info, "All systems closed.")

	shutdown_ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdown_ctx)
	system_logger.Log(logger.Info, "Server gracefully stopped.")
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

	var accountInfo privilege.AccountInfo
	accountInfo.Privilege = -1
	if req.Action != "LogIn" {
		accountInfo, err = privilege.UserAccess(req.Token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	}

	switch req.Action {
	case "Register":
		HandleRegister(w, req.Parameters, accountInfo.Privilege)
	case "Remove":
		HandleRemove(w, req.Parameters, accountInfo.Privilege)
	case "LogIn":
		HandleLogIn(w, req.Parameters)
	case "LogOut":
		HandleLogOut(w, req.Token)
	case "ModifyPassword":
		HandleModifyPassword(w, req.Parameters, accountInfo.UserName)
	case "GetUserInfo":
		HandleGetUserInfo(w, req.Parameters, accountInfo.Privilege)
	case "GetAllUsersInfo":
		HandleGetAllUsersInfo(w, req.Parameters, accountInfo.Privilege)
	case "GetPartUsersInfo":
		HandleGetPartUsersInfo(w, req.Parameters, accountInfo.Privilege)
	case "AddCourse":
		HandleAddCourse(w, req.Parameters, accountInfo.Privilege)
	case "ModifyCourse":
		HandleModifyCourse(w, req.Parameters, accountInfo.Privilege)
	case "LaunchCourse":
		HandleLaunchCourse(w, req.Parameters, accountInfo.Privilege)
	case "GetAllCoursesInfo":
		HandleGetAllCoursesInfo(w, req.Parameters)
	case "SelectCourse":
		HandleSelectCourse(w, req.Parameters, accountInfo)
	case "DropCourse":
		HandleDropCourse(w, req.Parameters, accountInfo)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

type UserInfoJson struct {
	UserName      string `json:"username"`
	Password      string `json:"password"`
	Identity_info struct {
		Class struct {
			Grade int `json:"grade"`
			Class int `json:"class"`
		}
		Privilege string `json:"privilege"`
	}
}

func userInfoJsonConstruct(userInfo *account.UserInfo) UserInfoJson {
	var user UserInfoJson
	user.UserName = userInfo.Uid
	user.Password = userInfo.Password
	user.Identity_info.Class.Grade = userInfo.Classid.Grade
	user.Identity_info.Class.Class = userInfo.Classid.Class
	user.Identity_info.Privilege = account.PrivilegeToString(userInfo.Privilege)
	return user
}

func userInfoJsonDeconstruct(userInfo *UserInfoJson) account.UserInfo {
	var user account.UserInfo
	user.Uid = userInfo.UserName
	user.Password = userInfo.Password
	user.Classid.Grade = userInfo.Identity_info.Class.Grade
	user.Classid.Class = userInfo.Identity_info.Class.Class
	user.Privilege = account.StringToPrivilege(userInfo.Identity_info.Privilege)
	return user
}

// Typical logic for my work handler: (check privilege), decode parameters, execute the corresponding function and write back http reponse.
func HandleRegister(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		UserInfo UserInfoJson `json:"userInfo"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeAdmin {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = account.Register(userInfoJsonDeconstruct(&params.UserInfo))
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
	if privilege < account.PrivilegeAdmin {
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
		privilegeLevel, err := account.LogIn(params.User_name, params.Password)
		if err != nil {
			response.Message = err.Error()
		} else {
			accountInfo := privilege.AccountInfo{UserName: params.User_name, Privilege: privilegeLevel}
			response.Token = privilege.UserLogIn(accountInfo)
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

func HandleModifyPassword(w http.ResponseWriter, parameters json.RawMessage, uid string) {
	type Parameters struct {
		Password string `json:"password"`
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
		err = account.ModifyPassword(uid, params.Password)
		if err != nil {
			response.Message = err.Error()
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleGetUserInfo(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		UserName string `json:"name"`
	}
	type Response struct {
		UserInfo UserInfoJson `json:"userInfo"`
		Message  string       `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeTeacher {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			userInfo, err := account.GetUserInfo(params.UserName)
			if err != nil {
				response.Message = err.Error()
			} else {
				response.UserInfo = userInfoJsonConstruct(userInfo)
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleGetAllUsersInfo(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Response struct {
		Users   []UserInfoJson `json:"users"`
		Message string         `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeAdmin {
		response.Message = "Permission denied"
	} else {
		allUsers := account.GetAllUsersInfo()
		for _, userInfo := range allUsers {
			response.Users = append(response.Users, userInfoJsonConstruct(userInfo))
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
		Users   []UserInfoJson `json:"users"`
		Message string         `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeTeacher {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			if params.Way == 0 {
				classid := account.ClassID{Grade: params.Class.Grade, Class: params.Class.Class}
				users, err := account.GetClassUsersInfo(classid)
				if err != nil {
					response.Message = err.Error()
				} else {
					for _, userInfo := range users {
						user := userInfoJsonConstruct(userInfo)
						response.Users = append(response.Users, user)
					}
				}
			} else if params.Way == 1 {
				users, err := account.GetCourseUsersInfo(params.Course_id)
				if err != nil {
					response.Message = err.Error()
				} else {
					for _, userInfo := range users {
						user := userInfoJsonConstruct(userInfo)
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
	CourseName  string `json:"name"`
	TeacherName string `json:"teacherName"`
	Max_student int    `json:"maximum"`
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
	if privilege < account.PrivilegeAdmin {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.AddCourse(params.Course_Info.CourseName, params.Course_Info.TeacherName, params.Course_Info.Max_student)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleModifyCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		CourseName string     `json:"courseName"`
		CourseInfo CourseInfo `json:"courseInfo"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeAdmin {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.ModifyCourse(params.CourseName, params.CourseInfo.TeacherName, params.CourseInfo.Max_student)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleLaunchCourse(w http.ResponseWriter, parameters json.RawMessage, privilege int) {
	type Parameters struct {
		CourseName string `json:"courseName"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if privilege < account.PrivilegeAdmin {
		response.Message = "Permission denied"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.LaunchCourse(params.CourseName)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

type CourseFullInfo struct {
	CourseName  string `json:"name"`
	TeacherName string `json:"teacherName"`
	MaxStudents int    `json:"maximum"`
	NowStudents int    `json:"current"`
	Lanuched    bool   `json:"launched"`
}

func courseFullInfoConstruct(course_info *course.CourseInfo) CourseFullInfo {
	var course CourseFullInfo
	course.CourseName = course_info.CourseName
	course.TeacherName = course_info.Teacher
	course.MaxStudents = course_info.MaxStudents
	course.NowStudents = course_info.NowStudents
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

func HandleSelectCourse(w http.ResponseWriter, parameters json.RawMessage, accountInfo privilege.AccountInfo) {
	type Parameters struct {
		CourseName string `json:"courseName"`
	}
	type Response struct {
		Message string `json:"errorMessage"`
	}

	w.Header().Set("Content-Type", "application/json")
	var response Response
	if accountInfo.Privilege != account.PrivilegeStudent {
		response.Message = "Only students can select courses"
	} else {
		var params Parameters
		err := json.Unmarshal(parameters, &params)
		if err != nil {
			response.Message = "Invalid parameters"
		} else {
			err = course.SelectCourse(accountInfo.UserName, params.CourseName)
			if err != nil {
				response.Message = err.Error()
			}
		}
	}
	json.NewEncoder(w).Encode(response)
}

func HandleDropCourse(w http.ResponseWriter, parameters json.RawMessage, accountInfo privilege.AccountInfo) {
	type Response struct {
		Message string `json:"errorMessage"`
	}
	w.Header().Set("Content-Type", "application/json")
	var response Response
	if accountInfo.Privilege != account.PrivilegeStudent {
		response.Message = "Only students can drop courses"
	} else {
		err := course.DropCourse(accountInfo.UserName)
		if err != nil {
			response.Message = err.Error()
		}
	}
	json.NewEncoder(w).Encode(response)
}
