package course

import (
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

type CourseInfo struct {
	Course_name  string
	Teacher      string
	Max_students int
	Now_students int
	Lanuched     bool
}

var (
	course_info_map *concurrentmap.ConcurrentMap[string, CourseInfo]
	launched_map    *concurrentmap.ConcurrentMap[string, struct{}]
	course_user_map *concurrentmap.ConcurrentMap[string, *concurrentmap.ConcurrentMap[string, struct{}]]
	user_course_map *concurrentmap.ConcurrentMap[string, string]
	course_logger   *logger.Logger
)

func InitCourseSystem() {
	course_info_map = concurrentmap.NewConcurrentMap[string, CourseInfo]()
	launched_map = concurrentmap.NewConcurrentMap[string, struct{}]()
	course_user_map = concurrentmap.NewConcurrentMap[string, *concurrentmap.ConcurrentMap[string, struct{}]]()
	user_course_map = concurrentmap.NewConcurrentMap[string, string]()
	course_logger = logger.GetLogger()
	course_info_map.Load("data/course_info.json")
	launched_map.Load("data/launched_courses.json")
	course_user_map.Load("data/course_user.json")
	user_course_map.Load("data/user_course.json")
	// Check for consistency
	// 1. All launched courses must exist in course_info_map
	all_launched_course := launched_map.ReadAll()
	for course_id := range all_launched_course {
		if _, ok := course_info_map.ReadPair(course_id); !ok {
			course_logger.Log(logger.ERROR, "Inconsistent state: Launched course %s does not exist in course_info_map, removing", course_id)
			launched_map.DeletePair(course_id)
		}
	}
	// 2. All courses in course_user_map must exist in launched_map.
	all_selected_course := course_user_map.ReadAll()
	for course_id := range all_selected_course {
		if _, ok := launched_map.ReadPair(course_id); !ok {
			course_logger.Log(logger.ERROR, "Inconsistent state: Course %s in course_user_map does not exist in launched_map, removing", course_id)
			course_user_map.DeletePair(course_id)
		}
	}
	// 3. The information in course_user_map and user_course_map must be reflective, and I use the course_user_map as the standard.
	for course_id, user_map := range all_selected_course {
		for uid := range user_map.ReadAll() {
			if cid, ok := user_course_map.ReadPair(uid); !ok || cid != course_id {
				course_logger.Log(logger.ERROR, "Inconsistent state: User %s in course_user_map[%s] but not in user_course_map or inconsistent, removing", uid, course_id)
				user_course_map.DeletePair(uid)
				user_course_map.WritePair(uid, &course_id)
			}
		}
	}
}

func StoreCourseData() {
	err := course_info_map.Store("data/course_info.json")
	if err != nil {
		course_logger.Log(logger.ERROR, "Failed to store course info: %v", err)
	}
	err = launched_map.Store("data/launched_courses.json")
	if err != nil {
		course_logger.Log(logger.ERROR, "Failed to store launched courses: %v", err)
	}
	err = course_user_map.Store("data/course_user.json")
	if err != nil {
		course_logger.Log(logger.ERROR, "Failed to store course user map: %v", err)
	}
	err = user_course_map.Store("data/user_course.json")
	if err != nil {
		course_logger.Log(logger.ERROR, "Failed to store user course map: %v", err)
	}
	course_logger.Log(logger.INFO, "Course data stored successfully")
}

func AddCourse(course_name string, teacher string, max_students int) error {
	if _, ok := course_info_map.ReadPair(course_name); ok {
		course_logger.Log(logger.WARN, "Addition failed: Course %s already exists", course_name)
		return fmt.Errorf("course %s already exists", course_name)
	}
	new_course := CourseInfo{
		Course_name:  course_name,
		Teacher:      teacher,
		Max_students: max_students,
		Now_students: 0,
		Lanuched:     false,
	}
	course_info_map.WritePair(course_name, &new_course)
	return nil
}

func ModifyCourse(course_id string, course_name string, teacher string, max_students int) error {
	if _, exist := launched_map.ReadPair(course_id); exist {
		course_logger.Log(logger.WARN, "Modification failed: Course %s is already launched", course_id)
		return fmt.Errorf("course %s is already launched", course_id)
	}
	course_info, ok := course_info_map.ReadPair(course_id)
	if !ok {
		course_logger.Log(logger.WARN, "Modification failed: Course %s does not exist", course_id)
		return fmt.Errorf("course %s does not exist", course_id)
	}
	course_info.Course_name = course_name
	course_info.Teacher = teacher
	course_info.Max_students = max_students
	course_info_map.WritePair(course_id, &course_info)
	return nil
}

func LaunchCourse(course_id string) error {
	_, ok := course_info_map.ReadPair(course_id)
	if !ok {
		course_logger.Log(logger.WARN, "Launch failed: Course %s does not exist", course_id)
		return fmt.Errorf("course %s does not exist", course_id)
	}
	if _, exist := launched_map.ReadPair(course_id); exist {
		course_logger.Log(logger.WARN, "Launch failed: Course %s is already launched", course_id)
		return fmt.Errorf("course %s is already launched", course_id)
	}
	launched_map.WritePair(course_id, &struct{}{})
	temp_map := concurrentmap.NewConcurrentMap[string, struct{}]()
	course_user_map.WritePair(course_id, &temp_map)
	course_logger.Log(logger.INFO, "Course %s launched successfully", course_id)
	return nil
}

func SelectCourse(uid string, course_id string) error {
	if _, ok := user_course_map.ReadPair(uid); ok {
		course_logger.Log(logger.WARN, "Selection failed: User %s has already selected a course", uid)
		return fmt.Errorf("user %s has already selected a course", uid)
	}
	if _, exist := launched_map.ReadPair(course_id); !exist {
		course_logger.Log(logger.WARN, "Selection failed: Course %s is not launched", course_id)
		return fmt.Errorf("course %s is not launched", course_id)
	}
	course_info, _ := course_info_map.ReadPair(course_id)
	if course_info.Now_students >= course_info.Max_students {
		course_logger.Log(logger.WARN, "Selection failed: Course %s is full", course_id)
		return fmt.Errorf("course %s is full", course_id)
	}
	user_map, _ := course_user_map.ReadPair(course_id)
	user_map.WritePair(uid, &struct{}{})
	user_course_map.WritePair(uid, &course_id)
	course_info.Now_students++
	course_info_map.WritePair(course_id, &course_info)
	course_logger.Log(logger.INFO, "User %s selected course %s successfully", uid, course_id)
	return nil
}

func DropCourse(uid string) error {
	course_id, ok := user_course_map.ReadPair(uid)
	if !ok {
		course_logger.Log(logger.WARN, "Drop failed: User %s has not selected any course", uid)
		return fmt.Errorf("user %s has not selected any course", uid)
	}
	user_map, ok := course_user_map.ReadPair(course_id)
	if !ok {
		course_logger.Log(logger.ERROR, "Inconsistent state: Course %s for user %s does not exist in course_user_map", course_id, uid)
		return fmt.Errorf("inconsistent state: course %s for user %s does not exist in course_user_map", course_id, uid)
	}
	user_map.DeletePair(uid)
	user_course_map.DeletePair(uid)
	course_info, _ := course_info_map.ReadPair(course_id)
	course_info.Now_students--
	course_info_map.WritePair(course_id, &course_info)
	course_logger.Log(logger.INFO, "User %s dropped course %s successfully", uid, course_id)
	return nil
}

func GetAllCoursesInfo() []*CourseInfo {
	result_map := course_info_map.ReadAll()
	result := make([]*CourseInfo, 0, len(result_map))
	for _, course_info := range result_map {
		result = append(result, &course_info)
	}
	return result
}

func GetCourseUsers(course_id string) []string {
	user_map, ok := course_user_map.ReadPair(course_id)
	if !ok {
		course_logger.Log(logger.WARN, "GetCourseUsers failed: Course %s is not launched or does not exist", course_id)
		return nil
	}
	user_names := user_map.ReadAll()
	result := make([]string, 0, len(user_names))
	for uid := range user_names {
		result = append(result, uid)
	}
	return result
}
