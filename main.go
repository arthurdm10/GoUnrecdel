package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	unrecdel "./unrecdel/src"
)

func main() {
	runtime.GOMAXPROCS(0)

	rand.Seed(time.Now().UnixNano())

	filePath := flag.String("path", "", "Path to file or directory to be deleted")

	flag.Parse()

	if *filePath == "" || !unrecdel.PathExists(*filePath) {
		fmt.Println("Invalid path")
	} else {
		stat, err := os.Stat(*filePath)
		if err == nil {
			if stat.IsDir() {
				start := time.Now()
				unrecdel.DeleteDir(*filePath)
				done := time.Now().Sub(start)
				fmt.Println(done.String())
			} else {
				if unrecdel.DeleteFile(*filePath) {
					fmt.Println("Done")
				}
			}
		} else {
			fmt.Println(err.Error())
		}
	}
}
