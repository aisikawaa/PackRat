package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func Ulimit() int64 {
	out, err := exec.Command("bash", "-c", "ulimit -n").Output()

	if err != nil {
		panic(err)
	}
	s := strings.TrimSpace(string(out))
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}
func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		}
		return
	}
	conn.Close()
	fmt.Println(port, "open")
}
func (ps *PortScanner) Start(fst, last int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	for port := fst; port <= last; port++ {
		wg.Add(1)
		ps.lock.Acquire(context.TODO(), 1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func main() {
	var ipAddr string
	print("Ip address: ")
	fmt.Scan(&ipAddr)
	ps := &PortScanner{
		ip:   ipAddr,
		lock: semaphore.NewWeighted(Ulimit()),
	}

	var fst, last int
	print("First port: ")
	fmt.Scan(&fst)
	print("Last port: ")
	fmt.Scan(&last)

	ps.Start(fst, last, 1000*time.Millisecond)

}
