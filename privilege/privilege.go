package privilege

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

const LENGTH = 16

var (
	privilege_map    *concurrentmap.ConcurrentMap[string, int]
	privilege_logger *logger.Logger
)

func InitPrivilegeSystem() {
	privilege_map = concurrentmap.NewConcurrentMap[string, int]()
	privilege_logger = logger.GetLogger()
	// The privilege_map is not persisted.
}

func generateToken() string {
	// Generate a random token of LENGTH bytes and return its hex encoding, which written by the Gemini.
	randomBytes := make([]byte, LENGTH)
	_, err := rand.Read(randomBytes)
	if err != nil {
		privilege_logger.Log(logger.ERROR, "Failed to generate random bytes: %v", err)
		return ""
	}
	return hex.EncodeToString(randomBytes)
}

func UserLogIn(privilege int) string {
	token := generateToken()
	privilege_map.WritePair(token, &privilege)
	privilege_logger.Log(logger.INFO, "An user with privilege %d get token %s", privilege, token)
	return token
}

func UserAccess(token string) (int, error) {
	if privilege, ok := privilege_map.ReadPair(token); ok {
		return privilege, nil
	}
	privilege_logger.Log(logger.WARN, "Access denied: Invalid token %s", token)
	return 0, fmt.Errorf("invalid token %s", token)
}

func UserLogOut(token string) error {
	if _, ok := privilege_map.ReadPair(token); !ok {
		privilege_map.DeletePair(token)
		privilege_logger.Log(logger.INFO, "Token %s logged out successfully", token)
		return nil
	}
	privilege_logger.Log(logger.WARN, "Logout failed: Invalid token %s", token)
	return fmt.Errorf("invalid token %s", token)
}
