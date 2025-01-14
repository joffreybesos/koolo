package helper

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
)

type GameManager struct {
	gr *reader.GameReader
}

func NewGameManager(gr *reader.GameReader) *GameManager {
	return &GameManager{gr: gr}
}

func (gm *GameManager) ExitGame() error {
	// First try to exit game as fast as possible, without any check, useful when chickening
	hid.PressKey("esc")
	hid.Click(hid.LeftButton, hid.GameAreaSizeX/2, int(float64(hid.GameAreaSizeY)/2.2))

	for range 5 {
		if !gm.gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	// If we are still in game, probably character is dead, so let's do it nicely.
	// Probably closing the socket is more reliable, but was not working properly for me on singleplayer.
	for range 10 {
		if gm.gr.GetData(false).OpenMenus.QuitMenu {
			hid.Click(hid.LeftButton, hid.GameAreaSizeX/2, int(float64(hid.GameAreaSizeY)/2.2))

			for range 5 {
				if !gm.gr.InGame() {
					return nil
				}
				Sleep(1000)
			}
		}
		hid.PressKey("esc")
		Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

func (gm *GameManager) NewGame() error {
	if gm.gr.InGame() {
		return errors.New("character still in a game")
	}

	for range 30 {
		gm.gr.InGame()
		if gm.gr.InCharacterSelectionScreen() {
			Sleep(2000) // Wait for character selection screen to load
			break
		}
		Sleep(500)
	}

	difficultyPosition := map[difficulty.Difficulty]struct {
		X, Y int
	}{
		difficulty.Normal:    {X: 640, Y: 311},
		difficulty.Nightmare: {X: 640, Y: 355},
		difficulty.Hell:      {X: 640, Y: 403},
	}

	createX := difficultyPosition[config.Config.Game.Difficulty].X
	createY := difficultyPosition[config.Config.Game.Difficulty].Y
	hid.Click(hid.LeftButton, 600, 650)
	Sleep(250)
	hid.Click(hid.LeftButton, createX, createY)

	for range 30 {
		if gm.gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error creating game! Timeout")
}

func (gm *GameManager) clearGameNameOrPasswordField() {
	for range 16 {
		hid.PressKey("backspace")
	}
}

func (gm *GameManager) CreateOnlineGame(gameCounter int) (string, error) {
	// Enter bnet lobby
	hid.Click(hid.LeftButton, 744, 650)
	Sleep(1200)

	// Click "Create game" tab
	hid.Click(hid.LeftButton, 845, 54)
	Sleep(200)

	// Click the game name textbox, delete text and type new game name
	hid.Click(hid.LeftButton, 1000, 116)
	gm.clearGameNameOrPasswordField()
	gameName := config.Config.Companion.GameNameTemplate + fmt.Sprintf("%d", gameCounter)
	for _, ch := range gameName {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	hid.Click(hid.LeftButton, 1000, 161)
	Sleep(200)
	gamePassword := config.Config.Companion.GamePassword
	if gamePassword != "" {
		gm.clearGameNameOrPasswordField()
		for _, ch := range gamePassword {
			hid.PressKey(fmt.Sprintf("%c", ch))
		}
	}
	hid.PressKey("enter")

	for range 30 {
		if gm.gr.InGame() {
			return gameName, nil
		}
		Sleep(1000)
	}

	return gameName, errors.New("error creating game! Timeout")
}

func (gm *GameManager) JoinOnlineGame(gameName, password string) error {
	// Enter bnet lobby
	hid.Click(hid.LeftButton, 744, 650)
	Sleep(1200)

	// Click "Join game" tab
	hid.Click(hid.LeftButton, 977, 54)
	Sleep(200)

	// Click the game name textbox, delete text and type new game name
	hid.Click(hid.LeftButton, 950, 100)
	Sleep(200)
	gm.clearGameNameOrPasswordField()
	Sleep(200)
	for _, ch := range gameName {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	hid.Click(hid.LeftButton, 1130, 100)
	Sleep(200)
	gm.clearGameNameOrPasswordField()
	Sleep(200)
	for _, ch := range password {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}
	hid.PressKey("enter")

	for range 30 {
		if gm.gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error joining game! Timeout")
}

func (gm *GameManager) InGame() bool {
	return gm.gr.InGame()
}
