package game

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"sync"
)

var userID string
var userIDOnce sync.Once

func generateUserID() string {
	userIDOnce.Do(func() {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		hash := md5.Sum([]byte(hostname))
		userID = "PLAYER_" + hex.EncodeToString(hash[:])[:6]
	})
	return userID
}
