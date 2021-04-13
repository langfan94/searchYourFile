package main

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

type workerStatus struct {
	id     int
	status string
}

var query = "main.go"

var workerId = 1

// var workers = make(map[int]string)
var workers = sync.Map{}
var maxWorkers = 32

var leaveDir = 0
var doneDir = 0
var resNum = 0

var tasks = make(chan string, 32)
var newTasks = make(chan string, 32)

// var extendTasks = make(chan string)

var status = make(chan workerStatus)
var workDone = make(chan bool)
var matches = make(chan bool)
var result = make(chan int)

func main() {
	start := time.Now()
	go monitorTasks()
	// go monitorExtendTasks()
	go monitorResult()
	go monitorWorkDone()
	go monitorStatus()
	newTasks <- "/Users/xf/Documents/works/"
	fmt.Println(<-result, "<-result")
	fmt.Println(time.Since(start))
}

func monitorResult() {
	for {
		<-matches
		resNum++
	}
}

func monitorWorkDone() {
	for {
		<-workDone
		doneDir++
		fmt.Println("doneDir", doneDir, leaveDir)
		if doneDir == leaveDir {
			result <- resNum
		}
	}
}

func monitorStatus() {
	for {
		wStatus := <-status
		fmt.Println("wStatus", wStatus.id, wStatus.status)
		// workers[wStatus.id] = wStatus.status
		workers.Store(wStatus.id, wStatus.status)
	}
}

func monitorTasks() {
	for {
		newT := <-newTasks
		leaveDir++
		// workerNum := len(workers)
		workerNum := 0
		var hasWaiting = false
		workers.Range(func(id, status interface{}) bool {
			workerNum++
			fmt.Println("id", id, "status", status)
			if status == "waiting" {
				hasWaiting = true
			}
			return true
		})
		fmt.Println("hasWaiting", hasWaiting, workers)
		if !hasWaiting && workerId <= maxWorkers {
			go worker(workerId)
			workerId++
			tasks <- newT
		} else {
			select {
			case tasks <- newT:
				fmt.Println("run task")
			default:
				// tasks <- newT
				leaveDir--
				fmt.Println("out of memory", newT)
			}
			// tasks <- newT
		}
	}
}

// func monitorExtendTasks() {
// 	var id = 0
// 	var extendsMap = make(map[int]string)
// 	for {
// 		select {
// 		case eTask := <-extendTasks:
// 			if eTask != "" {
// 				extendsMap[id] = eTask
// 				id++
// 			}
// 		default:
// 			workerNum := 0
// 			var waitingNum = 0
// 			workers.Range(func(id, status interface{}) bool {
// 				workerNum++
// 				if status == "waiting" {
// 					waitingNum++
// 				}
// 				return true
// 			})
// 			if waitingNum > 0 {
// 				for i := 0; i < waitingNum; i++ {
// 					tasks <- extendsMap[i]
// 					delete(extendsMap, i)
// 				}
// 			}
// 		}
// 	}
// }

func worker(workerId int) {
	for {
		// status <- workerStatus{workerId, "waiting"}
		task := <-tasks
		status <- workerStatus{workerId, "working"}
		files, err := ioutil.ReadDir(task)
		if err == nil {
			for _, file := range files {
				name := file.Name()
				isDir := file.IsDir()
				fmt.Println("name", task+name)
				if name == query {
					matches <- true
				}
				if isDir {
					newTasks <- task + name + "/"
				}
			}
			workDone <- true
		}
		status <- workerStatus{workerId, "waiting"}
	}
}
