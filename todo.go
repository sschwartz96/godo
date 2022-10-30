package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
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
	width, _, err := terminal.GetSize(0)
	if err != nil {
		return fmt.Sprintf("Error encounter while trying to get terminal size: %v", err)
	}
	oneThirdWidth := width / 3
	// leftOverSpace := width - (width / 3 * 2) + 4 + 5
	sb := strings.Builder{}
	sb.WriteString(createDividerLine(width))
	sb.WriteString("\n\n")
	sb.WriteString("Todo")
	addSpace(&sb, oneThirdWidth-4)
	sb.WriteString("Doing")
	addSpace(&sb, oneThirdWidth-5)
	sb.WriteString("Done\n")
	sb.WriteString(createDividerLine(width))
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
		prefixLength := 0
		if !printedTodoID && bufTodo.hasNext() {
			idStr := fmt.Sprintf("%d. ", curTodo.Id)
			prefixLength = len(idStr)
			sb.WriteString(idStr)
			printedTodoID = true
		} else {
			sb.WriteString("   ")
			prefixLength = 3
		}

		bufTodoNext := bufTodo.getNext(oneThirdWidth - prefixLength - 2)
		sb.WriteString(bufTodoNext)
		addSpace(&sb, oneThirdWidth-len(bufTodoNext)-prefixLength)

		if len(g.Doing) > iDoing && !bufDoing.hasNext() {
			curDoing = g.Doing[iDoing]
			bufDoing = bufferedString{s: curDoing.Text}
			iDoing++
			printedDoingID = false
		}
		if !printedDoingID && bufDoing.hasNext() {
			idStr := fmt.Sprintf("%d. ", curDoing.Id)
			prefixLength = len(idStr)
			sb.WriteString(idStr)
			printedDoingID = true
		} else {
			sb.WriteString("   ")
			prefixLength = 3
		}

		bufDoingNext := bufDoing.getNext(oneThirdWidth - prefixLength - 2)
		sb.WriteString(bufDoingNext)
		addSpace(&sb, oneThirdWidth-len(bufDoingNext)-prefixLength)

		if len(g.Done) > iDone && !bufDone.hasNext() {
			curDone = g.Done[iDone]
			bufDone = bufferedString{s: curDone.Text}
			iDone++
			printedDoneID = false
		}
		if !printedDoneID && bufDone.hasNext() {
			idStr := fmt.Sprintf("%d. ", curDone.Id)
			prefixLength = len(idStr)
			sb.WriteString(idStr)
			printedDoneID = true
		} else {
			sb.WriteString("   ")
			prefixLength = 3
		}

		sb.WriteString(bufDone.getNext(oneThirdWidth - 2))

		sb.WriteString("\n")
	}
	return sb.String()
}

// addSpace adds space to s strings.Builder for n length
func addSpace(builder *strings.Builder, n int) {
	for i := 0; i < n; i++ {
		builder.WriteString(" ")
	}
	return
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

func (g *godo) matchPrefixTodo(prefix string) (int, error) {
	return matchPrefix(prefix, g.Todo)
}

func (g *godo) matchPrefixDoing(prefix string) (int, error) {
	return matchPrefix(prefix, g.Doing)
}

func (g *godo) matchPrefixDone(prefix string) (int, error) {
	return matchPrefix(prefix, g.Done)
}

func matchPrefix(prefix string, list []todoItem) (int, error) {
	for _, item := range list {
		if strings.HasPrefix(item.Text, prefix) {
			return item.Id, nil
		}
	}
	return 0, errors.New("no prefix match")
}
