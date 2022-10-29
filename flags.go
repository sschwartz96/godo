package main

import (
	"fmt"
	"strings"
)

// parseFlags takes os.Args
func parseFlags(args []string) *cmdFlags {
	cmdFlags := &cmdFlags{
		programName: args[0],
		args:        args,
		flags:       map[string]string{},
	}
	extraSlice := []string{}
	wasLastFlag := false
	for i := range args {
		if i == 0 {
			continue
		}

		flag, value, err := getFlag(i, args)
		if err != nil {
			if i == 1 {
				cmdFlags.cmd = args[i]
				wasLastFlag = false
				continue
			}
			// append too the extra if last argument was not flag or is the last argument
			if !wasLastFlag || len(args)-1 == i {
				extraSlice = append(extraSlice, args[i])
			}
			wasLastFlag = false
			continue
		}
		wasLastFlag = true
		cmdFlags.flags[flag] = value
	}
	cmdFlags.extra = strings.Join(extraSlice, " ")
	return cmdFlags
}

// getFlag returns flag name and value, error if not a flag
// must handle the following cases:
// 1. -p		{flag: "p": value: ""}
// 2. -p tw		{flag: "p": value: "tw"}
// 3. -pa		{flag: "p": value: "", flag: "a", value: ""}
// 4. --ppa		{flag: "ppa", value: ""}
// 5. --ppa tw	{flag: "ppa", value: "tw"}
// 6. --ppa=tw	{flag: "ppa", value: "tw"}
// 7. -p=tws	{flag: "p", value: "tws"}
func getFlag(flagLoc int, args []string) (string, string, error) {
	if !strings.HasPrefix(args[flagLoc], "-") {
		return "", "", fmt.Errorf("no '-' prefix, not flag")
	}

	// satifies case 6, 7
	if strings.Contains(args[flagLoc], "=") {
		s := strings.SplitN(args[flagLoc], "=", 2)
		return strings.ReplaceAll(s[0], "-", ""), s[1], nil
	}

	// satifies cases 1, 3, 4
	if len(args) == flagLoc+1 || strings.Contains(args[flagLoc+1], "-") {
		return strings.ReplaceAll(args[flagLoc], "-", ""), "", nil
	}

	// satifies cases 2, 5
	return strings.ReplaceAll(args[flagLoc], "-", ""), args[flagLoc+1], nil
}

type cmdFlags struct {
	programName string
	args        []string
	flags       map[string]string // empty string no value to flag
	extra       string            // extra random strings of text separated by space
	cmd         string            // the command used not a flag
}

// getValue gets the value for given flag name, error if flag does not exist
func (f *cmdFlags) getValue(flag string) (string, error) {
	value, ok := f.flags[flag]
	if !ok {
		return "", fmt.Errorf("flag does not exist")
	}
	return value, nil
}

func (f *cmdFlags) hasFlag(flag string) bool {
	_, ok := f.flags[flag]
	return ok
}

func (f *cmdFlags) toString() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Program Name:\t%s\n", f.programName))
	if len(f.cmd) > 0 {
		sb.WriteString(fmt.Sprintf("Command:\t%s\n", f.cmd))
	}
	sb.WriteString("Flags:\n")
	for flag, value := range f.flags {
		sb.WriteString(fmt.Sprintf("\t%s:\t%s\n", flag, value))
	}
	if len(f.extra) > 0 {
		sb.WriteString(fmt.Sprintf("Extra:\n\t%s", f.extra))
	}
	return sb.String()
}
