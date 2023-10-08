package id

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"volo_meeting/consts"
)

func GetMeetingId() (string, error) {
	return gonanoid.Generate(consts.MeetingIdReader, consts.DefaultMeetingIdSize)
}

func GetFriendlyId() (string, error) {
	return gonanoid.Generate(consts.FriendlyIdReader, consts.DefaultFriendlyIdSize)
}

func Must() string {
	return gonanoid.Must()
}
func MustWithConfig(alphabet string, size int) string {
	return gonanoid.MustGenerate(alphabet, size)
}
