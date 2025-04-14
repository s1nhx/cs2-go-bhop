package main

import (
	"encoding/binary" // Turning []byte into Uint
	"fmt"             // Printing in console
	"strings"         // Checking for strings equality
	"syscall"         // Turning UTF-16 strings to Go-strings
	"time"            // Sleeping
	"unsafe"          // Getting pointers
)

// Stores client.dll handle
var client uintptr

// Max amount of modules in process
const maxModules = 1024

func main() {
	hProcess := get_process_handle(get_process_id_by_name("cs2.exe"))

	/* A great example of calling EnumProcessModules can be found here: https://learn.microsoft.com/en-us/windows/win32/psapi/enumerating-all-modules-for-a-process */

	var moduleHandles [maxModules]uintptr
	// The amount of bytes needed to represent all the modules
	var cbNeeded uint32

	// Get a list of process' modules and store them in a buffer
	ret, _, err := procEnumProcessModules.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&moduleHandles[0])),
		uintptr(unsafe.Sizeof(moduleHandles)),
		uintptr(unsafe.Pointer(&cbNeeded)),
	)
	if ret == 0 {
		fmt.Printf("Having a hard time calling EnumProcessModules: %s\n", err)
		fmt.Printf("Cancelling..\n")
		return
	}

	moduleCount := cbNeeded / uint32(unsafe.Sizeof(moduleHandles[0]))
	fmt.Printf("Found %d modules\n", moduleCount)

	// Loop through modules
	for i := 0; i < int(moduleCount); i++ {
		var name [MAX_PATH]uint16

		// Get module's filename (e.g. kernel32.dll, client.dll, engine.dll, etc)
		ret, _, err := procGetModuleBaseName.Call(
			uintptr(hProcess),
			uintptr(moduleHandles[i]),
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(len(name)),
		)
		if ret == 0 {
			fmt.Printf("Having a hard time calling GetModuleBaseName: %s\n", err)
			fmt.Printf("Cancelling..\n")
			return
		}

		// Turn UTF-16 to Go-string
		base_name := syscall.UTF16ToString(name[:])

		if strings.EqualFold(strings.ToLower(base_name), "client.dll") {
			fmt.Printf("Found client.dll, sitting in %d place\n", i)
			client = moduleHandles[i]
		}
	}

	off_CSPlayerPawn := client + 0x1874050 // dwLocalPlayerPawn (see https://github.com/a2x/cs2-dumper/blob/main/output/offsets.hpp#L20)
	force_jump_off := client + 0x186CD60   // dwForceJump

	fmt.Printf("Ready to hop!\n")

	for {
		C_CSPlayerPawn, err := read(hProcess, uintptr(off_CSPlayerPawn), 8)
		if err != nil {
			fmt.Printf("Couldn't read C_CSPlayerPawn (0x%X): %s\n", uintptr(off_CSPlayerPawn), err)
   break
		}

		localplayer := binary.LittleEndian.Uint64(C_CSPlayerPawn)

		is_in_air_flag, err := read(hProcess, uintptr(localplayer+0x450), 4)
		if err != nil {
			fmt.Printf("Couldn't read is_in_air_flag (0x%X): %s\n", uintptr(localplayer+0x450), err)
   break
		}

		// Two possible values:
		// 0x00008000 - On ground
		// 0xFFFFFFFF - In air
		is_in_air := binary.LittleEndian.Uint32(is_in_air_flag) & 1

		if space_pressed() && is_in_air == 0 {
			time.Sleep(25 * time.Millisecond) // You might want to play around with that 25ms value
			safe_copy(hProcess, uintptr(force_jump_off), put_bytes(0x10001), 4)
		} else {
			safe_copy(hProcess, uintptr(force_jump_off), put_bytes(0x1000100), 4)
		}
	}
}
