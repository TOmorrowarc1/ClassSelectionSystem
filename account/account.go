package account

import (
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/course"
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
	account_logger *logger.Logger
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
	account_logger = logger.GetLogger()
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
	account_logger.Log(logger.INFO, "Account system initialized")
}

func StoreAccountData() {
	err := user_info_map.Store(USERINFOPATH)
	if err != nil {
		account_logger.Log(logger.ERROR, "Failed to store user info: %v", err)
	}
	err = class_user_map.Store(CLASSUSERPATH)
	if err != nil {
		account_logger.Log(logger.ERROR, "Failed to store class user info: %v", err)
	}
	account_logger.Log(logger.INFO, "Account data stored successfully")
}

func Register(uid string, password string, grade int, class int, priviledge string) error {
	if _, ok := user_info_map.ReadPair(uid); ok {
		account_logger.Log(logger.WARN, "Registration failed: User %s already exists", uid)
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
		account_logger.Log(logger.WARN, "Removal failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	user_info_map.DeletePair(uid)
	class_id := user_info.class_id_
	class_map, ok := class_user_map.ReadPair(class_id)
	if !ok {
		account_logger.Log(logger.ERROR, "Inconsistent state: Class %v for user %s does not exist", class_id, uid)
		return fmt.Errorf("inconsistent state: class %v for user %s does not exist", class_id, uid)
	}
	class_map.DeletePair(uid)
	class_user_map.WritePair(class_id, &class_map)
	return nil
}

func LogIn(uid string, password string) (int, error) {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		account_logger.Log(logger.WARN, "Login failed: User %s does not exist", uid)
		return 0, fmt.Errorf("user %s does not exist", uid)
	}
	if user_info.password_ != password {
		account_logger.Log(logger.WARN, "Login failed: Incorrect password for user %s", uid)
		return 0, fmt.Errorf("incorrect password for user %s", uid)
	}
	account_logger.Log(logger.INFO, "User %s logged in successfully", uid)
	return user_info.priviledge, nil
}

func ModifyPassword(uid string, new_password string) error {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		account_logger.Log(logger.WARN, "Password modification failed: User %s does not exist", uid)
		return fmt.Errorf("user %s does not exist", uid)
	}
	user_info.password_ = new_password
	user_info_map.WritePair(uid, &user_info)
	account_logger.Log(logger.INFO, "Password for user %s modified successfully", uid)
	return nil
}

func GetUserInfo(uid string) (*UserInfo, error) {
	user_info, ok := user_info_map.ReadPair(uid)
	if !ok {
		account_logger.Log(logger.WARN, "GetUserInfo failed: User %s does not exist", uid)
		return nil, fmt.Errorf("user %s does not exist", uid)
	}
	return &user_info, nil
}

func GetClassUsersInfo(grade int, class int) ([]*UserInfo, error) {
	class_id := ClassID{grade_: grade, class_: class}
	class_map, ok := class_user_map.ReadPair(class_id)
	if !ok {
		account_logger.Log(logger.WARN, "GetClassUsers failed: Class %v does not exist", class_id)
		return nil, fmt.Errorf("class %v does not exist", class_id)
	}
	users_names := class_map.ReadAll()
	result := make([]*UserInfo, 0, len(users_names))
	for uid := range users_names {
		user_info, ok := user_info_map.ReadPair(uid)
		if ok {
			result = append(result, &user_info)
		} else {
			account_logger.Log(logger.ERROR, "Inconsistent state: User %s in class %v does not exist", uid, class_id)
			return nil, fmt.Errorf("inconsistent state: user %s in class %v does not exist", uid, class_id)
		}
	}
	return result, nil
}

func GetCourseUsersInfo(uid string) ([]*UserInfo, error) {
	course_id := course.GetCourseUsers(uid)
	result := make([]*UserInfo, 0, len(course_id))
	for _, cid := range course_id {
		user_info, ok := user_info_map.ReadPair(cid)
		if ok {
			result = append(result, &user_info)
		} else {
			account_logger.Log(logger.ERROR, "Inconsistent state: User %s in course %s does not exist", cid, course_id)
			return nil, fmt.Errorf("inconsistent state: user %s in course %s does not exist", cid, course_id)
		}
	}
	return result, nil
}

func GetAllUsersInfo() []*UserInfo {
	all_users := user_info_map.ReadAll()
	result := make([]*UserInfo, 0, len(all_users))
	for _, user_info := range all_users {
		result = append(result, &user_info)
	}
	return result
}
