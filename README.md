# Gosh Interpreter

Gosh is a custom interpreter built on top of Go, allowing you to execute Go-like code with custom features such as shell command execution using a custom syntax. 

Key Features

- Execute Go code: Write Go-like code in Gosh and execute it using the Go toolchain.
- Custom Shell Command Syntax: Use $ to execute shell commands directly in your script.
- Implicit Main Function: No need to define the main function in your script—Gosh will automatically handle that for you.

How to Use

Writing a Gosh Script
Gosh scripts use Go syntax but also introduce custom syntax for executing shell commands. Here's an example:

```
ls := $`ls`
lsSlice := strings.Split(ls, "\n")

for _, file := range lsSlice {
  print(file)
}
```


The $`` syntax executes the shell command within the backticks and stores the output in the variable ls. You can then use this variable like any other Go variable.

Running Gosh Scripts

To run a Gosh script:

1. `npm install gosh-interpreter`
1. Create a .gosh file with your Gosh code.
2. Run the script using the gosh interpreter:

gosh <filename.gosh>

Example:
gosh myscript.gosh

This will read your .gosh file, convert it into valid Go code, and execute it using Go’s go run command.