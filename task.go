package main

import (
	"bufio"
	"encoding/json"
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

var tasks Tasks
var logFile *os.File

type Task struct {
	Name       string    `json:"name"`
	Deadline   time.Time `json:"deadline"`
	Todo       string    `json:"todo"`
	Importance string    `json:"importance"`
}
type Tasks struct {
	Title string
	Tasks []Task
}

const layout = "20060102"

func init() {
	err := initTaskList()
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
			Usage:       "show your tasks",
			Description: "`show` command shows all the tasks that have been recorded",
			Action:      showTask,
		},
		{
			Name:   "add",
			Usage:  "add task",
			Action: addTask,
		},
	}
	app.ExitErrHandler = func(c *cli.Context, err error) {
		logFile.Close()
	}
	app.Run(os.Args)
}

func initTaskList() error {
	tasks.Tasks = make([]Task, 0)
	tasks.Title = "Tasks"
	var err error
	logFile, err = os.OpenFile("F\\GoWorkspace\\src\\task\\log\\log.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetPrefix("[task]")
	log.SetFlags(log.Ldate | log.Ltime)
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
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func showTask(c *cli.Context) error {
	if len(tasks.Tasks) == 0 {
		fmt.Println("nothing to do right now")
		return nil
	}
	color.Cyan("%-30s%-30s%-30s%-30s", "Name", "ToDo", "Deadline", "Importance")
	for _, task := range tasks.Tasks {
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
	tasks.Tasks = append(tasks.Tasks, task)
	data, err := json.Marshal(tasks)
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
