package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type Assignment struct {
	Source   *multipart.FileHeader `form:"source"`
	Language string                `form:"language"`
	Input    string                `form:"input"`
}

// albums slice to seed record album data.
var assignments = []Assignment{}

func prepare_input(input string) (mod_input string) {
	inputs := strings.Split(input, ",")

	for i := 0; i < len(inputs); i++ {

		mod_input += inputs[i]
		mod_input += "\n"
	}
	fmt.Println(mod_input)
	return
}

func garbage_collector(file_path string) {

	prg := "rm"
	arg1 := file_path

	cmd := exec.Command(prg, arg1)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print(string(stdout))

}

func run_file(path string, filename string, language string, input string) (output string) {
	if language == "java" {
		cmd := exec.Command("java", path+filename+".java")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		output = string(out)

	} else if language == "python" {
		cmd := exec.Command("python3", path+filename+".py")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		output = string(out)
	}
	return
}
func main() {
	server := gin.Default()
	server.GET("/compile", getAssignments)
	server.POST("/compile", postAssignments)

	server.Run("localhost:8081")
}

func getAssignments(c *gin.Context) {
	c.JSON(http.StatusOK, assignments)
}

func postAssignments(c *gin.Context) {

	var assignment Assignment

	if err := c.ShouldBind(&assignment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	assignments = append(assignments, assignment)

	file, _ := c.FormFile("source")
	full_filename := strings.Split(file.Filename, ".")
	filename := full_filename[0]
	language := c.Request.FormValue("language")
	input := c.Request.FormValue("input")

	// Upload the file to specific dst.
	storage_path := "C:/Users/npt/project/storage/"
	dst := storage_path + file.Filename
	c.SaveUploadedFile(file, dst)

	mod_input := prepare_input(input)
	output := run_file(storage_path, filename, language, mod_input)
	garbage_collector(dst)
	c.JSON(http.StatusOK, output)
}
