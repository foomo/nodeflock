package nodeflock

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type server struct {
	flock *Flock
}

func (s *server) handleConnection(conn net.Conn) error {
	var headerBuffer [1]byte
	header := ""
	//iconn := 0
	for {
		_, readErr := conn.Read(headerBuffer[0:])
		if readErr != nil {
			return readErr
		}
		// let us read with 1 byte steps on conn until we find "{"
		if headerBuffer[0:1][0] == 123 {
			jsonLength, strconvErr := strconv.Atoi(string(header))
			if strconvErr != nil {
				return strconvErr
			}
			bytesRead := 1
			jsonBytes := []byte{123}
			for bytesRead < jsonLength {
				nextJSONBytes := make([]byte, 1024)
				n, readErr := conn.Read(nextJSONBytes)
				if readErr != nil && readErr != io.EOF {
					return readErr
				}
				if readErr == io.EOF && bytesRead < jsonLength {
					return errors.New(fmt.Sprintln("eof, but not done, expected", jsonLength, "from", header, "got", bytesRead, "::", string(jsonBytes)))
				}
				jsonBytes = append(jsonBytes, nextJSONBytes[0:n]...)
				bytesRead += n

			}
			//fmt.Println("raw call", string(jsonBytes), jsonBytes)
			resultBytes, errRawCall := s.flock.CallRaw(jsonBytes)
			if errRawCall != nil {
				return errRawCall
			}
			resultBytes = append([]byte(fmt.Sprint(len(resultBytes))), resultBytes...)
			//fmt.Println("ready to reply", string(resultBytes))
			bytesWritten := 0
			for bytesWritten < len(resultBytes) {
				// respond
				n, writeErr := conn.Write(resultBytes[bytesWritten:])
				if writeErr != nil {
					return writeErr
				}
				bytesWritten += n
			}
			//iconn++
			//fmt.Println("responded", iconn, conn)
			header = ""
			continue
		}
		header += string(headerBuffer[0:])
	}
}

func RunServer(addr string, flockSize int, jsModuleFile string, maxExecutionTime time.Duration) error {
	flock, flockErr := NewFlock(jsModuleFile, flockSize, maxExecutionTime)
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
		nextConn, err := ln.Accept()
		if err != nil {
			log.Println("RunSocketServer: could not accept connection", fmt.Sprint(err))
			continue
		}
		// a goroutine handles conn so that the loop can accept other connections
		go func(conn net.Conn) {
			handlingErr := s.handleConnection(conn)
			if handlingErr != nil && handlingErr != io.EOF {
				log.Println("could not handle request in socket server::", handlingErr)
				// log.Println("connection was closed by the client")
			}
			closeErr := conn.Close()
			if closeErr != nil {
				log.Println("could not close connection", closeErr)
			}
		}(nextConn)
	}
}
