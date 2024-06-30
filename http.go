package main

import (
	"golang.org/x/sys/unix"
	"time"
)

func main() {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)

	if err != nil {
		panic(err)
	}

	if err := unix.Bind(fd, &unix.SockaddrInet4{
		Port: 36895,
		Addr: [4]byte{127, 0, 0, 1},
	}); err != nil {
		panic(err)
	}

	if err := unix.Listen(fd, 10); err != nil {
		panic(err)
	}

	for {
		socketAddress, _, err := unix.Accept(fd)

		if err != nil {
			panic(err)
		}

		for {
			now := time.Now()
			buf := make([]byte, 1024)

			if _, err := unix.Read(socketAddress, buf); err != nil {
				panic(err)
			}

			print("Received: " + string(buf) + "\n")

			if _, err := unix.Write(socketAddress, []byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
				panic(err)
			}

			unix.Close(socketAddress)
			print(time.Now().Sub(now).String() + "\n")

			break
		}
	}

	if err := unix.Close(fd); err != nil {
		panic(err)
	}
}
