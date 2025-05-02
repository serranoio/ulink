package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

var replacedFiles map[string]string

const replacedFilesFilePath = ".replaced-files.yaml"

var blue = lipgloss.NewStyle().Foreground(lipgloss.Color("#3498db")).Bold(true)
var green = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ecc71")).Bold(true)
var red = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)

func findStringInFile(filePath string, searchString string, replacePath string) (bool, error) {
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	newString := strings.ReplaceAll(string(fileContents), searchString, replacePath)

	if newString == string(fileContents) {
		return false, nil
	}
	os.WriteFile(filePath, []byte(newString), 0644)

	if replacedFiles != nil {
		replacedFiles[filePath] = replacePath
	}

	return true, nil
}

func createRelativePath(root, path string) string {
	relPath := strings.ReplaceAll(path, root, "")

	exp := regexp.MustCompile(`[^\/]*\/`)

	newString := exp.ReplaceAllString(relPath, "../")

	return newString[0 : strings.LastIndex(newString, "/")+1]
}

// now that I have the ../, I need to save the file where I replaced it with the string

func walkThroughFiles(root string, searchString string, localPath string) error {
	replacedFiles = make(map[string]string)

	matches := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") || strings.Contains(path, "dist") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		found, err := findStringInFile(path, searchString, createRelativePath(root, path)+localPath)
		if err != nil {
			return err
		}

		if found {
			fmt.Println(blue.Render((fmt.Sprintf("file: %s replaced %s with %s", info.Name(), searchString, localPath))))
			matches++
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println(green.Render(fmt.Sprintf("total matches: %d", matches)))
	return nil
}

func savePathsToFile() error {
	b, err := yaml.Marshal(replacedFiles)

	if err != nil {
		return fmt.Errorf("error marshalling replaced files: %v", err)
	}

	err = os.WriteFile(replacedFilesFilePath, b, 0644)
	if err != nil {
		return fmt.Errorf("error writing replaced files to file: %v", err)
	}

	return nil
}

func switchToRelativePath(nodeModules, localPath string) {
	wd, _ := os.Getwd()

	fmt.Println(blue.Render("looking in ", wd))
	if err := walkThroughFiles(wd, nodeModules, localPath); err != nil {
		fmt.Printf("Error walking through files: %v\n", err)
	}

	savePathsToFile()
}

func relativePathToOriginalPath(contents map[string]string, nodeModule string) []error {
	var errs []error
	// file -> searchString
	for file, searchString := range contents {
		// get file
		fileContents, err := os.ReadFile(file)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// replace
		newFile := strings.ReplaceAll(string(fileContents), searchString, nodeModule)

		// write
		err = os.WriteFile(file, []byte(newFile), 0644)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func switchBackToOriginalPath() (bool, error) {

	if file, err := os.OpenFile(replacedFilesFilePath, os.O_RDWR, 0755); file != nil {
		fmt.Println(blue.Render("Detected .yaml file, switching back to the original path..."))

		if err != nil {
			fmt.Print(red.Render(fmt.Sprintf("you f'ed up your file bro, %s", err.Error())))
			return true, err
		}

		defer file.Close()

		contents, err := os.ReadFile(file.Name())

		var obj map[string]string
		yaml.Unmarshal(contents, &obj)

		if err != nil {
			fmt.Print(red.Render(fmt.Sprintf("you f'ed up your file bro, %s", err.Error())))
			return true, err
		}

		if len(os.Args) <= 1 {
			fmt.Println(red.Render("To return your files back to normal: go run <executable> <node_module>"))
			return true, nil
		}

		nodeModule := os.Args[1]

		relativePathToOriginalPath(obj, nodeModule)

		os.Remove(replacedFilesFilePath)

		return true, nil
	}

	return false, nil
}

func main() {
	if action, err := switchBackToOriginalPath(); action {
		if err != nil {
			fmt.Printf("error switching back to original path: %s", err)
		}
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <searchString> <localPath>")
		return
	}
	nodeModules := os.Args[1]
	localPath := os.Args[2]
	fmt.Println(blue.Render(fmt.Sprintf("Switching %s to %s", nodeModules, localPath)))

	switchToRelativePath(nodeModules, localPath)

}
