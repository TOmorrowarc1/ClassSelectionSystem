package privilege

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/concurrentmap"
	"github.com/TOmorrowarc1/ClassSelectionSystem/utils/logger"
)

type AccountInfo struct {
	UserName  string
	Privilege int
}

const Length = 16

var (
	privilegeMap    *concurrentmap.ConcurrentMap[string, AccountInfo]
	privilegeLogger *logger.Logger
)

func InitPrivilegeSystem() {
	privilegeMap = concurrentmap.NewConcurrentMap[string, AccountInfo]()
	privilegeLogger = logger.GetLogger()
	// The privilegeMap is not persisted.
}

func generateToken() string {
	// Generate a random token of LENGTH bytes and return its hex encoding, which written by the Gemini.
	randomBytes := make([]byte, Length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		privilegeLogger.Log(logger.Error, "Failed to generate random bytes: %v", err)
		return ""
	}
	return hex.EncodeToString(randomBytes)
}

func UserLogIn(accountInfo AccountInfo) string {
	token := generateToken()
	privilegeMap.WritePair(token, &accountInfo)
	privilegeLogger.Log(logger.Info, "User %s with privilege %d get token %s", accountInfo.UserName, accountInfo.Privilege, token)
	return token
}

func UserAccess(token string) (AccountInfo, error) {
	if accountInfo, ok := privilegeMap.ReadPair(token); ok {
		return accountInfo, nil
	}
	privilegeLogger.Log(logger.Warn, "Access denied: Invalid token %s", token)
	return AccountInfo{}, fmt.Errorf("invalid token %s", token)
}

func UserLogOut(token string) error {
	if _, ok := privilegeMap.ReadPair(token); ok {
		privilegeMap.DeletePair(token)
		privilegeLogger.Log(logger.Info, "Token %s logged out successfully", token)
		return nil
	}
	privilegeLogger.Log(logger.Warn, "Logout failed: Invalid token %s", token)
	return fmt.Errorf("invalid token %s", token)
}
