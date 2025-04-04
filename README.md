# Test1-SP

This is a simple concurrency tool, it allows users to scan a specific range of ports.
This program can scan scan specific and a range of ports making it good for networking
tasks

You can:
-scan single or multiple ports 
-specify a port range
-customizable the timeout time
-grab banners from open ports
-output files in json format

How to run:
-simply build the go program with: go build main.go
-run it with this command: ./main -target scanme.nmap.org -start-port 80 -end-port 443 -workers 50 -timeout 3 -json

Your terminal should scan through ports then output your file in a json format 
For example:

Scanning example.com on port 80/443
Scanning example.com on port 443/443
Total ports scanned: 2
Number of open ports: 1
Total time taken: 1.234s

[
  {
    "target": "scanme.nmap.org",
    "port": 80,
    "open": true,
  }
[

https://youtu.be/1AJzJadcLRM
