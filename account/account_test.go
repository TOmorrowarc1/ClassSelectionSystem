package account

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

// setupAccountTest 是一个辅助函数，用于在每个测试之前初始化或重置系统状态。
// 这确保了测试的独立性，避免了对真实文件的读写依赖。
func setupAccountTest() {
	userInfoMap = concurrentmap.NewConcurrentMap[string, UserInfo]()
	classUserMap = concurrentmap.NewConcurrentMap[ClassID, *concurrentmap.ConcurrentMap[string, struct{}]]()
	accountLogger = logger.GetLogger() // 假设 GetLogger 可以安全地重复调用
	os.Remove("account_test.log")      // 删除旧的日志文件以避免干扰
	accountLogger.SetLogFile("account_test.log")

	// 为测试添加一个默认的 admin 用户，模拟 InitAccountSystem 的行为
	adminInfo := UserInfo{
		Uid:       "admin",
		Password:  "123456",
		Privilege: PrivilegeAdmin,
		Classid:   ClassID{Grade: 0, Class: 0},
	}
	userInfoMap.WritePair("admin", &adminInfo)
}

// TestPrivilegeConversion 测试权限与字符串之间的转换函数。
func TestPrivilegeConversion(t *testing.T) {
	t.Run("StringToPrivilege", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected int
		}{
			{"student", PrivilegeStudent},
			{"teacher", PrivilegeTeacher},
			{"admin", PrivilegeAdmin},
			{"invalid", PrivilegeStudent}, // 默认情况
		}

		for _, tc := range testCases {
			if result := StringToPrivilege(tc.input); result != tc.expected {
				t.Errorf("StringToPrivilege(%q) = %d; 期望 %d", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("PrivilegeToString", func(t *testing.T) {
		testCases := []struct {
			input    int
			expected string
		}{
			{PrivilegeStudent, "student"},
			{PrivilegeTeacher, "teacher"},
			{PrivilegeAdmin, "admin"},
			{99, "unknown"}, // 无效权限
		}

		for _, tc := range testCases {
			if result := PrivilegeToString(tc.input); result != tc.expected {
				t.Errorf("PrivilegeToString(%d) = %q; 期望 %q", tc.input, result, tc.expected)
			}
		}
	})
}

// TestRegister 测试用户注册功能。
func TestRegister(t *testing.T) {
	setupAccountTest()

	user := UserInfo{
		Uid:       "student1",
		Password:  "pass123",
		Classid:   ClassID{Grade: 1, Class: 1},
		Privilege: PrivilegeStudent,
	}

	// 场景1: 成功注册新用户
	t.Run("SuccessfulRegistration", func(t *testing.T) {
		err := Register(user)
		if err != nil {
			t.Fatalf("注册新用户失败: %v", err)
		}

		// 验证用户信息是否已写入
		info, ok := userInfoMap.ReadPair(user.Uid)
		if !ok {
			t.Fatal("注册后，在 userInfoMap 中找不到用户")
		}
		if !reflect.DeepEqual(info, user) {
			t.Errorf("注册后存储的用户信息不正确。得到 %+v, 期望 %+v", info, user)
		}

		// 验证班级信息是否已更新
		classMap, ok := classUserMap.ReadPair(user.Classid)
		if !ok {
			t.Fatal("注册后，在 classUserMap 中找不到对应的班级")
		}
		if _, userInClass := classMap.ReadPair(user.Uid); !userInClass {
			t.Error("注册后，在班级映射中找不到对应的用户ID")
		}
	})

	// 场景2: 注册已存在的用户
	t.Run("RegisterExistingUser", func(t *testing.T) {
		// user 已经在上一个子测试中注册过了
		err := Register(user)
		if err == nil {
			t.Error("注册已存在的用户时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestRemoveUser 测试用户移除功能。
func TestRemoveUser(t *testing.T) {
	setupAccountTest()

	user := UserInfo{
		Uid:       "student2",
		Password:  "pass456",
		Classid:   ClassID{Grade: 2, Class: 2},
		Privilege: PrivilegeStudent,
	}
	// 先注册一个用户以便移除
	Register(user)

	// 场景1: 成功移除用户
	t.Run("SuccessfulRemoval", func(t *testing.T) {
		err := RemoveUser(user.Uid)
		if err != nil {
			t.Fatalf("移除用户失败: %v", err)
		}

		// 验证用户是否已从 userInfoMap 中移除
		if _, ok := userInfoMap.ReadPair(user.Uid); ok {
			t.Error("用户被移除后，仍然存在于 userInfoMap 中")
		}

		// 验证用户是否已从 classUserMap 中移除
		classMap, ok := classUserMap.ReadPair(user.Classid)
		if !ok {
			t.Fatal("移除用户后，班级映射不应消失")
		}
		if _, userInClass := classMap.ReadPair(user.Uid); userInClass {
			t.Error("用户被移除后，仍然存在于班级映射中")
		}
	})

	// 场景2: 移除不存在的用户
	t.Run("RemoveNonExistentUser", func(t *testing.T) {
		err := RemoveUser("nonexistentuser")
		if err == nil {
			t.Error("移除不存在的用户时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestLogIn 测试用户登录功能。
func TestLogIn(t *testing.T) {
	setupAccountTest()

	user := UserInfo{
		Uid:       "student3",
		Password:  "pass789",
		Classid:   ClassID{Grade: 3, Class: 3},
		Privilege: PrivilegeTeacher,
	}
	Register(user)

	// 场景1: 登录成功
	t.Run("SuccessfulLogin", func(t *testing.T) {
		privilege, err := LogIn(user.Uid, user.Password)
		if err != nil {
			t.Fatalf("使用正确的凭据登录失败: %v", err)
		}
		if privilege != user.Privilege {
			t.Errorf("登录后返回的权限不正确。得到 %d, 期望 %d", privilege, user.Privilege)
		}
	})

	// 场景2: 用户不存在
	t.Run("LoginNonExistentUser", func(t *testing.T) {
		_, err := LogIn("nonexistentuser", "somepassword")
		if err == nil {
			t.Error("使用不存在的用户名登录时，期望得到一个错误，但实际为 nil")
		}
	})

	// 场景3: 密码错误
	t.Run("LoginWrongPassword", func(t *testing.T) {
		_, err := LogIn(user.Uid, "wrongpassword")
		if err == nil {
			t.Error("使用错误的密码登录时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestModifyPassword 测试修改密码功能。
func TestModifyPassword(t *testing.T) {
	setupAccountTest()

	user := UserInfo{
		Uid:      "student4",
		Password: "oldPassword",
		Classid:  ClassID{Grade: 4, Class: 4},
	}
	Register(user)

	newPassword := "newPassword"

	// 场景1: 成功修改密码
	t.Run("SuccessfulPasswordModification", func(t *testing.T) {
		err := ModifyPassword(user.Uid, newPassword)
		if err != nil {
			t.Fatalf("修改密码失败: %v", err)
		}

		// 验证新密码是否生效
		info, _ := userInfoMap.ReadPair(user.Uid)
		if info.Password != newPassword {
			t.Errorf("修改密码后，存储的密码不正确。得到 %q, 期望 %q", info.Password, newPassword)
		}

		// 尝试用新密码登录
		_, loginErr := LogIn(user.Uid, newPassword)
		if loginErr != nil {
			t.Errorf("修改密码后，无法使用新密码登录: %v", loginErr)
		}
	})

	// 场景2: 修改不存在用户的密码
	t.Run("ModifyPasswordForNonExistentUser", func(t *testing.T) {
		err := ModifyPassword("nonexistentuser", "somepassword")
		if err == nil {
			t.Error("为不存在的用户修改密码时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestGetUserInfo 测试获取单个用户信息的函数。
func TestGetUserInfo(t *testing.T) {
	setupAccountTest()

	user := UserInfo{
		Uid:      "student5",
		Password: "password",
		Classid:  ClassID{Grade: 5, Class: 5},
	}
	Register(user)

	// 场景1: 获取存在的用户信息
	t.Run("GetExistingUserInfo", func(t *testing.T) {
		info, err := GetUserInfo(user.Uid)
		if err != nil {
			t.Fatalf("获取存在的用户信息失败: %v", err)
		}
		if !reflect.DeepEqual(*info, user) {
			t.Errorf("获取到的用户信息不正确。得到 %+v, 期望 %+v", *info, user)
		}
	})

	// 场景2: 获取不存在的用户信息
	t.Run("GetNonExistentUserInfo", func(t *testing.T) {
		_, err := GetUserInfo("nonexistentuser")
		if err == nil {
			t.Error("获取不存在的用户信息时，期望得到一个错误，但实际为 nil")
		}
	})
}

// TestGetClassAndAllUsersInfo 测试获取班级用户和所有用户信息的函数。
func TestGetClassAndAllUsersInfo(t *testing.T) {
	setupAccountTest()

	class1 := ClassID{Grade: 10, Class: 1}
	class2 := ClassID{Grade: 10, Class: 2}

	userC1_1 := UserInfo{Uid: "userC1_1", Classid: class1}
	userC1_2 := UserInfo{Uid: "userC1_2", Classid: class1}
	userC2_1 := UserInfo{Uid: "userC2_1", Classid: class2}

	Register(userC1_1)
	Register(userC1_2)
	Register(userC2_1)

	t.Run("GetClassUsersInfo", func(t *testing.T) {
		users, err := GetClassUsersInfo(class1)
		if err != nil {
			t.Fatalf("获取班级用户信息失败: %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("期望获取到 2 个用户，但实际得到 %d 个", len(users))
		}

		// 因为返回顺序不确定，使用 map 进行验证
		expectedUIDs := map[string]bool{"userC1_1": true, "userC1_2": true}
		for _, u := range users {
			if !expectedUIDs[u.Uid] {
				t.Errorf("在班级用户列表中发现了不期望的用户: %s", u.Uid)
			}
		}
	})

	t.Run("GetNonExistentClassUsersInfo", func(t *testing.T) {
		_, err := GetClassUsersInfo(ClassID{Grade: 99, Class: 99})
		if err == nil {
			t.Error("获取不存在的班级信息时，期望得到一个错误，但实际为 nil")
		}
	})

	t.Run("GetAllUsersInfo", func(t *testing.T) {
		allUsers := GetAllUsersInfo()
		// 3 个学生 + 1 个 admin
		if len(allUsers) != 4 {
			t.Fatalf("期望获取到 4 个用户，但实际得到 %d 个", len(allUsers))
		}
	})
}

// TestFullConcurrency 运行一个高强度的并发测试，模拟多种操作（读、写、修改、删除）
// 同时发生在同一组用户数据上，以暴露潜在的竞态条件和数据不一致问题。
func TestFullConcurrency(t *testing.T) {
	setupAccountTest()

	// --- 配置 ---
	const numGoroutines = 100  // 模拟的并发用户数
	const opsPerGoroutine = 50 // 每个用户执行的操作次数
	const userPoolSize = 20    // 操作将集中在这一小部分用户上，以增加冲突概率

	var wg sync.WaitGroup

	// --- 执行并发操作 ---
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			// 每个 goroutine 将随机对 userPoolSize 个用户中的一个执行操作
			for j := 0; j < opsPerGoroutine; j++ {
				// 随机选择一个用户进行操作
				uid := fmt.Sprintf("concurrent_user_%d", j%userPoolSize)
				classID := ClassID{Grade: 100, Class: j % 5}

				// 随机选择一个操作来执行
				op := j % 5
				switch op {
				case 0: // 注册
					user := UserInfo{
						Uid:      uid,
						Password: "password",
						Classid:  classID,
					}
					// 注册可能会因为用户已存在而失败，这是预期的行为，所以我们不检查错误。
					// 我们关心的是这个操作会不会导致程序崩溃或数据损坏。
					_ = Register(user)

				case 1: // 登录 (读取)
					// 登录可能会因为用户不存在或密码错误而失败，这也是预期的。
					_, _ = LogIn(uid, "password")

				case 2: // 修改密码 (读取 + 写入)
					newPassword := fmt.Sprintf("new_pass_%d", j)
					// 修改可能会因为用户不存在而失败，这也是预期的。
					_ = ModifyPassword(uid, newPassword)

				case 3: // 删除
					// 删除可能会因为用户不存在而失败，这也是预期的。
					_ = RemoveUser(uid)

				case 4: // 获取用户信息 (读取)
					_, _ = GetUserInfo(uid)
				}
			}
		}()
	}

	// 等待所有并发操作完成
	wg.Wait()

	// --- 一致性验证 ---
	// 在所有操作完成后，检查两个 map 之间的数据是否一致。
	// 这是最关键的一步，用于发现那些没有导致程序崩溃但破坏了数据完整性的 bug。
	t.Run("FinalStateConsistencyCheck", func(t *testing.T) {
		// 检查1: 任何存在于 classUserMap 中的用户，必须也存在于 userInfoMap 中。
		allClasses := classUserMap.ReadAll()
		for classID, classMap := range allClasses {
			allUIDsInClass := classMap.ReadAll()
			for uid := range allUIDsInClass {
				if _, ok := userInfoMap.ReadPair(uid); !ok {
					t.Errorf("数据不一致: 用户 %s 存在于班级 %v 的列表中, 但在主用户列表中不存在!", uid, classID)
				}
			}
		}

		// 检查2: 任何存在于 userInfoMap 中的用户 (非特殊用户如 admin)，必须也存在于其对应的 classUserMap 中。
		allUsers := userInfoMap.ReadAll()
		for uid, userInfo := range allUsers {
			// 跳过没有班级的特殊用户
			if userInfo.Classid.Grade == 0 && userInfo.Classid.Class == 0 {
				continue
			}

			classMap, ok := classUserMap.ReadPair(userInfo.Classid)
			if !ok {
				t.Errorf("数据不一致: 用户 %s 的班级 %v 不存在于班级映射中!", uid, userInfo.Classid)
				continue // 如果班级本身不存在，就无法进行下一步检查
			}

			if _, userInClass := classMap.ReadPair(uid); !userInClass {
				t.Errorf("数据不一致: 用户 %s 不存在于其声称的班级 %v 的列表中!", uid, userInfo.Classid)
			}
		}
	})
}
