package consts

import (
	"path"
	"time"
)

const (
	defaultLogFileDir     = "docs/logs"
	FriendlyIdExpire      = 24 * 30 * time.Hour
	MeetingIdReader       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	DefaultMeetingIdSize  = 21
	FriendlyIdReader      = "0123456789"
	DefaultFriendlyIdSize = 8
	KeepaliveInterval     = 10 * time.Second
)

// descp immutable constants
var (
	DefaultConfigFileType = "json"
	DefaultConfigFileName = "config"
	DefaultLogFilePath    = logFilePath("zap.log")
	DefaultSQLLogFilePath = logFilePath("labor.log")
)

func logFilePath(filename string) string {
	return path.Join(defaultLogFileDir, filename)
}
