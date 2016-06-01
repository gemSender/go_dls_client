package dls_cli

import (
	"testing"
	"sync"
	"log"
	"time"
)

var list *safe_list = &safe_list{arr:make([]int,0,64)}
var taskCount int = 0
var retChan chan int = make(chan int, 100)
type safe_list struct {
	sync.Mutex
	arr []int
}

func (this *safe_list) Append(i int) {
	this.Lock()
	this.arr = append(this.arr, i)
	this.Unlock()
}

func Test_Pool(t *testing.T)  {
	CreatePool("localhost:5556", 5, 30)
	defer StopPool()
	taskCount ++
	go SyncDo("abcde1", action)
	taskCount ++
	go SyncDo("abcde2", action)
	taskCount ++
	go SyncDo("abcd3e", action)
	taskCount ++
	go SyncDo("abcde`", action)
	taskCount ++
	go SyncDo("a`bcde", action)
	taskCount ++
	go SyncDo("abc6de", action)
	taskCount ++
	go SyncDo("abc7de", action)
	taskCount ++
	go SyncDo("ab8cde", action)
	taskCount ++
	go SyncDo("ab9cde", action)
	taskCount ++
	go SyncDo("hello`", action)
	for i:=0; i < taskCount; i++{
		<- retChan
	}
	log.Println("concurrent", list.arr)
	log.Println(GetPool().Len())
	list.arr = list.arr[:0]
	taskCount = 0
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	taskCount ++
	go SyncDo("hello", action)
	for i:=0; i < taskCount; i++{
		<- retChan
	}
	log.Println("consistent", list.arr)
	log.Println(GetPool().Len())
}

func action()  {
	for i := 0; i < 10; i++{
		list.Append(i)
		time.Sleep(time.Millisecond)
	}
	retChan <- 0
}