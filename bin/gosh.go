package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {

	// Check if a filename was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: gosh <filename.gosh>")
		return
	}

	// Read the file
	filename := os.Args[1]
	content, err := os.ReadFile(filename)

	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Convert the content into a Go equivalent
	goshCode := string(content)
	goCode := processGoshCode(goshCode)

	tempFileName := filename + ".go"
	err = os.WriteFile(tempFileName, []byte(goCode), 0644)

	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	defer os.Remove(tempFileName)

	cmd := exec.Command("go", "run", tempFileName)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error getting stdout pipe: %v\n", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error getting stderr pipe: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Command finished with error: %v\n", err)
	}
}

func processGoshCode(goshCode string) string {
	lines := strings.Split(goshCode, "\n")

	var goTopLevel []string
	var goMainFunction []string
	inFunction := false
	inGroupImport := false

	goCode := []string{
		"package main",
		"import \"fmt\"",
		"import \"os\"",
		"import" + " \"os/exec\"",
		"import \"time\"",
	}

	reShell := regexp.MustCompile(`(\w+)?\s*(:=)?\s*\$\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "import") {
			goTopLevel = append(goTopLevel, line)
			if strings.HasSuffix(line, "(") {
				inGroupImport = true
			}
			continue
		}

		if inGroupImport {
			goTopLevel = append(goTopLevel, line)
			if strings.HasPrefix(line, ")") {
				inGroupImport = false
			}
			continue
		}

		if strings.HasPrefix(line, "func ") {
			inFunction = true
			goTopLevel = append(goTopLevel, line)
			continue
		}

		if inFunction {
			goTopLevel = append(goTopLevel, line)

			if strings.HasSuffix(line, "}") {
				inFunction = false
			}

			continue
		}

		if match := reShell.FindStringSubmatch(line); match != nil {
			identifier := match[1]
			assignment := match[2]
			cmd := match[3]

			cmd = strings.TrimSuffix(cmd, ")")
			cmd = strings.TrimPrefix(cmd, "(")

			trimmedCmd := strings.Join(strings.Split(strings.Trim(cmd, "`"), " "), "")

			if assignment == "" {
				line = strings.ReplaceAll(line, "$"+cmd, `func() string {
					cmd`+trimmedCmd+` := exec.Command(`+cmd+`)
					outputShell, errorShell = cmd`+trimmedCmd+`.CombinedOutput()
					if errorShell != nil {
						fmt.Printf("Error executing command: %v\nOutput: %s\n", errorShell, string(outputShell))
						return string(outputShell)
					}
					
					return string(outputShell)
				}() `+func() string {
					if strings.HasSuffix(line, ")") {
						return ","
					}
					return ""
				}()+`
				`)

				goMainFunction = append(goMainFunction, line)
				continue
			}

			if cmd != "" {
				line = strings.ReplaceAll(line, identifier+" := $"+cmd, `
					`+identifier+`cmd := exec.Command(`+cmd+`)
					outputShell, errorShell = `+identifier+`cmd.CombinedOutput()
					if errorShell != nil {
						fmt.Printf("Error executing command: %v\nOutput: %s\n", errorShell, string(outputShell))
						return
					}
					
					`+identifier+` := string(outputShell)
					_ = `+identifier+`
				`)

				goMainFunction = append(goMainFunction, line)
				continue
			}
		}

		goMainFunction = append(goMainFunction, line)
	}

	globarVar := `var (
		pwd 	= os.Getwd
		chdir = os.Chdir
		mkdir = os.Mkdir
		rmdir = os.Remove
		sleep = time.Sleep
		cd = os.Chdir
		touch = os.Create
		outputShell []byte
		errorShell error
		print = fmt.Println
		execCommand = exec.Command
	)`

	goTopLevel = append(goTopLevel, globarVar)

	goCode = append(goCode, goTopLevel...)
	goCode = append(goCode, "func main() {")
	goCode = append(goCode, goMainFunction...)
	goCode = append(goCode, "}")

	return strings.Join(goCode, "\n")
}
