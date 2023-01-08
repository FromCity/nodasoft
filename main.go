package main

import (
	"fmt"
	"time"
	"sync"
	"context"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	timeCreate string // время создания
	timeFinish string // время выполнения
	taskRESULT []byte
}

func taskCreator(ctx context.Context, t chan Ttype) {
	defer close(t)
	ticker := time.NewTicker(time.Nanosecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
				ft := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
					ft = "Some error occured"
				}
				t <- Ttype{timeCreate: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
		}
	}
}

func taskWorker(a Ttype, tOut chan Ttype) {
	tt, _ := time.Parse(time.RFC3339, a.timeCreate)
	if tt.After(time.Now().Add(-20 * time.Second)) {
		a.taskRESULT = []byte("task has been successed")
	} else {
		a.taskRESULT = []byte("something went wrong")
	}
	a.timeFinish = time.Now().Format(time.RFC3339Nano)

	time.Sleep(time.Millisecond * 150)

	tOut <- a
}

func taskSorter(tChan chan Ttype, doneTasks chan Ttype, undoneTasks chan error) {
	for t := range tChan {
		if string(t.taskRESULT[14:]) == "successed" {
			doneTasks <- t
		} else {
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.timeCreate, t.taskRESULT)
		}
	}
}

func main() {

	superChan := make(chan Ttype, 10)
	ctx, finishCreator := context.WithCancel(context.Background())
	go taskCreator(ctx, superChan)

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)
	tOut := make(chan Ttype)

	go taskSorter(tOut, doneTasks, undoneTasks)
	go func() {
		// получение тасков
		for t := range superChan {
			go taskWorker(t, tOut)
		}
		finishCreator()
	}()


	result := map[int]Ttype{}
	err := []error{}
	go func() {
		muDoneTasks := &sync.Mutex{}
		muUndoneTasks := &sync.Mutex{}
		for r := range doneTasks {
			go func(mu *sync.Mutex) {
				mu.Lock()
				result[r.id] = r
				mu.Unlock()
			}(muDoneTasks)
		}
		for r := range undoneTasks {
			go func(mu *sync.Mutex) {
				mu.Lock()
				err = append(err, r)
				mu.Unlock()
			}(muUndoneTasks)
		}
		close(doneTasks)
		close(undoneTasks)
		close(tOut)
	}()

	time.Sleep(time.Second * 3)
	println("Errors:")
	for r := range err {
		println(r)
	}
	println("Done tasks:")
	for r := range result {
		println(r)
	}
}
