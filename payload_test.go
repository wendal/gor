package gor

import (
	"testing"
)

func TestDebug(*testing.T) {
	D("ABC", ">>>>>>>>>>", 3452145)
	D("ABC", ">>>>>>>>>>", 3452145)
	D(nil)
}

func TestPayLoad(t *testing.T) {
	_, err := MakePayLoad("H:/wendal_net")
	if err != nil {
		t.Error(err)
	}
}
