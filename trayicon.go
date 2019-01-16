// +build !cli

package main

import (
	"fmt"
	"os"

	"github.com/3devo/feconnector/icon"
	"github.com/3devo/feconnector/models"
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
	mAbout := systray.AddMenuItem("About", "About this application")
	mReset := systray.AddMenuItem("Reset user", "Resets the user accounts")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				launchBrowserWithToken()
			case <-mAbout.ClickedCh:
				open.Run("https://3devo.com/support/")
			case <-mReset.ClickedCh:
				var users []models.User
				env.Db.All(&users)
				for _, user := range users {
					env.Db.DeleteStruct(&user)
				}
				env.HasAuth = false

				// set network back to false when deleting a user
				var config models.Config
				db.One("ID", 1, &config)
				config.OpenNetwork = false
				db.Save(&config)
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				fmt.Println("Finished quitting")
				os.Exit(0)
			}
		}
	}()
}

func onExit() {
	// clean up here
}
