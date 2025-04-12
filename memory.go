package main

import (
	"syscall" // Getting NewLazyDLL
	"unsafe"  // Getting pointers
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procWriteProcessMemory = modkernel32.NewProc("WriteProcessMemory")
	procReadProcessMemory  = modkernel32.NewProc("ReadProcessMemory")
	procVirtualProtectEx   = modkernel32.NewProc("VirtualProtectEx")
)

const (
	PAGE_EXECUTE_READWRITE = 0x40
)

// Writes data to an area of memory in a specified process. See https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-writeprocessmemory
func wpm(process uintptr, baseAddress uintptr, buffer *byte, size uintptr) (err error) {
	var numberOfBytesWritten *uintptr

	ret, _, err := procWriteProcessMemory.Call(uintptr(process),
		uintptr(baseAddress),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(size),
		uintptr(unsafe.Pointer(numberOfBytesWritten)))

	if ret == 0 {
		return err
	}
	return nil
}

// Reads data from an area of memory in a specified process. See https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-readprocessmemory
func rpm(process uintptr, baseAddress uintptr, buffer *byte, size uintptr) (err error) {
	var numberOfBytesRead *uintptr

	ret, _, err := procReadProcessMemory.Call(uintptr(process),
		uintptr(baseAddress),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(size),
		uintptr(unsafe.Pointer(numberOfBytesRead)))

	if ret == 0 {
		return err
	}
	return nil
}

// Changes the state of memory protection. See https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualprotectex
func virtprot(hProcess uintptr, lpAddress uintptr, dwSize int, flNewProtect uint32, lpflOldProtect *uint32) bool {
	ret, _, _ := procVirtualProtectEx.Call(
		uintptr(hProcess),
		lpAddress,
		uintptr(dwSize),
		uintptr(flNewProtect),
		uintptr(unsafe.Pointer(lpflOldProtect)))
	return ret != 0
}

// Unprotects region of memory by setting it RWX permission. Old protection is NOT being restored.
// Using PAGE_EXECUTE_READWRITE is not really secure?..
func unprot(handle uintptr, dest uintptr, size int) (old_protect uint32) {
	virtprot(handle, dest, size, PAGE_EXECUTE_READWRITE, &old_protect)
	return old_protect
}

// Protects region of memory. Note: use `defer` keyword to automatically protect memory region back at the end of the scope
func prot(handle uintptr, dest uintptr, size int, old_protect uint32) {
	var dummy uint32
	virtprot(handle, dest, size, old_protect, &dummy)
}

// Safely copies binary data to process' memory
func safe_copy(handle uintptr, dest uintptr, data []byte, size uint) (err error) {
	// Unprotect memory to avoid access violation
	old_protect := unprot(handle, uintptr(dest), int(size))
	defer prot(handle, uintptr(dest), int(size), old_protect)

	return wpm(handle, uintptr(dest), &data[0], uintptr(size))
}

// Safely sets `size` amount of bytes to process' memory
func safe_set(handle uintptr, dest uintptr, value byte, size uint) (err error) {
	// Unprotect memory to avoid access violation
	old_protect := unprot(handle, uintptr(dest), int(size))
	defer prot(handle, uintptr(dest), int(size), old_protect)

	// Fill up byte array with the same bytes
	data := make([]byte, size)
	for i := uint(0); i < size; i++ {
		data[i] = value
	}

	return wpm(handle, uintptr(dest), &data[0], uintptr(size))
}

// Safely reads region of memory, returning []byte containing data
func read(handle uintptr, dest uintptr, size uint) (data []byte, err error) {
	// Unprotect memory to avoid access violation
	old_protect := unprot(handle, uintptr(dest), int(size))
	defer prot(handle, uintptr(dest), int(size), old_protect)

	data = make([]byte, size)

	err = rpm(handle, uintptr(dest), &data[0], uintptr(size))
	return data, err
}

// Puts byte representation of value into array of bytes.
// Size of array is being determined on call
func put_bytes(value uint64) []byte {
	if value == 0 {
		return []byte{0}
	}

	// Determine the size of array
	var size int
	temp := value
	for temp > 0 {
		size++
		temp >>= 8 // Shift `value` by 8 bits (1 byte) to the right to read next byte
	}

	// Make an array and fill it with bytes in little-endian order
	byte_array := make([]byte, size)
	for i := 0; i < size; i++ {
		byte_array[i] = byte(value >> (8 * i))
	}

	return byte_array
}
