package completers

import (
	"fmt"

	"github.com/johnrichardrinehart/readline"
	"github.com/johnrichardrinehart/readline/runes"
)

type LongShortCompleter struct {
	Name     []rune
	Text     []rune
	Callback DynamicCompleteFunc
	Children []LongShortCompleter
}

func (p *LongShortCompleter) IsDynamic() bool {
	return p.Callback != nil
}

func (p *LongShortCompleter) GetName() []rune {
	return p.Name
}

func (p *LongShortCompleter) GetDynamicNames(line []rune) [][]rune {
	var names = [][]rune{}
	for _, name := range p.Callback(string(line)) {
		names = append(names, []rune(name+" "))
	}
	return names
}

func NewLongShortCompleter(pc ...LongShortCompleter) LongShortCompleter {
	return LongShortCompleterItem("", "", pc...)
}

func LongShortCompleterItem(name, text string, children ...LongShortCompleter) LongShortCompleter {
	name += ""
	return LongShortCompleter{
		Name:     []rune(name),
		Text:     []rune(text),
		Children: children,
	}
}

func LongShortCompleterItemDynamic(callback DynamicCompleteFunc, children ...LongShortCompleter) LongShortCompleter {
	return LongShortCompleter{
		Callback: callback,
		Children: children,
	}
}

func (p *LongShortCompleter) Do(line []rune, pos int, long bool) (newLine [][]rune, offset int) {
	return doLongShortInternal(*p, line, pos, line, long)
}

func doLongShortInternal(p LongShortCompleter, line []rune, pos int, origLine []rune, long bool) ([][]rune, int) {
	// return values
	var newLine [][]rune
	var offset int

	// clean input
	line = readline.TrimSpaceLeft(line[:pos])

	// defaults
	var goNext bool
	var lineCompleter LongShortCompleter

	for _, child := range p.Children {
		childNames := make([]string, 0)

		if child.Callback != nil {
			childNames = child.Callback(string(origLine))
		} else {
			childNames = append(childNames, string(child.Name))
		}

		for _, childName := range childNames {
			// try to match the line prefix with the first matching child (lengths and characters match)
			if len(line) >= len(childName) {
				if runes.HasPrefix(line, []rune(childName)) {
					if len(line) == len(childName) {
						if child.Callback != nil {
							goNext = true
							newLine = append(newLine, []rune{' '})
						}
					} else {
						if long {
							newLine = append(newLine, []rune(fmt.Sprintf("%s\t\t%s", childName, string(child.Text))))
						} else {
							if newLine == nil {
								newLine = [][]rune{[]rune(childName)}
							} else {
								newLine[0] = []rune(fmt.Sprintf("%s\t%s", string(newLine[0]), childName))
							}
						}
					}
					offset = len(childName)
					lineCompleter = child
					goNext = true
				}
			} else {
				// check if whole line matches a child
				if runes.HasPrefix([]rune(childName), line) {
					// the entire line already matches a child
					if long {
						newLine = append(newLine, []rune(fmt.Sprintf("%s\t\t%s", childName[len(line):], string(child.Text))))
					} else {
						if newLine == nil {
							newLine = [][]rune{[]rune(childName)}
						} else {
							newLine[0] = []rune(fmt.Sprintf("%s\t%s", string(newLine[0]), childName))
						}
					}

					offset = len(line)
					lineCompleter = child
				}
			}
		}
	}
	if len(newLine) != 1 {
		return newLine, offset
	}

	var tmpLine []rune
	for i := offset; i < len(line); i++ {
		if line[i] == ' ' {
			continue
		}

		tmpLine = append(tmpLine, line[i:]...)
		return doLongShortInternal(lineCompleter, tmpLine, len(tmpLine), origLine, long)
	}

	// process next word in line
	if goNext {
		return doLongShortInternal(lineCompleter, nil, 0, origLine, long)
	}

	return newLine, offset
}
