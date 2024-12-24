package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	_ "path/filepath"
	"strconv"
	"strings"
	_ "syscall"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "\\quit" {
			break
		}
		tokens := strings.Fields(input)
		if len(tokens) == 0 {
			continue
		}
		cmd := tokens[0]
		args := tokens[1:]

		switch cmd {
		case "cd":
			cd(args)
		case "pwd":
			pwd()
		case "echo":
			echo(args)
		case "kill":
			kill(args)
		case "ps":
			ps()
		default:
			if strings.Contains(input, "|") {
				handlePipeline(input)
			} else {
				executeExternalCommand(cmd, args)
			}
		}
	}
}

func cd(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: cd <directory>")
		return
	}
	dir := args[0]
	if dir == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		dir = home
	}
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
}

func pwd() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cwd)
}

func echo(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func kill(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: kill <pid>")
		return
	}
	pid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("invalid PID")
		return
	}
	p := os.Process{Pid: pid}
	err = p.Kill()
	if err != nil {
		fmt.Println(err)
	}
}

func ps() {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(output))
}

func handlePipeline(input string) {
	parts := strings.Split(input, "|")
	if len(parts) < 2 {
		executeExternalCommand(strings.TrimSpace(parts[0]), []string{})
		return
	}
	var prevReader *os.File
	var prevWriter *os.File
	for i := 0; i < len(parts)-1; i++ {
		var err error
		prevReader, prevWriter, err = os.Pipe()
		if err != nil {
			fmt.Println(err)
			return
		}
		go executeCommandWithPipes(strings.TrimSpace(parts[i]), []string{}, prevWriter, prevReader)
	}
	executeCommandWithPipes(strings.TrimSpace(parts[len(parts)-1]), []string{}, os.Stdout, prevReader)
}

func executeCommandWithPipes(cmd string, args []string, stdout, stdin *os.File) {
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 0 {
		return
	}
	command := exec.Command(cmdParts[0], cmdParts[1:]...)
	command.Stdout = stdout
	command.Stderr = os.Stderr
	command.Stdin = stdin
	err := command.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func executeExternalCommand(cmd string, args []string) {
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		fmt.Println(err)
	}
}
