// +build !cli

package main

import (
	"fmt"
	"os"

	"github.com/3devo/feconnector/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func setupSysTray(onInit func()) {
	systray.Run(onInit, onExit)
}

func fillSysTray() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("3devo serial monitor")
	mOpen := systray.AddMenuItem("Open Monitor", "Opens the serial monitor")
	go func() {
		<-mOpen.ClickedCh
		open.Run("http://localhost:8989")
	}()
	mAbout := systray.AddMenuItem("About", "About this application")
	go func() {
		<-mAbout.ClickedCh
		open.Run("https://3devo.com/support/")
	}()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
		os.Exit(0)
	}()
}

func onExit() {
	// clean up here
}
