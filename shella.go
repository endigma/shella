package shella

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/chzyer/readline"
)

type Context struct {
	Input string
	Args  []string
	Shell *Shell
}

type Shell struct {
	handler     func(*Context)
	interrupt   func()
	Cmds        []*Cmd
	reader      *readline.Instance
	readlinecfg readline.Config
}

type Cmd struct {
	Name    string
	Help    string
	Handler func(*Context)
}

func (s *Shell) AddCmd(cmd *Cmd) {
	s.Cmds = append(s.Cmds, cmd)
}

func New() *Shell {
	return &Shell{
		handler:   defaultHandler,
		interrupt: defaultInterrupt,
		readlinecfg: readline.Config{
			Prompt:            fmt.Sprintf("\033[31m%s\033[0m", "→ "),
			HistoryFile:       "",
			InterruptPrompt:   "^C",
			EOFPrompt:         "exit",
			HistorySearchFold: true,
		},
	}
}

// Defaults

func defaultInterrupt() {
	fmt.Println("\rBye!")
	os.Exit(0)
}

func defaultHandler(c *Context) {
	fmt.Println("Default handler must be replaced!")
}

// Setters

func (s *Shell) SetHandler(f func(*Context)) {
	s.handler = f
}

func (s *Shell) SetInterruptHandler(f func()) {
	s.interrupt = f
}

func (s *Shell) SetHomeHistoryFile(p string) {
	s.readlinecfg.HistoryFile = fmt.Sprintf("%s/%s", userHomeDir(), p)
}

func (s *Shell) SetHistoryFile(p string) {
	s.readlinecfg.HistoryFile = p
}

func (s *Shell) SetPrompt(p string) {
	s.readlinecfg.Prompt = p
}

// Helpers

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}

func checkErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func (s *Shell) readline() *Context {
	inp, err := s.reader.Readline()
	switch err {
	case readline.ErrInterrupt:
		s.interrupt()
	default:
		checkErr(err)
	}

	return &Context{
		Input: strings.Replace(inp, "\n", "", -1),
		Args:  strings.Split(strings.Replace(inp, "\n", "", -1), " "),
		Shell: s,
	}
}

func (s *Shell) handle(ctx *Context) {
	for _, command := range s.Cmds {
		if ctx.Args[0] == command.Name {
			command.Handler(ctx)
			return
		}
	}
	s.handler(ctx)
}

// Main funcs

func (s *Shell) Interrupt() {
	s.interrupt()
}

func (s *Shell) Process(args ...string) {
	s.handler(&Context{
		Input: strings.Join(args, " "),
		Args:  args,
		Shell: s,
	})
}

func (s *Shell) Run() {
	var err error
	s.reader, err = readline.NewEx(&s.readlinecfg)
	checkErr(err)

	for {
		s.handle(s.readline())
	}
}
