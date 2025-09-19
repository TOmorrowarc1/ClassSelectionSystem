package priviledge

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

const LENGTH = 16

var (
	priviledge_map    *concurrentmap.ConcurrentMap[string, int]
	priviledge_logger *logger.Logger
)

func InitPriviledgeSystem() {
	priviledge_map = concurrentmap.NewConcurrentMap[string, int]()
	priviledge_logger = logger.GetLogger()
	// The priviledge_map is not persisted.
}

func generateToken() string {
	// Generate a random token of LENGTH bytes and return its hex encoding, which written by the Gemini.
	randomBytes := make([]byte, LENGTH)
	_, err := rand.Read(randomBytes)
	if err != nil {
		priviledge_logger.Log(logger.ERROR, "Failed to generate random bytes: %v", err)
		return ""
	}
	return hex.EncodeToString(randomBytes)
}

func UserLogIn(priviledge int) {
	token := generateToken()
	priviledge_map.WritePair(token, &priviledge)
	priviledge_logger.Log(logger.INFO, "An user with priviledge %d get token %s", priviledge, token)
}

func UserAccess(token string) (int, error) {
	if priviledge, ok := priviledge_map.ReadPair(token); ok {
		return priviledge, nil
	}
	priviledge_logger.Log(logger.WARN, "Access denied: Invalid token %s", token)
	return 0, fmt.Errorf("invalid token %s", token)
}

func UserLogOut(token string) error {
	if _, ok := priviledge_map.ReadPair(token); !ok {
		priviledge_map.DeletePair(token)
		priviledge_logger.Log(logger.INFO, "Token %s logged out successfully", token)
		return nil
	}
	priviledge_logger.Log(logger.WARN, "Logout failed: Invalid token %s", token)
	return fmt.Errorf("invalid token %s", token)
}
