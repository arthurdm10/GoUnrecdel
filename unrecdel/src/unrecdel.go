package unrecdel

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const BUFF_SIZE = 2048

const alpha string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const alphaLen = len(alpha)

func ok(err error) bool {
	return err == nil
}

func PathExists(path string) bool {
	_, err := os.Stat(filepath.Dir(path))
	return !os.IsNotExist(err)
}

func DeleteFile(path string) bool {
	for i := 0; i < 3; i++ {
		if _, err := randomWrite(path); !ok(err) {
			fmt.Printf("Failed to write to file: %s -- Error %s\n", path, err.Error())
		}
	}

	for i := 0; i < 3; i++ {
		newPath, err := randomRename(path)

		if !ok(err) {
			fmt.Printf("Failed to rename file: %s -- Error: %s", path, err.Error())
		}

		path = newPath
	}

	return ok(os.Remove(path))
}

func deleteFileRoutine(fileChannel <-chan string, totalErased chan<- int64) {
	var bytesWritten int64

	for path := range fileChannel {
		for i := 0; i < 3; i++ {
			newPath, err := randomRename(path)
			if !ok(err) {
				fmt.Printf("Failed to rename file: %s -- Error: %s\n", path, err.Error())
				continue
			}

			path = newPath
			if totalWritten, err := randomWrite(path); !ok(err) {
				fmt.Printf("Failed to write to file: %s -- Error %s\n", path, err.Error())
			} else {
				bytesWritten += totalWritten
			}
		}

		err := os.Remove(path)
		if !ok(err) {
			fmt.Printf("Failed to remove file: %s -- Error: %s\n", path, err.Error())
		}
	}

	totalErased <- bytesWritten / 3
}

func DeleteDir(path string) {
	fileChannel := make(chan string, 100)
	totalErased := make(chan int64)
	var wg sync.WaitGroup

	numCpu := runtime.NumCPU()

	for i := 0; i < numCpu; i++ {
		go deleteFileRoutine(fileChannel, totalErased)
		fmt.Printf("Routine %d started\n", i)
	}

	wg.Add(1)
	dirWalk(fileChannel, path, &wg)

	wg.Wait()
	close(fileChannel)

	var bytesErased int64

	for i := 0; i < numCpu; i++ {
		routineErased := <-totalErased
		bytesErased += routineErased
	}

	fmt.Printf("%d bytes erased!\n", bytesErased)
}

func dirWalk(fChannel chan string, dirPath string, wg *sync.WaitGroup) {
	dir, err := os.Open(dirPath)
	defer wg.Done()
	defer dir.Close()

	if ok(err) {
		files, err := dir.Readdir(0)

		if ok(err) {
			for _, file := range files {
				path := filepath.Join(dirPath, file.Name())
				if file.IsDir() {
					wg.Add(1)

					go dirWalk(fChannel, path, wg)
				} else {
					fChannel <- path
				}
			}
		} else {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(err.Error())
	}

}

func randomBytes(size int64) []byte {
	data := make([]byte, size)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	err := binary.Read(rnd, binary.LittleEndian, data)

	if !ok(err) {
		panic(err.Error())
	}

	return data
}

func randomString(size int64) string {
	runes := make([]rune, size)
	var i int64

	for i = 0; i < size; i++ {
		runes[i] = rune(alpha[rand.Intn(alphaLen)])
	}

	return string(runes)
}

func randomRename(path string) (string, error) {

	randomName := filepath.Join(filepath.Dir(path), randomString(10))

	err := os.Rename(path, randomName)

	if !ok(err) {
		return "", err
	}

	return randomName, nil
}

func randomWrite(path string) (int64, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()

	if !ok(err) {
		return 0, err
	}

	var totalBytes int64 = 0

	fileInfo, err := file.Stat()

	if ok(err) {
		size := fileInfo.Size()

		for totalBytes < size {
			n, err := file.Write(randomBytes(BUFF_SIZE))
			if !ok(err) {
				return 0, err
			}
			totalBytes += int64(n)
		}
		return totalBytes, nil
	}

	return 0, err
}
