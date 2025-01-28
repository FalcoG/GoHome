package main

import (
	"embed"
	"log"

	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/kvstore"

	"github.com/adrg/xdg"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed MacTrayIcon.png
var SystrayMac []byte

func populateMenu(kvStore *kvstore.KeyValueStore, app *application.App) *application.Menu {
	println("populateMenu")
	menu := app.NewMenu()

	address := kvStore.Get("ha-address").(string)
	token := kvStore.Get("ha-token").(string)

	statusCode := verifyHomeConnection(
		address,
		token,
	)

	if statusCode != 200 {
		menu.Add("Scenes unavailable").SetEnabled(false)
		println("Failed to load scenes, http", statusCode)
	} else {
		statusCode, homeScenes := getHomeScenes(address, token)

		if statusCode == 200 {
			for _, value := range homeScenes {
				println(value.Name)

				menu.Add(value.Name).OnClick(func(ctx *application.Context) {
					activateHomeScene(address, token, value.EntityID)
				})
			}
		}
	}

	return menu
}

func constructTrayMenu(app *application.App, window *application.WebviewWindow) *application.Menu {
	println("Construct tray menu")
	menuAppControl := app.NewMenu()

	menuAppControl.Add("Preferences...").SetAccelerator("CmdOrCtrl+,").OnClick(func(ctx *application.Context) {
		println("Open settings")
		window.Show()
	})

	quit := menuAppControl.Add("Quit GoHome").SetAccelerator("CmdOrCtrl+q")
	quit.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	return menuAppControl
}

func main() {
	configFilePath, configFileError := xdg.ConfigFile("GoHome/app.json")

	if configFileError != nil {
		println("Should be a fatal error because we need to write to the disk")
		// todo: crash the application
	}

	configKvStore := kvstore.New(&kvstore.Config{
		Filename: configFilePath,
		AutoSave: true,
	})

	app := application.New(application.Options{
		Name:        "GoHome",
		Description: "Control Home Assistant from your desktop",
		Services: []application.Service{
			// application.NewService(&GreetService{}),
			application.NewService(configKvStore),
		},
		// Assets: application.AlphaAssets,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	myMenu := app.NewMenu()

	systemTray := app.NewSystemTray()

	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:        "GoHome preferences",
		AlwaysOnTop: true,
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
		},
		KeyBindings: map[string]func(window *application.WebviewWindow){
			"F12": func(window *application.WebviewWindow) {
				systemTray.OpenMenu()
			},
		},
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	app.OnEvent("saved-preferences", func(e *application.CustomEvent) {
		statusCode := verifyHomeConnection(
			configKvStore.Get("ha-address").(string),
			configKvStore.Get("ha-token").(string),
		)

		println("Saved preferences")

		myMenu.Clear()
		myMenu.Append(populateMenu(configKvStore, app))
		myMenu.AddSeparator()
		myMenu.Append(constructTrayMenu(app, window))
		myMenu.Update()

		app.EmitEvent("ha-status", statusCode)

		app.Logger.Info("[Go] CustomEvent received", "name", e.Name, "data", e.Data, "sender", e.Sender, "cancelled", e.Cancelled)
	})

	app.OnEvent("get-status", func(e *application.CustomEvent) {
		statusCode := verifyHomeConnection(
			configKvStore.Get("ha-address").(string),
			configKvStore.Get("ha-token").(string),
		)

		app.EmitEvent("ha-status", statusCode)
	})

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	window.OnWindowEvent(events.Common.WindowFocus, func(e *application.WindowEvent) {
		app.Logger.Info("[ApplicationEvent] Window focus!")
	})

	if runtime.GOOS == "darwin" {
		systemTray.SetTemplateIcon(SystrayMac)
	}

	myMenu.Append(populateMenu(configKvStore, app))
	myMenu.AddSeparator()
	myMenu.Append(constructTrayMenu(app, window))

	systemTray.SetMenu(myMenu)
	systemTray.OnClick(func() {
		systemTray.OpenMenu()
		app.EmitEvent("myevent")
	})

	systemTray.AttachWindow(window).WindowOffset(5)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
