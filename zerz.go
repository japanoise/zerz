package main

import (
	termutil "github.com/japanoise/termbox-util"
	termbox "github.com/nsf/termbox-go"
)

const (
	ZBgColor       termbox.Attribute = termbox.ColorDefault
	ZFgColor                         = termbox.ColorDefault
	ZBgUPColor                       = ZBgColor
	ZFgUPColor                       = termbox.ColorRed
	ZBoxColor                        = ZFgColor
	ZCursorChar                      = termbox.ColorGreen
	ZCursorPattern                   = termbox.ColorBlue
	ZCursorInt                       = termbox.ColorMagenta
	ZCursorUInt                      = termbox.ColorYellow
	ZCursorFg                        = termbox.ColorBlack
	ZStatBg                          = ZBgColor
	ZStatFg                          = termbox.AttrReverse
	ZFlagColorL                      = termbox.ColorGreen
	ZFlagColorR                      = termbox.ColorBlack
	ZHelpBg                          = termbox.ColorWhite
	ZHelpFg                          = termbox.ColorBlack
)

const (
	ZBoxHor   rune = '═'
	ZBoxVer        = '║'
	ZBoxTopL       = '╔'
	ZBoxTopR       = '╗'
	ZBoxBotL       = '╚'
	ZBoxBotR       = '╝'
	ZLineVert      = '│'
	ZLineTUH       = '┴'
	ZLineHor       = '─'
	ZFlagTri       = '◢'
)

func initTerm() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)
}

func box(x1, y1, x2, y2 int) {
	termbox.SetCell(x1, y1, ZBoxTopL, ZBoxColor, ZBgColor)
	termbox.SetCell(x1, y2, ZBoxBotL, ZBoxColor, ZBgColor)
	termbox.SetCell(x2, y1, ZBoxTopR, ZBoxColor, ZBgColor)
	termbox.SetCell(x2, y2, ZBoxBotR, ZBoxColor, ZBgColor)
	for x := x1 + 1; x < x2; x++ {
		termbox.SetCell(x, y1, ZBoxHor, ZBoxColor, ZBgColor)
		termbox.SetCell(x, y2, ZBoxHor, ZBoxColor, ZBgColor)
	}
	for y := y1 + 1; y < y2; y++ {
		termbox.SetCell(x1, y, ZBoxVer, ZBoxColor, ZBgColor)
		termbox.SetCell(x2, y, ZBoxVer, ZBoxColor, ZBgColor)
	}
}

func infoBox(title, prompt string, messages []string) {
	done := false
	sx, sy := termbox.Size()
	lens := make([]int, 0, len(messages))
	for _, message := range messages {
		lens = append(lens, len(message))
	}

	for !done {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		box(1, 1, sx-2, sy-2)
		termutil.PrintStringFgBg(2, 1, title, ZFgColor, ZBgColor)
		termutil.PrintStringFgBg(2, 2, prompt, ZFgColor, ZBgColor)
		for i, message := range messages {
			if lens[i] > sx-4 {
				termutil.PrintStringFgBg(2, 4+i, message[:sx-4], ZFgColor, ZBgColor)
			} else {
				termutil.PrintStringFgBg(2, 4+i, message, ZFgColor, ZBgColor)
			}
		}
		termbox.Flush()

		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventResize:
			termbox.Sync()
			sx, sy = termbox.Size()
		case termbox.EventKey:
			if ev.Key == termbox.KeyEnter || ev.Key ==
				termbox.KeyCtrlC || ev.Key == termbox.KeyCtrlG ||
				ev.Ch|0x20 == 'q' {
				done = true
			}
		}
	}
}

func showErrorList(title, prompt string, errors []error) {
	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Error())
	}
	infoBox(title, prompt, messages)
}

func zerz(global *ZerzEditor, startupErrors []error) {
	initTerm()
	defer termbox.Close()
	if len(startupErrors) > 0 {
		showErrorList("Zerz", "Some files had errors:", startupErrors)
	}
	done := false
	sx, sy := termbox.Size()
	tabbarscroll := 0
	for !done {
		bufareatop, bufareabot := 2, sy-2

		global.Draw(tabbarscroll, sx, sy, 0, bufareatop, sx, bufareabot)
		termbox.Flush()

		event := termbox.PollEvent()
		if event.Type == termbox.EventResize {
			termbox.Sync()
			sx, sy = termbox.Size()
		} else if event.Type == termbox.EventKey {
			if event.Ch == 0 {
				switch event.Key {
				case termbox.KeyCtrlC:
					done = true
				case termbox.KeyCtrlF, termbox.KeyArrowRight:
					if event.Mod == termbox.ModAlt {
						global.FocusBuf().ForwardDWord()
					} else {
						global.FocusBuf().ForwardByte()
					}
				case termbox.KeyCtrlN, termbox.KeyArrowDown:
					global.FocusBuf().ForwardParagraph()
				case termbox.KeyCtrlB, termbox.KeyArrowLeft:
					if event.Mod == termbox.ModAlt {
						global.FocusBuf().BackwardDWord()
					} else {
						global.FocusBuf().BackwardByte()
					}
				case termbox.KeyCtrlP, termbox.KeyArrowUp:
					global.FocusBuf().BackwardParagraph()
				case termbox.KeyPgup:
					global.PageUp(0, bufareatop, sx, bufareabot)
				case termbox.KeyCtrlV, termbox.KeyPgdn:
					global.PageDown(0, bufareatop, sx, bufareabot)
				case termbox.KeyCtrlA, termbox.KeyHome:
					global.FocusBuf().StartOfLine()
				case termbox.KeyCtrlE, termbox.KeyEnd:
					global.FocusBuf().EndOfLine()
				case termbox.KeyF1:
					helpscreen()
					termbox.Sync()
					sx, sy = termbox.Size()
				case termbox.KeyEnter:
					global.FocusBuf().Edit(global, tabbarscroll)
					termbox.Sync()
					sx, sy = termbox.Size()
				}
			} else if event.Mod == termbox.ModAlt {
				if '1' <= event.Ch && event.Ch <= '9' && int(event.Ch-'1') < len(global.Buffers) {
					global.SwitchBuf(int(event.Ch - '1'))
				}
				switch event.Ch {
				case '<':
					global.FocusBuf().StartOfFile()
				case '>':
					global.FocusBuf().EndOfFile()
				case 'v':
					global.PageUp(0, bufareatop, sx, bufareabot)
				case 'f':
					global.FocusBuf().ForwardWord()
				case 'b':
					global.FocusBuf().BackwardWord()
				case 'g':
					global.FocusBuf().GoTo(global, tabbarscroll)
					termbox.Sync()
					sx, sy = termbox.Size()
				case '-', '_':
					global.VSplit()
				case '|':
					global.HSplit()
				case 'u', 'U':
					global.SplitUp()
				case 'd', 'D':
					global.SplitDown()
				case 'l', 'L':
					global.SplitLeft()
				case 'r', 'R':
					global.SplitRight()
				case '0':
					global.KillSplit()
				}
			} else {
				switch event.Ch {
				case 'H':
					if Int8 < global.FocusBuf().IntWidth {
						global.FocusBuf().IntWidth--
					}
				case 'L':
					if global.FocusBuf().IntWidth < Int64 {
						global.FocusBuf().IntWidth++
					}
				case 'c', 'C':
					global.FocusBuf().Mode = ModeChar
				case 'p', 'P':
					global.FocusBuf().Mode = ModePattern
				case 'i', 'I':
					global.FocusBuf().Mode = ModeInt
				case 'u', 'U':
					global.FocusBuf().Mode = ModeUInt
				case 'e', 'E':
					global.FocusBuf().BigEndian = !global.FocusBuf().BigEndian
				case '?':
					helpscreen()
					termbox.Sync()
					sx, sy = termbox.Size()
				case 'q', 'Q':
					done = true
				}
			}
			global.DoScroll(0, bufareatop, sx, bufareabot)
		} else if event.Type == termbox.EventMouse {
			if event.MouseY < 1 {
				if (event.Key == termbox.MouseWheelUp || (event.Key == termbox.MouseLeft &&
					event.MouseX == 0)) && tabbarscroll > 0 {
					tabbarscroll--
				} else if (event.Key == termbox.MouseWheelDown || (event.Key == termbox.MouseLeft &&
					event.MouseX == sx-1)) && tabbarscroll < len(global.Buffers)-1 {
					tabbarscroll++
				} else if event.Key == termbox.MouseLeft {
					otbx := 0
					tbx := 1
					i := tabbarscroll
					for _, buf := range global.Buffers[tabbarscroll:] {
						tbx += buf.File.FilenameWidth + 3
						if otbx < event.MouseX && event.MouseX < tbx {
							global.SwitchBuf(i)
							continue
						}
						i++
						otbx = tbx
					}
				}
			} else if 1 < event.MouseY && event.MouseY <= bufareabot {
				if event.Key == termbox.MouseWheelDown {
					global.FocusBuf().ScrollDown()
				} else if event.Key == termbox.MouseWheelUp {
					global.ScrollUp(0, bufareatop, sx, bufareabot)
				} else if event.Key == termbox.MouseLeft {
					// Currently doesn't take splits into account
					global.FocusBuf().Click(0, bufareatop, event.MouseX, event.MouseY)
				}
			}
		}
	}
}
