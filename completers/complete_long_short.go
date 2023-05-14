package completers

import (
	"github.com/johnrichardrinehart/readline"
	"github.com/johnrichardrinehart/readline/runes"
)

// // Caller type for dynamic completion
// type DynamicCompleteFunc func(string) []string

type LongShortCompleterInterface interface {
	// Print(prefix string, level int, buf *bytes.Buffer)
	Do(line []rune, pos int, long bool) (newLine [][]rune, length int)
	GetName() []rune
	GetChildren() []LongShortCompleterInterface
	SetChildren(children []LongShortCompleterInterface)
}

type DynamicLongShortCompleterInterface interface {
	LongShortCompleterInterface
	IsDynamic() bool
	GetDynamicNames(line []rune) [][]rune
}

type LongShortCompleter struct {
	Name     []rune
	Text     []rune
	Dynamic  bool
	Callback DynamicCompleteFunc
	Children []LongShortCompleterInterface
}

// func (p *LongShortCompleter) Tree(prefix string) string {
// 	buf := bytes.NewBuffer(nil)
// 	p.Print(prefix, 0, buf)
// 	return buf.String()
// }

// func Print(p LongShortCompleterInterface, prefix string, level int, buf *bytes.Buffer) {
// 	if strings.TrimSpace(string(p.GetName())) != "" {
// 		buf.WriteString(prefix)
// 		if level > 0 {
// 			buf.WriteString("├")
// 			buf.WriteString(strings.Repeat("─", (level*4)-2))
// 			buf.WriteString(" ")
// 		}
// 		buf.WriteString(string(p.GetName()) + "\n")
// 		level++
// 	}
// 	for _, ch := range p.GetChildren() {
// 		ch.Print(prefix, level, buf)
// 	}
// }

// func (p *LongShortCompleter) Print(prefix string, level int, buf *bytes.Buffer) {
// 	Print(p, prefix, level, buf)
// }

func (p *LongShortCompleter) IsDynamic() bool {
	return p.Dynamic
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

func (p *LongShortCompleter) GetChildren() []LongShortCompleterInterface {
	return p.Children
}

func (p *LongShortCompleter) SetChildren(children []LongShortCompleterInterface) {
	p.Children = children
}

func NewLongShortCompleter(pc ...LongShortCompleterInterface) *LongShortCompleter {
	return LongShortCompleterItem("", "", pc...)
}

func LongShortCompleterItem(name, text string, children ...LongShortCompleterInterface) *LongShortCompleter {
	name += ""
	return &LongShortCompleter{
		Name:     []rune(name),
		Text:     []rune(text),
		Dynamic:  false,
		Children: children,
	}
}

func LongShortCompleterItemDynamic(callback DynamicCompleteFunc, children ...LongShortCompleterInterface) *LongShortCompleter {
	return &LongShortCompleter{
		Callback: callback,
		Dynamic:  true,
		Children: children,
	}
}

func (p *LongShortCompleter) Do(line []rune, pos int, long bool) (newLine [][]rune, offset int) {
	return doLongShortInternal(p, line, pos, line, long)
}

// func Do(p LongShortCompleterInterface, line []rune, pos int) (newLine [][]rune, offset int) {
// 	return doInternal(p, line, pos, line)
// }

func doLongShortInternal(p LongShortCompleterInterface, line []rune, pos int, origLine []rune, long bool) ([][]rune, int) {
	// return values
	var newLine [][]rune
	var offset int

	// clean input
	line = readline.TrimSpaceLeft(line[:pos])

	// defaults
	var goNext bool
	var lineCompleter LongShortCompleterInterface
	for _, child := range p.GetChildren() {
		childNames := make([][]rune, 1)

		childDynamic, ok := child.(DynamicLongShortCompleterInterface)
		if ok && childDynamic.IsDynamic() {
			childNames = childDynamic.GetDynamicNames(origLine)
		} else {
			childNames[0] = child.GetName()
		}

		for _, childName := range childNames {
			// try to match the line prefix with the first matching child (lengths and characters match)
			if len(line) >= len(childName) {
				if runes.HasPrefix(line, childName) {
					if len(line) == len(childName) {
						if len(child.GetChildren()) != 0 {
							goNext = true
							newLine = append(newLine, []rune{' '})
						}
					} else {
						newLine = append(newLine, childName)
					}
					offset = len(childName)
					lineCompleter = child
					// goNext = true
				}
			} else {
				// check if whole line matches a child
				if runes.HasPrefix(childName, line) {
					// the entire line already matches a child
					newLine = append(newLine, childName[len(line):])
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
