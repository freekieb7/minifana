package main

import (
	"syscall"
	"time"
	"unsafe"
)

func main() {
	registerSocket, _, sysSocketCode := syscall.Syscall(syscall.SYS_SOCKET, uintptr(syscall.AF_INET), uintptr(syscall.SOCK_STREAM), uintptr(0))

	if sysSocketCode != 0 {
		panic("Socket creation failed")
	}

	defer func() {
		print("Socket closing\n")
		syscall.Syscall(syscall.SYS_CLOSE, registerSocket, 0, 0)
		print("Socket closed\n")
	}()

	socketAddr := syscall.RawSockaddrInet4{
		Family: syscall.AF_INET,
		Port:   8080,
		Addr:   [4]byte{127, 0, 0, 1},
		Zero:   [8]uint8{},
	}

	_, _, sysBindCode := syscall.Syscall(syscall.SYS_BIND, registerSocket, uintptr(unsafe.Pointer(&socketAddr)), uintptr(syscall.SizeofSockaddrInet4))

	if sysBindCode != 0 {
		panic("Socket bind failed")
	}

	_, _, sysListenCode := syscall.Syscall(syscall.SYS_LISTEN, registerSocket, uintptr(10), 0)

	if sysListenCode != 0 {
		panic("Socket listen failed")
	}

	sizeofSockaddrAny := syscall.SizeofSockaddrAny

	for {
		registerSocketAddress, _, registerSocketAddressCode := syscall.Syscall6(syscall.SYS_ACCEPT4, registerSocket, uintptr(unsafe.Pointer(&syscall.RawSockaddrAny{})), uintptr(unsafe.Pointer(&sizeofSockaddrAny)), uintptr(0), 0, 0)

		if registerSocketAddressCode != 0 {
			panic("Socket accept failed")
		}

		//defer func() {
		//	print("Socket address closing\n")
		//	syscall.Syscall(syscall.SYS_CLOSE, registerSocketAddress, 0, 0)
		//	print("Socket address closed\n")
		//}()
		timeout := 100

		go func() {
			for {
				requestMessage := make([]byte, 1024)

				_, _, syscallPollCode := syscall.Syscall(syscall.SYS_POLL, registerSocketAddress, uintptr(1), uintptr(unsafe.Pointer(&timeout)))

				if syscallPollCode != 0 {
					panic("Socket poll failed")
				}

				//_, _, syscallReadCode := syscall.Syscall(syscall.SYS_POLL, registerSocketAddress, uintptr(unsafe.Pointer(&requestMessage[0])), uintptr(len(requestMessage)))

				print("Before read\n")
				_, _, syscallReadCode := syscall.Syscall(syscall.SYS_READ, registerSocketAddress, uintptr(unsafe.Pointer(&requestMessage[0])), uintptr(len(requestMessage)))
				print("After read\n")

				if syscallReadCode != 0 {
					panic("Socket read failed")
				}

				message := "HTTP/1.1 200 OK\r\n"
				message += "Date: " + time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT") + "\r\n"
				message += "Content-Length: 0\r\n"
				message += "\r\n" // Empty line before body
				responseMessage := []byte(message)

				_, _, syscallWriteCode := syscall.Syscall(syscall.SYS_WRITE, registerSocketAddress, uintptr(unsafe.Pointer(&responseMessage[0])), uintptr(len(responseMessage)))

				if syscallWriteCode != 0 {
					panic(syscallWriteCode)
				}

				//break
				//timer.Reset(5 * time.Millisecond)
			}
		}()
	}
}
