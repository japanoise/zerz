package main

import (
	"fmt"

	termutil "github.com/japanoise/termbox-util"
	termbox "github.com/nsf/termbox-go"
)

type ZBufTree struct {
	Split   bool
	Hor     bool
	Focused bool
	Buf     int
	ChildLT *ZBufTree
	ChildRB *ZBufTree
}

func (tree *ZBufTree) GetFocus() *ZBufTree {
	if tree.Split {
		ret := tree.ChildLT.GetFocus()
		if ret == nil {
			return tree.ChildRB.GetFocus()
		} else {
			return ret
		}
	} else if tree.Focused {
		return tree
	} else {
		return nil
	}
}

func (tree *ZBufTree) SetFocus(which int) {
	tree.GetFocus().Buf = which
}

func (tree *ZBufTree) Draw(zed *ZerzEditor, x1, y1, x2, y2 int) {
	if tree.Split {
		if tree.Hor {
			chunk := ((x2 - x1) / 2)
			tree.ChildLT.Draw(zed, x1, y1, (x2-1)-chunk, y2)
			tree.ChildRB.Draw(zed, (x2-chunk)+1, y1, x2, y2)
		} else {
			chunk := ((y2 - y1) / 2)
			tree.ChildLT.Draw(zed, x1, y1, x2, (y2-1)-chunk)
			tree.ChildRB.Draw(zed, x1, (y2-chunk)+1, x2, y2)
		}
	} else {
		zed.Buffers[tree.Buf].DrawBuffer(x1, y1, x2, y2)
	}
}

func (tree *ZBufTree) GetFocusBufDimensions(x1, y1, x2, y2 int) (bool, int, int, int, int) {
	if tree.Split {
		if tree.Hor {
			chunk := ((x2 - x1) / 2)
			ret, xx1, yy1, xx2, yy2 := tree.ChildLT.GetFocusBufDimensions(x1, y1, (x2-3)-chunk, y2)
			if ret {
				return true, xx1, yy1, xx2, yy2
			} else {
				return tree.ChildRB.GetFocusBufDimensions((x2-chunk)+1, y1, x2, y2)
			}
		} else {
			chunk := ((y2 - y1) / 2) - 2
			ret, xx1, yy1, xx2, yy2 := tree.ChildLT.GetFocusBufDimensions(x1, y1, x2, (y2-3)-chunk)
			if ret {
				return true, xx1, yy1, xx2, yy2
			} else {
				return tree.ChildRB.GetFocusBufDimensions(x1, (y2-chunk)+1, x2, y2)
			}
		}
	} else {
		return tree.Focused, x1, y1, x2, y2
	}
}

type ZerzEditor struct {
	Buffers []*ZerzBuffer
	CurBuf  int
	Tree    *ZBufTree
}

func InitEditor(filenames []string) (*ZerzEditor, []error) {
	buffers := make([]*ZerzBuffer, 0, len(filenames))
	errors := make([]error, 0, len(filenames))

	for _, filename := range filenames {
		buf, err := CreateBuffer(filename)
		if err == nil {
			buffers = append(buffers, buf)
		} else {
			errors = append(errors, fmt.Errorf("%s: %s", filename, err.Error()))
		}
	}

	if len(buffers) >= 1 {
		buffers[0].Focused = true
	}

	return &ZerzEditor{buffers, 0, &ZBufTree{false, false, true, 0, nil, nil}},
		errors
}

func (zed *ZerzEditor) Close() {
	for _, buffer := range zed.Buffers {
		buffer.DestroyBuffer()
	}
}

func (zed *ZerzEditor) FocusBuf() *ZerzBuffer {
	return zed.Buffers[zed.CurBuf]
}

func (zed *ZerzEditor) SwitchBuf(which int) {
	zed.Buffers[zed.CurBuf].Focused = false
	zed.Buffers[which].Focused = true
	zed.CurBuf = which
	zed.Tree.SetFocus(which)
}

func (zed *ZerzEditor) Draw(tabbarscroll, sx, sy, x1, y1, x2, y2 int) {
	termbox.Clear(ZFgColor, ZBgColor)
	for i := 0; i < sx; i++ {
		termbox.SetCell(i, sy-1, ' ', ZStatFg, ZStatBg)
		termbox.SetCell(i, 1, ZLineHor, ZFgColor, ZBgColor)
	}
	termutil.PrintStringFgBg(0, sy-1, fmt.Sprintf("%s | %s | Offset: %016x | %s",
		zed.FocusBuf().GetCursorData(), zed.FocusBuf().File.Filename,
		zed.FocusBuf().Offset, zed.FocusBuf().EndStr()),
		ZStatFg, ZStatBg)
	i := tabbarscroll
	tbx := 1
	for j, buf := range zed.Buffers[tabbarscroll:] {
		if i+j == zed.CurBuf {
			termutil.PrintStringFgBg(tbx, 0, buf.File.Filename, ZStatFg, ZStatBg)
		} else {
			termutil.PrintStringFgBg(tbx, 0, buf.File.Filename, ZFgColor, ZBgColor)
		}
		tbx += buf.File.FilenameWidth
		termbox.SetCell(tbx+1, 0, ZLineVert, ZFgColor, ZBgColor)
		termbox.SetCell(tbx+1, 1, ZLineTUH, ZFgColor, ZBgColor)
		tbx += 3
		if tbx >= sx-1 {
			termbox.SetCell(sx-1, 0, '→', ZFgColor, ZBgColor)
		}
	}
	if tabbarscroll > 0 {
		termbox.SetCell(0, 0, '←', ZFgColor, ZBgColor)
	}
	zed.Tree.Draw(zed, x1, y1, x2, y2)
}

func (zed *ZerzEditor) PageUp(x1, y1, x2, y2 int) {
	_, _, yy1, _, yy2 := zed.Tree.GetFocusBufDimensions(x1, y1, x2, y2)
	zed.FocusBuf().PageUp(yy1, yy2)
}

func (zed *ZerzEditor) PageDown(x1, y1, x2, y2 int) {
	_, _, yy1, _, yy2 := zed.Tree.GetFocusBufDimensions(x1, y1, x2, y2)
	zed.FocusBuf().PageDown(yy1, yy2)
}

func (zed *ZerzEditor) DoScroll(x1, y1, x2, y2 int) {
	_, _, yy1, _, yy2 := zed.Tree.GetFocusBufDimensions(x1, y1, x2, y2)
	zed.FocusBuf().DoScroll(yy1, yy2)
}

func (zed *ZerzEditor) ScrollUp(x1, y1, x2, y2 int) {
	_, _, yy1, _, yy2 := zed.Tree.GetFocusBufDimensions(x1, y1, x2, y2)
	zed.FocusBuf().ScrollUp(yy1, yy2)
}

func (zed *ZerzEditor) VSplit() {
	ftree := zed.Tree.GetFocus()
	ftree.Split = true
	ftree.Hor = false
	ftree.ChildLT = &ZBufTree{false, false, true, ftree.Buf, nil, nil}
	ftree.ChildRB = &ZBufTree{false, false, false, ftree.Buf, nil, nil}
}

func (zed *ZerzEditor) HSplit() {
	ftree := zed.Tree.GetFocus()
	ftree.Split = true
	ftree.Hor = true
	ftree.ChildLT = &ZBufTree{false, false, true, ftree.Buf, nil, nil}
	ftree.ChildRB = &ZBufTree{false, false, false, ftree.Buf, nil, nil}
}
