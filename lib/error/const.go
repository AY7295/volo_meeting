package error

import (
	"errors"
	"volo_meeting/consts"
)

var (
	InvalidContext      = New(consts.PermissionDenied, errors.New("context is invalid"))
	InvalidMeetingId    = New(consts.ParamError, errors.New("meeting id is invalid"))
	InvalidTypeAssert   = New(consts.CacheError, errors.New("type assert error"))
	InvalidClosedSocket = New(consts.WSError, errors.New("channel has been closed"))
	EndedMeeting        = New(consts.MeetingError, errors.New("meeting has been ended"))
)

func NotFound(msg string) error {
	return New(consts.NotFound, errors.New(msg))
}
