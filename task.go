package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var taskList Tasks
var logFile *os.File

type Task struct {
	Name       string    `json:"name"`
	Deadline   time.Time `json:"deadline"`
	Todo       string    `json:"todo"`
	Importance string    `json:"importance"`
}
type Tasks struct {
	Tasks []Task
}

const layout = "20060102"

func init() {
	var err error
	logFile, err = os.OpenFile("F:\\GoWorkspace\\src\\task\\log\\log.log", os.O_APPEND|os.O_RDWR, 0777)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetPrefix("[task]")
	log.SetFlags(log.Ldate | log.Ltime)
	err = initTaskList()
	if err != nil {
		log.Println(err)
	}

}
func main() {
	app := cli.NewApp()
	app.Name = "task"
	app.Usage = "record your task list"

	app.Version = "1.0.0"
	app.Description = "a simple tool that record your todo list"
	app.Authors = append(app.Authors, &cli.Author{
		Name: "tttturtle-russ",
	})
	app.Commands = cli.Commands{
		{
			Name:        "show",
			Usage:       "show your taskList",
			Description: "`show` command shows all the taskList that have been recorded",
			Action:      showTask,
		},
		{
			Name:   "add",
			Usage:  "add task",
			Action: addTask,
		},
		{
			Name:   "remove",
			Usage:  "remove task",
			Action: removeTask,
		},
	}
	app.ExitErrHandler = func(c *cli.Context, err error) {
		logFile.Close()
	}
	app.Run(os.Args)
}

func initTaskList() error {
	taskList.Tasks = make([]Task, 0)
	var err error
	list, err := os.OpenFile("F:\\GoWorkspace\\src\\task\\list.json", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer list.Close()
	var data []byte
	data, err = io.ReadAll(list)
	if err != nil {
		log.Println(err)
		return err
	}
	if len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, &taskList)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func showTask(c *cli.Context) error {
	if len(taskList.Tasks) == 0 {
		fmt.Println("nothing to do right now")
		return nil
	}
	color.Cyan("%-30s%-30s%-30s%-30s", "Name", "ToDo", "Deadline", "Importance")
	for _, task := range taskList.Tasks {
		var importanceString string
		switch task.Importance {
		case "\u001B[31m‼️very important\u001B[0m":
			importanceString = color.RedString(task.Importance)
		case "\u001B[36m❗️just so so\u001B[0m":
			importanceString = color.CyanString(task.Importance)
		case "\u001B[30m❕ doesn't matter\u001B[0m":
			importanceString = color.BlackString(task.Importance)
		}
		color.Yellow("%-30s%-30s%-30s%-30s", task.Name, task.Todo, task.Deadline.Format("2006-01-02 15-04"), importanceString)
	}
	return nil
}

func addTask(c *cli.Context) error {
	// TODO 参数列表实现
	var task Task
	fmt.Println("input task name")
	_, err := fmt.Scanln(&task.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("input what todo")
	reader := bufio.NewReader(os.Stdin)
	task.Todo, err = reader.ReadString('\r')
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("input task deadline format:YYYYMMdd")
	var ddl string
	_, err = fmt.Scanln(&ddl)
	if err != nil {
		log.Println(err)
		return err
	}
	task.Deadline, err = time.ParseInLocation(layout, ddl, time.Local)
	if err != nil {
		log.Println(err)
		return err
	}
	importantStr := color.RedString("‼️very important")
	sosoStr := color.CyanString("❗️just so so")
	dontMatterStr := color.BlackString("❕ doesn't matter")
	p := promptui.Select{
		Label: "how important it is?",
		Items: []string{
			importantStr, sosoStr, dontMatterStr,
		},
	}
	_, result, err := p.Run()
	if err != nil {
		log.Println(err)
		return err
	}
	task.Importance = result
	list, err := os.OpenFile("F:\\GoWorkspace\\src\\task\\list.json", os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer list.Close()
	task.Todo = strings.TrimSuffix(task.Todo, "\r")
	taskList.Tasks = append(taskList.Tasks, task)
	data, err := json.Marshal(taskList)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = list.Write(data)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func removeTask(c *cli.Context) error {
	color.Red("which task to delete?")
	var taskName string
	_, err := fmt.Scanln(&taskName)
	if err != nil {
		log.Println(err)
		return err
	}
	ifExist := containsTask(taskList.convert2Set(), taskName)
	if ifExist {
		confirmation := askForConfirmation()
		if !confirmation {
			return nil
		}
		err = taskList.remove(taskName)
		if err != nil {
			log.Println(err)
			return err
		}
		color.Cyan("task removed successfully!")
	}
	return nil
}

func (t Tasks) convert2Set() map[string]struct{} {
	set := make(map[string]struct{}, len(t.Tasks))
	for _, task := range t.Tasks {
		set[task.Name] = struct{}{}
	}
	return set
}

func (t Tasks) remove(taskName string) error {
	index := -1
	for i, task := range t.Tasks {
		if task.Name == taskName {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("task not found")
	}
	if len(t.Tasks) == 1 {
		t.Tasks = nil
		err := os.Truncate("F:\\GoWorkspace\\src\\task\\list.json", 0)
		if err != nil {
			return err
		}
		return nil
	}
	t.Tasks = append(t.Tasks[:index], t.Tasks[index+1:]...)
	file, err := os.OpenFile("F:\\GoWorkspace\\src\\task\\list.json", os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()
	bytes, err := json.Marshal(taskList)
	if err != nil {
		log.Println(err)
		return err
	}
	file.Write(bytes)
	return nil
}

func containsTask(s map[string]struct{}, taskName string) bool {
	_, ok := s[taskName]
	return ok
}

func askForConfirmation() bool {
	confirmStr := color.RedString("Are you sure to remove this task?(y/n)")
	fmt.Printf("%s", confirmStr)
	var response string
	_, _ = fmt.Scanf("%s", &response)
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES", ""}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func containsString(okayResponses []string, response string) bool {
	for _, ok := range okayResponses {
		if ok == response {
			return true
		}
	}
	return false
}
