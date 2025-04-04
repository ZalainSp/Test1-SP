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
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, openports *int, mutex *sync.Mutex) {
	defer wg.Done()
	maxRetries := 3 //set the max amount of retries for connection attemps
    for addr := range tasks {
		var success bool
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr) //attempt tcp connection
		if err == nil { // if statement if the connection is successful
			conn.Close()
			success = true
			break //exit loop
		}
		backoff := time.Duration(1<<i) * time.Second //if there is no connection calculate total backoff and retry
		time.Sleep(backoff) //wait before trying again
	    }
		if !success { //if everything fails print error message
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
		if success { //if sucessful increment the openports count
			mutex.Lock() //lock to ensure multiple goroutines do not interfering 
			*openports++ //increment openports
			mutex.Unlock() //unlock after incerement 
		}else{
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries) //if the connection fails print error message
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
	workers := flag.Int("workers", 100, "Number of concurrent workers (default: 100)") //number of workers
	timeout:= flag.Int("timeout", 5, "connection timeout for each port in seconds(default: 5)")
	flag.Parse() //parse command line flags
	

	dialer := net.Dialer { //handle TCP connections 
		Timeout: time.Duration(*timeout)*time.Second, //timout for each connection
	}
  
	var openports int //keep track of open ports
	var mutex sync.Mutex //to help with the access to openports
	

    for i := 1; i <= *workers; i++ { //increment to the amount of workers
		wg.Add(1) 
		go worker(&wg, tasks, dialer, &openports, &mutex) //start the worker go routine
	}
	
	starttime := time.Now() //start time of the scan


	for p := *startport; p <= *endport; p++ {
		port := strconv.Itoa(p) //convert port number to a string
        address := net.JoinHostPort(*target, port) //put target and port into a address
		tasks <- address //send address to worker channel
	}
	close(tasks) //close tasks
	wg.Wait() //wait for all worker go routines to finish

	calculatedtime := time.Since(starttime) 

	//scan summary
	fmt.Printf("Total ports scanned: %d\n",*endport-*startport+1)
	fmt.Printf("Number of open ports: %d\n", openports)
	fmt.Printf("Total time taken: %d\n", calculatedtime)
}