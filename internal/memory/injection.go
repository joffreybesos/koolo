package memory

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"strings"
	"syscall"
)

const fullAccess = windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE | windows.PROCESS_VM_READ

var handle windows.Handle
var HWND win.HWND
var getCursorPosAddr uintptr
var getCursorPosOrigBytes [32]byte
var trackMouseEventAddr uintptr
var trackMouseEventBytes [32]byte
var getKeyStateAddr uintptr
var getKeyStateOrigBytes [18]byte

func InjectorInit(pid uint32) error {
	pHandle, err := windows.OpenProcess(fullAccess, false, pid)
	if err != nil {
		return fmt.Errorf("error opening process: %w", err)
	}
	handle = pHandle

	modules, err := memory.GetProcessModules(pid)
	if err != nil {
		return fmt.Errorf("error getting process modules: %w", err)
	}

	syscall.MustLoadDLL("USER32.dll")

	for _, module := range modules {
		// GetCursorPos
		if strings.Contains(strings.ToLower(module.ModuleName), "user32.dll") {
			getCursorPosAddr, err = syscall.GetProcAddress(module.ModuleHandle, "GetCursorPos")
			getKeyStateAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "GetKeyState")
			trackMouseEventAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "TrackMouseEvent")

			err = windows.ReadProcessMemory(handle, getCursorPosAddr, &getCursorPosOrigBytes[0], uintptr(len(getCursorPosOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}

			err = stopTrackingMouseLeaveEvents()
			if err != nil {
				return err
			}

			err = windows.ReadProcessMemory(handle, getKeyStateAddr, &getKeyStateOrigBytes[0], uintptr(len(getKeyStateOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}
		}
	}
	if getCursorPosAddr == 0 || getKeyStateAddr == 0 {
		return errors.New("could not find GetCursorPos address")
	}

	return nil
}

func InjectorUnload() error {
	err := RestoreGetCursorPosAddr()
	if err != nil {
		return fmt.Errorf("error writing to memory: %w", err)
	}

	err = RestoreGetKeyState()
	if err != nil {
		return err
	}

	return windows.CloseHandle(handle)
}

func InjectCursorPos(x, y int) error {
	/*
		push rax
		mov rax, rcx
		mov dword ptr [rax], 1 // X
		mov dword ptr [rax+4], 2 // Y
		pop rax
		mov al, 1
		ret
	*/
	bytes := []byte{0x50, 0x48, 0x89, 0xC8, 0xC7, 0x00, 0x01, 0x00, 0x00, 0x00, 0xC7, 0x40, 0x04, 0x02, 0x00, 0x00, 0x00, 0x58, 0xB0, 0x01, 0xC3}

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(x))
	copy(bytes[6:], buff)

	binary.LittleEndian.PutUint32(buff, uint32(y))
	copy(bytes[13:], buff)

	return windows.WriteProcessMemory(handle, getCursorPosAddr, &bytes[0], uintptr(len(bytes)), nil)
}

func OverrideGetKeyState(key int) error {
	/*
		cmp rcx, 0x12
		mov rax, 0x8000
		ret
	*/
	bytes := []byte{0x48, 0x81, 0xF9, byte(key), 0x00, 0x00, 0x00, 0x48, 0xB8, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC3}
	return windows.WriteProcessMemory(handle, getKeyStateAddr, &bytes[0], uintptr(len(bytes)), nil)
}

func RestoreGetKeyState() error {
	return windows.WriteProcessMemory(handle, getKeyStateAddr, &getKeyStateOrigBytes[0], uintptr(len(getKeyStateOrigBytes)), nil)
}

func RestoreGetCursorPosAddr() error {
	return windows.WriteProcessMemory(handle, getCursorPosAddr, &getCursorPosOrigBytes[0], uintptr(len(getCursorPosOrigBytes)), nil)
}

// This is needed in order to let the game keep processing mouse events even if the mouse is not over the window
func stopTrackingMouseLeaveEvents() error {
	err := windows.ReadProcessMemory(handle, trackMouseEventAddr, &trackMouseEventBytes[0], uintptr(len(trackMouseEventBytes)), nil)
	if err != nil {
		return err
	}

	// and dword ptr [rcx+4], 0xFFFFFFFD
	// Modify TRACKMOUSEEVENT struct to disable mouse leave events, since we are injecting our events even if the mouse is not over the window
	disableMouseLeaveRequest := []byte{0x81, 0x61, 0x04, 0xFD, 0xFF, 0xFF, 0xFF}

	// Already hooked
	if bytes.Contains(trackMouseEventBytes[:], disableMouseLeaveRequest) {
		return nil
	}

	// We need to move back the pointer 7 bytes to get the correct position, since we are injecting 7 bytes in front of it
	num := int32(binary.LittleEndian.Uint32(trackMouseEventBytes[2:6]))
	num -= 7
	numberBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numberBytes, uint32(num))
	injectBytes := append(trackMouseEventBytes[0:2], numberBytes...)

	hook := append(disableMouseLeaveRequest, injectBytes...)

	return windows.WriteProcessMemory(handle, trackMouseEventAddr, &hook[0], uintptr(len(hook)), nil)
}
