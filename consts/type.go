package consts

type Event string

const (
	Description Event = "description"
	Candidate   Event = "iceCandidate"
	Device      Event = "device"
	KeepAlive   Event = "keepAlive"
	Join        Event = "join"
	Leave       Event = "leave"
	Error       Event = "error"
)

type EmitType string

const (
	Message EmitType = "message"
	Err     EmitType = "error"
	Close   EmitType = "close"
)
