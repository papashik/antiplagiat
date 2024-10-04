package ant

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	//"sync"
)

func CreateTree(log_file_path, source_code, language, file_path, file_name string) error {
	// logging settings
	log_file, err := os.OpenFile(log_file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	defer log_file.Close()
	logger := log.New(log_file, file_name+": ", 7)

	// parsing code
	stdout := new(strings.Builder)
	stderr := new(strings.Builder)

	parser_path, grammar_name, first_rule := getParseInfo(language)
	if grammar_name == "" {
		logger.Println("Wrong parse language: " + language)
		return errors.New("Wrong parse language: " + language)
	}

	cmd := exec.Command("java", "-Xmx500M", "-cp", "/usr/local/lib/antlr-4.13.0-complete.jar:"+parser_path, "org.antlr.v4.gui.TestRig", grammar_name, first_rule, "-tree")
	cmd.Stdin = strings.NewReader(source_code)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		logger.Println(err)
		return err
	}

	if err := cmd.Wait(); err != nil {
		logger.Printf("-----> Error Message: %s", stderr.String())
		if exiterr, ok := err.(*exec.ExitError); ok {
			logger.Printf("Exit Code: %d", exiterr.ExitCode())
			logger.Printf("Exit Message: %s", exiterr.Error())
		} else {
			logger.Printf("Error: %v", err)
		}
		return err
	}

	if stderr.String() != "" {
		logger.Print("-----> ERROR making tree, stderr stream:\n", stderr.String())
		return errors.New(stderr.String())
	}

	logger.Println("Success tree")

	// saving tree
	var tree_file *os.File
	if tree_file, err = os.Create(file_path + "/" + file_name); err != nil {
		logger.Println(err)
		return err
	}
	defer tree_file.Close()

	if _, err := tree_file.WriteString(stdout.String()); err != nil {
		logger.Println(err)
		return err
	}

	return err
}

func CompareTrees(log_file_path, comparing_file_path, file_path_1, file_path_2 string) (compare_result float64, res error) {
	log_file, err := os.OpenFile(log_file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return -1, err
	}
	defer log_file.Close()
	logger := log.New(log_file, "comparing "+file_path_1+" and "+file_path_2+": ", 7)

	stdout := new(strings.Builder)

	cmd := exec.Command("python3", comparing_file_path, strings.Replace(file_path_1, "'", "", -1), strings.Replace(file_path_2, "'", "", -1))
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	if err := cmd.Start(); err != nil {
		logger.Println(err)
		return -1, err
	}

	if err := cmd.Wait(); err != nil {
		logger.Printf("-----> ERROR message: %s", stdout.String())
		if exiterr, ok := err.(*exec.ExitError); ok {
			logger.Printf("Exit code: %d, exit message: %s", exiterr.ExitCode(), exiterr.Error())
		} else {
			logger.Printf("Error: %v", err)
		}
		return -1, err
	}

	compare_result, err = strconv.ParseFloat(stdout.String(), 32)
	if err != nil {
		logger.Println(err)
		return -1, err
	}
	logger.Printf("Result = %f", compare_result)

	return
}

func getParseInfo(language string) (parser_path, grammar_name, first_rule string) {
	parser_path = "/usr/local/lib/antlr/" + language + ".jar"
	switch language {
	case "go":
		return parser_path, "Go", "sourceFile"
	case "java":
		return parser_path, "Java9", "compilationUnit"
	case "cpp":
		return parser_path, "CPP14", "translationUnit"
	default:
		//log.Println("wrong language: " + language)
		return
	}
}
