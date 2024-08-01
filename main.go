package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type File struct {
	FileName  string   `json:"file_name"`
	OutputArr []string `json:"outputArr"`
}

type Flags struct {
	Begin         string `json:"begin"`
	End           string `json:"end"`
	Command       string `json:"command"`
	OutputFile    string `json:"output_file"`
	FileNameStart string `json:"file_name_start"`
	FileNameEnd   string `json:"file_name_end"`
	Options       string `json:"options"`
	Help          bool   `json:"help"`
}

func parseOptions(options string, flags *Flags) error {
	data, err := os.ReadFile(options)
	if err != nil {
		return fmt.Errorf("failed to read options file: %v", err)
	}

	if err := json.Unmarshal(data, flags); err != nil {
		return fmt.Errorf("failed to unmarshal options file: %v", err)
	}

	return nil
}

func parseFile(data string, flags *Flags, outputChan chan<- File) {
	inCode := false
	fileName := ""
	localOutputMap := make(map[string]File)
	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, flags.Begin) {
			inCode = true
			fileName = strings.Split(line, flags.FileNameStart)[1]
			fileName = strings.Split(fileName, flags.FileNameEnd)[0]
			continue
		}
		if strings.Contains(line, flags.End) {
			inCode = false
			continue
		}
		if inCode {
			if _, ok := localOutputMap[fileName]; !ok {
				localOutputMap[fileName] = File{FileName: fileName}
			}
			entry := localOutputMap[fileName]
			entry.OutputArr = append(entry.OutputArr, line)
			localOutputMap[fileName] = entry
		}
	}

	for _, file := range localOutputMap {
		outputChan <- file
	}
}

func openFile(fileName string, flags *Flags, outputChan chan<- File, wg *sync.WaitGroup) {
	defer wg.Done()
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("failed to read file %s: %v\n", fileName, err)
		return
	}
	parseFile(string(data), flags, outputChan)
}

func writeOutput(outputChan <-chan File, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range outputChan {
		outputFile, err := os.Create(file.FileName)
		if err != nil {
			fmt.Printf("failed to create output file %s: %v\n", file.FileName, err)
			continue
		}
		defer outputFile.Close()

		for _, line := range file.OutputArr {
			if _, err := outputFile.WriteString(line + "\n"); err != nil {
				fmt.Printf("failed to write to output file %s: %v\n", file.FileName, err)
				break
			}
		}
	}
}

func runCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command %s: %v\nOutput: %s", command, err, string(out))
	}
	fmt.Println(string(out))
	return nil
}

func main() {
	flags := Flags{}
	flag.StringVar(&flags.Begin, "begin", "\\begin{code}", "Begin string")
	flag.StringVar(&flags.Options, "options", "", "Options file")
	flag.StringVar(&flags.OutputFile, "o", "output.txt", "Output file")
	flag.StringVar(&flags.FileNameStart, "fstart", "{", "Specify start of file name")
	flag.StringVar(&flags.FileNameEnd, "fend", "}", "Specify end of file name")
	flag.StringVar(&flags.Command, "command", "", "Command to run after output")
	flag.StringVar(&flags.End, "end", "\\end{code}", "End string")
	flag.BoolVar(&flags.Help, "help", false, "Shows help")

	flag.Parse()

	if flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	if flags.Options != "" {
		if err := parseOptions(flags.Options, &flags); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}

	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("Please provide at least one file name")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	outputChan := make(chan File, len(files))

	// Start goroutines to process each file
	for _, file := range files {
		wg.Add(1)
		go openFile(file, &flags, outputChan, &wg)
	}

	// Wait for all file processing to complete and then close the channel
	go func() {
		wg.Wait()
		close(outputChan)
	}()

	// Start goroutine to write output files
	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go writeOutput(outputChan, &writeWg)

	// Wait for all writing to complete
	writeWg.Wait()

	if flags.Command != "" {
		if err := runCommand(flags.Command); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}
}

