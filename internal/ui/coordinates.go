package ui

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/lxn/win"
)

const (
	itemBoxSize            = 33
	inventoryTopLeftX      = 846
	inventoryTopLeftY      = 369
	topCornerVendorWindowX = 109
	topCornerVendorWindowY = 147
	MercAvatarPositionX    = 36
	MercAvatarPositionY    = 39

	CubeTransmuteBtnX = 273
	CubeTransmuteBtnY = 411

	AnvilCenterX = 272
	AnvilCenterY = 333
	AnvilBtnX    = 272
	AnvilBtnY    = 450

	MainSkillButtonX = 596
	MainSkillButtonY = 693

	SecondarySkillButtonX = 686
	SecondarySkillButtonY = 693

	GambleRefreshButtonX = 390
	GambleRefreshButtonY = 515

	SkillListFirstSkillX = 687
	SkillListFirstSkillY = 590
	SkillListSkillOffset = 45

	FirstMercFromContractorListX = 175
	FirstMercFromContractorListY = 142
)

func GetScreenCoordsForItem(itm data.Item) data.Position {
	switch itm.Location {
	case item.LocationVendor, item.LocationStash, item.LocationSharedStash1, item.LocationSharedStash2, item.LocationSharedStash3:
		x := topCornerVendorWindowX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
		y := topCornerVendorWindowY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
	y := inventoryTopLeftY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

	return data.Position{X: x, Y: y}
}

func GameWindowScale() float64 {
	dpi := win.GetDpiForWindow(memory.HWND)
	return float64(dpi) / 96.0
}
