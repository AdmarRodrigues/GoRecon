package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func worker(target string, ports <-chan int, results chan<- int, wg *sync.WaitGroup, withGrabBanner bool) {
	defer wg.Done()
	timeout := 800 * time.Millisecond

	for port := range ports {
		addr := fmt.Sprintf("%s:%d", target, port)
		conn, err := net.DialTimeout("tcp", addr, timeout)

		if err != nil {
			//log.Printf("Error: %v", err)
			continue
		}

		if withGrabBanner {
			resultGrab := bannerGrab(conn, 2*time.Second)
			if resultGrab != "" {
				fmt.Println("banner TCP grab result: ", resultGrab)
			} else {
				fmt.Println("No TCP banner grabbed...")
			}
		}

		conn.Close()
		results <- port
	}

}

func bannerGrab(conn net.Conn, timeout time.Duration) string {
	conn.SetDeadline(time.Now().Add(timeout))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return ""
	}
	banner := string(buf[:n])
	for i, c := range banner {
		if c == '\n' || c == '\r' {
			banner = banner[:i]
			break
		}
	}
	if len(banner) > 80 {
		banner = banner[:80] + "..."
	}
	return banner
}

func bannerFlagHTTP(target string) {
	resp, err := http.Get(target)
	if err != nil {
		log.Fatalf("Error making request to %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code : %v", resp.StatusCode)
	}
	server := resp.Header.Get("Server")
	if server != "" {
		fmt.Println("[+] Server: ", server)
	} else {
		fmt.Println("[-] Server header not found.")
	}

	x := resp.Header.Get("X-Powered-By")
	if x != "" {
		fmt.Println("[-] X-Powered-by: ", x)
	}

}

func main() {
	target := "scanme.nmap.org"
	workers := 100

	//bannerFlagHTTP("http://scanme.nmap.org/")

	ports := make(chan int, workers)
	result := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(target, ports, result, &wg, true)
	}

	go func() {
		for i := 1; i < 1024; i++ {
			ports <- i
		}
		close(ports)
	}()

	go func() {
		wg.Wait()
		close(result)
	}()

	fmt.Printf("Scan ports in: %s\n", target)
	for openPorts := range result {
		fmt.Printf("[+] Port %d is open...\n", openPorts)
	}

}
