package main

import (
	"github.com/gen2brain/beeep"
)

// Notify -
func Notify(title, message, iconPath string, priority int) (err error) {
	err = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	if err != nil {
		return
	}

	switch {
	case priority >= 8 && priority <= 10:
		err = beeep.Alert(title, message, iconPath)
	default:
		err = beeep.Notify(title, message, iconPath)
	}

	if err != nil {
		return
	}
	return
}
