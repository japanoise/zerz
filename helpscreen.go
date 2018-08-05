package main

import (
	termutil "github.com/japanoise/termbox-util"
	termbox "github.com/nsf/termbox-go"
)

func helpscreen() {
	motto := "“For the Wild!”"
	motw := termutil.RunewidthStr(motto)
	arroww := 40
	lownybw := 23
	fh, fw := 11, 40
	th, tw := 5, 27
	for {
		termbox.Sync()
		sx, _ := termbox.Size()
		termbox.Clear(ZHelpFg, ZHelpBg)
		// Title & motto
		tx, ty := (sx/2)-(tw/2), 0
		termutil.PrintStringFgBg(tx, ty, "mmmmmm mmmmmm mmmmm  mmmmmm",
			ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(tx, ty+1, "    #\" #      #   \"#     #\"",
			ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(tx, ty+2, "  m#   #mmmmm #mmmm\"   m#",
			ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(tx, ty+3, " m\"    #      #   \"m  m\"",
			ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(tx, ty+4, "##mmmm #mmmmm #    \" ##mmmm",
			ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg((sx/2)-(motw/2), ty+5, motto, ZHelpFg, ZHelpBg)
		// Green anarchist flag >:)
		fx, fy := (sx/2)-(fw/2), th+2
		i := fw - 1
		for y := 0; y < fh; y++ {
			for x := 0; x < fw; x++ {
				if x == i {
					termbox.SetCell(fx+x, fy+y, ZFlagTri,
						ZFlagColorR|termbox.AttrUnderline, ZFlagColorL)
				} else if x < i {
					termbox.SetCell(fx+x, fy+y, ' ',
						ZFlagColorR, ZFlagColorL)
				} else {
					termbox.SetCell(fx+x, fy+y, ' ',
						ZFlagColorL, ZFlagColorR)
				}
			}
			i -= 4
		}
		// Actual help
		xanc := (sx / 2) - ((arroww + lownybw) / 2)
		termutil.PrintStringFgBg(xanc, fy+fh+1, " BYTE: ←    ^B →    ^F |  Endian:    e | Beg of Line: Home/^A", ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(xanc, fy+fh+2, " WORD: ←   M-b →   M-f | Jump to:  M-g | End of Line:  End/^E", ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(xanc, fy+fh+3, "DWORD: ← C-M-b → C-M-f |  Search:  C-s | Beg of File:     M-<", ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(xanc, fy+fh+4, "PARAG: ↓    ^N ↑    ^B | Size-/+:  H/L | End of File:     M->", ZHelpFg, ZHelpBg)
		termutil.PrintStringFgBg(xanc, fy+fh+5, "MODES:    BITS: p |   INT: i |   UINT: u |   CHAR: c", ZHelpFg, ZHelpBg)
		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			return
		}
	}
}
