package main

import (
	"strings" // Checking for strings equality
	"syscall" // Getting NewLazyDLLs and turning UTF-16 strings to Go-strings
	"unsafe"  // Getting pointers and data sizes
)

const (
	MAX_PATH = 260
)

type PROCESSENTRY32 struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [MAX_PATH]uint16
}

// Stores Win32 functions
type win32 struct{}

var (
	modPsapi  = syscall.NewLazyDLL("psapi.dll")
	moduser32 = syscall.NewLazyDLL("user32.dll")

	procOpenProcess              = modkernel32.NewProc("OpenProcess")
	procGetAsyncKeyState         = moduser32.NewProc("GetAsyncKeyState")
	procCreateToolhelp32Snapshot = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procCloseHandle              = modkernel32.NewProc("CloseHandle")
	procProcess32First           = modkernel32.NewProc("Process32FirstW")
	procProcess32Next            = modkernel32.NewProc("Process32NextW")

	procEnumProcessModules = modPsapi.NewProc("EnumProcessModules")
	procGetModuleBaseName  = modPsapi.NewProc("GetModuleBaseNameW")

	// Why would GetAsyncKeyState() accept any value above uint8?
	VK_SPACE           int    = 0x20
	TH32CS_SNAPPROCESS uint32 = 0x2

	Win32 win32
)

func (win32) GetAsyncKeyState(vKey int) (keyState uint16) {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vKey))
	return uint16(ret)
}

func (win32) Process32First(snapshot uintptr, pe *PROCESSENTRY32) bool {
	if pe.Size == 0 {
		pe.Size = uint32(unsafe.Sizeof(*pe))
	}
	ret, _, _ := procProcess32First.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(pe)))

	return ret != 0
}

func (win32) Process32Next(snapshot uintptr, pe *PROCESSENTRY32) bool {
	if pe.Size == 0 {
		pe.Size = uint32(unsafe.Sizeof(*pe))
	}
	ret, _, _ := procProcess32Next.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(pe)))

	return ret != 0
}

func (win32) CreateToolhelp32Snapshot(flags, processId uint32) uintptr {
	ret, _, _ := procCreateToolhelp32Snapshot.Call(
		uintptr(flags),
		uintptr(processId))

	if ret <= 0 {
		return uintptr(0)
	}

	return uintptr(ret)
}

func (win32) CloseHandle(object uintptr) bool {
	ret, _, _ := procCloseHandle.Call(
		uintptr(object))
	return ret != 0
}

func (win32) OpenProcess(desiredAccess uint32, inheritHandle bool, processId uint32) (handle uintptr) {
	inherit := 0
	if inheritHandle {
		inherit = 1
	}

	ret, _, _ := procOpenProcess.Call(
		uintptr(desiredAccess),
		uintptr(inherit),
		uintptr(processId))

	return uintptr(ret)
}

// Gets ID of Win32 process by its name
func get_process_id_by_name(process_name string) uint32 {
	// Make a snapshot to get all processes
	handle := Win32.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)

	// Close handle after it gets out of scope
	defer Win32.CloseHandle(handle)

	// Create entry to store process' data
	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))

	// Get the first process
	Win32.Process32First(handle, &entry)

	// Loop through all the processes
	for {
		// Turn UTF-16 string to Go-string
		exe_name := syscall.UTF16ToString(entry.ExeFile[:])

		// Check if the name of required process checks out
		if strings.EqualFold(strings.ToLower(exe_name), strings.ToLower(process_name)) {
			return entry.ProcessID
		}

		// If it don't, search again
		any_processes_left := Win32.Process32Next(handle, &entry)

		// See https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32next#return-value
		if !any_processes_left {
			return 0
		}
	}
}

// Gets Win32 process' handle
func get_process_handle(pid uint32) uintptr {
	res := Win32.OpenProcess(0xFFFF, false, pid)
	return res
}

func space_pressed() bool {
	return Win32.GetAsyncKeyState(VK_SPACE) != 0
}
