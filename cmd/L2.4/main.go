package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type sortOptions struct {
	column      int
	numeric     bool
	reverse     bool
	unique      bool
	month       bool
	ignoreSpace bool
	checkSorted bool
	humanSort   bool
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func writeFile(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func sortLines(lines []string, opts sortOptions) ([]string, error) {
	if opts.unique {
		lines = uniqueLines(lines)
	}

	if opts.ignoreSpace {
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " ")
		}
	}

	sorter := &lineSorter{
		lines: lines,
		opts:  opts,
	}

	if opts.checkSorted {
		if sort.IsSorted(sorter) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("lines are not sorted")
		}
	}

	sort.Sort(sorter)
	return sorter.lines, nil
}

type lineSorter struct {
	lines []string
	opts  sortOptions
}

func (ls *lineSorter) Len() int {
	return len(ls.lines)
}

func (ls *lineSorter) Swap(i, j int) {
	ls.lines[i], ls.lines[j] = ls.lines[j], ls.lines[i]
}

func (ls *lineSorter) Less(i, j int) bool {
	colI := getColumn(ls.lines[i], ls.opts.column)
	colJ := getColumn(ls.lines[j], ls.opts.column)

	if ls.opts.numeric {
		numI, errI := strconv.ParseFloat(colI, 64)
		numJ, errJ := strconv.ParseFloat(colJ, 64)
		if errI == nil && errJ == nil {
			if ls.opts.reverse {
				return numI > numJ
			}
			return numI < numJ
		}
	}

	if ls.opts.reverse {
		return colI > colJ
	}
	return colI < colJ
}

func getColumn(line string, column int) string {
	columns := strings.Fields(line)
	if column > 0 && column <= len(columns) {
		return columns[column-1]
	}
	return line
}

func uniqueLines(lines []string) []string {
	lineSet := make(map[string]struct{})
	var unique []string
	for _, line := range lines {
		if _, exists := lineSet[line]; !exists {
			lineSet[line] = struct{}{}
			unique = append(unique, line)
		}
	}
	return unique
}

func main() {
	inputFile := flag.String("i", "", "Input file path")
	outputFile := flag.String("o", "", "Output file path")
	column := flag.Int("k", 0, "Column number for sorting (default: entire line)")
	numeric := flag.Bool("n", false, "Sort by numeric value")
	reverse := flag.Bool("r", false, "Sort in reverse order")
	unique := flag.Bool("u", false, "Remove duplicate lines")
	month := flag.Bool("M", false, "Sort by month name")
	ignoreSpace := flag.Bool("b", false, "Ignore trailing spaces")
	checkSorted := flag.Bool("c", false, "Check if lines are sorted")
	humanSort := flag.Bool("h", false, "Sort by human-readable numbers")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Input file is required")
		flag.Usage()
		os.Exit(1)
	}

	lines, err := readFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	opts := sortOptions{
		column:      *column,
		numeric:     *numeric,
		reverse:     *reverse,
		unique:      *unique,
		month:       *month,
		ignoreSpace: *ignoreSpace,
		checkSorted: *checkSorted,
		humanSort:   *humanSort,
	}

	sortedLines, err := sortLines(lines, opts)
	if err != nil {
		fmt.Printf("Error sorting lines: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := writeFile(*outputFile, sortedLines); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}
	} else {
		for _, line := range sortedLines {
			fmt.Println(line)
		}
	}
}
