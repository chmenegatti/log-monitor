package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var processedTransactions = make(map[string]bool)

func main() {
	logFile := "/var/log/messages"
	linesChannel := make(chan string)

	go monitorLog(logFile, linesChannel)

	for line := range linesChannel {
		processLogLine(line, logFile)
	}
}

func monitorLog(logFile string, linesChannel chan string) {

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		close(linesChannel)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file: ", err)
		}
	}(file)

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking to end of file: ", err)
		close(linesChannel)
		return
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("error reading line: ", err)
				continue
			}
			fmt.Println("erro ao ler o arquivo", err)
			break
		}
		linesChannel <- strings.TrimSpace(line)
	}
}

func processLogLine(line, filePath string) {
	if strings.Contains(line, "level=error") {
		trasactionID := extractTransactionID(line)
		if trasactionID == "" || processedTransactions[trasactionID] {
			return
		}

		processedTransactions[trasactionID] = true

		lines := findTransactionsLines(filePath, trasactionID)

		saveTransactionLines(trasactionID, lines)

	}
}

func extractTransactionID(line string) string {
	paths := strings.Split(line, " ")
	for _, part := range paths {
		if strings.Contains(part, "transaction_id") {
			return strings.TrimPrefix(part, "transaction_id=")
		}
	}
	return ""
}

func findTransactionsLines(filepath, transactionID string) []string {
	var lines []string
	file, err := os.Open(filepath)

	if err != nil {
		fmt.Println("error opening file: ", err)
		return lines
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("error closing file: ", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, transactionID) {
			lines = append(lines, line)
		}
	}

	return lines
}

func saveTransactionLines(transactionID string, lines []string) {
	filename := fmt.Sprintf("erro_%s.log", transactionID)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("error creating file: ", err)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("error closing file: ", err)
		}
	}(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println("error writing line: ", err)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("error flushing writer: ", err)
	}
}
