package consts

type Event string

const (
	Description Event = "description"
	Candidate   Event = "iceCandidate"
	Device      Event = "device"
	KeepAlive   Event = "keepAlive"
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
