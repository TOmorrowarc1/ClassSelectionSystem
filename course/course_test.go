package course

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"

	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

// setupCourseTest 是一个辅助函数，用于在每个测试之前初始化或重置系统状态。
// 它在内存中创建所有需要的数据结构，避免了对文件系统的依赖。
func setupCourseTest() {
	courseInfoMap = concurrentmap.NewConcurrentMap[string, CourseInfo]()
	// 注意：您的代码中存在拼写错误 "Lanuched"，为了让测试能够编译通过，这里保持一致。
	// 建议在未来将其更正为 "Launched"。
	lanuchedMap = concurrentmap.NewConcurrentMap[string, struct{}]()
	courseUserMap = concurrentmap.NewConcurrentMap[string, *concurrentmap.ConcurrentMap[string, struct{}]]()
	userCourseMap = concurrentmap.NewConcurrentMap[string, string]()
	courseLogger = logger.GetLogger()
	os.Remove("course_test.log") // 删除旧的日志文件
	courseLogger.SetLogFile("course_test.log")
}

// TestAddAndModifyCourse 测试课程的创建和修改功能。
func TestAddAndModifyCourse(t *testing.T) {
	setupCourseTest()
	courseName := "Intro to Go"
	teacher := "Prof. Gopher"

	t.Run("SuccessfulAdd", func(t *testing.T) {
		err := AddCourse(courseName, teacher, 30)
		if err != nil {
			t.Fatalf("添加新课程失败: %v", err)
		}
		info, ok := courseInfoMap.ReadPair(courseName)
		if !ok {
			t.Fatal("添加课程后，在 courseInfoMap 中找不到该课程")
		}
		if info.Teacher != teacher || info.MaxStudents != 30 {
			t.Errorf("存储的课程信息不正确")
		}
	})

	t.Run("AddExistingCourse", func(t *testing.T) {
		err := AddCourse(courseName, teacher, 30)
		if err == nil {
			t.Error("添加已存在的课程时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("SuccessfulModifyUnlaunched", func(t *testing.T) {
		newTeacher := "Dr. Gemini"
		err := ModifyCourse(courseName, newTeacher, 40)
		if err != nil {
			t.Fatalf("修改未发布的课程失败: %v", err)
		}
		info, _ := courseInfoMap.ReadPair(courseName)
		if info.Teacher != newTeacher || info.MaxStudents != 40 {
			t.Errorf("修改后，课程信息未正确更新")
		}
	})

	t.Run("ModifyNonExistent", func(t *testing.T) {
		err := ModifyCourse("NonExistentCourse", "Some Teacher", 20)
		if err == nil {
			t.Error("修改不存在的课程时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestLaunchCourse 测试课程发布功能。
func TestLaunchCourse(t *testing.T) {
	setupCourseTest()
	courseName := "Advanced Go"
	AddCourse(courseName, "Prof. Gopher", 25)

	t.Run("SuccessfulLaunch", func(t *testing.T) {
		err := LaunchCourse(courseName)
		if err != nil {
			t.Fatalf("发布课程失败: %v", err)
		}
		if _, ok := lanuchedMap.ReadPair(courseName); !ok {
			t.Error("课程发布后，在 lanuchedMap 中找不到该课程")
		}
		if _, ok := courseUserMap.ReadPair(courseName); !ok {
			t.Error("课程发布后，在 courseUserMap 中找不到对应的学生名册")
		}
	})

	t.Run("ModifyLaunchedCourse", func(t *testing.T) {
		// 此时课程已发布
		err := ModifyCourse(courseName, "New Teacher", 30)
		if err == nil {
			t.Error("修改已发布的课程时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("LaunchNonExistent", func(t *testing.T) {
		err := LaunchCourse("FakeCourse")
		if err == nil {
			t.Error("发布不存在的课程时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("LaunchAlreadyLaunched", func(t *testing.T) {
		err := LaunchCourse(courseName)
		if err == nil {
			t.Error("重复发布课程时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestSelectAndDropCourse 测试学生选课和退课的核心流程。
func TestSelectAndDropCourse(t *testing.T) {
	setupCourseTest()
	courseFull := "Full Course"
	courseAvailable := "Available Course"
	courseNotLaunched := "Not Launched Course"
	student1, student2 := "student1", "student2"

	// 设置场景
	AddCourse(courseFull, "teacher", 1)
	AddCourse(courseAvailable, "teacher", 2)
	AddCourse(courseNotLaunched, "teacher", 5)
	LaunchCourse(courseFull)
	LaunchCourse(courseAvailable)

	t.Run("SuccessfulSelect", func(t *testing.T) {
		err := SelectCourse(student1, courseAvailable)
		if err != nil {
			t.Fatalf("学生 %s 选课 %s 失败: %v", student1, courseAvailable, err)
		}
		info, _ := courseInfoMap.ReadPair(courseAvailable)
		if info.NowStudents != 1 {
			t.Errorf("选课后，课程人数应为 1, 实际为 %d", info.NowStudents)
		}
		cName, _ := userCourseMap.ReadPair(student1)
		if cName != courseAvailable {
			t.Errorf("选课后，userCourseMap 记录不正确")
		}
	})

	t.Run("SelectCourseIsFull", func(t *testing.T) {
		// 先让一个学生占满名额
		_ = SelectCourse("temp_student", courseFull)
		// student2 尝试选择满员课程
		err := SelectCourse(student2, courseFull)
		if err == nil {
			t.Error("选择满员课程时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("SelectAlreadyHasCourse", func(t *testing.T) {
		// student1 已经选了 courseAvailable
		err := SelectCourse(student1, courseFull)
		if err == nil {
			t.Error("已选课的学生再次选课时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("SelectNotLaunchedCourse", func(t *testing.T) {
		err := SelectCourse(student2, courseNotLaunched)
		if err == nil {
			t.Error("选择未发布的课程时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("SuccessfulDrop", func(t *testing.T) {
		err := DropCourse(student1)
		if err != nil {
			t.Fatalf("学生 %s 退课失败: %v", student1, err)
		}
		info, _ := courseInfoMap.ReadPair(courseAvailable)
		if info.NowStudents != 0 {
			t.Errorf("退课后，课程人数应为 0, 实际为 %d", info.NowStudents)
		}
		if _, ok := userCourseMap.ReadPair(student1); ok {
			t.Error("退课后，userCourseMap 中仍存在该学生的记录")
		}
	})

	t.Run("DropWithoutCourse", func(t *testing.T) {
		// student2 从未成功选课
		err := DropCourse(student2)
		if err == nil {
			t.Error("未选课的学生退课时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestGetters 测试数据获取功能。
func TestGetters(t *testing.T) {
	setupCourseTest()
	c1, c2 := "Course1", "Course2"
	s1, s2 := "student1", "student2"
	AddCourse(c1, "t1", 3)
	AddCourse(c2, "t2", 2)
	LaunchCourse(c1)
	SelectCourse(s1, c1)
	SelectCourse(s2, c1)

	t.Run("GetAllCoursesInfo", func(t *testing.T) {
		allCourses := GetAllCoursesInfo()
		if len(allCourses) != 2 {
			t.Fatalf("期望获取到 2 门课程，实际得到 %d", len(allCourses))
		}
	})

	t.Run("GetCourseUsers", func(t *testing.T) {
		users := GetCourseUsers(c1)
		if len(users) != 2 {
			t.Fatalf("期望获取到 2 个学生，实际得到 %d", len(users))
		}
		// 排序以确保测试结果的确定性
		sort.Strings(users)
		expected := []string{s1, s2}
		if users[0] != expected[0] || users[1] != expected[1] {
			t.Errorf("获取到的学生列表不正确。得到 %v, 期望 %v", users, expected)
		}

		// 测试未发布或不存在的课程
		if GetCourseUsers(c2) != nil {
			t.Error("获取未发布课程的学生列表时，应返回 nil")
		}
		if GetCourseUsers("fake") != nil {
			t.Error("获取不存在课程的学生列表时，应返回 nil")
		}
	})
}

// TestFullConcurrencyCourseSelection 运行一个高强度的并发测试，
// 模拟大量学生同时抢一门容量有限的课程，以验证选课和退课操作的原子性。
func TestFullConcurrencyCourseSelection(t *testing.T) {
	setupCourseTest()

	// --- 配置 ---
	const courseCapacity = 50
	const concurrentStudents = 200 // 学生数远大于课程容量，以制造高竞争
	courseName := "Concurrent Programming"

	// --- 场景设置 ---
	AddCourse(courseName, "Prof. Race", courseCapacity)
	LaunchCourse(courseName)

	var wg sync.WaitGroup
	wg.Add(concurrentStudents)

	// --- 执行并发操作 ---
	for i := 0; i < concurrentStudents; i++ {
		go func(i int) {
			defer wg.Done()
			uid := fmt.Sprintf("student_%d", i)

			// 每个学生尝试选课，如果成功，则立即退课
			// 这种高频的“选-退”操作可以最大化对课程人数计数的压力
			err := SelectCourse(uid, courseName)
			if err == nil {
				// 选课成功，立即退课
				_ = DropCourse(uid)
			}
		}(i)
	}

	// 等待所有学生操作完成
	wg.Wait()

	// --- 一致性验证 ---
	// 在所有并发操作结束后，系统应该回到一个稳定且一致的状态。
	t.Run("FinalStateConsistencyCheck", func(t *testing.T) {
		finalInfo, ok := courseInfoMap.ReadPair(courseName)
		if !ok {
			t.Fatal("测试结束后，课程信息丢失")
		}

		// 1. 最终人数检查：由于每个成功的选课都伴随着一次退课，最终人数应该为 0。
		if finalInfo.NowStudents != 0 {
			t.Errorf("最终课程人数应为 0, 但实际为 %d。这表明选课和退课操作存在计数错误。", finalInfo.NowStudents)
		}

		// 2. 学生名册检查：最终应该没有学生在课程中。
		usersInCourse := GetCourseUsers(courseName)
		if len(usersInCourse) != 0 {
			t.Errorf("最终课程学生名册应为空，但仍有 %d 个学生: %v", len(usersInCourse), usersInCourse)
		}

		// 3. 反向映射检查：最终不应有任何学生记录指向该课程。
		if length := len(userCourseMap.ReadAll()); length != 0 {
			t.Errorf("最终 userCourseMap 应为空，但仍有 %d 个条目。", length)
		}
	})
}
