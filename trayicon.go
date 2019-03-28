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
	systray.SetTitle("DevoVision")
	mOpen := systray.AddMenuItem("Open DevoVision", "Opens the DevoVision browser interface")
	mHelp := systray.AddMenuItem("Help", "Get help using this application")
	mReset := systray.AddMenuItem("Reset user (Requires restart)", "Resets the user accounts (Requires restart)")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				launchBrowserWithToken()
			case <-mHelp.ClickedCh:
				open.Run("https://3devo.com/devovision-help")
			case <-mReset.ClickedCh:
				var users []models.User
				env.Db.All(&users)
				for _, user := range users {
					env.Db.DeleteStruct(&user)
				}
				// set network back to false when deleting a user
				var config models.Config
				db.One("ID", 1, &config)
				config.OpenNetwork = false
				config.UserCreated = false
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
