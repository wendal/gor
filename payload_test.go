package gor

import (
	"os"
	"testing"
)

func TestDebug(*testing.T) {
	D("ABC", ">>>>>>>>>>", 3452145)
	D("ABC", ">>>>>>>>>>", 3452145)
	D(nil)
}

func TestPayLoad(t *testing.T) {
	root := "H:/wendal_net"
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("skipping: test data directory not available:", root)
	}
	_, err := MakePayLoad(root)
	if err != nil {
		t.Error(err)
	}
}
