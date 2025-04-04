// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

//worker function 
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer) {
	defer wg.Done()
	maxRetries := 3 //set the max amount of retries for connection attemps
    for addr := range tasks {
		var success bool
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr) //attempt tcp connection
		if err == nil { // if statement if the connection is successful
			conn.Close()
			fmt.Printf("Connection to %s was successful\n", addr)
			success = true
			break //exit loop
		}
		backoff := time.Duration(1<<i) * time.Second //if there is no connection calculate total backoff and retry
		fmt.Printf("Attempt %d to %s failed. Waiting %v...\n", i+1,  addr, backoff)
		time.Sleep(backoff) //wait before trying again
	    }
		if !success { //if everything fails print error message
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
	}
}

func main() {

	var wg sync.WaitGroup //waitgroup to manage go routines
	tasks := make(chan string, 100) //

	//command line flags
    target := flag.String("target","scanme.nmap.org", "specify the IP address or hostname" ) //target IP address or hostname
	startport := flag.Int("start port", 1, "Start port (default: 1) ") //starting port for scanning
	endport := flag.Int("end port", 1042, "End port(default: 1024)") //ending port for scanning
	flag.Parse() //parse command line flags

	dialer := net.Dialer { //handle TCP connections 
		Timeout: 5 * time.Second, //timout for each connection
	}
  
	workers := 100 //number of worker go routines

    for i := 1; i <= workers; i++ {
		wg.Add(1) 
		go worker(&wg, tasks, dialer) //start the worker go routine
	}


	for p := *startport; p <= *endport; p++ {
		port := strconv.Itoa(p) //convert port number to a string
        address := net.JoinHostPort(*target, port) //put target and port into a address
		tasks <- address //send address to worker channel
	}
	close(tasks) //close tasks
	wg.Wait() //wait for all worker go routines to finish
}