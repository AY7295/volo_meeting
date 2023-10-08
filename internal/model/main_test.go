package model

import (
	"os"
	"testing"
	"volo_meeting/config"
	"volo_meeting/lib/db/mysql"
)

func TestMain(m *testing.M) {
	config.Init()
	mysql.Init()

	m.Run()

	os.Exit(0)
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TestInit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
		})
	}
}
