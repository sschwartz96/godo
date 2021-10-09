package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type status string

const (
	todo status = "todo"
	done status = "done"
)

type todoItem struct {
	Id     int    `json:"id"`
	Text   string `json:"text"`
	Status status `json:"status"`
}

func genID() string {
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	return strconv.FormatUint(r.Uint64(), 36)[:3]
}

type godo struct {
	Todo         []todoItem `json:"todo"`
	Doing        []todoItem `json:"doing"`
	Done         []todoItem `json:"done"`
	CurrentIndex int        `json:"current_index"`
}

func (g *godo) String() string {
	sb := strings.Builder{}
	sb.WriteString(createDividerLine(32 * 3))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%-32s%-32s%-32s\n", "Todo", "Doing", "Done"))
	sb.WriteString(createDividerLine(32 * 3))
	sb.WriteString("\n\n")

	var iTodo, iDoing, iDone int
	var curTodo, curDoing, curDone todoItem
	var bufTodo, bufDoing, bufDone bufferedString
	var printedTodoID, printedDoingID, printedDoneID bool
	for len(g.Todo) > iTodo || len(g.Doing) > iDoing || len(g.Done) > iDone ||
		bufTodo.hasNext() || bufDoing.hasNext() || bufDone.hasNext() {
		if len(g.Todo) > iTodo && !bufTodo.hasNext() {
			curTodo = g.Todo[iTodo]
			bufTodo = bufferedString{s: curTodo.Text}
			iTodo++
			printedTodoID = false
		}
		if !printedTodoID && bufTodo.hasNext() {
			sb.WriteString(fmt.Sprintf("%d. ", curTodo.Id))
			printedTodoID = true
		} else {
			sb.WriteString("   ")
		}

		sb.WriteString(fmt.Sprintf("%-29s", bufTodo.getNext(27)))

		if len(g.Doing) > iDoing && !bufDoing.hasNext() {
			curDoing = g.Doing[iDoing]
			bufDoing = bufferedString{s: curDoing.Text}
			iDoing++
			printedDoingID = false
		}
		if !printedDoingID && bufDoing.hasNext() {
			sb.WriteString(fmt.Sprintf("%d. ", curDoing.Id))
			printedDoingID = true
		} else {
			sb.WriteString("   ")
		}

		sb.WriteString(fmt.Sprintf("%-29s", bufDoing.getNext(27)))

		if len(g.Done) > iDone && !bufDone.hasNext() {
			curDone = g.Done[iDone]
			bufDone = bufferedString{s: curDone.Text}
			iDone++
			printedDoneID = false
		}
		if !printedDoneID && bufDone.hasNext() {
			sb.WriteString(fmt.Sprintf("%d. ", curDone.Id))
			printedDoneID = true
		} else {
			sb.WriteString("   ")
		}
		sb.WriteString(fmt.Sprintf("%-29s", bufDone.getNext(27)))

		sb.WriteString("\n")
	}
	return sb.String()
}

type bufferedString struct {
	s string
}

func (b *bufferedString) hasNext() bool {
	return len(b.s) > 0
}

func (b *bufferedString) getNext(size int) string {
	var s string
	if len(b.s) < size {
		s = b.s
		b.s = ""
		return strings.TrimSpace(s)
	}

	spaceIndex := strings.LastIndex(b.s[:size], " ")
	if spaceIndex == -1 {
		s = b.s[:size]
		b.s = b.s[size:]
	} else {
		s = b.s[:spaceIndex]
		b.s = b.s[spaceIndex:]
	}
	return strings.TrimSpace(s)
}

func (g *godo) getNextID() int {
	nextID := g.CurrentIndex
	g.CurrentIndex++
	return nextID
}

func (g *godo) Remove(id int) error {
	var err error
	g.Todo, _, err = removeTodoFromSlice(id, g.Todo)
	if err == nil {
		return nil
	}
	g.Doing, _, err = removeTodoFromSlice(id, g.Doing)
	if err == nil {
		return nil
	}
	g.Done, _, err = removeTodoFromSlice(id, g.Done)
	if err == nil {
		return nil
	}
	return fmt.Errorf("Could not find todo with that id")
}

func removeTodoFromSlice(id int, todos []todoItem) ([]todoItem, *todoItem, error) {
	for i, todo := range todos {
		if todo.Id == id {
			newTodos := todos[:i]
			newTodos = append(newTodos, todos[i+1:]...)
			return newTodos, &todo, nil
		}
	}
	return todos, nil, fmt.Errorf("Id not found")
}

func fmtTodoID(index int, list []todoItem) string {
	if len(list) > index {
		return fmt.Sprintf("%d. ", list[index].Id)
	}
	return ""
}

func fmtTodoText(index int, list []todoItem) string {
	if len(list) > index {
		return list[index].Text
	}
	return ""
}

func loadGodoList(reader io.Reader) (*godo, error) {
	decoder := json.NewDecoder(reader)
	var g godo
	err := decoder.Decode(&g)
	return &g, err
}

func transferTodoItem(id int, from []todoItem, to []todoItem) ([]todoItem, []todoItem, error) {
	from, item, err := removeTodoFromSlice(id, from)
	if err != nil {
		return to, from, err
	}
	from = append(from, *item)
	return to, from, err
}

func createDividerLine(length int) string {
	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		sb.WriteString("_")
	}
	return sb.String()
}
