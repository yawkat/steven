// Copyright 2015 Matthew Collins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/thinkofdeath/steven/chat"
	"github.com/thinkofdeath/steven/render"
)

const (
	chatHistoryLines = 10
	maxLineWidth     = 300
)

type ChatUI struct {
	Elements []*chatUIElement

	dirty    bool
	Lines    [chatHistoryLines]chat.AnyComponent
	lineFade [chatHistoryLines]float64

	lineLength float64
}

type chatUIElement struct {
	text   *render.UIText
	offset int
	line   int
}

func (c *ChatUI) render(delta float64) {
	for i := range c.lineFade {
		if c.lineFade[i] > 0 {
			c.lineFade[i] -= 0.005 * delta
		} else {
			c.lineFade[i] = 0
		}
	}
	if c.dirty {
		c.dirty = false
		for _, e := range c.Elements {
			if e.text != nil {
				e.text.Free()
			}
		}
		c.Elements = nil

		for i, line := range c.Lines {
			c.newLine()
			if line.Value == nil {
				continue
			}
			c.lineLength = 0
			c.renderComponent(i, line.Value, nil)
		}
	}
	for _, e := range c.Elements {
		if e.text != nil {
			e.text.Alpha(c.lineFade[e.line])
		}
	}
}

func (c *ChatUI) renderComponent(line int, co interface{}, color chatGetColorFunc) {
	switch co := co.(type) {
	case *chat.TextComponent:
		getColor := chatGetColor(&co.Component, color)
		width := 0
		runes := []rune(co.Text)
		r, g, b := chatColorRGB(getColor())
		for i := 0; i < len(runes); i++ {
			size := render.SizeOfCharacter(runes[i])
			if width+size > maxLineWidth {
				c.appendText(line, string(runes[:i]), r, g, b)
				c.lineLength = 0
				runes = runes[i:]
				i = 0
				width = 0
				c.newLine()
			}
			width += size
		}
		c.lineLength += c.appendText(line, string(runes), r, g, b)
		for _, e := range co.Extra {
			c.renderComponent(line, e.Value, getColor)
		}
	default:
		fmt.Printf("Can't handle %T\n", co)
	}
}

func (c *ChatUI) appendText(line int, str string, r, g, b int) float64 {
	e := &chatUIElement{
		text:   render.AddUIText(str, 2+c.lineLength, 480-18, r, g, b),
		offset: 0,
		line:   line,
	}
	c.Elements = append(c.Elements, e)
	return e.text.Width
}

type chatGetColorFunc func() chat.Color

func chatGetColor(c *chat.Component, parent chatGetColorFunc) chatGetColorFunc {
	return func() chat.Color {
		if c.Color != "" {
			return c.Color
		}
		if parent != nil {
			return parent()
		}
		return chat.White
	}
}

func chatColorRGB(c chat.Color) (r, g, b int) {
	switch c {
	case chat.Black:
		return 0, 0, 0
	case chat.DarkBlue:
		return 0, 0, 170
	case chat.DarkGreen:
		return 0, 170, 0
	case chat.DarkAqua:
		return 0, 170, 170
	case chat.DarkRed:
		return 170, 0, 0
	case chat.DarkPurple:
		return 170, 0, 170
	case chat.Gold:
		return 255, 170, 0
	case chat.Gray:
		return 170, 170, 170
	case chat.DarkGray:
		return 85, 85, 85
	case chat.Blue:
		return 85, 85, 255
	case chat.Green:
		return 85, 255, 85
	case chat.Aqua:
		return 85, 255, 255
	case chat.Red:
		return 255, 85, 85
	case chat.LightPurple:
		return 255, 85, 255
	case chat.Yellow:
		return 255, 255, 85
	case chat.White:
		return 255, 255, 255

	}
	return 255, 255, 255
}

func (c *ChatUI) newLine() {
	for _, e := range c.Elements {
		if e.text == nil {
			continue
		}
		e.offset++
		if e.offset > 6 {
			e.text.Free()
			e.text = nil
			continue
		}
		e.text.Shift(0, -18)
	}
}

func (c *ChatUI) Add(msg chat.AnyComponent) {
	copy(c.Lines[0:chatHistoryLines-1], c.Lines[1:])
	copy(c.lineFade[0:chatHistoryLines-1], c.lineFade[1:])
	c.Lines[chatHistoryLines-1] = msg
	c.lineFade[chatHistoryLines-1] = 3.0
	c.dirty = true
}
