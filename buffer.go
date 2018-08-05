package main

import (
	"fmt"
	"math"
	"strconv"

	termutil "github.com/japanoise/termbox-util"
	termbox "github.com/nsf/termbox-go"
)

type ZerzMode uint8
type ZerzIntWidth uint8

const (
	ModePattern ZerzMode = iota
	ModeInt
	ModeUInt
	ModeChar
)

const (
	Int8 ZerzIntWidth = iota
	Int16
	Int32
	Int64
)

type ZerzBuffer struct {
	File      *ZerzFile
	Offset    int64
	Scroll    int64
	Mode      ZerzMode
	IntWidth  ZerzIntWidth
	BigEndian bool
	Focused   bool
}

func CreateBuffer(filename string) (*ZerzBuffer, error) {
	file, err := OpenFile(filename)
	if err != nil {
		return nil, err
	}
	return &ZerzBuffer{File: file}, nil
}

func (zbuf *ZerzBuffer) DestroyBuffer() {
	zbuf.File.Close()
}

func IsPrintableAscii(c byte) bool {
	return 0x20 <= c && c <= 0x7e
}

func (zbuf *ZerzBuffer) CursColor() termbox.Attribute {
	switch zbuf.Mode {
	case ModeInt:
		return ZCursorInt
	case ModeUInt:
		return ZCursorUInt
	case ModeChar:
		return ZCursorChar
	default:
		return ZCursorPattern
	}

}

func (zbuf *ZerzBuffer) drawByte(offset, j int64, x1, y, k, nearOffset, farOffset int) {
	bg := ZBgColor
	fg := ZFgColor
	if !zbuf.Focused {
		// Blank; just don't draw cursors if we're not focused.
	} else if offset == zbuf.Offset {
		bg = zbuf.CursColor()
		fg = ZCursorFg
	} else if offset > zbuf.Offset && (zbuf.Mode == ModeInt || zbuf.Mode == ModeUInt) {
		if zbuf.IntWidth > Int8 && offset == zbuf.Offset+1 {
			bg = zbuf.CursColor()
			fg = ZCursorFg
		} else if zbuf.IntWidth > Int16 && offset <= zbuf.Offset+3 {
			bg = zbuf.CursColor()
			fg = ZCursorFg
		} else if zbuf.IntWidth > Int32 && offset <= zbuf.Offset+7 {
			bg = zbuf.CursColor()
			fg = ZCursorFg
		}
	}
	termutil.PrintStringFgBg(x1+nearOffset+int(10+j)+(3*k), y,
		fmt.Sprintf("%02x", zbuf.File.Bytes[offset]),
		fg, bg)
	if IsPrintableAscii(zbuf.File.Bytes[offset]) {
		termbox.SetCell(x1+farOffset+(2*k), y, rune(zbuf.File.Bytes[offset]),
			fg, bg)
	} else if zbuf.File.Bytes[offset] < 0x20 {
		termbox.SetCell(x1+farOffset+(2*k), y, rune(zbuf.File.Bytes[offset]|0x40),
			fg|termbox.AttrReverse, bg)
	} else {
		termbox.SetCell(x1+farOffset+(2*k), y, '.',
			ZFgUPColor, bg)
	}
}

func (zbuf *ZerzBuffer) DrawBuffer(x1, y1, x2, y2 int) {
	_, sy := x2-x1, y2-y1
	boty := int64(sy)<<4 + zbuf.Scroll
	if boty > zbuf.File.Size {
		boty = (zbuf.File.Size - 1) & (math.MaxInt64 - 0x0F)
	}
	y := y1
	for i := zbuf.Scroll; i <= boty; i += 0x10 {
		termutil.PrintStringFgBg(x1, y, fmt.Sprintf("%08x:", i),
			ZFgColor, ZBgColor)
		k := 0
		for j := int64(0); j < 0x10; j += 2 {
			if i+j >= zbuf.File.Size {
				break
			}
			zbuf.drawByte(i+j, j, x1, y, k, 0, 51)

			if i+j+1 >= zbuf.File.Size {
				break
			}
			zbuf.drawByte(i+j+1, j, x1, y, k, 2, 52)

			k++
		}
		y++
	}
}

func (zbuf *ZerzBuffer) EndStr() string {
	if zbuf.BigEndian {
		return "BIGEND"
	}
	return "lilend"
}

func (zbuf *ZerzBuffer) ForwardByte() {
	if zbuf.Offset < zbuf.File.Size-1 {
		zbuf.Offset++
	}
}

func (zbuf *ZerzBuffer) BackwardByte() {
	if zbuf.Offset > 0 {
		zbuf.Offset--
	}
}

func (zbuf *ZerzBuffer) ForwardWord() {
	if zbuf.Offset < zbuf.File.Size-2 {
		zbuf.Offset += 2
	} else {
		zbuf.Offset = zbuf.File.Size - 1
	}
}

func (zbuf *ZerzBuffer) BackwardWord() {
	if zbuf.Offset > 1 {
		zbuf.Offset -= 2
	} else {
		zbuf.Offset = 0
	}
}

func (zbuf *ZerzBuffer) ForwardDWord() {
	if zbuf.Offset < zbuf.File.Size-4 {
		zbuf.Offset += 4
	} else {
		zbuf.Offset = zbuf.File.Size - 1
	}
}

func (zbuf *ZerzBuffer) BackwardDWord() {
	if zbuf.Offset > 3 {
		zbuf.Offset -= 4
	} else {
		zbuf.Offset = 0
	}
}

func (zbuf *ZerzBuffer) ForwardParagraph() {
	if zbuf.Offset < zbuf.File.Size-0x10 {
		zbuf.Offset += 0x10
	}
}

func (zbuf *ZerzBuffer) BackwardParagraph() {
	if zbuf.Offset >= 0x10 {
		zbuf.Offset -= 0x10
	}
}

func (zbuf *ZerzBuffer) StartOfFile() {
	zbuf.Offset = 0
}

func (zbuf *ZerzBuffer) EndOfFile() {
	zbuf.Offset = zbuf.File.Size - 1
}

func (zbuf *ZerzBuffer) PageUp(y1, y2 int) {
	sy := int64(y2 - y1)
	zbuf.Scroll -= sy << 4
	if zbuf.Scroll < 0 {
		zbuf.Scroll = 0
	}
	zbuf.Offset = zbuf.Scroll
}

func (zbuf *ZerzBuffer) PageDown(y1, y2 int) {
	sy := int64(y2 - y1)
	zbuf.Scroll += sy << 4
	if zbuf.Scroll >= zbuf.File.Size {
		zbuf.Scroll = (zbuf.File.Size - 1) & (math.MaxInt64 - 0x0F)
		zbuf.Offset = zbuf.File.Size - 1
	} else {
		zbuf.Offset = zbuf.Scroll
	}
}

func (zbuf *ZerzBuffer) StartOfLine() {
	// Drop the final nybble
	zbuf.Offset >>= 4
	zbuf.Offset <<= 4
}

func (zbuf *ZerzBuffer) EndOfLine() {
	// Set final nybble to F
	zbuf.Offset |= 0x0f
	if zbuf.Offset >= zbuf.File.Size {
		zbuf.Offset = zbuf.File.Size - 1
	}
}

func (zbuf *ZerzBuffer) ScrollUp(y1, y2 int) {
	sy := int64(y2 - y1)
	if zbuf.Scroll > 0 {
		zbuf.Scroll -= 0x10
		if zbuf.Offset > zbuf.Scroll+(sy<<4) {
			zbuf.Offset = zbuf.Scroll + (sy << 4)
		}
	}
}

func (zbuf *ZerzBuffer) ScrollDown() {
	if zbuf.Scroll+0x10 < zbuf.File.Size {
		zbuf.Scroll += 0x10
		if zbuf.Offset < zbuf.Scroll {
			zbuf.Offset = zbuf.Scroll
		}
	}
}

func (zbuf *ZerzBuffer) Click(x1, y1, mousex, mousey int) {
	offsetx, offsety := mousex-x1, mousey-y1
	zbuf.Offset = zbuf.Scroll + int64(offsety)*0x10
	if 52 <= offsetx && offsetx <= 67 {
		zbuf.Offset += int64(offsetx - 52)
	} else if offsetx > 66 {
		zbuf.Offset += 0x0f
	} else if 10 <= offsetx && offsetx < 50 {
		// offset x into hex
		oxix := int64(offsetx - 10)
		zbuf.Offset += oxix * 2 / 5
	} else if 50 <= offsetx {
		zbuf.Offset += 0x0f
	}
	if zbuf.Offset >= zbuf.File.Size {
		zbuf.Offset = zbuf.File.Size - 1
	}
}

func (zbuf *ZerzBuffer) DoScroll(y1, y2 int) {
	sy := int64(y2 - y1)
	cy := zbuf.Offset & (math.MaxInt64 - 0x0F)
	if cy < zbuf.Scroll {
		zbuf.Scroll = cy
	} else if cy > zbuf.Scroll+(sy<<4) {
		zbuf.Scroll = cy - ((sy) << 4)
	}
}

func (zbuf *ZerzBuffer) interpretBytesAsInteger(data []byte) uint64 {
	var integer uint64
	if zbuf.BigEndian {
		for i := 0; i < len(data); i++ {
			integer = (integer * 256) + uint64(data[i])
		}
	} else {
		for i := len(data) - 1; i >= 0; i-- {
			integer = (integer * 256) + uint64(data[i])
		}
	}
	return integer
}

func (zbuf *ZerzBuffer) GetCursorData() string {
	switch zbuf.Mode {
	case ModeInt:
		switch zbuf.IntWidth {
		case Int8:
			return fmt.Sprintf("int8: %d", int8(zbuf.File.Bytes[zbuf.Offset]))
		case Int16:
			if zbuf.Offset+1 < zbuf.File.Size {
				return fmt.Sprintf("int16: %d", int16(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+2])))
			}
		case Int32:
			if zbuf.Offset+3 < zbuf.File.Size {
				return fmt.Sprintf("int32: %d", int32(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+4])))
			}
		case Int64:
			if zbuf.Offset+7 < zbuf.File.Size {
				return fmt.Sprintf("int64: %d", int64(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+8])))
			}
		}
	case ModeUInt:
		switch zbuf.IntWidth {
		case Int8:
			return fmt.Sprintf("uint8: %d", uint8(zbuf.File.Bytes[zbuf.Offset]))
		case Int16:
			if zbuf.Offset+1 < zbuf.File.Size {
				return fmt.Sprintf("uint16: %d", uint16(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+2])))
			}
		case Int32:
			if zbuf.Offset+3 < zbuf.File.Size {
				return fmt.Sprintf("uint32: %d", uint32(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+4])))
			}
		case Int64:
			if zbuf.Offset+7 < zbuf.File.Size {
				return fmt.Sprintf("uint64: %d", uint64(zbuf.interpretBytesAsInteger(
					zbuf.File.Bytes[zbuf.Offset:zbuf.Offset+8])))
			}
		}
	case ModePattern:
		return fmt.Sprintf("pattern: %08b", zbuf.File.Bytes[zbuf.Offset])
	case ModeChar:
		return fmt.Sprintf("char: %c", zbuf.File.Bytes[zbuf.Offset])
	}
	return "???"
}

func (zbuf *ZerzBuffer) Edit(zed *ZerzEditor, tabbarscroll int) {
	value := termutil.Prompt("value", func(sx, sy int) {
		zed.Draw(tabbarscroll, sx, sy, 0, 2, sx, sy-2)
	})
	termbox.HideCursor()
	if value == "" {
		return
	}
	switch zbuf.Mode {
	case ModePattern:
		result, err := strconv.ParseUint(value, 16, 8)
		if err == nil {
			zbuf.File.Bytes[zbuf.Offset] = byte(result)
		}
	case ModeChar:
		zbuf.File.Bytes[zbuf.Offset] = value[0]
	case ModeUInt:
		var result uint64
		var err error
		switch zbuf.IntWidth {
		case Int8:
			result, err = strconv.ParseUint(value, 0, 8)
		case Int16:
			result, err = strconv.ParseUint(value, 0, 16)
		case Int32:
			result, err = strconv.ParseUint(value, 0, 32)
		case Int64:
			result, err = strconv.ParseUint(value, 0, 64)
		}
		if err != nil {
			return
		}
		if zbuf.BigEndian {
			for i := int64(1<<uint(zbuf.IntWidth)) - 1; i >= 0; i-- {
				roffset := zbuf.Offset + i
				if roffset >= zbuf.File.Size {
					break
				}
				zbuf.File.Bytes[roffset] = byte(result & 0xFF)
				result >>= 8
			}
		} else {
			for i := int64(0); i < int64(1<<uint(zbuf.IntWidth)); i++ {
				roffset := zbuf.Offset + i
				if roffset >= zbuf.File.Size {
					break
				}
				zbuf.File.Bytes[roffset] = byte(result & 0xFF)
				result >>= 8
			}
		}
	case ModeInt:
		var result int64
		var err error
		switch zbuf.IntWidth {
		case Int8:
			result, err = strconv.ParseInt(value, 0, 8)
		case Int16:
			result, err = strconv.ParseInt(value, 0, 16)
		case Int32:
			result, err = strconv.ParseInt(value, 0, 32)
		case Int64:
			result, err = strconv.ParseInt(value, 0, 64)
		}
		if err != nil {
			return
		}
		if zbuf.BigEndian {
			for i := int64(1<<uint(zbuf.IntWidth)) - 1; i >= 0; i-- {
				roffset := zbuf.Offset + i
				if roffset >= zbuf.File.Size {
					break
				}
				zbuf.File.Bytes[roffset] = byte(result & 0xFF)
				result >>= 8
			}
		} else {
			for i := int64(0); i < int64(1<<uint(zbuf.IntWidth)); i++ {
				roffset := zbuf.Offset + i
				if roffset >= zbuf.File.Size {
					break
				}
				zbuf.File.Bytes[roffset] = byte(result & 0xFF)
				result >>= 8
			}
		}
	}
}

func (zbuf *ZerzBuffer) GoTo(zed *ZerzEditor, tabbarscroll int) {
	value := termutil.Prompt("value", func(sx, sy int) {
		zed.Draw(tabbarscroll, sx, sy, 0, 2, sx, sy-2)
	})
	termbox.HideCursor()
	if value == "" {
		return
	}

	result, err := strconv.ParseInt(value, 0, 64)
	if err != nil || result < 0 {
		return
	}

	zbuf.Offset = result
	if zbuf.Offset >= zbuf.File.Size {
		zbuf.Offset = zbuf.File.Size - 1
	}
	zbuf.Scroll = zbuf.Offset
}
