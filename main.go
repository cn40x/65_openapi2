package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// model of each object
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

// assignment data.
var assignments = []Assignment{}



func garbage_collector(file_path string) {

	prg := "rm"
	arg1 := file_path

	cmd := exec.Command(prg, arg1)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print("stdout:" + string(stdout))

}

func run_file(path string, filename string, language string, input string, c *gin.Context) (output string) {
	if language == "java" {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "java", path+filename+".java")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not execute",
			})
			log.Panic(err)

		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		outChan := make(chan []byte)
		errChan := make(chan error)


		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- err
			} else {
				outChan <- out
			}
		}()

		select {

		case err := <-errChan:
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not execute",
			})
			log.Panic(err)
		case <-ctx.Done():
			log.Println("Timeout, killing process...")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Timeout",
			})
			log.Panic()
			err := cmd.Process.Kill()
			if err != nil {
				log.Panic(err)
			}
		case out := <-outChan:
			output = string(out)
		}



	} else if language == "python" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "python3", path+filename+".py")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			fmt.Println("This error1")
			log.Fatal(err)

		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		outChan := make(chan []byte)
		errChan := make(chan error)

		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- err
			} else {
				outChan <- out
			}
		}()

		select {

		case err := <-errChan:
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not execute",
			})
			log.Panic(err)
		case <-ctx.Done():
			log.Println("Timeout, killing process...")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Timeout",
			})
			log.Panic()
			err := cmd.Process.Kill()
			if err != nil {
				log.Panic(err)
			}
		case out := <-outChan:
			output = string(out)
		}

	}

	return
}

func run_file_problem(path string, filename string, language string, input string, c *gin.Context) (output string, num_error int) {

	if language == "java" {
		num := 0
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "java", path+filename+".java")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		outChan := make(chan []byte)
		errChan := make(chan error)

	
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- err
			} else {
				outChan <- out
			}
		}()

		select {

		case err := <-errChan:

			num += 1
			log.Print(err)
		case <-ctx.Done():
			log.Println("Timeout, killing process...")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Timeout",
			})
			log.Panic()
			err := cmd.Process.Kill()
			if err != nil {
				log.Panic(err)
			}
		case out := <-outChan:
			output = string(out)
		}
		num_error = num
		fmt.Print(num)

	} else if language == "python" {
		num := 0
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		cmd := exec.CommandContext(ctx, "python3", path+filename+".py")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			fmt.Println("This error1")
			log.Fatal(err)

		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()

		outChan := make(chan []byte)
		errChan := make(chan error)

		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- err
			} else {
				outChan <- out
			}
		}()

		select {

		case err := <-errChan:

			num += 1
			log.Print(err)
		case <-ctx.Done():
			log.Println("Timeout, killing process...")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Timeout",
			})
			log.Panic()
			err := cmd.Process.Kill()
			if err != nil {
				log.Panic(err)
			}
		case out := <-outChan:
			output = string(out)
		}

		num_error = num
		fmt.Println(num_error)

	}
	fmt.Println("num_error:")
	fmt.Println(num_error)
	return
}

func main() {

	server := gin.Default()
	server.GET("/compile", getAssignments)
	server.POST("/compile", postAssignments)
	server.POST("/problem", postProblems)

	server.Run("0.0.0.0:8081")

}

func getAssignments(c *gin.Context) {
	c.JSON(http.StatusOK, assignments)
}

func postAssignments(c *gin.Context) {

	var assignment Assignment

	if err := c.ShouldBind(&assignment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Please fill sourcefile",
		})
		return
	}

	assignments = append(assignments, assignment)

	file, _ := c.FormFile("source")
	full_filename := strings.Split(file.Filename, ".")
	filename := full_filename[0]
	language := strings.ToLower(c.Request.FormValue("language"))
	input := c.Request.FormValue("input")
	// fmt.Print(full_filename[1])
	if language == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill language",
		})
	} else if full_filename[1] != "py" && full_filename[1] != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill only python or java file",
		})
	} else if full_filename[1] == "py" && language != "python" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill sourcefile and language in same language",
		})
	} else if full_filename[1] != "py" && language == "python" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill sourcefile and language in same language",
		})

	} else if full_filename[1] == "java" && language != "java" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill sourcefile and language in same language",
		})

	} else if full_filename[1] != "java" && language == "java" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill sourcefile and language in same language",
		})

	} else {

		// Upload the file to specific dst.
		storage_path := "C:/Users/MY PC/Desktop/golang/storage/"
		dst := storage_path + file.Filename
		c.SaveUploadedFile(file, dst)

		// mod_input := prepare_input(input)
		// output := run_file(storage_path, filename, language, mod_input)
		mod_input := strings.Replace(input, `\n`, "\n", -1)
		fmt.Println(input)
		fmt.Println(mod_input)
		output := run_file(storage_path, filename, language, mod_input, c)
		fmt.Println("output: " + output)

		garbage_collector(dst)
		c.JSON(http.StatusOK, output)
	}
}


func postProblems(c *gin.Context) {

	var problem Problem

	if err := c.ShouldBind(&problem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Please fill sourcefile",
		})
		return
	}

	problem_file, _ := c.FormFile("problem_file")
	answer_file, _ := c.FormFile("answer_file")
	problem_fullfilename := strings.Split(problem_file.Filename, ".")
	answer_fullfilename := strings.Split(answer_file.Filename, ".")
	language := strings.ToLower(c.Request.FormValue("language"))
	input := c.Request.FormValue("input")

	if language == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill language",
		})

	} else if language != "python" && language != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill only python and java",
		})

	} else if problem_fullfilename[1] != answer_fullfilename[1] {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill file in same language",
		})

	} else if language == "java" && answer_fullfilename[1] != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill file and language in same language",
		})

	} else if language == "java" && problem_fullfilename[1] != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill file and language in same language",
		})

	} else if language == "python" && answer_fullfilename[1] != "py" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill file and language in same language",
		})
	} else if language == "python" && problem_fullfilename[1] != "py" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill file and language in same language",
		})
	} else if problem_fullfilename[1] != "py" && problem_fullfilename[1] != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill only python or java file",
		})

	} else if answer_fullfilename[1] != "py" && answer_fullfilename[1] != "java" {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please fill only python or java file",
		})

	} else {

		// Upload the file to specific dst.
		problem_file_storage := "C:/Users/MY PC/Desktop/golang/problem_storage/"
		answer_file_storage := "C:/Users/MY PC/Desktop/golang/answer_storage/"
		problem_file_path := problem_file_storage + problem_file.Filename
		answer_file_path := answer_file_storage + answer_file.Filename
		c.SaveUploadedFile(problem_file, problem_file_path)
		c.SaveUploadedFile(answer_file, answer_file_path)

		mod_input := strings.Replace(input, `\n`, "\n", -1)

		problem_output, problem_num_error := run_file_problem(problem_file_storage, problem_fullfilename[0], language, mod_input, c)
		answer_output, answer_num_error := run_file_problem(answer_file_storage, answer_fullfilename[0], language, mod_input, c)

		garbage_collector(problem_file_path)
		garbage_collector(answer_file_path)
		fmt.Println("-----------------------------------------------------")
		fmt.Println(problem_output)
		fmt.Println("-----------------------------------------------------")
		fmt.Println(answer_output)
		fmt.Println("-----------------------------------------------------")

		fmt.Println("problem_num_error:")
		fmt.Println(problem_num_error)
		fmt.Println("answer_num_error:")
		fmt.Println(answer_num_error)

		var result bool
		//answer file คือไฟล์คำตอบ
		//problem file คือไฟล์ที่จะนำมาตรวจ

		if answer_num_error == 1 && problem_num_error == 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not execute",
			})

		} else if answer_num_error == 1 && problem_num_error == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "answerfile can not execute",
			})

		} else if answer_num_error == 0 && problem_num_error == 1 {
			result = false
			fmt.Println(result)

			c.JSON(http.StatusOK, result)
		} else {
			result = problem_output == answer_output
			fmt.Println(result)

			c.JSON(http.StatusOK, result)
		}
	}
}
