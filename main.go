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
	// "bytes"
	// "io/ioutil"
	// "os"
	// "time"
)

type Assignment struct {
	Source   *multipart.FileHeader `form:"source"`
	Language string                `form:"language"`
	Input    string                `form:"input"`
}

type Problem struct {
	Problem_file *multipart.FileHeader `form:"problem_file"`
	Answer_file  *multipart.FileHeader `form:"answer_file"`
	Language     string                `form:"language"`
	Input        string                `form:"input"`
}

var assignments = []Assignment{}
var problems = []Problem{}

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
	server.POST("/problem", postProblems)

	server.Run("localhost:8081")

}

func getAssignments(c *gin.Context) {
	// call("http://localhost:8081/compile", "POST")
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

	mod_input := strings.Replace(input, `\n`, "\n", -1)
	output := run_file(storage_path, filename, language, mod_input)
	garbage_collector(dst)
	fmt.Println(output)
	c.JSON(http.StatusOK, output)
}

// func call(url string, method string, language string, input string) error {
// 	client := &http.Client{
// 		Timeout: time.Second * 10,
// 	}
// 	// New multipart writer.
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	fw, err := writer.CreateFormField("language")
// 	if err != nil {
// 		fmt.Println("language session error")
// 	}
// 	_, err = io.Copy(fw, strings.NewReader("python"))
// 	if err != nil {
// 		return err
// 	}

// 	fw, err = writer.CreateFormField("input")
// 	if err != nil {
// 		fmt.Println("input session error")
// 	}
// 	_, err = io.Copy(fw, strings.NewReader("23"))
// 	if err != nil {
// 		return err
// 	}

// 	fw, err = writer.CreateFormFile("source", "test.py")
// 	if err != nil {
// 		fmt.Println("file session error")
// 	}
// 	file, err := os.Open("test.py")
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = io.Copy(fw, file)
// 	if err != nil {
// 		return err
// 	}

// 	// Close multipart writer.
// 	writer.Close()
// 	req, err := http.NewRequest(method, url, bytes.NewReader(body.Bytes()))
// 	if err != nil {
// 		return err
// 	}
// 	req.Header.Set("Content-Type", writer.FormDataContentType())
// 	rsp, _ := client.Do(req)
// 	if rsp.StatusCode != http.StatusOK {
// 		log.Printf("Request failed with response code: %d", rsp.StatusCode)
// 	}

// 	data, _ := ioutil.ReadAll(rsp.Body)
// 	fmt.Println(string(data))

// 	return nil
// }

func postProblems(c *gin.Context) {

	var problem Problem

	if err := c.ShouldBind(&problem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	problem_file, _ := c.FormFile("problem_file")
	answer_file, _ := c.FormFile("answer_file")
	problem_fullfilename := strings.Split(problem_file.Filename, ".")
	answer_fullfilename := strings.Split(answer_file.Filename, ".")
	language := c.Request.FormValue("language")
	input := c.Request.FormValue("input")

	// Upload the file to specific dst.
	problem_file_storage := "C:/Users/npt/project/problem_storage/"
	answer_file_storage := "C:/Users/npt/project/answer_storage/"
	problem_file_path := problem_file_storage + problem_file.Filename
	answer_file_path := answer_file_storage + answer_file.Filename
	c.SaveUploadedFile(problem_file, problem_file_path)
	c.SaveUploadedFile(answer_file, answer_file_path)

	mod_input := strings.Replace(input, `\n`, "\n", -1)

	problem_output := run_file(problem_file_storage, problem_fullfilename[0], language, mod_input)
	answer_output := run_file(problem_file_storage, answer_fullfilename[0], language, mod_input)

	garbage_collector(problem_file_path)
	garbage_collector(answer_file_path)
	fmt.Println("-----------------------------------------------------")
	fmt.Println(problem_output)
	fmt.Println(answer_output)
	fmt.Println("-----------------------------------------------------")

	result := problem_output == answer_output

	fmt.Println(result)

	c.JSON(http.StatusOK, result)
}
