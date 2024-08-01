package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type File struct {
	FileName  string   `json:"file_name"`
	OutputArr []string `json:"outputArr"`
}

type Flags struct {
	Begin         string            `json:"begin"`
	End           string            `json:"end"`
	Command       string            `json:"command"`
	OutputFile    string            `json:"output_file"`
	FileNameStart string            `json:"file_name_start"`
	FileNameEnd   string            `json:"file_name_end"`
	Options       string            `json:"options"`
	OutputMap     map[string]File   `json:"output_map"`
	Help          bool              `json:"help"`
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

func parseFile(data string, flags *Flags) {
	inCode := false
	fileName := ""
	flags.OutputMap = make(map[string]File)
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
			if _, ok := flags.OutputMap[fileName]; !ok {
				flags.OutputMap[fileName] = File{FileName: fileName}
			}
			entry := flags.OutputMap[fileName]
			entry.OutputArr = append(entry.OutputArr, line)
			flags.OutputMap[fileName] = entry
		}
	}
}

func openFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", fileName, err)
	}
	return string(data), nil
}

func writeOutput(flags Flags) error {
	for fileName, file := range flags.OutputMap {
		outputFile, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %v", fileName, err)
		}
		defer outputFile.Close()

		for _, line := range file.OutputArr {
			if _, err := outputFile.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write to output file %s: %v", fileName, err)
			}
		}
	}
	return nil
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

	for _, file := range files {
		data, err := openFile(file)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		parseFile(data, &flags)
		if err := writeOutput(flags); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}

	if flags.Command != "" {
		if err := runCommand(flags.Command); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}
}

