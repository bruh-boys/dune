package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/filesystem"
	"github.com/mattn/go-runewidth"

	termbox "github.com/nsf/termbox-go"
)

const PROMPT = "> "

var stdOut *bytes.Buffer

func startREPL() error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	s := NewScreen()

	defer func() {
		termbox.Close()
		s.saveHistory()
	}()

	termbox.SetInputMode(termbox.InputEsc)

	s.Redraw()

	// until stdout is redirected
	dune.AddBuiltinFunc("print")
	dune.AddNativeFunc(dune.NativeFunction{
		Name:      "print",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			for _, v := range args {
				var text string
				switch v.Type {
				case dune.String, dune.Int, dune.Float, dune.Bool:
					text = v.String()
				default:
					b, err := json.MarshalIndent(v.Export(0), "", "    ")
					if err != nil {
						return dune.NullValue, err
					}
					text = string(b)
				}

				s.Print(text)
			}
			return dune.NullValue, nil
		},
	})

	p, err := dune.CompileStr("")
	if err != nil {
		return err
	}

	p.AddPermission("trusted")

	stdOut = &bytes.Buffer{}
	vm := dune.NewVM(p)
	vm.FileSystem = filesystem.OS
	vm.Stdout = stdOut
	vm.Stderr = stdOut

	if _, err = vm.Run(); err != nil {
		return err
	}

	defer func() {
		vm.FinalizeGlobals()
	}()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				break loop

			case termbox.KeyArrowLeft:
				s.MoveCursorOneRuneBackward()

			case termbox.KeyArrowRight:
				s.MoveCursorOneRuneForward()

			case termbox.KeyBackspace, termbox.KeyBackspace2:
				s.DeleteRuneBackward()

			case termbox.KeyDelete:
				s.DeleteRuneForward()

			case termbox.KeyTab:
				s.InsertRune(' ')
				s.InsertRune(' ')
				s.InsertRune(' ')
				s.InsertRune(' ')

			case termbox.KeySpace:
				s.InsertRune(' ')

			case termbox.KeyArrowUp:
				s.HistoryUp()

			case termbox.KeyArrowDown:
				s.HistoryDown()

			case termbox.KeyCtrlK:
				s.DeleteTheRestOfTheLine()

			case termbox.KeyHome, termbox.KeyCtrlA:
				s.MoveCursorToBeginningOfTheLine()

			case termbox.KeyEnd, termbox.KeyCtrlE:
				s.MoveCursorToEndOfTheLine()

			case termbox.KeyCtrlL:
				if err := s.Clear(); err != nil {
					return err
				}

			case termbox.KeyCtrlD:
				s.pasteMode = false
				s.Prompt = PROMPT
				code := strings.Join(s.pasted, "\n")
				s.History = append(s.History, s.pasted...)
				s.historyIndex = len(s.History)
				s.pasted = nil
				s.Print("// Exiting paste mode, now interpreting.")

				if err := dune.Eval(vm, code); err != nil {
					s.PrintError(err)
					continue
				}

				if stdOut.Len() > 0 {
					s.Print(stdOut.String())
					stdOut.Reset()
				}

				s.Source = append(s.Source, code)
				continue

			case termbox.KeyEnter:
				code := string(s.text)

				if s.pasteMode {
					s.Print(code)
					s.pasted = append(s.pasted, code)
					continue
				}

				s.Print(s.Prompt + code)
				if code == "" {
					continue
				}

				s.AddToHistory(code)
				s.text = nil
				s.lastText = nil

				// custom commands from the REPL
				switch code {

				case ":paste":
					s.pasteMode = true
					s.pasted = nil
					s.Prompt = ""
					s.Print("// Entering paste mode (ctrl-D to finish)")
					continue

				case "list":
					s.Print(strings.Join(s.Source, "\n"))
					continue

				case "asm":
					var b bytes.Buffer
					dune.Fprint(&b, vm.Program)
					asm := strings.Trim(b.String(), "\n")
					s.Print(asm)
					continue

				case "quit":
					return nil
				}

				if strings.HasPrefix(code, "help ") {
					s.showHelp(strings.TrimPrefix(code, "help "))
					continue
				}

				if err := dune.Eval(vm, code); err != nil {
					s.PrintError(err)
					continue
				}

				if stdOut.Len() > 0 {
					s.Print(stdOut.String())
					stdOut.Reset()
				}

				s.Source = append(s.Source, code)

			default:
				if ev.Ch != 0 {
					s.InsertRune(ev.Ch)
				}
			}
		case termbox.EventResize:
			s.Redraw()

		case termbox.EventError:
			return ev.Err
		}
	}

	return nil
}

func NewScreen() *screen {
	s := &screen{
		ColorFG: termbox.ColorDefault,
		ColorBG: termbox.ColorDefault,
		Prompt:  PROMPT,
		Lines: []string{
			dune.VERSION,
			"commands: :paste, help, list, asm, quit",
			"",
		},
	}

	s.loadHistory()

	s.Width, s.Height = termbox.Size()
	return s
}

type screen struct {
	Width        int
	Height       int
	Lines        []string
	History      []string
	Source       []string
	cursorX      int
	cursorY      int
	ColorFG      termbox.Attribute
	ColorBG      termbox.Attribute
	Prompt       string
	pasteMode    bool
	pasted       []string
	text         []byte
	lastText     []byte
	historyIndex int
	historyStart int
}

func (s *screen) Clear() error {
	if err := termbox.Clear(s.ColorFG, s.ColorBG); err != nil {
		return err
	}

	s.Lines = nil
	s.cursorX = 0
	s.cursorY = 0

	s.RedrawLine(true)

	s.Flush()
	return nil
}

func (s *screen) Flush() {
	s.DrawCursor()
	termbox.Flush()
}

func (s *screen) PrintError(err error) {
	text := err.Error()

	// remove the stacktrace
	i := strings.IndexRune(text, '\n')
	if i != -1 {
		text = text[:i]
	}

	s.Print(text)
}

func (s *screen) Redraw() error {
	if err := termbox.Clear(s.ColorFG, s.ColorBG); err != nil {
		return err
	}

	s.Width, s.Height = termbox.Size()

	// split lines based on terminal width
	var lines []string
	for _, line := range s.Lines {
		lines = append(lines, splitLineWidth(line, s.Width)...)
	}

	// take the last ones that fit in the screen
	ln := len(lines)
	if ln >= s.Height {
		start := ln - s.Height + 1
		lines = lines[start:]
	}

	for i, line := range lines {
		s.printline(i, "", line)
	}

	// print the promt line
	y := len(lines)
	s.printline(y, s.Prompt, string(s.text))
	s.MoveCursorTo(0, y)

	s.Flush()
	return nil
}

func (s *screen) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)

	i := s.byteIndexAtCell(s.text, s.cursorX)

	s.text = byte_slice_insert(s.text, i, buf[:n])
	s.lastText = s.text
	s.cursorX += runewidth.RuneWidth(r)
	s.RedrawLine(false)
}

func (s *screen) DebugMsg(msg string) {
	y := s.Height - 1
	s.printline(y, "[*] ", msg)
	s.Flush()
}

func (s *screen) DrawCursor() {
	termbox.SetCursor(s.cursorX+len(s.Prompt), s.cursorY)
}

func (s *screen) RedrawLine(moveCursorToEnd bool) {
	s.printline(s.cursorY, s.Prompt, string(s.text))

	if moveCursorToEnd {
		s.cursorX = utf8.RuneCount(s.text)
	}

	s.Flush()
}

func (s *screen) Print(text string) {
	s.Lines = append(s.Lines, strings.Split(text, "\n")...)
	s.ResetText()
	s.Redraw()
}

func splitLineWidth(text string, width int) []string {
	ln := utf8.RuneCountInString(text)
	if ln <= width {
		return []string{text}
	}

	var lines []string
	var line []rune

	i := 0
	for _, r := range text {
		cn := runewidth.RuneWidth(r)
		if i+cn > width {
			lines = append(lines, string(line))
			i = cn
			line = []rune{r}
			continue
		}
		line = append(line, r)
		i += cn
	}

	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	return lines
}

func (s *screen) ResetText() {
	s.text = nil
	s.cursorX = 0
}

func (s *screen) printline(y int, prompt string, text string) {
	text = prompt + text

	x := 0
	for _, r := range text {
		termbox.SetCell(x, y, r, s.ColorFG, s.ColorBG)
		x += runewidth.RuneWidth(r)
	}

	// clear the rest of the line
	for ; x < s.Width; x++ {
		termbox.SetCell(x, y, ' ', s.ColorFG, s.ColorBG)
	}
}

func (s *screen) AddToHistory(line string) {
	// ignore repeated lines
	ln := len(s.History)
	if ln > 0 && s.History[ln-1] == line {
		return
	}

	s.History = append(s.History, line)
	s.historyIndex = len(s.History)
}

func (s *screen) HistoryUp() {
	if s.historyIndex == 0 {
		return
	}
	s.text = []byte(s.History[s.historyIndex-1])
	s.historyIndex--
	s.RedrawLine(true)
}

func (s *screen) HistoryDown() {
	if s.historyIndex >= len(s.History)-1 {
		s.text = s.lastText
		s.historyIndex = len(s.History)
	} else {
		s.historyIndex++
		s.text = []byte(s.History[s.historyIndex])
	}
	s.RedrawLine(true)
}

func (s *screen) MoveCursorTo(x, y int) {
	s.cursorX = x
	s.cursorY = y
}

func (s *screen) RuneUnderCursor() (rune, int) {
	i := s.byteIndexAtCell(s.text, s.cursorX)
	return utf8.DecodeRune(s.text[i:])
}

func (s *screen) RuneBeforeCursor() (rune, int) {
	i := s.byteIndexAtCell(s.text, s.cursorX)
	return utf8.DecodeLastRune(s.text[:i])
}

func (s *screen) MoveCursorOneRuneBackward() {
	if s.cursorX == 0 {
		return
	}

	r, _ := s.RuneBeforeCursor()
	w := runewidth.RuneWidth(r)
	s.MoveCursorTo(s.cursorX-w, s.cursorY)
	s.Flush()
}

func (s *screen) MoveCursorOneRuneForward() {
	if s.cursorX == len(s.text) {
		return
	}

	r, _ := s.RuneBeforeCursor()
	w := runewidth.RuneWidth(r)
	s.MoveCursorTo(s.cursorX+w, s.cursorY)
	s.Flush()
}

func (s *screen) MoveCursorToBeginningOfTheLine() {
	s.MoveCursorTo(0, s.cursorY)
	s.Flush()
}

func (s *screen) MoveCursorToEndOfTheLine() {
	s.MoveCursorTo(len(s.text), s.cursorY)
	s.Flush()
}

func (s *screen) DeleteRuneBackward() {
	if s.cursorX == 0 {
		return
	}
	s.MoveCursorOneRuneBackward()
	_, size := s.RuneUnderCursor()

	i := s.byteIndexAtCell(s.text, s.cursorX)

	s.text = byte_slice_remove(s.text, i, i+size)
	s.RedrawLine(false)
}

func (s *screen) DeleteRuneForward() {
	if s.cursorX == len(s.text) {
		return
	}

	i := s.byteIndexAtCell(s.text, s.cursorX)

	_, size := s.RuneUnderCursor()
	s.text = byte_slice_remove(s.text, i, i+size)
	s.RedrawLine(false)
}

func (s *screen) DeleteTheRestOfTheLine() {
	i := s.byteIndexAtCell(s.text, s.cursorX)
	s.text = s.text[:i]
}

func historyFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".dune_history"), nil
}

func (s *screen) saveHistory() {
	if len(s.History)-s.historyStart == 0 {
		return
	}

	name, err := historyFilePath()
	if err != nil {
		return
	}

	f, err := os.OpenFile(name, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer f.Close()

	lines := s.History[s.historyStart:]

	if len(lines) > 1000 {
		lines = lines[len(lines)-1000:]
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		_, err = f.WriteString(line + "\n")
		if err != nil {
			return
		}
	}
}

func (s *screen) loadHistory() {
	name, err := historyFilePath()
	if err != nil {
		return
	}

	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	lines := strings.Split(string(b), "\n")
	s.History = lines
	s.historyStart = len(lines)
	s.historyIndex = s.historyStart - 1
}

func (s *screen) showHelp(value string) {
	for _, f := range dune.AllNativeFuncs() {
		n := strings.TrimPrefix(f.Name, "->")
		if strings.HasPrefix(n, value) {
			s.Print(n)
		}
	}
}

func (s *screen) byteIndexAtCell(text []byte, cell int) int {
	var i = 0
	for j := 0; j < cell; j++ {
		_, w := utf8.DecodeRune(text[i:])
		i += w
	}
	return i
}

func byte_slice_grow(s []byte, desired_cap int) []byte {
	if cap(s) < desired_cap {
		ns := make([]byte, len(s), desired_cap)
		copy(ns, s)
		return ns
	}
	return s
}

func byte_slice_remove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

func byte_slice_insert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byte_slice_grow(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}
