package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	var prog *cmdFlags
	var f *os.File
	var g *godo
	var err error
	var newFile bool

	prog = parseFlags(os.Args)

	f, err = os.OpenFile("godo.json", os.O_RDWR, 0)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("godo.json does not exist, creating new file")
			newFile = true
			f, err = os.Create("godo.json")
			if err != nil {
				fmt.Println("Error creating godo.json file:", err)
				return
			}
		} else {
			fmt.Println("Error opening godo.json file:", err)
		}
	}

	defer func() {
		// encode to file
		err := f.Truncate(0)
		if err != nil {
			fmt.Println("Unable to truncate godo.json file size:", err)
			return
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			fmt.Println("Unable to seek to beginning of godo.json file:", err)
			return
		}

		encoder := json.NewEncoder(f)
		err = encoder.Encode(g)
		if err != nil {
			fmt.Println("Unable to encode godo struct into godo.json:", err)
		}
		err = f.Close()
		if err != nil {
			fmt.Println("Unable to close godo.json file:", err)
		}
	}()

	if !newFile {
		g, err = loadGodoList(f)
		if err != nil {
			fmt.Println("Error parsing godo.json:", err)
		}
	} else {
		g = &godo{
			Todo:  []todoItem{},
			Doing: []todoItem{},
			Done:  []todoItem{},
		}
	}

	switch prog.cmd {
	case "ls":
		fmt.Println(g.String())
		return

	case "add":
		if len(prog.extra) == 0 {
			fmt.Println("Please add todo text at the end of command")
			return
		}

		item := todoItem{
			Id:     g.getNextID(),
			Text:   prog.extra,
			Status: todo,
		}

		if prog.hasFlag("d") {
			g.Doing = append(g.Doing, item)
		} else {
			g.Todo = append(g.Todo, item)
		}

	case "mv":
		split := strings.Split(prog.extra, " ")
		if len(split) != 2 {
			fmt.Println("Invalid arguments of mv:")
			fmt.Println("\tmv [ID] [LIST_NAME]")
			return
		}
		id, err := strconv.Atoi(split[0])
		if err != nil {
			fmt.Println("Error parsing id:", err)
			return
		}
		toListName := strings.ToLower(split[1])

		var item *todoItem
		if toListName == "todo" || toListName == "doing" || toListName == "done" {
			g.Todo, item, err = removeTodoFromSlice(id, g.Todo)
			if err != nil {
				g.Doing, item, err = removeTodoFromSlice(id, g.Doing)
				if err != nil {
					g.Done, item, _ = removeTodoFromSlice(id, g.Done)
				}
			}
		} else {
			fmt.Printf("%s not a list\n", split[1])
			return
		}
		switch strings.ToLower(split[1]) {
		case "todo":
			g.Todo = append(g.Todo, *item)

		case "doing":
			g.Doing = append(g.Doing, *item)

		case "done":
			g.Done = append(g.Done, *item)
		}

	case "rm":
		id, err := strconv.Atoi(prog.extra)
		if err != nil {
			fmt.Println("Error parsing id:", err)
			return
		}

		err = g.Remove(id)
		if err != nil {
			fmt.Printf("Error removing id %d:%v\n", id, err)
			return
		}

	default:
		fmt.Println("Usage: godo [COMMAND] [OPTION]... [TEXT|ID]]")
		fmt.Println("Adds simple todo tracking to the command line.")
		fmt.Println("Creates a godo.json file in current directory")
		fmt.Println("\nCOMMAND can be any of the following:")
		fmt.Println("  ls,\n  add [TEXT],\n  rm [ID],\n  mv [ID] [LIST_NAME]")
		fmt.Println("\nadd can have the following OPTIONs:")
		fmt.Println("  -d adds new task to the doing column")
		return
	}

	fmt.Println(g.String())
}
