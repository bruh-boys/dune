package lib

import (
	"fmt"

	"github.com/dunelang/dune"
	"github.com/nsf/termbox-go"
)

func init() {
	dune.RegisterLib(Terminal, `

declare namespace terminal {
	export function init(): void
	export function close(): void
	export function sync(): void
	export function setInputMode(v: number): void
	export function setOutputMode(v: number): void
	export function size(): { width: number, height: number }
	export function flush(): void
	export function clear(fg?: number, bg?: number): void
	export function setCursor(x: number, y: number): void
	export function hideCursor(): void
	export function setCell(x: number, y: number, r: number | string, fg: number, bg: number): void
	export function pollEvent(): Event

	interface Event {
		type: number
		mod: number
		key: number
		ch: number
		chStr: string
		width: number
		height: number
		err: string
		mouseX: number
		mouseY: number
	}

	const ColorDefault = 0
	const ColorBlack = 1
	const ColorRed = 2
	const ColorGreen = 3
	const ColorYellow = 4
	const ColorBlue = 5
	const ColorMagenta = 6
	const ColorCyan = 7
	const ColorWhite = 8

	const EventKey = 0
	const EventResize = 1
	const EventMouse = 2
	const EventError = 3
	const EventInterrupt = 4
	const EventRaw = 5
	const EventNone = 6

	const InputCurrent = 0
	const InputEsc = 1
	const InputAlt = 2
	const InputMouse = 3

	const OutputCurrent = 0
	const OutputNormal = 1
	const Output256 = 2
	const Output216 = 3
	const OutputGrayscale = 4

	const AttrBold = 512
	const AttrUnderline = 1024
	const AttrReverse = 2048

	const ModAlt = 1
	const ModMotion = 2

	const KeyCtrlTilde = 0x00
	const KeyCtrl2 = 0x00
	const KeyCtrlSpace = 0x00
	const KeyCtrlA = 0x01
	const KeyCtrlB = 0x02
	const KeyCtrlC = 0x03
	const KeyCtrlD = 0x04
	const KeyCtrlE = 0x05
	const KeyCtrlF = 0x06
	const KeyCtrlG = 0x07
	const KeyBackspace = 0x08
	const KeyCtrlH = 0x08
	const KeyTab = 0x09
	const KeyCtrlI = 0x09
	const KeyCtrlJ = 0x0A
	const KeyCtrlK = 0x0B
	const KeyCtrlL = 0x0C
	const KeyEnter = 0x0D
	const KeyCtrlM = 0x0D
	const KeyCtrlN = 0x0E
	const KeyCtrlO = 0x0F
	const KeyCtrlP = 0x10
	const KeyCtrlQ = 0x11
	const KeyCtrlR = 0x12
	const KeyCtrlS = 0x13
	const KeyCtrlT = 0x14
	const KeyCtrlU = 0x15
	const KeyCtrlV = 0x16
	const KeyCtrlW = 0x17
	const KeyCtrlX = 0x18
	const KeyCtrlY = 0x19
	const KeyCtrlZ = 0x1A
	const KeyEsc = 0x1B
	const KeyCtrlLsqBracket = 0x1B
	const KeyCtrl3 = 0x1B
	const KeyCtrl4 = 0x1C
	const KeyCtrlBackslash = 0x1C
	const KeyCtrl5 = 0x1D
	const KeyCtrlRsqBracket = 0x1D
	const KeyCtrl6 = 0x1E
	const KeyCtrl7 = 0x1F
	const KeyCtrlSlash = 0x1F
	const KeyCtrlUnderscore = 0x1F
	const KeySpace = 0x20
	const KeyBackspace2 = 0x7F
	const KeyCtrl8 = 0x7F


	const KeyF1 = 0
	const KeyF2 = 1
	const KeyF3 = 2
	const KeyF4 = 3
	const KeyF5 = 4
	const KeyF6 = 5
	const KeyF7 = 6
	const KeyF8 = 7
	const KeyF9 = 8
	const KeyF10 = 9
	const KeyF11 = 10
	const KeyF12 = 11
	const KeyInsert = 12
	const KeyDelete = 13
	const KeyHome = 14
	const KeyEnd = 15
	const KeyPgup = 16
	const KeyPgdn = 17
	const KeyArrowUp = 18
	const KeyArrowDown = 19
	const KeyArrowLeft = 20
	const KeyArrowRight = 21
	const MouseLeft = 22
	const MouseMiddle = 23
	const MouseRight = 24
	const MouseRelease = 25
	const MouseWheelUp = 26
	const MouseWheelDown = 27

}
`)
}

var Terminal = []dune.NativeFunction{
	{
		Name:      "terminal.EventType",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := termbox.Init(); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.init",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := termbox.Init(); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.close",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			termbox.Close()
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.sync",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := termbox.Sync(); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.flush",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := termbox.Flush(); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.setInputMode",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			mode := termbox.InputMode(args[0].ToInt())
			termbox.SetInputMode(mode)

			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.setOutputMode",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			mode := termbox.InputMode(args[0].ToInt())
			termbox.SetInputMode(mode)

			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.size",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			w, h := termbox.Size()

			m := dune.NewMap(2)
			mm := m.ToMap().Map
			mm[dune.NewString("width")] = dune.NewInt(w)
			mm[dune.NewString("height")] = dune.NewInt(h)

			return m, nil
		},
	},
	{
		Name:      "terminal.clear",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			var fg, bg termbox.Attribute
			switch len(args) {
			case 0:
				fg = termbox.ColorDefault
				bg = termbox.ColorDefault
			case 1:
				fg = termbox.Attribute(args[0].ToInt())
				bg = termbox.ColorDefault
			case 2:
				fg = termbox.Attribute(args[0].ToInt())
				bg = termbox.Attribute(args[1].ToInt())
			}

			if err := termbox.Clear(fg, bg); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.setCursor",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			x := int(args[0].ToInt())
			y := int(args[1].ToInt())

			termbox.SetCursor(x, y)

			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.hideCursor",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			termbox.HideCursor()
			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.setCell",
		Arguments: 5,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("invalid x type: %s", args[0].Type)
			}
			if args[1].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("invalid x type: %s", args[1].Type)
			}

			x := int(args[0].ToInt())
			y := int(args[1].ToInt())

			var r rune
			switch args[2].Type {
			case dune.Rune, dune.Int:
				r = args[2].ToRune()
			case dune.String:
				s := args[2].ToString()
				if len(s) != 1 {
					return dune.NullValue, fmt.Errorf("invalid rune: %s", args[2].Type)
				}
				r = rune(s[0])
			default:
				return dune.NullValue, fmt.Errorf("invalid rune: %s", args[2].Type)
			}

			if args[3].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("invalid x type: %s", args[3].Type)
			}
			if args[4].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("invalid x type: %s", args[4].Type)
			}
			fg := termbox.Attribute(args[3].ToInt())
			bg := termbox.Attribute(args[4].ToInt())

			termbox.SetCell(x, y, r, fg, bg)

			return dune.NullValue, nil
		},
	},
	{
		Name:      "terminal.pollEvent",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ev := termbox.PollEvent()
			return dune.NewObject(termboxEvent{ev}), nil
		},
	},
	{
		Name: "->terminal.KeyF1",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF1)), nil
		},
	},
	{
		Name: "->terminal.ColorDefault",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorDefault)), nil
		},
	},
	{
		Name: "->terminal.ColorBlack",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorBlack)), nil
		},
	},
	{
		Name: "->terminal.ColorRed",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorRed)), nil
		},
	},
	{
		Name: "->terminal.ColorGreen",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorGreen)), nil
		},
	},
	{
		Name: "->terminal.ColorYellow",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorYellow)), nil
		},
	},
	{
		Name: "->terminal.ColorBlue",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorBlue)), nil
		},
	},
	{
		Name: "->terminal.ColorMagenta",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorMagenta)), nil
		},
	},
	{
		Name: "->terminal.ColorCyan",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorCyan)), nil
		},
	},
	{
		Name: "->terminal.ColorWhite",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ColorWhite)), nil
		},
	},
	{
		Name: "->terminal.EventKey",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventKey)), nil
		},
	},
	{
		Name: "->terminal.EventResize",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventResize)), nil
		},
	},
	{
		Name: "->terminal.EventMouse",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventMouse)), nil
		},
	},
	{
		Name: "->terminal.EventError",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventError)), nil
		},
	},
	{
		Name: "->terminal.EventInterrupt",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventInterrupt)), nil
		},
	},
	{
		Name: "->terminal.EventRaw",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventRaw)), nil
		},
	},
	{
		Name: "->terminal.EventNone",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.EventNone)), nil
		},
	},
	{
		Name: "->terminal.InputCurrent",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.InputCurrent)), nil
		},
	},
	{
		Name: "->terminal.InputEsc",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.InputEsc)), nil
		},
	},
	{
		Name: "->terminal.InputAlt",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.InputAlt)), nil
		},
	},
	{
		Name: "->terminal.InputMouse",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.InputMouse)), nil
		},
	},
	{
		Name: "->terminal.OutputCurrent",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.OutputCurrent)), nil
		},
	},
	{
		Name: "->terminal.OutputNormal",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.OutputNormal)), nil
		},
	},
	{
		Name: "->terminal.Output256",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.Output256)), nil
		},
	},
	{
		Name: "->terminal.Output216",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.Output216)), nil
		},
	},
	{
		Name: "->terminal.OutputGrayscale",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.OutputGrayscale)), nil
		},
	},
	{
		Name: "->terminal.AttrBold",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.AttrBold)), nil
		},
	},
	{
		Name: "->terminal.AttrUnderline",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.AttrUnderline)), nil
		},
	},
	{
		Name: "->terminal.AttrReverse",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.AttrReverse)), nil
		},
	},
	{
		Name: "->terminal.ModAlt",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ModAlt)), nil
		},
	},
	{
		Name: "->terminal.ModMotion",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.ModMotion)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlTilde",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlTilde)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl2",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl2)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlSpace",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlSpace)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlA",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlA)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlB",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlB)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlC",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlC)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlD",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlD)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlE",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlE)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlF",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlF)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlG",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlG)), nil
		},
	},
	{
		Name: "->terminal.KeyBackspace",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyBackspace)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlH",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlH)), nil
		},
	},
	{
		Name: "->terminal.KeyTab",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyTab)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlI",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlI)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlJ",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlJ)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlK",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlK)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlL",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlL)), nil
		},
	},
	{
		Name: "->terminal.KeyEnter",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyEnter)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlM",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlM)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlN",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlN)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlO",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlO)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlP",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlP)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlQ",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlQ)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlR",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlR)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlS",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlS)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlT",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlT)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlU",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlU)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlV",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlV)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlW",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlW)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlX",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlX)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlY",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlY)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlZ",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlZ)), nil
		},
	},
	{
		Name: "->terminal.KeyEsc",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyEsc)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlLsqBracket",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlLsqBracket)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl3",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl3)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl4",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl4)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlBackslash",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlBackslash)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl5",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl5)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlRsqBracket",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlRsqBracket)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl6",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl6)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl7",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl7)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlSlash",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlSlash)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrlUnderscore",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrlUnderscore)), nil
		},
	},
	{
		Name: "->terminal.KeySpace",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeySpace)), nil
		},
	},
	{
		Name: "->terminal.KeyBackspace2",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyBackspace2)), nil
		},
	},
	{
		Name: "->terminal.KeyCtrl8",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyCtrl8)), nil
		},
	},
	{
		Name: "->terminal.KeyF1",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF1)), nil
		},
	},
	{
		Name: "->terminal.KeyF2",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF2)), nil
		},
	},
	{
		Name: "->terminal.KeyF3",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF3)), nil
		},
	},
	{
		Name: "->terminal.KeyF4",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF4)), nil
		},
	},
	{
		Name: "->terminal.KeyF5",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF5)), nil
		},
	},
	{
		Name: "->terminal.KeyF6",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF6)), nil
		},
	},
	{
		Name: "->terminal.KeyF7",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF7)), nil
		},
	},
	{
		Name: "->terminal.KeyF8",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF8)), nil
		},
	},
	{
		Name: "->terminal.KeyF9",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF9)), nil
		},
	},
	{
		Name: "->terminal.KeyF10",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF10)), nil
		},
	},
	{
		Name: "->terminal.KeyF11",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF11)), nil
		},
	},
	{
		Name: "->terminal.KeyF12",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyF12)), nil
		},
	},
	{
		Name: "->terminal.KeyInsert",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyInsert)), nil
		},
	},
	{
		Name: "->terminal.KeyDelete",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyDelete)), nil
		},
	},
	{
		Name: "->terminal.KeyHome",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyHome)), nil
		},
	},
	{
		Name: "->terminal.KeyEnd",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyEnd)), nil
		},
	},
	{
		Name: "->terminal.KeyPgup",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyPgup)), nil
		},
	},
	{
		Name: "->terminal.KeyPgdn",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyPgdn)), nil
		},
	},
	{
		Name: "->terminal.KeyArrowUp",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyArrowUp)), nil
		},
	},
	{
		Name: "->terminal.KeyArrowDown",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyArrowDown)), nil
		},
	},
	{
		Name: "->terminal.KeyArrowLeft",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyArrowLeft)), nil
		},
	},
	{
		Name: "->terminal.KeyArrowRight",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.KeyArrowRight)), nil
		},
	},
	{
		Name: "->terminal.MouseLeft",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseLeft)), nil
		},
	},
	{
		Name: "->terminal.MouseMiddle",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseMiddle)), nil
		},
	},
	{
		Name: "->terminal.MouseRight",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseRight)), nil
		},
	},
	{
		Name: "->terminal.MouseRelease",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseRelease)), nil
		},
	},
	{
		Name: "->terminal.MouseWheelUp",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseWheelUp)), nil
		},
	},
	{
		Name: "->terminal.MouseWheelDown",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(termbox.MouseWheelDown)), nil
		},
	},
}

type termboxEvent struct {
	event termbox.Event
}

func (termboxEvent) Type() string {
	return "terminal.Event"
}

func (e termboxEvent) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "type":
		return dune.NewInt(int(e.event.Type)), nil
	case "mod":
		return dune.NewInt(int(e.event.Mod)), nil
	case "key":
		return dune.NewInt(int(e.event.Key)), nil
	case "ch":
		return dune.NewRune(e.event.Ch), nil
	case "chStr":
		return dune.NewString(string(e.event.Ch)), nil
	case "width":
		return dune.NewInt(e.event.Width), nil
	case "height":
		return dune.NewInt(e.event.Height), nil
	case "err":
		err := e.event.Err
		if err == nil {
			return dune.NullValue, nil
		}
		return dune.NewString(err.Error()), nil
	case "mouseX":
		return dune.NewInt(e.event.MouseX), nil
	case "mouseY":
		return dune.NewInt(e.event.MouseY), nil
	}
	return dune.UndefinedValue, nil
}
