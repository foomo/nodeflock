package nodeflock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
)

type process struct {
	//sourceFile string
	cmd         *exec.Cmd
	pipeStdIn   io.WriteCloser
	chanStderr  chan []byte
	chanStdout  chan []byte
	chanPipeErr chan error
}

type ProcessCall struct {
	Func string        `json:"func"`
	Args []interface{} `json:"args"`
}

func newProcess(sourceFile string) (p *process, err error) {
	p = &process{
		chanStdout:  make(chan []byte),
		chanStderr:  make(chan []byte),
		chanPipeErr: make(chan error),
	}
	err = p.start(sourceFile)
	return
}

func read(pipe io.Reader, pipeInfo string, pipeChan chan []byte, errChan chan error) {
	for {
		next := make([]byte, 1024)
		len, err := pipe.Read(next)
		// fmt.Println(pipeInfo, string(next[0:len]))
		pipeChan <- next[0:len]

		if err != nil {
			if err == io.EOF {
				log.Println("pipe", pipeInfo, "eof", err)
			}
			if err != nil {
				log.Println("pipe", pipeInfo, "wtf", err)
			}
			errChan <- err
			break
		}
	}
}

func (p *process) start(sourceFile string) error {
	cmd := exec.Command("nodeflock-process/process.js", sourceFile)
	pipeStdout, pipeStdoutErr := cmd.StdoutPipe()
	if pipeStdoutErr != nil {
		return pipeStdoutErr
	}
	pipeStderr, pipeStderrErr := cmd.StderrPipe()
	if pipeStderrErr != nil {
		return pipeStderrErr
	}

	pipeStdIn, pipeStinErr := cmd.StdinPipe()
	if pipeStinErr != nil {
		return pipeStinErr
	}
	p.pipeStdIn = pipeStdIn
	go read(pipeStderr, "std err", p.chanStderr, p.chanPipeErr)
	go read(pipeStdout, "std out", p.chanStdout, p.chanPipeErr)

	startErr := cmd.Start()
	if startErr != nil {
		return startErr
	}
	p.cmd = cmd
	return nil
}

func (p *process) kill() error {
	if p.cmd != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *process) rawCallJS(callBytes []byte) (callResultBytes []byte, err error) {
	// send data
	written := 0

	//fmt.Println("writing", len(callBytes), string(callBytes))

	// prepend the length
	callBytes = append([]byte(fmt.Sprint(len(callBytes))), callBytes...)

	// write to stdin
	for written < len(callBytes) {
		n, writeErr := p.pipeStdIn.Write(callBytes[written:])
		if writeErr != nil {
			err = writeErr
			return
		}
		written += n
	}

	// receive data
	callResultBytes = []byte{}
	length := 0
	headerLen := 0
	for {
		//fmt.Println("waiting for data from stdin")
		callResultBytes = append(callResultBytes, <-p.chanStdout...)
		if length == 0 {
			lenBytes := []byte{}
			for i, b := range callResultBytes {
				if string(b) == "{" {
					parsedLen, strconvErr := strconv.Atoi(string(lenBytes))
					if strconvErr != nil {
						err = strconvErr
						return
					}
					length = parsedLen
					headerLen = i
					break
				}
				lenBytes = append(lenBytes, b)
			}
		}
		if length > 0 && len(callResultBytes) >= headerLen+length {
			callResultBytes = callResultBytes[headerLen : headerLen+length]
			break
		}
	}
	return
}

// call the running js process
func (p *process) callJS(funcResult interface{}, funcName string, args ...interface{}) (result *CallResult, err error) {
	if p.cmd == nil {
		err = errors.New("process needs to be started first")
		return
	}
	call := &ProcessCall{
		Func: funcName,
		Args: args,
	}
	callBytes, jsonErr := json.Marshal(call)
	if jsonErr != nil {
		err = jsonErr
		return
	}
	callResultBytes, rawCallError := p.rawCallJS(callBytes)
	if rawCallError != nil {
		err = rawCallError
		return
	}

	result = &CallResult{
		Result: funcResult,
	}
	jsonErr = json.Unmarshal(callResultBytes, result)
	if jsonErr != nil {
		result = nil
		err = errors.New("callJS:: " + jsonErr.Error())
		return
	}
	return
}
