// Copyright © 2017 David King <davidgwking@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sqsjanitor

import (
	"errors"
	"fmt"
	"log"
	"math"
	"unicode/utf8"

	"github.com/gosuri/uitable"
	"github.com/jroimartin/gocui"
)

type QueueListModel []*QueueDetails

type QueueListView struct {
	cursorPosition int
	rows           int
	columns        int
	contents       string
}

func NewQueueListView() *QueueListView {
	return &QueueListView{1, 0, 0, ""}
}

func (v *QueueListView) MoveCursorUp() error {
	if v.cursorPosition-1 <= 0 {
		return errors.New("invalid position")
	}

	v.cursorPosition--
	return nil
}

func (v *QueueListView) MoveCursorDown() error {
	if v.cursorPosition >= v.rows {
		return errors.New("invalid position")
	}

	v.cursorPosition++
	return nil
}

func (v *QueueListView) UpdateQueues(queues QueueListModel) {
	v.rows = len(queues)

	v.columns = 0
	for _, queue := range queues {
		runeCount := utf8.RuneCountInString(queue.QueueURL)
		if runeCount > v.columns {
			v.columns = runeCount
		}
	}

	v.contents = ""
	for _, queue := range queues {
		v.contents += fmt.Sprintf("- url=%s; messageCount=%d\n", queue.QueueURL, queue.MessageCount)
	}
}

type QueueListController struct {
	queues    *QueueListModel
	view      *QueueListView
	queueURLs chan<- string
	active    bool
}

func NewQueueListController(queues *QueueListModel, queueURLs chan<- string) *QueueListController {
	ctrl := &QueueListController{queues, NewQueueListView(), queueURLs, true}

	ctrl.view.UpdateQueues(*queues)

	return ctrl
}

func (ctrl *QueueListController) GetCurrentSelection() *QueueDetails {
	return (*ctrl.queues)[ctrl.view.cursorPosition-1]
}

func (ctrl *QueueListController) ConfigureView(v *gocui.View) {
	v.SelBgColor, v.SelFgColor = gocui.ColorBlack, gocui.ColorWhite
	v.SelBgColor, v.SelFgColor = gocui.ColorWhite, gocui.ColorBlack

	v.Frame = false
	v.Editable = true
	v.Editor = ctrl

	v.Wrap = false
	v.Highlight = true

	v.SetCursor(0, ctrl.view.cursorPosition)
}

func (ctrl *QueueListController) WriteViewBytesTo(v *gocui.View) {
	table := uitable.New()

	table.AddRow("MESSAGE COUNT", "SQS QUEUE URL")
	for _, queue := range *ctrl.queues {
		table.AddRow(queue.MessageCount, queue.QueueURL)
	}
	rendered := table.String()

	v.Write([]byte(rendered))
}

func (ctrl *QueueListController) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyEnter:
		details := ctrl.GetCurrentSelection()
		ctrl.queueURLs <- details.QueueURL
	case gocui.KeyArrowUp:
		if err := ctrl.view.MoveCursorUp(); err == nil {
			v.MoveCursor(0, -1, false)
		}
	case gocui.KeyArrowDown:
		if err := ctrl.view.MoveCursorDown(); err == nil {
			v.MoveCursor(0, 1, false)
		}
	}
}

func InitTerminalInterface(queues *QueueListModel, queueURLs chan<- string, exit chan<- error) {
	ctrl := NewQueueListController(queues, queueURLs)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		exit <- err
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error {
		// this is executed by the main loop
		maxX, maxY := g.Size()
		x, y := int(math.Floor(float64(maxX)*0.85)), int(math.Floor(float64(maxY)*0.9))
		if v, err := g.SetView("queues", 0, 0, x, y); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			ctrl.ConfigureView(v)
			ctrl.WriteViewBytesTo(v)
		}

		if _, err := g.SetCurrentView("queues"); err != nil {
			return err
		}

		return nil
	})

	quit := func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		exit <- err
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil {
		exit <- err
		if err != gocui.ErrQuit {
			log.Panicln(err)
		}
	}
}
