package account

import (
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

type ClassID struct {
	grade_ int
	class_ int
}

type UserInfo struct {
	uid_       string
	password_  string
	class_id_  ClassID
	priviledge int
}

var (
	user_info_map  *concurrentmap.ConcurrentMap[string, UserInfo] // uid -> UserInfo
	class_user_map *concurrentmap.ConcurrentMap[ClassID, *concurrentmap.ConcurrentMap[string, struct{}]]
	logger_        *logger.Logger
)

const (
	USERINFOPATH  = "data/user_info.json"
	CLASSUSERPATH = "data/class_user.json"
)

const (
	PRIVILEDGE_STUDENT = 0
	PRIVILEDGE_TEACHER = 1
	PRIVILEDGE_ADMIN   = 2
)

func stringToPriviledge(priv string) int {
	switch priv {
	case "student":
		return PRIVILEDGE_STUDENT
	case "teacher":
		return PRIVILEDGE_TEACHER
	case "admin":
		return PRIVILEDGE_ADMIN
	default:
		return PRIVILEDGE_STUDENT
	}
}

func InitAccountSystem() {
	user_info_map = concurrentmap.NewConcurrentMap[string, UserInfo]()
	class_user_map = concurrentmap.NewConcurrentMap[ClassID, *concurrentmap.ConcurrentMap[string, struct{}]]()
	logger_ = logger.GetLogger()
	user_info_map.Load(USERINFOPATH)
	if _, ok := user_info_map.ReadPair("admin"); !ok {
		admin_info := &UserInfo{
			uid_:      "admin",
			password_: "123456",
			class_id_: ClassID{grade_: 0, class_: 0},
		}
		user_info_map.WritePair("admin", admin_info)
	}
	class_user_map.Load(CLASSUSERPATH)
	logger_.Log(logger.INFO, "Account system initialized")
}

func StoreAccountData() error {
	err := user_info_map.Store(USERINFOPATH)
	if err != nil {
		logger_.Log(logger.ERROR, "Failed to store user info: %v", err)
		return err
	}
	err = class_user_map.Store(CLASSUSERPATH)
	if err != nil {
		logger_.Log(logger.ERROR, "Failed to store class user info: %v", err)
		return err
	}
	logger_.Log(logger.INFO, "Account data stored successfully")
	return nil
}

func Register(uid string, password string, grade int, class int, priviledge string) error {
	if _, ok := user_info_map.ReadPair(uid); ok {
		logger_.Log(logger.WARN, "Registration failed: User %s already exists", uid)
		return fmt.Errorf("user %s already exists", uid)
	}
	new_user := UserInfo{
		uid_:       uid,
		password_:  password,
		class_id_:  ClassID{grade_: grade, class_: class},
		priviledge: stringToPriviledge(priviledge),
	}
	user_info_map.WritePair(uid, &new_user)
	class_id := ClassID{grade_: grade, class_: class}
	class_map, ok := class_user_map.ReadPair(class_id)
	if !ok {
		class_map = concurrentmap.NewConcurrentMap[string, struct{}]()
	}
	class_map.WritePair(uid, &struct{}{})
	class_user_map.WritePair(class_id, &class_map)
	return nil
}

func RemoveUser(uid string) error {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		logger_.Log(logger.WARN, "Removal failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	user_info_map.DeletePair(uid)
	class_id := user_info.class_id_
	class_map, ok := class_user_map.ReadPair(class_id)
	if !ok {
		logger_.Log(logger.ERROR, "Inconsistent state: Class %v for user %s does not exist", class_id, uid)
		return fmt.Errorf("inconsistent state: class %v for user %s does not exist", class_id, uid)
	}
	class_map.DeletePair(uid)
	class_user_map.WritePair(class_id, &class_map)
	return nil
}

func LogIn(uid string, password string) (bool, error) {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		logger_.Log(logger.WARN, "Login failed: User %s does not exist", uid)
		return false, fmt.Errorf("user %s does not exist", uid)
	}
	if user_info.password_ != password {
		logger_.Log(logger.WARN, "Login failed: Incorrect password for user %s", uid)
		return false, fmt.Errorf("incorrect password for user %s", uid)
	}
	logger_.Log(logger.INFO, "User %s logged in successfully", uid)
	return true, nil
}

func ModifyPassword(uid string, new_password string) error {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		logger_.Log(logger.WARN, "Password modification failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	user_info.password_ = new_password
	user_info_map.WritePair(uid, &user_info)
	logger_.Log(logger.INFO, "Password for user %s modified successfully", uid)
	return nil
}

func GetUserInfo(uid string) (*UserInfo, error) {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		logger_.Log(logger.WARN, "GetUserInfo failed: User %s does not exist", uid)
		return nil, fmt.Errorf("user %s does not exist", uid)
	}
	return &user_info, nil
}

func GetClassUsersInfo(grade int, class int) ([]*UserInfo, error) {
	class_id := ClassID{grade_: grade, class_: class}
	class_map, ok := class_user_map.ReadPair(class_id)
	if !ok {
		logger_.Log(logger.WARN, "GetClassUsers failed: Class %v does not exist", class_id)
		return nil, fmt.Errorf("class %v does not exist", class_id)
	}
	result := make([]*UserInfo, 0, 32)
	users_names := class_map.ReadAll()
	for uid := range users_names {
		user_info, ok := user_info_map.ReadPair(uid)
		if ok {
			result = append(result, &user_info)
		} else {
			logger_.Log(logger.ERROR, "Inconsistent state: User %s in class %v does not exist", uid, class_id)
			return nil, fmt.Errorf("inconsistent state: user %s in class %v does not exist", uid, class_id)
		}
	}
	return result, nil
}

func GetAllUsersInfo() []*UserInfo {
	result := make([]*UserInfo, 0, 128)
	all_users := user_info_map.ReadAll()
	for _, user_info := range all_users {
		result = append(result, &user_info)
	}
	return result
}
