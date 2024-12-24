package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Flags
	var (
		after      = flag.Int("A", 0, "print N lines after match")
		before     = flag.Int("B", 0, "print N lines before match")
		context    = flag.Int("C", 0, "print N lines around match")
		count      = flag.Bool("c", false, "print only count of matching lines")
		ignoreCase = flag.Bool("i", false, "ignore case")
		invert     = flag.Bool("v", false, "invert match")
		fixed      = flag.Bool("F", false, "fixed string match")
		lineNum    = flag.Bool("n", false, "print line numbers")
	)
	flag.Parse()

	// If -C is set, override -A and -B
	if *context > 0 {
		*before = *context
		*after = *context
	}

	args := flag.Args()
	var files []*os.File
	if len(args) == 0 {
		files = append(files, os.Stdin)
	} else {
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", arg, err)
				continue
			}
			files = append(files, f)
		}
	}

	for _, file := range files {
		lines, err := readLines(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			continue
		}

		pattern := ""
		if flag.NArg() > 0 {
			pattern = flag.Arg(0)
		}

		var re *regexp.Regexp
		if !*fixed {
			rePattern := pattern
			if *ignoreCase {
				rePattern = "(?i)" + pattern
			}
			re, err = regexp.Compile(rePattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid regex pattern: %s\n", err)
				continue
			}
		}

		var matches []int
		for idx, line := range lines {
			match := false
			if *fixed {
				if *ignoreCase {
					match = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
				} else {
					match = strings.Contains(line, pattern)
				}
			} else {
				match = re.MatchString(line)
			}

			if *invert {
				match = !match
			}

			if match {
				matches = append(matches, idx)
			}
		}

		if *count {
			fmt.Println(len(matches))
			continue
		}

		// Apply context
		var contextLines []int
		for _, m := range matches {
			start := m - *before
			if start < 0 {
				start = 0
			}
			end := m + *after
			if end >= len(lines) {
				end = len(lines) - 1
			}
			for i := start; i <= end; i++ {
				contextLines = append(contextLines, i)
			}
		}

		// Remove duplicates and sort
		unique := make(map[int]bool)
		for _, l := range contextLines {
			unique[l] = true
		}
		var sortedLines []int
		for l := range unique {
			sortedLines = append(sortedLines, l)
		}
		// Sort the slice
		for i := 0; i < len(sortedLines); i++ {
			for j := i + 1; j < len(sortedLines); j++ {
				if sortedLines[i] > sortedLines[j] {
					sortedLines[i], sortedLines[j] = sortedLines[j], sortedLines[i]
				}
			}
		}

		// Print lines
		for _, l := range sortedLines {
			line := lines[l]
			if *lineNum {
				fmt.Printf("%d:%s\n", l+1, line)
			} else {
				fmt.Println(line)
			}
		}
	}
}

func readLines(file *os.File) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
