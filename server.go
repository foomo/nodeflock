package nodeflock

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type server struct {
	flock *Flock
}

func (s *server) handleConnection(conn net.Conn) {
	var headerBuffer [1]byte
	header := ""
	for {
		_, readErr := conn.Read(headerBuffer[0:])
		if readErr != nil {
			return
		}
		// let us read with 1 byte steps on conn until we find "{"
		if string(headerBuffer[0:1]) == "{" {
			jsonLength, strconvErr := strconv.Atoi(string(header))
			if strconvErr != nil {
				return
			}
			bytesRead := 0
			jsonBytes := []byte{}
			for bytesRead < jsonLength {
				n, readErr := conn.Read(jsonBytes)
				if readErr != nil {
					return
				}
				bytesRead += n
			}
			//resultBytes, errs.flock.CallRaw()
			continue
		}
		header += string(headerBuffer[0:])
	}
}

func RunServer(addr string, flockSize int, jsModuleFile string) error {
	flock, flockErr := NewFlock(jsModuleFile, flockSize)
	if flockErr != nil {
		return flockErr
	}
	s := &server{
		flock: flock,
	}
	return s.run(addr, flockSize, jsModuleFile)
}

func (s *server) run(addr string, flockSize int, jsModuleFile string) error {

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		// this blocks until connection or error
		conn, err := ln.Accept()
		if err != nil {
			log.Println("RunSocketServer: could not accept connection", fmt.Sprint(err))
			continue
		}
		// a goroutine handles conn so that the loop can accept other connections
		go func() {
			s.handleConnection(conn)
			conn.Close()
		}()
	}
}
