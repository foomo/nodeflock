package nodeflock

import (
	"fmt"
	"os"
	"time"
)

type CallResult struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
	Log    []struct {
		Level string        `json:"level"`
		Data  []interface{} `json:"data"`
		Stack []string      `json:"stack"`
	} `json:"log"`
}
type processRentalApplicationResponse struct {
	err     error
	process *process
}

type processRentalApplication struct {
	chanResponse chan processRentalApplicationResponse
}

type Flock struct {
	size             int
	processes        []*process
	chanRentalReturn chan *process
	chanRentalApply  chan chan processRentalApplicationResponse //chan *processRentalApplication
	idCounter        int
}

func NewFlock(jsModuleFile string, size int) (f *Flock, err error) {
	f = &Flock{
		size:             size,
		processes:        []*process{},
		chanRentalReturn: make(chan *process),
		chanRentalApply:  make(chan chan processRentalApplicationResponse),
	}

	go func() {
		busy := map[int]bool{}
		applications := []chan processRentalApplicationResponse{} //[]*processRentalApplication{}
		chanProcessDied := make(chan int)
		lastSourceChange := getFileChange(jsModuleFile)
		for {
			// is anyone obsolote
			numOkProcesses := 0
			for _, deathRowProcess := range f.processes {
				if _, isBusy := busy[deathRowProcess.id]; !isBusy && deathRowProcess.sourceFileChange < lastSourceChange {
					deathRowProcess.kill()
				} else {
					numOkProcesses++
				}
			}

			// spawn processes, if we lost some
			for i := numOkProcesses; i < f.size; i++ {
				p, processErr := newProcess(jsModuleFile, lastSourceChange, f.getId(), chanProcessDied)
				if processErr != nil {
					err = processErr
					return
				}
				f.processes = append(f.processes, p)
			}

			select {
			case <-time.After(time.Millisecond):
				newSourceChange := getFileChange(jsModuleFile)
				if newSourceChange != lastSourceChange {
					fmt.Println("source file update", jsModuleFile, lastSourceChange, newSourceChange)
				}
				lastSourceChange = newSourceChange
				//fmt.Println("checking source", lastSourceChange)
			case deadProcessID := <-chanProcessDied:
				// a process died
				// handle error
				// remove process
				f.processes = filterProcesses(f.processes, func(process *process) bool {
					return process.id != deadProcessID
				})
				delete(busy, deadProcessID)
			case rentalApplication := <-f.chanRentalApply:
				//fmt.Println("incoming application")
				applications = append(applications, rentalApplication)
			case returnedProcess := <-f.chanRentalReturn:
				// back to the
				//fmt.Println("got one back")
				delete(busy, returnedProcess.id)
			}
			// let us see if anyone has time
			processedApplications := []chan processRentalApplicationResponse{}
		applicationLoop:
			for _, application := range applications {

				if len(busy) == len(f.processes) {
					// they are all busy
					fmt.Println("everybody is busy")
					break
				}
				var lonelyProcess *process
				lonelyProcess = nil
				for _, process := range f.processes {
					_, isBusy := busy[process.id]
					if !isBusy && process.sourceFileChange == lastSourceChange {
						if lonelyProcess == nil {
							lonelyProcess = process
						}
						if lonelyProcess.lastCall < process.lastCall {
							lonelyProcess = process
						}
					}
				}

				// this one is not busy
				processedApplications = append(processedApplications, application)
				// make the applicant happy
				//fmt.Println("sending process to applicant", process.id)
				if lonelyProcess != nil {
					busy[lonelyProcess.id] = true
					application <- processRentalApplicationResponse{
						process: lonelyProcess,
						err:     nil,
					}
					continue applicationLoop

				} else {
					fmt.Println("could not find lonely process")
				}

			}
			// clean up
			cleanedApplications := []chan processRentalApplicationResponse{}
		applicationLoopClean:
			for _, application := range applications {
				for _, processedApplication := range processedApplications {
					if processedApplication == application {
						continue applicationLoopClean
					}
				}
				cleanedApplications = append(cleanedApplications, application)
			}
			applications = cleanedApplications
			// fmt.Println(busy)
			//fmt.Println(time.Now().Sub(start))

		}
	}()

	return
}

func getFileChange(file string) int64 {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return int64(0)
	}
	return fileInfo.ModTime().UnixNano()
}

func filterProcesses(processes []*process, filter func(*process) bool) (cleaned []*process) {
	survivors := []*process{}
	for _, process := range processes {
		if filter(process) {
			survivors = append(survivors, process)
		}
	}
	return survivors
}

func (f *Flock) getId() int {
	f.idCounter++
	return f.idCounter
}

func (f *Flock) borrowProcess() (p *process, err error) {
	// dummy implementation
	application := make(chan processRentalApplicationResponse)
	f.chanRentalApply <- application
	response := <-application
	return response.process, response.err
}

func (f *Flock) returnProcess(p *process) {
	//fmt.Println("returning process", p.id)
	f.chanRentalReturn <- p
}

func (f *Flock) CallRaw(jsonBytes []byte) (rawResult []byte, err error) {
	p, processErr := f.borrowProcess()
	if processErr != nil {
		err = processErr
		return
	}
	//fmt.Println("got a process", p.id)
	rawResult, callErr := p.rawCallJS(jsonBytes)
	f.returnProcess(p)
	return rawResult, callErr
}

func (f *Flock) Call(funcName string, args ...interface{}) (result CallResult, err error) {
	return
}
