package main

import (
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// 1 for service1 2 for service2 etc
type service int

var (
	mutex = &sync.Mutex{}
	// 6 services
	srvFd     = []*os.File{nil, nil, nil, nil, nil, nil}
	testTable = []struct {
		selectedService service
		traceOperation  func(serviceSelected service, msg string)
		msg             string
		repeatTime      int
		backoffTime     int
	}{
		{1, Info, "First Info ra 1", 20, 40},
		{5, Info, "First Info ra 5", 200, 4},
		{5, Error, "First Error ra 5", 20, 50},
		{1, Warn, "First Warn ra 1", 20, 10},
		{1, Error, "First Error ra 1", 20, 5},
		{2, Info, "First Info ra 2", 10, 40},
		{3, Info, "First Info ra 3", 100, 10},
		{4, Warn, "First Warn ra 4", 1000, 1},
		{2, Warn, "First Warn ra 2", 50, 10},
		{2, Error, "First Error ra 2", 5, 5},
		{5, Warn, "First Warn ra 5", 400, 100},
		{1, Info, "Second Info ra 11", 50, 0},
		{6, Error, "First Error ra 6", 1000, 1},
	}
)

// selectService switch to proper log file
func selectService(selectedService service) *log.Entry {
	// test service index
	index := int(selectedService)
	if index > len(srvFd) || index <= 0 {
		errorMsg := strconv.Itoa(index) + " is an invalid Service Index"
		panic(errorMsg)
	}

	serv := srvFd[index-1]
	serviceName := "service" + strconv.Itoa(index)

	log.SetOutput(serv)
	return log.WithFields(log.Fields{"service": serviceName})
}

// Info to trace info
func Info(serviceSelected service, msg string) {
	mutex.Lock()
	defer mutex.Unlock()
	selectService(serviceSelected).Info(msg)
}

// Warn to trace warning
func Warn(serviceSelected service, msg string) {
	mutex.Lock()
	defer mutex.Unlock()
	selectService(serviceSelected).Warn(msg)
}

// Error to trace errors
func Error(serviceSelected service, msg string) {
	mutex.Lock()
	defer mutex.Unlock()
	selectService(serviceSelected).Error(msg)
}

func main() {
	var err error
	var wg sync.WaitGroup

	// open log files fds
	for i := 0; i < len(srvFd); i++ {
		srvFd[i], err = os.OpenFile("service"+strconv.Itoa(i+1)+".log", os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err.Error())
		}
	}
	// close fds
	defer func() {
		for i := 0; i < len(srvFd); i++ {
			if srvFd[i] != nil {
				srvFd[i].Close()
			}
		}
	}()
	// set JSON formatter and log level
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	// init wait group
	wg.Add(len(testTable))
	for i := 0; i < len(testTable); i++ {
		go func(i int) {
			for k := 0; k < testTable[i].repeatTime; k++ {
				msg := testTable[i].msg + " for the " + strconv.Itoa(k+1) + " time(s)"
				// Call tracing function
				testTable[i].traceOperation(testTable[i].selectedService, msg)
				// delay a bit
				time.Sleep(time.Duration(testTable[i].backoffTime) * time.Millisecond)
			}
			wg.Done()
		}(i)
	}

	// wait for all to complete
	wg.Wait()
}
