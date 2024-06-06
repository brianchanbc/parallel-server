package server

import (
	"encoding/json"
	"io"
	"log"
	"parserver/feed"
	"parserver/queue"
	"parserver/semaphore"
	"sync"
)

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}

type Response struct {
	Success bool                     `json:"success,omitempty"`
	Id      int64                    `json:"id,omitempty"`
	Feed    []map[string]interface{} `json:"feed,omitempty"`
}

// Run starts up the twitter server based on the configuration
// information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	mode := config.Mode
	numRoutines := config.ConsumersCount
	decoder := config.Decoder
	encoder := config.Encoder
	masterFeed := feed.NewFeed()
	// Run Sequential version
	if mode == "s" {
		runSequentialTasks(masterFeed, decoder, encoder)
		return
	}
	// Run parallel version
	newQueue := queue.NewLockFreeQueue()
	sema := semaphore.NewSemaphore(numRoutines)
	var wg sync.WaitGroup
	// Spawn go routine, each go routine call consumer and wait for work
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go consumer(newQueue, sema, &wg, masterFeed, encoder)
	}
	// Launch producer routine for processing requests and placing them to the queue
	wg.Add(1)
	go producer(newQueue, sema, &wg, decoder)
	wg.Wait()
}

// Producer reads in from decoder for requests (tasks)
// and place tasks inside a queue, if there is consumer waiting for work
// place a task inside the queue and wake one consumer up
// otherwise continue to place tasks into the queue
func producer(taskQueue *queue.LockFreeQueue, sema *semaphore.Semaphore, waitgp *sync.WaitGroup, dec *json.Decoder) {
	for {
		var req queue.Request
		// Decode request
		err := dec.Decode(&req)
		if err != nil && err != io.EOF {
			// Something is wrong with decoding
			log.Fatal(err)
		}
		if err == nil {
			// If this is a last request, wake all consumers up to finish all tasks and
			// end the program
			if req.Command == "DONE" {
				sema.Finish = true
				sema.Cond.Broadcast()
				break
			}
			// No error, so there is a request decoded
			// Enqueue the new request
			taskQueue.Enqueue(&req)
			// If there is a consumer waiting for work, wake one consumer up
			sema.Cond.Signal()
		}
	}
	waitgp.Done()
}

// Each go routine will try to grab task, process the request and send response back.
// When a consumer finishes the task, checks the queue to grab another task
// If no more tasks then wait for more tasks to process or exit if no remaining task
func consumer(taskQueue *queue.LockFreeQueue, sema *semaphore.Semaphore, waitgp *sync.WaitGroup, masterFeed feed.Feed, enc *json.Encoder) {
	for {
		// Attempt to get a new task
		sema.Down()
		// If nothing is in the queue but not received finish signal yet, wait
		for taskQueue.Length() == 0 && !sema.Finish {
			sema.Cond.Wait()
		}
		// Take out task from front of the queue
		task := taskQueue.Dequeue()
		// If nothing is taken out and received finish signal, unlock, break out and return
		if task == nil && sema.Finish {
			sema.Up()
			break
		}
		// Unlock
		sema.Up()
		// Work on the task
		performTask(task, masterFeed, enc)
	}
	waitgp.Done()
}

// performTask works on processing a request and prepares a Response to be placed on Encoder
func performTask(task *queue.Request, masterFeed feed.Feed, enc *json.Encoder) {
	// Create and initialize new Response struct
	res := Response{}
	res.Id = task.Id
	if task.Command == "ADD" {
		// Process ADD request
		masterFeed.Add(task.Body, task.Timestamp)
		res.Success = true
	} else if task.Command == "REMOVE" {
		// Process REMOVE request
		status := masterFeed.Remove(task.Timestamp)
		res.Success = status
	} else if task.Command == "CONTAINS" {
		// Process CONTAINS request
		status := masterFeed.Contains(task.Timestamp)
		res.Success = status
	} else if task.Command == "FEED" {
		// Process FEED request
		feeds := masterFeed.GetFeeds()
		res.Feed = feeds
	}
	// Place Response into encoder
	err := enc.Encode(&res)
	if err != nil {
		// Something is wrong with encoding
		log.Fatal(err)
	}
}

// runSequentialTasks runs all tasks sequentially
func runSequentialTasks(masterFeed feed.Feed, dec *json.Decoder, enc *json.Encoder) {
	for {
		var req queue.Request
		// Decode request
		decodeErr := dec.Decode(&req)
		if decodeErr != nil && decodeErr != io.EOF {
			log.Fatal(decodeErr)
		}
		if decodeErr == nil {
			// Once received DONE request, break out and return
			if req.Command == "DONE" {
				break
			}
		}
		// Process various types of requests and place into encoder
		res := Response{}
		res.Id = req.Id
		if req.Command == "ADD" {
			masterFeed.Add(req.Body, req.Timestamp)
			res.Success = true
		} else if req.Command == "REMOVE" {
			status := masterFeed.Remove(req.Timestamp)
			res.Success = status
		} else if req.Command == "CONTAINS" {
			status := masterFeed.Contains(req.Timestamp)
			res.Success = status
		} else if req.Command == "FEED" {
			feeds := masterFeed.GetFeeds()
			res.Feed = feeds
		}
		encodeErr := enc.Encode(&res)
		if encodeErr != nil {
			log.Fatal(encodeErr)
		}
	}
}
