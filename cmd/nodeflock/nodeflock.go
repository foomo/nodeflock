package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

func read(comment string, pipe io.Reader) {
	fmt.Println("enter read for", comment)
	out := []byte{}
	for {
		fmt.Println("reading from", comment)
		next := make([]byte, 8)
		len, err := pipe.Read(next)

		out = append(out, next[0:len]...)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if len == 0 {
			break
		}
		fmt.Println("got", string(out))
	}

	log.Println(comment, ":", pipe, "data:", string(out))

}

func main() {
	fmt.Println("launching", os.Args[1:])
	cmd := exec.Command("node", os.Args[1:]...)
	//cmd := exec.Command("echo", "hello here i am")
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("no stdin pipe for me", err)
	}
	outPipe, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		log.Fatalln("no outpipe for me", pipeErr)
	}
	errPipe, pipeErr := cmd.StderrPipe()
	if pipeErr != nil {
		log.Fatalln("no err pipe", pipeErr)
	}
	err = cmd.Start()
	/*
		_ := func() {
			err := cmd.Wait()
			fmt.Println("game over process died", err)
			os.Exit(1)
		}
	*/
	go read("stdout", outPipe)
	go read("stderr", errPipe)
	if err != nil {
		log.Fatal("could not start command", err)
	}
	for {
		inPipe.Write([]byte(fmt.Sprint("Go time is ", time.Now())))
		time.Sleep(time.Second)
	}
}
