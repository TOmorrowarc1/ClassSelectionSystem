package account

import (
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/course"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

type ClassID struct {
	Grade int
	Class int
}

type UserInfo struct {
	Uid       string
	Password  string
	Classid   ClassID
	Privilege int
}

var (
	userInfoMap   *concurrentmap.ConcurrentMap[string, UserInfo] // uid -> UserInfo
	classUserMap  *concurrentmap.ConcurrentMap[ClassID, *concurrentmap.ConcurrentMap[string, struct{}]]
	accountLogger *logger.Logger
)

const (
	userInfoPath  = "data/userInfo.json"
	classUserPath = "data/classUser.json"
)

const (
	PrivilegeStudent = 0
	PrivilegeTeacher = 1
	PrivilegeAdmin   = 2
)

func stringToPrivilege(privilege string) int {
	switch privilege {
	case "student":
		return PrivilegeStudent
	case "teacher":
		return PrivilegeTeacher
	case "admin":
		return PrivilegeAdmin
	default:
		return PrivilegeStudent
	}
}

func PrivilegeToString(privilege int) string {
	switch privilege {
	case PrivilegeStudent:
		return "student"
	case PrivilegeTeacher:
		return "teacher"
	case PrivilegeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func InitAccountSystem() {
	userInfoMap = concurrentmap.NewConcurrentMap[string, UserInfo]()
	classUserMap = concurrentmap.NewConcurrentMap[ClassID, *concurrentmap.ConcurrentMap[string, struct{}]]()
	accountLogger = logger.GetLogger()
	userInfoMap.Load(userInfoPath)
	if _, ok := userInfoMap.ReadPair("admin"); !ok {
		admin_info := &UserInfo{
			Uid:      "admin",
			Password: "123456",
			Classid:  ClassID{Grade: 0, Class: 0},
		}
		userInfoMap.WritePair("admin", admin_info)
	}
	classUserMap.Load(classUserPath)
	accountLogger.Log(logger.Info, "Account system initialized")
}

func StoreAccountData() {
	err := userInfoMap.Store(userInfoPath)
	if err != nil {
		accountLogger.Log(logger.Error, "Failed to store user info: %v", err)
	}
	err = classUserMap.Store(classUserPath)
	if err != nil {
		accountLogger.Log(logger.Error, "Failed to store class user info: %v", err)
	}
	accountLogger.Log(logger.Info, "Account data stored successfully")
}

func Register(userInfo UserInfo) error {
	if _, ok := userInfoMap.ReadPair(userInfo.Uid); ok {
		accountLogger.Log(logger.Warn, "Registration failed: User %s already exists", userInfo.Uid)
		return fmt.Errorf("user %s already exists", userInfo.Uid)
	}
	userInfoMap.WritePair(userInfo.Uid, &userInfo)
	classMap, ok := classUserMap.ReadPair(userInfo.Classid)
	if !ok {
		classMap = concurrentmap.NewConcurrentMap[string, struct{}]()
	}
	classMap.WritePair(userInfo.Uid, &struct{}{})
	classUserMap.WritePair(userInfo.Classid, &classMap)
	return nil
}

func RemoveUser(uid string) error {
	userInfo, ok := userInfoMap.ReadPair(uid)
	if !ok {
		accountLogger.Log(logger.Warn, "Removal failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	userInfoMap.DeletePair(uid)
	classid := userInfo.Classid
	classMap, ok := classUserMap.ReadPair(classid)
	if !ok {
		accountLogger.Log(logger.Error, "Inconsistent state: Class %v for user %s does not exist", classid, uid)
		return fmt.Errorf("inconsistent state: class %v for user %s does not exist", classid, uid)
	}
	classMap.DeletePair(uid)
	classUserMap.WritePair(classid, &classMap)
	return nil
}

func LogIn(uid string, password string) (int, error) {
	userInfo, ok := userInfoMap.ReadPair(uid)
	if !ok {
		accountLogger.Log(logger.Warn, "Login failed: User %s does not exist", uid)
		return 0, fmt.Errorf("user %s does not exist", uid)
	}
	if userInfo.Password != password {
		accountLogger.Log(logger.Warn, "Login failed: Incorrect password for user %s", uid)
		return 0, fmt.Errorf("incorrect password for user %s", uid)
	}
	accountLogger.Log(logger.Info, "User %s logged in successfully", uid)
	return userInfo.Privilege, nil
}

func ModifyPassword(uid string, newPassword string) error {
	userInfo, ok := userInfoMap.ReadPair(uid)
	if !ok {
		accountLogger.Log(logger.Warn, "Password modification failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	userInfo.Password = newPassword
	userInfoMap.WritePair(uid, &userInfo)
	accountLogger.Log(logger.Info, "Password for user %s modified successfully", uid)
	return nil
}

func GetUserInfo(uid string) (*UserInfo, error) {
	userInfo, ok := userInfoMap.ReadPair(uid)
	if !ok {
		accountLogger.Log(logger.Warn, "GetUserInfo failed: User %s does not exist", uid)
		return nil, fmt.Errorf("user %s does not exist", uid)
	}
	return &userInfo, nil
}

func GetClassUsersInfo(classid ClassID) ([]*UserInfo, error) {
	classMap, ok := classUserMap.ReadPair(classid)
	if !ok {
		accountLogger.Log(logger.Warn, "GetClassUsers failed: Class %v does not exist", classid)
		return nil, fmt.Errorf("class %v does not exist", classid)
	}
	users_names := classMap.ReadAll()
	result := make([]*UserInfo, 0, len(users_names))
	for uid := range users_names {
		userInfo, ok := userInfoMap.ReadPair(uid)
		if ok {
			result = append(result, &userInfo)
		} else {
			accountLogger.Log(logger.Error, "Inconsistent state: User %s in class %v does not exist", uid, classid)
			return nil, fmt.Errorf("inconsistent state: user %s in class %v does not exist", uid, classid)
		}
	}
	return result, nil
}

func GetCourseUsersInfo(uid string) ([]*UserInfo, error) {
	course_id := course.GetCourseUsers(uid)
	result := make([]*UserInfo, 0, len(course_id))
	for _, cid := range course_id {
		userInfo, ok := userInfoMap.ReadPair(cid)
		if ok {
			result = append(result, &userInfo)
		} else {
			accountLogger.Log(logger.Error, "Inconsistent state: User %s in course %s does not exist", cid, course_id)
			return nil, fmt.Errorf("inconsistent state: user %s in course %s does not exist", cid, course_id)
		}
	}
	return result, nil
}

func GetAllUsersInfo() []*UserInfo {
	all_users := userInfoMap.ReadAll()
	result := make([]*UserInfo, 0, len(all_users))
	for _, userInfo := range all_users {
		result = append(result, &userInfo)
	}
	return result
}
