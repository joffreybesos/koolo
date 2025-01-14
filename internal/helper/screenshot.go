package helper

import (
	"github.com/hectorgimenez/koolo/internal/helper/winproc"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/ui"
	"image"
	"image/jpeg"
	"os"
	"unsafe"
)

func Screenshot() image.Image {
	sizeX := int(float64(hid.GameAreaSizeX) * ui.GameWindowScale())
	sizeY := int(float64(hid.GameAreaSizeY) * ui.GameWindowScale())

	// Create a device context compatible with the window
	hdcWindow, _, _ := winproc.GetWindowDC.Call(uintptr(memory.HWND))
	hdcMem, _, _ := winproc.CreateCompatibleDC.Call(hdcWindow)
	hbmMem, _, _ := winproc.CreateCompatibleBitmap.Call(hdcWindow, uintptr(sizeX), uintptr(sizeY))
	_, _, _ = winproc.SelectObject.Call(hdcMem, hbmMem)

	// Use PrintWindow to copy the window into the bitmap
	winproc.PrintWindow.Call(uintptr(memory.HWND), hdcMem, 3) // use 3 to get window content only

	// map the bitmap structure
	bmpInfo := struct {
		BiSize            uint32
		BiWidth, BiHeight int32
		BiPlanes          uint16
		BiBitCount        uint16
		BiCompression     uint32
		BiSizeImage       uint32
		BiXPelsPerMeter   int32
		BiYPelsPerMeter   int32
		BiClrUsed         uint32
		BiClrImportant    uint32
	}{
		BiSize:        40, // The size of the BITMAPINFOHEADER structure
		BiWidth:       int32(sizeX),
		BiHeight:      -int32(sizeY), // negative to indicate top-down bitmap
		BiPlanes:      1,
		BiBitCount:    32, // 32 bits-per-pixel
		BiCompression: 0,  // BI_RGB, no compression
		BiSizeImage:   0,  // 0 for BI_RGB
	}

	bufSize := sizeX * sizeY * 4
	buf := make([]byte, bufSize)
	winproc.GetDIBits.Call(
		hdcMem,
		hbmMem,
		0,
		uintptr(sizeY),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bmpInfo)),
		0, // DIB_RGB_COLORS
	)

	// Convert raw bytes to *image.RGBA
	img := image.NewRGBA(image.Rect(0, 0, sizeX, sizeY))
	copy(img.Pix, buf)

	// Windows is using BRG instead of RGB, let's swap red and blue layers
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			idx := y*img.Stride + x*4 // Calculate index for the start of the pixel
			// Swap red and blue (at idx and idx+2)
			img.Pix[idx], img.Pix[idx+2] = img.Pix[idx+2], img.Pix[idx]
		}
	}

	// Cleanup
	_, _, _ = winproc.DeleteObject.Call(hbmMem)
	_, _, _ = winproc.DeleteDC.Call(hdcMem)

	return img
}

func SaveImageJPEG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 80})
}
