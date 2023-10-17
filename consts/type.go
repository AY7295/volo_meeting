package consts

type ErrorId = int32

const (
	WrongMessageModel ErrorId = -1
	WrongMeeting      ErrorId = -2
	InvalidId         ErrorId = -3
)

type Event string

const (
	Description Event = "description"
	Candidate   Event = "iceCandidate"
	Device      Event = "device"
	Member      Event = "member"
	Leave       Event = "leave"
	Error       Event = "error"
)

type EmitEvent int

const (
	Message EmitEvent = iota
	Join
	Err
	Close
)
