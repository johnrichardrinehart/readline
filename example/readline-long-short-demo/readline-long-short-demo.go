package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/johnrichardrinehart/readline"
	"github.com/johnrichardrinehart/readline/completers"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	// io.WriteString(w, completer.Tree("    "))
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = completers.NewLongShortCompleter(
	completers.LongShortCompleterItem("mode", "this is a mode",
		completers.LongShortCompleterItem("vi", "this is vi"),
		completers.LongShortCompleterItem("emacs", "this is emacs"),
	),
	completers.LongShortCompleterItem("login", "login"),
	completers.LongShortCompleterItem("say", "say",
		completers.LongShortCompleterItemDynamic(listFiles("./")),
	),
	// completers.LongShortCompleterItem("setprompt"),
	// completers.LongShortCompleterItem("setpassword"),
	completers.LongShortCompleterItem("bye", "exit the application (alias for bye)"),
	completers.LongShortCompleterItem("quit", "exit the application (alias for 'exit')"),
	completers.LongShortCompleterItem("exit", "exit the application "),
	// completers.LongShortCompleterItem("help"),
	// completers.LongShortCompleterItem("go",
	// 	completers.LongShortCompleterItem("build", completers.LongShortCompleterItem("-o"), completers.LongShortCompleterItem("-v")),
	// 	completers.LongShortCompleterItem("install",
	// 		completers.LongShortCompleterItem("-v"),
	// 		completers.LongShortCompleterItem("-vv"),
	// 		completers.LongShortCompleterItem("-vvv"),
	// 	),
	// 	completers.LongShortCompleterItem("test"),
	// ),
	// completers.LongShortCompleterItem("sleep"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	l, err := readline.NewEx(&readline.Config{
		SortFunction: func(in [][]rune) {
			sort.Slice(in, func(l1, l2 int) bool {
				min := func(a, b int) int {
					if a < b {
						return a
					}
					return b
				}

				line1 := in[l1]
				line2 := in[l2]

				shortest := min(len(line1), len(line2))
				for p := 0; p < shortest; p++ {
					if line1[p] < line2[p] {
						return true
					} else if line1[p] > line2[p] {
						return false
					}
				}

				return false
			})
		},
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    &completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		AutoCompleteKey:        readline.CharQuestion,
		IsVerticalAutocomplete: true,

		IsAutoComplete: true,

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "login":
			pswd, err := l.ReadPassword("please enter your password: ")
			if err != nil {
				break
			}
			println("you enter:", strconv.Quote(string(pswd)))
		case line == "help":
			usage(l.Stderr())
		case line == "setpassword":
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				println("you set:", strconv.Quote(string(pswd)))
			}
		case strings.HasPrefix(line, "setprompt"):
			if len(line) <= 10 {
				log.Println("setprompt <prompt>")
				break
			}
			l.SetPrompt(line[10:])
		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				log.Println("say what?")
				break
			}
			go func() {
				for range time.Tick(time.Second) {
					log.Println(line)
				}
			}()
		case line == "bye" || line == "exit" || line == "quit":
			goto exit
		case line == "sleep":
			log.Println("sleep 4 second")
			time.Sleep(4 * time.Second)
		case line == "":
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}
exit:
}
