package privilege

import (
	"fmt"
	"sync"
	"testing"
)

// TestMain 函数用于在所有测试运行前进行设置。
// 这里我们调用 InitPrivilegeSystem 来确保测试环境的一致性。
func TestMain(m *testing.M) {
	InitPrivilegeSystem()
	m.Run()
}

// TestGenerateToken 测试令牌生成功能。
// 它验证生成的令牌长度是否正确，并且每次调用都会生成唯一的令牌。
func TestGenerateToken(t *testing.T) {
	t.Run("TokenLength", func(t *testing.T) {
		token := generateToken()
		// 令牌是 16 字节数据的十六进制编码，所以长度应该是 32。
		expectedLength := Length * 2
		if len(token) != expectedLength {
			t.Errorf("期望的令牌长度为 %d, 但实际得到 %d", expectedLength, len(token))
		}
	})

	t.Run("TokenUniqueness", func(t *testing.T) {
		token1 := generateToken()
		token2 := generateToken()
		if token1 == "" || token2 == "" {
			t.Error("生成的令牌不应为空字符串")
		}
		if token1 == token2 {
			t.Error("连续生成的两个令牌不应相同，表明随机性不足")
		}
	})
}

// TestUserLoginAndAccess 测试用户登录和访问的核心流程。
// 它确保用户登录后可以获得一个有效的令牌，并使用该令牌成功获取用户信息。
// 同时，它也测试了使用无效令牌访问会被拒绝。
func TestUserLogInAndAccess(t *testing.T) {
	// 为测试用例重置环境
	InitPrivilegeSystem()

	userInfo := AccountInfo{UserName: "testuser", Privilege: 1}
	token := UserLogIn(userInfo)

	if token == "" {
		t.Fatal("UserLogIn 函数不应返回空令牌")
	}

	// 场景1: 使用有效令牌进行访问
	t.Run("ValidTokenAccess", func(t *testing.T) {
		retrievedInfo, err := UserAccess(token)
		if err != nil {
			t.Fatalf("使用有效令牌访问失败: %v", err)
		}
		if retrievedInfo.UserName != userInfo.UserName || retrievedInfo.Privilege != userInfo.Privilege {
			t.Errorf("获取到的用户信息 %+v 与原始信息 %+v 不匹配", retrievedInfo, userInfo)
		}
	})

	// 场景2: 使用无效令牌进行访问
	t.Run("InvalidTokenAccess", func(t *testing.T) {
		_, err := UserAccess("this_is_an_invalid_token")
		if err == nil {
			t.Error("使用无效令牌进行访问应返回错误，但实际没有")
		}
	})
}

// TestUserLogOut 测试用户登出功能。
// 它验证了使用有效令牌可以成功登出，并且登出后该令牌失效。
// 它也测试了使用无效令牌登出会失败。
// 注意：这个测试将会失败，因为它会暴露原始代码中的一个逻辑错误。
func TestUserLogOut(t *testing.T) {
	// 为测试用例重置环境
	InitPrivilegeSystem()

	userInfo := AccountInfo{UserName: "logout_user", Privilege: 2}
	token := UserLogIn(userInfo)

	// 场景1: 使用有效令牌登出
	t.Run("ValidTokenLogout", func(t *testing.T) {
		err := UserLogOut(token)
		if err != nil {
			// 根据原始代码的逻辑错误，这里会失败。
			t.Fatalf("使用有效令牌登出失败: %v", err)
		}

		// 验证登出后令牌是否失效
		_, err = UserAccess(token)
		if err == nil {
			t.Error("用户登出后，其令牌应立即失效，但访问依然成功")
		}
	})

	// 场景2: 使用无效令牌登出
	t.Run("InvalidTokenLogout", func(t *testing.T) {
		err := UserLogOut("this_is_an_invalid_token")
		if err == nil {
			// 根据原始代码的逻辑错误，这里也会失败。
			t.Error("使用无效令牌登出应返回错误，但实际没有")
		}
	})
}

// TestConcurrency 测试在高并发场景下系统的稳定性。
// 它模拟了大量用户同时登录、访问和登出的情况，以确保没有竞态条件并且功能正常。
func TestConcurrency(t *testing.T) {
	// 为测试用例重置环境
	InitPrivilegeSystem()
	
	numGoroutines := 100
	var wg sync.WaitGroup

	tokens := make(chan string, numGoroutines)

	// 并发登录
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			userInfo := AccountInfo{UserName: fmt.Sprintf("concurrent_user_%d", i), Privilege: i}
			token := UserLogIn(userInfo)
			if token == "" {
				// t.Errorf 是线程安全的，可以在 goroutine 中直接使用
				t.Errorf("并发登录时，用户 %d 获取到了空令牌", i)
			}
			tokens <- token
		}(i)
	}
	wg.Wait()
	close(tokens)

	tokenMap := make(map[string]bool)
	for token := range tokens {
		if tokenMap[token] {
			t.Errorf("发现了重复的令牌: %s", token)
		}
		tokenMap[token] = true
	}

	if len(tokenMap) != numGoroutines {
		t.Errorf("期望生成 %d 个唯一令牌，但实际只生成了 %d 个", numGoroutines, len(tokenMap))
	}

	// 并发访问和登出
	for token := range tokenMap {
		wg.Add(1)
		go func(tk string) {
			defer wg.Done()
			// 访问
			_, err := UserAccess(tk)
			if err != nil {
				t.Errorf("并发访问失败，令牌: %s, 错误: %v", tk, err)
			}
			// 登出
			err = UserLogOut(tk)
			// 注意：由于 UserLogOut 中的错误，这里会失败
			if err != nil {
				t.Errorf("并发登出失败，令牌: %s, 错误: %v", tk, err)
			}
		}(token)
	}
	wg.Wait()
}