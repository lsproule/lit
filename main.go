package main 

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
  "os/exec"
)

type Flags struct {
	Begin   string `json:"begin"`
	End     string `json:"end"`
	Command string `json:"command"`
	OutputFile  string `json:"output_file"`
	options string
	outputArr  []string
  output string
	help    bool
}

func ParseOptions(options string, flags *Flags) {
  options_file, err := os.Open(options)
  if err != nil {
    fmt.Println("Error: ", err)
    os.Exit(1)
  }

  stat, err := options_file.Stat()
  if err != nil {
    fmt.Println("Error: ", err)
    os.Exit(1)
  }

  bs := make([]byte, stat.Size())
  _, err = options_file.Read(bs)

  if err != nil {
    fmt.Println("Error: ", err)
    os.Exit(1)
  }
  json.Unmarshal([]byte(bs), flags)

}

func ParseFile(data string, flags *Flags) string {
	in_code := false
	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, flags.Begin) {
			in_code = true
			continue
		}
		if strings.Contains(line, flags.End) {
			in_code = false
			continue
		}
		if in_code {
			flags.outputArr = append(flags.outputArr, line)
		}
	}
  flags.output = strings.Join(flags.outputArr[:], "\n")
	return flags.output 
}

func Openfile(file_name string) (string, error) {
	file, err := os.Open(file_name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Error: ", err)
		return "", err
	}

	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil {
		fmt.Println("Error: ", err)
		return "", err
	}
	str := string(bs)
	return str, nil
}

func WriteOutput(flags Flags) {
  os.WriteFile(flags.OutputFile, []byte(flags.output), 0644)
}

func RunCommand(command string) {
  cmd := exec.Command("bash", "-c", command)
  out, err := cmd.CombinedOutput()
  if err != nil{
    fmt.Println("Error: ", err)
    fmt.Println("Output: ", string(out))
    os.Exit(1)
  }
  fmt.Println(string(out))
}

func main() {
	args := os.Args[1:]
	flags := Flags{}
	if len(args) == 0 {
		fmt.Println("Please provide a file name")
		os.Exit(1)
	}
	flag.StringVar(&flags.Begin, "begin", "\\begin{code}", "Begin string")
	flag.StringVar(&flags.options, "options", "", "Options file")
	flag.StringVar(&flags.OutputFile , "o", "output.txt", "Output file")
  flag.StringVar(&flags.Command, "command", "", "Command to run after output")
	flag.StringVar(&flags.End, "end", "\\end{code}", "End string")
	flag.BoolVar(&flags.help, "help", false, "shows help")
  
	flag.Parse()

  if flags.help{
    os.Exit(0)
  }

	var file string
  
	if flags.options != "" {
    ParseOptions(flags.options, &flags)
	}

	for index, arg := range args {
		if arg == "--" {
			file = args[index+1]
		}
	}

	data, err := Openfile(file)

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	ParseFile(data, &flags)
  WriteOutput(flags)
  if flags.Command != "" {
    RunCommand(flags.Command)
  }
}
