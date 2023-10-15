package model

import (
	"testing"
	"volo_meeting/config"
	"volo_meeting/lib/db/mysql"
)

func TestMeeting_AppendDevice(t *testing.T) {
	config.Init()
	mysql.Init()

	device := &Device{Id: "12243ewsd"}
	meeting := &Meeting{Id: "1qaz2wsx3edc", Devices: []Device{*device}}

	err := mysql.Instance().Model(meeting).Association("Devices").Replace(&device)
	if err != nil {
		t.Error(err)
	}
}
