package course

import (
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
	"sync"
)

type CourseInfo struct {
	CourseName  string
	Teacher     string
	MaxStudents int
	NowStudents int
	Lanuched    bool
}

var (
	courseInfoMap *concurrentmap.ConcurrentMap[string, CourseInfo]
	lanuchedMap   *concurrentmap.ConcurrentMap[string, struct{}]
	courseUserMap *concurrentmap.ConcurrentMap[string, *concurrentmap.ConcurrentMap[string, struct{}]]
	userCourseMap *concurrentmap.ConcurrentMap[string, string]
	courseLogger  *logger.Logger
	// For only one concurrent operation holds(or rely on the consistency of) data outside the map: SelectCourse() and DropCourse().
	courseMutex sync.Mutex
)

func InitCourseSystem() {
	courseInfoMap = concurrentmap.NewConcurrentMap[string, CourseInfo]()
	lanuchedMap = concurrentmap.NewConcurrentMap[string, struct{}]()
	courseUserMap = concurrentmap.NewConcurrentMap[string, *concurrentmap.ConcurrentMap[string, struct{}]]()
	userCourseMap = concurrentmap.NewConcurrentMap[string, string]()
	courseLogger = logger.GetLogger()
	courseInfoMap.Load("data/course_Info.json")
	lanuchedMap.Load("data/launched_courses.json")
	courseUserMap.Load("data/course_user.json")
	userCourseMap.Load("data/user_course.json")
	// Check for consistency
	// 1. All launched courses must exist in courseInfoMap
	all_launched_course := lanuchedMap.ReadAll()
	for courseName := range all_launched_course {
		if _, ok := courseInfoMap.ReadPair(courseName); !ok {
			courseLogger.Log(logger.Error, "Inconsistent state: Launched course %s does not exist in courseInfoMap, removing", courseName)
			lanuchedMap.DeletePair(courseName)
		}
	}
	// 2. All courses in courseUserMap must exist in lanuchedMap.
	all_selected_course := courseUserMap.ReadAll()
	for courseName := range all_selected_course {
		if _, ok := lanuchedMap.ReadPair(courseName); !ok {
			courseLogger.Log(logger.Error, "Inconsistent state: Course %s in courseUserMap does not exist in lanuchedMap, removing", courseName)
			courseUserMap.DeletePair(courseName)
		}
	}
	// 3. The Information in courseUserMap and userCourseMap must be reflective, and I use the courseUserMap as the standard.
	for courseName, user_map := range all_selected_course {
		for uid := range user_map.ReadAll() {
			if cid, ok := userCourseMap.ReadPair(uid); !ok || cid != courseName {
				courseLogger.Log(logger.Error, "Inconsistent state: User %s in courseUserMap[%s] but not in userCourseMap or inconsistent, removing", uid, courseName)
				userCourseMap.DeletePair(uid)
				userCourseMap.WritePair(uid, &courseName)
			}
		}
	}
}

func StoreCourseData() {
	err := courseInfoMap.Store("data/course_Info.json")
	if err != nil {
		courseLogger.Log(logger.Error, "Failed to store course Info: %v", err)
	}
	err = lanuchedMap.Store("data/launched_courses.json")
	if err != nil {
		courseLogger.Log(logger.Error, "Failed to store launched courses: %v", err)
	}
	err = courseUserMap.Store("data/course_user.json")
	if err != nil {
		courseLogger.Log(logger.Error, "Failed to store course user map: %v", err)
	}
	err = userCourseMap.Store("data/user_course.json")
	if err != nil {
		courseLogger.Log(logger.Error, "Failed to store user course map: %v", err)
	}
	courseLogger.Log(logger.Info, "Course data stored successfully")
}

func AddCourse(CourseName string, teacher string, MaxStudents int) error {
	if _, ok := courseInfoMap.ReadPair(CourseName); ok {
		courseLogger.Log(logger.Warn, "Addition failed: Course %s already exists", CourseName)
		return fmt.Errorf("course %s already exists", CourseName)
	}
	new_course := CourseInfo{
		CourseName:  CourseName,
		Teacher:     teacher,
		MaxStudents: MaxStudents,
		NowStudents: 0,
		Lanuched:    false,
	}
	courseInfoMap.WritePair(CourseName, &new_course)
	return nil
}

func ModifyCourse(courseName string, teacher string, MaxStudents int) error {
	if _, exist := lanuchedMap.ReadPair(courseName); exist {
		courseLogger.Log(logger.Warn, "Modification failed: Course %s is already launched", courseName)
		return fmt.Errorf("course %s is already launched", courseName)
	}
	course_Info, ok := courseInfoMap.ReadPair(courseName)
	if !ok {
		courseLogger.Log(logger.Warn, "Modification failed: Course %s does not exist", courseName)
		return fmt.Errorf("course %s does not exist", courseName)
	}
	course_Info.CourseName = courseName
	course_Info.Teacher = teacher
	course_Info.MaxStudents = MaxStudents
	courseInfoMap.WritePair(courseName, &course_Info)
	return nil
}

func LaunchCourse(courseName string) error {
	_, ok := courseInfoMap.ReadPair(courseName)
	if !ok {
		courseLogger.Log(logger.Warn, "Launch failed: Course %s does not exist", courseName)
		return fmt.Errorf("course %s does not exist", courseName)
	}
	if _, exist := lanuchedMap.ReadPair(courseName); exist {
		courseLogger.Log(logger.Warn, "Launch failed: Course %s is already launched", courseName)
		return fmt.Errorf("course %s is already launched", courseName)
	}
	lanuchedMap.WritePair(courseName, &struct{}{})
	temp_map := concurrentmap.NewConcurrentMap[string, struct{}]()
	courseUserMap.WritePair(courseName, &temp_map)
	courseLogger.Log(logger.Info, "Course %s launched successfully", courseName)
	return nil
}

func SelectCourse(uid string, courseName string) error {
	courseMutex.Lock()
	defer courseMutex.Unlock()
	if _, ok := userCourseMap.ReadPair(uid); ok {
		courseLogger.Log(logger.Warn, "Selection failed: User %s has already selected a course", uid)
		return fmt.Errorf("user %s has already selected a course", uid)
	}
	if _, exist := lanuchedMap.ReadPair(courseName); !exist {
		courseLogger.Log(logger.Warn, "Selection failed: Course %s is not launched", courseName)
		return fmt.Errorf("course %s is not launched", courseName)
	}
	courseInfo, _ := courseInfoMap.ReadPair(courseName)
	if courseInfo.NowStudents >= courseInfo.MaxStudents {
		courseLogger.Log(logger.Warn, "Selection failed: Course %s is full", courseName)
		return fmt.Errorf("course %s is full", courseName)
	}
	user_map, _ := courseUserMap.ReadPair(courseName)
	user_map.WritePair(uid, &struct{}{})
	userCourseMap.WritePair(uid, &courseName)
	courseInfo.NowStudents++
	courseInfoMap.WritePair(courseName, &courseInfo)
	courseLogger.Log(logger.Info, "User %s selected course %s successfully", uid, courseName)
	return nil
}

func DropCourse(uid string) error {
	courseMutex.Lock()
	defer courseMutex.Unlock()
	courseName, ok := userCourseMap.ReadPair(uid)
	if !ok {
		courseLogger.Log(logger.Warn, "Drop failed: User %s has not selected any course", uid)
		return fmt.Errorf("user %s has not selected any course", uid)
	}
	userMap, ok := courseUserMap.ReadPair(courseName)
	if !ok {
		courseLogger.Log(logger.Error, "Inconsistent state: Course %s for user %s does not exist in courseUserMap", courseName, uid)
		return fmt.Errorf("inconsistent state: course %s for user %s does not exist in courseUserMap", courseName, uid)
	}
	userMap.DeletePair(uid)
	userCourseMap.DeletePair(uid)
	courseInfo, _ := courseInfoMap.ReadPair(courseName)
	courseInfo.NowStudents--
	courseInfoMap.WritePair(courseName, &courseInfo)
	courseLogger.Log(logger.Info, "User %s dropped course %s successfully", uid, courseName)
	return nil
}

func GetAllCoursesInfo() []*CourseInfo {
	resultMap := courseInfoMap.ReadAll()
	result := make([]*CourseInfo, 0, len(resultMap))
	for _, courseInfo := range resultMap {
		result = append(result, &courseInfo)
	}
	return result
}

func GetCourseUsers(courseName string) []string {
	user_map, ok := courseUserMap.ReadPair(courseName)
	if !ok {
		courseLogger.Log(logger.Warn, "GetCourseUsers failed: Course %s is not launched or does not exist", courseName)
		return nil
	}
	userNames := user_map.ReadAll()
	result := make([]string, 0, len(userNames))
	for uid := range userNames {
		result = append(result, uid)
	}
	return result
}
