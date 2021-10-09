package main

import (
	"reflect"
	"testing"
)

func Test_parseFlags(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want *cmdFlags
	}{
		{
			name: "command only",
			args: args{[]string{"program_name", "command"}},
			want: &cmdFlags{
				programName: "program_name",
				args:        []string{"program_name", "command"},
				flags:       map[string]string{},
				extra:       "",
				cmd:         "command",
			},
		},
		{
			name: "argument only",
			args: args{[]string{"program_name", "-p", "--happy", "-test", "value", "--equals=equal"}},
			want: &cmdFlags{
				programName: "program_name",
				args:        []string{"program_name", "-p", "--happy", "-test", "value", "--equals=equal"},
				flags:       map[string]string{"p": "", "happy": "", "test": "value", "equals": "equal"},
				extra:       "",
				cmd:         "",
			},
		},
		{
			name: "command and extra",
			args: args{[]string{"program_name", "command", "this is extra"}},
			want: &cmdFlags{
				programName: "program_name",
				args:        []string{"program_name", "command", "this is extra"},
				flags:       map[string]string{},
				extra:       "this is extra",
				cmd:         "command",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFlags(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cmdFlags_getValue(t *testing.T) {
	type args struct {
		flag string
	}
	tests := []struct {
		name     string
		cmdFlags *cmdFlags
		args     args
		want     string
		wantErr  bool
		hasFlag  bool
	}{
		{
			name:     "flag and value exists",
			cmdFlags: parseFlags([]string{"programName", "command", "-p", "hello"}),
			args:     args{flag: "p"},
			want:     "hello",
			wantErr:  false,
			hasFlag:  true,
		},
		{
			name:     "flag exists",
			cmdFlags: parseFlags([]string{"programName", "command", "-p"}),
			args:     args{flag: "p"},
			want:     "",
			wantErr:  false,
			hasFlag:  true,
		},
		{
			name:     "flag does not exists",
			cmdFlags: parseFlags([]string{"programName", "command"}),
			args:     args{flag: "p"},
			want:     "",
			wantErr:  true,
			hasFlag:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmdFlags.getValue(tt.args.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdFlags.getValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cmdFlags.getValue() = %v, want %v", got, tt.want)
			}

			hasFlagTest := tt.cmdFlags.hasFlag(tt.args.flag)
			if hasFlagTest != tt.hasFlag {
				t.Errorf("cmdFlags.hasFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
