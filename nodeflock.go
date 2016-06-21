package nodeflock

import "errors"

type CallResult struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
	Log    []struct {
		Level   string   `json:"level"`
		Message string   `json:"message"`
		Stack   []string `json:"stack"`
	} `json:"log"`
}

type Flock struct {
	size            int
	processes       []*process
	chanProcessPool chan *process
}

func NewFlock(jsModuleFile string, size int) (f *Flock, err error) {
	f = &Flock{
		size:            size,
		processes:       []*process{},
		chanProcessPool: make(chan *process),
	}
	for i := 0; i < size; i++ {
		p, processErr := newProcess(jsModuleFile)
		if processErr != nil {
			err = processErr
			return
		}
		f.processes = append(f.processes, p)
	}
	/*
			go func() {
		        busy := map[string]*process
				for {
					select {
					case <-f.chanProcessPool:

					}
				}
			}()
	*/
	return
}

func (f *Flock) Call(funcName string, args ...interface{}) (result CallResult, err error) {
	f.chanProcessPool <- nil
	process := <-f.chanProcessPool
	if process == nil {
		err = errors.New("could not get process")
		return
	}
	return
}
