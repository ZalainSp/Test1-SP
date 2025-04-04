// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"os"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)
//store scan results
type Result struct {
	Target      string `json:"target"`
	Port        int    `json:"port"`
	Open        bool   `json:"open"`
	Banner      string `json:"banner,omitempty"`
}


// worker function
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, results *[]Result, mutex *sync.Mutex) {
	defer wg.Done()
	maxRetries := 3 //set the max amount of retries for connection attemps
	for addr := range tasks {
		var success bool
		
		for i := 0; i < maxRetries; i++ {
			conn, err := dialer.Dial("tcp", addr) //attempt tcp connection
			if err == nil {                       // if statement if the connection is successful
				conn.Close()
				success = true
				banner(conn)
				break //exit loop
			}
			backoff := time.Duration(1<<i) * time.Second //if there is no connection calculate total backoff and retry
			time.Sleep(backoff)                          //wait before trying again
		}
		if !success { //if everything fails print error message
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
		if success { //if sucessful increment the openports count
			mutex.Lock()   //lock to ensure multiple goroutines do not interfering
			port, err := strconv.Atoi(addr[strings.LastIndex(addr, ":")+1:])
			if err != nil {
				fmt.Printf("Failed to parse port from address %s: %v\n", addr, err)
				continue
			}
			*results = append(*results, Result{
				Target: addr[:strings.LastIndex(addr, ":")],
				Port:   port,
				Open:   true,
			})
			mutex.Unlock()
		}
	}
}

func banner(conn net.Conn) { //grab banner from a successful connection
	conn.SetDeadline(time.Now().Add(5 * time.Second)) //set a timeout for the banner grabbing
	buf := make([]byte, 1042)                         //buffer to hold the banner data
	n, err := conn.Read(buf)                          //read data from connection

	if err != nil { //if there is a error print it
		fmt.Printf("error handling banner: %s \n", err)
		return
	}
	if n > 0 { //print the banner
		fmt.Printf("Banner from %s to %s \n", conn.RemoteAddr(), string(buf[:n]))
		return
	}
}

func main() {

	var wg sync.WaitGroup           //waitgroup to manage go routines
	tasks := make(chan string, 100) //

	//command line flags
	target := flag.String("target", "scanme.nmap.org", "specify the IP address or hostname") //target IP address or hostname
	targetsFlag := flag.String("targets", "", "Comma-separated list of targets (overrides -target if provided)")
	startport := flag.Int("start-port", 1, "Start port (default: 1) ")                 //starting port for scanning
	endport := flag.Int("end-port", 1042, "End port(default: 1024)")                   //ending port for scanning
	workers := flag.Int("workers", 100, "Number of concurrent workers (default: 100)") //number of workers
	timeout := flag.Int("timeout", 5, "connection timeout for each port in seconds(default: 5)")
	jsonoutput := flag.Bool("json", false, "output results in JSON format")
	portsflag := flag.String("ports", "", "list of specific ports to scan")
	flag.Parse() //parse command line flags

	var targetList []string //determine the targets
	if *targetsFlag != "" {
		targetList = strings.Split(*targetsFlag, ",")
	} else {
		targetList = append(targetList, *target)
	}

	var portstoscan []int //list of ports to scan
	if *portsflag != "" {
		ports := strings.Split(*portsflag, ",")
		for _, port := range ports {
			portnum, err := strconv.Atoi(port)
			if err != nil {
				fmt.Printf("Invalid port number '%s'. Skipping...\n", port)
				continue
			}
			portstoscan = append(portstoscan,portnum)
		}
	}else {
		for p := *startport; p<= *endport; p++ {
			portstoscan = append(portstoscan, p)
		}
	}

	dialer := net.Dialer{ //handle TCP connections
		Timeout: time.Duration(*timeout) * time.Second, //timout for each connection
	}

	var results []Result    //keep track of open ports
	var mutex sync.Mutex //to help with the access to openports

	for i := 1; i <= *workers; i++ { //increment to the amount of workers
		wg.Add(1)
		go worker(&wg, tasks, dialer, &results, &mutex) //start the worker go routine
	}

	starttime := time.Now()                 //start time of the scan
	totalports := len(portstoscan) * len(targetList) //calculate total ports

	for _, t := range strings.Split(*target, ",") { //print multiple targets
		t = strings.TrimSpace(t) // Remove spaces
		for _, p := range portstoscan {
			port := strconv.Itoa(p)              //convert port number to a string
			address := net.JoinHostPort(t, port) //put target and port into a address
			tasks <- address                     //send address to worker channel

			fmt.Printf("Scanning %s on port %d/%d\n", t, p, *endport) //long scan feedback
		}
	}
	close(tasks) //close tasks
	wg.Wait()    //wait for all worker go routines to finish

	calculatedtime := time.Since(starttime)

	//scan summary
	fmt.Printf("Total ports scanned: %d\n", totalports)
	fmt.Printf("Number of open ports: %d\n", len(results))
	fmt.Printf("Total time taken: %d\n", calculatedtime)

//output results in json format
if *jsonoutput {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling results to JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData)) // print JSON results
}
}
