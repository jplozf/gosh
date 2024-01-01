// ****************************************************************************
//
//	 _____ _____ _____ _____
//	|   __|     |   __|  |  |
//	|  |  |  |  |__   |     |
//	|_____|_____|_____|__|__|
//
// ****************************************************************************
// G O S H   -   Copyright © JPL 2023
// ****************************************************************************
package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	suffixes [5]string
	CpuUsage float64
)

// ****************************************************************************
// Round()
// ****************************************************************************
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// ****************************************************************************
// HumanFileSize()
// ****************************************************************************
func HumanFileSize(size float64) string {
	if size == 0 {
		return "0 B"
	} else {
		suffixes[0] = "B"
		suffixes[1] = "KB"
		suffixes[2] = "MB"
		suffixes[3] = "GB"
		suffixes[4] = "TB"

		base := math.Log(size) / math.Log(1024)
		getSize := Round(math.Pow(1024, base-math.Floor(base)), .5, 2)
		getSuffix := suffixes[int(math.Floor(base))]
		return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + string(getSuffix)
	}
}

// ****************************************************************************
// IsTextFile()
// ****************************************************************************
func IsTextFile(fName string) bool {
	readFile, err := os.Open(fName)
	if err != nil {
		return false
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()

	return (utf8.ValidString(string(fileScanner.Text())))
}

// ****************************************************************************
// GetMimeType()
// ****************************************************************************
func GetMimeType(fName string) string {
	readFile, err := os.Open(fName)
	if err != nil {
		return "NIL"
	}
	defer readFile.Close()
	// Read the response body as a byte slice
	bytes, err := ioutil.ReadAll(readFile)
	if err != nil {
		return "NIL"
	}
	mimeType := http.DetectContentType(bytes)
	return mimeType
}

// ****************************************************************************
// NumberOfFilesAndFolders()
// ****************************************************************************
func NumberOfFilesAndFolders(path string) (int, int, error) {
	nFiles := 0
	nFolders := 0

	files, err := os.ReadDir(path)
	if err != nil {
		return 0, 0, err
	}
	for _, file := range files {
		if file.IsDir() {
			nFolders++
		} else {
			nFiles++
		}
	}
	return nFiles, nFolders, nil
}

// ****************************************************************************
// GetSha256()
// ****************************************************************************
func GetSha256(fName string) (string, error) {
	file, err := os.Open(fName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hashSHA256 := sha256.New()
	if _, err := io.Copy(hashSHA256, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hashSHA256.Sum(nil)), nil
}

// ****************************************************************************
// GetCPUSample()
// ****************************************************************************
func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

// ****************************************************************************
// GetCpuUsage()
// ****************************************************************************
func GetCpuUsage() {
	for {
		idle0, total0 := getCPUSample()
		time.Sleep(3 * time.Second)
		idle1, total1 := getCPUSample()
		idleTicks := float64(idle1 - idle0)
		totalTicks := float64(total1 - total0)
		CpuUsage = 100 * (totalTicks - idleTicks) / totalTicks
	}
}

// ****************************************************************************
// DirSize()
// ****************************************************************************
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// ****************************************************************************
// FilenameWithoutExtension()
// ****************************************************************************
func FilenameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
