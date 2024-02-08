package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"strings"
	// "slices"

	// "github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/bubbles/textinput"
)

// ~~~~~~~~~~
// datatypes
// ~~~~~~~~~~

type model struct {
	windows []window
	winCt   uint
	active  int
	visWS   byte
	width   int
	height  int
	bg      int
	ready   bool
	dt      time.Time
	currX   int
	currY   int
	action  action
}

type window struct {
	id    uint
	name  string
	cont  [] string
	onWS  byte
	top   int // index of first line
	lines int // how many lines to take
	left  int // index of leftmost char
	cols  int // length of lines
}

type action int

const (
	cursor action = iota
	move
	resize
)

func strAct (a action) string {
	switch a {
		case cursor: return "cursor"
		case move:   return "move"
		case resize: return "resize"
	}
	return ""
}

// ~~~~~~~~~~~~~~
// initial setup
// ~~~~~~~~~~~~~~

func initialModel() model {
	return model {
		windows: []window {},
		winCt  : 0,
		active : 0,
		visWS  : 0,
		bg     : 0,
		dt     : time.Now(),
		action : cursor,
	}
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		tea.EnterAltScreen,
		tea.SetWindowTitle("ttywm"),
		doTick(),
	)
}

// ~~~~~~~
// update
// ~~~~~~~

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
		case TickMsg:
			m.dt = time.Now()
			return m, doTick()
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.currX = m.width/2
			m.currY = m.height/2
			if !m.ready {
				m.ready = true
			}
			return m, nil
		case tea.KeyMsg:
			switch msg.String() {
				case "alt+esc":
					return m, tea.Quit
				case "alt+b":
					if m.bg == len(allBGs) - 1 {
						m.bg = 0
					} else {
						m.bg++
					}
					return m, nil
				case "alt+enter":
					newWin :=
						window {
							id    : m.winCt,
							name  : "",
							cont  : make([]string, 0),
							onWS  : 0,
							top   : m.currY,
							lines : 10,
							left  : m.currX,
							cols  : 20,
						}
					m.windows = append(m.windows, newWin)
					return m, nil
				case "alt+z": // lift window to top of stack
					cw := getCurWinInd (m.windows, m.currX, m.currY)
					if cw >= 0 && cw < len(m.windows) -1 {
						// only adjust stack if there is a window under the cursor
						// and it's not already on top of the stack
						new := append (m.windows[:cw], append(m.windows[cw+1:], m.windows[cw])...)
						m.windows = new
					}
					return m, nil
				case "alt+q": // delete window
					cw := getCurWinInd (m.windows, m.currX, m.currY)
					if cw >= 0 && cw < len(m.windows) -1 {
						// only adjust stack if there is a window under the cursor
						// and it's not already on top of the stack
						// closeterm(m.windows[cw])
						// will need to define that when I have actual termnials rendering
						new := append (m.windows[:cw], m.windows[cw+1:]...)
						m.windows = new
					}
					return m, nil

				case "alt+w": // cursor up
					switch m.action {
						case cursor:
							if m.currY > 0 {
								m.currY--
							}
						case move:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currY > 0 && cw >= 0 {
								m.windows[cw].top--
								m.currY--
							}
						case resize:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currY > 0 && cw >= 0 && m.windows[cw].lines>2 {
								m.windows[cw].lines--
								m.currY--
							}
					}
					return m, nil
				case "alt+s": // cursor down
					switch m.action{
						case cursor:
							if m.currY < m.height - 1 {
								m.currY++
							}
						case move:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currY < m.height - 1 && cw >= 0 {
								m.windows[cw].top++
								m.currY++
							}
						case resize:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currY < m.height - 1 && cw >= 0 {
								m.windows[cw].lines++
								m.currY++
							}
					}
					return m, nil
				case "alt+a": // cursor left
					switch m.action{
						case cursor:
							if m.currX > 0 {
								m.currX--
							}
						case move:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currX > 0 && cw >= 0 {
								m.windows[cw].left--
								m.currX--
							}
						case resize:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currX > 0 && cw >= 0 && m.windows[cw].cols > 2 {
								m.windows[cw].cols--
								m.currX--
							}
					}
					return m, nil
				case "alt+d": // cursor right
					switch m.action{
						case cursor:
							if m.currX < m.width - 1 {
								m.currX++
							}
						case move:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currX < m.width - 1 && cw >= 0 {
								m.windows[cw].left++
								m.currX++
							}
						case resize:
							cw := getCurWinInd (m.windows, m.currX, m.currY)
							if m.currX < m.width - 1 && cw >= 0 {
								m.windows[cw].cols++
								m.currX++
							}
					}
					return m, nil
				case "alt+e": // go to move mode
					if m.action == move {
						m.action = cursor
					} else {
						if getCurWinInd (m.windows, m.currX, m.currY) >= 0 {
							m.action = move
						}
					}
					return m, nil
				case "alt+r": // go to resize mode
					if m.action == resize {
						m.action = cursor
					} else {
						if getCurWinInd (m.windows, m.currX, m.currY) >= 0 {
							m.action = resize
						}
					}
					return m, nil
			}
	}
	return m, nil
}

// get index of window the cursor is currently over
// return -1 if cursor is not over any window
func getCurWinInd (ws []window, x, y int) int {
	index := -1
	for i, w := range ws {
		if currWin (x, y, w) {
			index = i
		}
	}
	return index
}

func currWin (x, y int, w window) bool {
	return x >= w.left &&
		x <= w.left - 1 + w.cols &&
		y >= w.top &&
		y <= w.top - 1 + w.lines
}

// ~~~~~
// view
// ~~~~~

func (m model) View() string {
	if !m.ready {
		return "still loading"
	}
	// first fill in bg
	finStrs := fillBG(m)
	// next add the info bars
	for k, v := range barFns {
		if k >= 0 && k < len(finStrs) {
			finStrs[k] = v(m, finStrs[k])
		}
		if k < 0 && k + len(finStrs) >= 0 {
			finStrs[k + len(finStrs)] = v(m, finStrs[k + len(finStrs)])
		}
	}
	// draw windows
	for _, w := range m.windows {
	finStrs = drawWin(finStrs, w)
	}
	// lastly draw the cursor on top
	for k, v := range finStrs {
		if k == m.currY {
			nstr := ""
			for i, r := range v {
				if i == m.currX {
					nstr += "🠭"
				} else {
					nstr += fmt.Sprintf("%c", r)
				}
			}
			finStrs[k] = nstr
		}
	}
	// return the final product
	return strings.Join(finStrs, "\n")
}

func fillBG(m model) []string {
	finStrs := make([] string, 0)
	fulStrs := make([] string, 0)
	// first stretch all lines to the m.width
	for _, str := range allBGs[m.bg] {
		fulStrs = append(fulStrs, strings.Repeat(str, m.width/len(str)) + str[0:m.width%len(str)])
	}// append the lines until less than one more set can fit
	// TODO, figure out why: for i:=2; i*len(allBGs[m.bg])<=m.height; i++ started eating ram like crazy

	for len(append(finStrs, fulStrs...)) <= m.height {
		finStrs = append(finStrs, fulStrs...)
	}
	// append enought of the lines to fill the bottom of the screen
	if len(finStrs) < m.height {
		finStrs = append(finStrs, finStrs[0:m.height%len(allBGs[m.bg])]...)
	}
	return finStrs
}

func drawWin (strs []string, w window) []string {
	for i, v := range strs {
		if i == w.top || i == w.top - 1 + w.lines {
			strs[i] = v[:w.left] + "+" + strings.Repeat("-", w.cols-2) + "+" + v[w.left + w.cols:]
		}
		if i > w.top && i < w.top - 1 + w.lines {
			strs[i] = v[:w.left] + "|" + strings.Repeat(" ", w.cols-2) + "|" + v[w.left + w.cols:]
		}
	}
	return strs
}

// ~~~~~
// main
// ~~~~~

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}

// ~~~~~~~~~~~~
// config vars
//~~~~~~~~~~~~~

var allBGs = [][]string {
	{
		"/|/ \\|\\ ",
	},
	{
		"_|__",
		"___|",
	},
	{
		" / __ \\ \\__/",
		"/ /  \\ \\____",
		"\\ \\__/ / __ ",
		" \\____/ /  \\",
	},
}

var barFns = map[int]func(model, string) string {
	0:
		func (m model, s string) string {
			wd := fmt.Sprint (
				"screen: ", m.width, " x ", m.height,
			)
			wd += s[len(wd):]
			hour, min, sec := m.dt.Clock()
			hms := stringTime (hour, min, sec)
			// making two changes to wd means hms could overwrite wd
			// but also avoids crash if total length is too long
			wd = wd[:len(wd)-len(hms)] + hms
			return wd
		},
	1:
		func (m model, s string) string {
			action := fmt.Sprint("action set to: ", strAct(m.action))
			fin := action + s[len(action):]
			return fin
		},
	-1:
		func (_ model, s string) string {
			fin := ""
			cmd := exec.Command("uptime", "-p")
			var out strings.Builder
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil {
				fin = "ext commnd failed"
			} else {
				fin = strings.TrimSpace(out.String())
			}
			fin += s[len(fin):]
			return fin
		},
}

func stringTime (hour, min, sec int) string {
	var h,m,s string
	if hour<10 {
		h = fmt.Sprintf ("0%d", hour)
	} else {
		h = fmt.Sprint (hour)
	}
	if min<10 {
		m = fmt.Sprintf ("0%d", min)
	} else {
		m = fmt.Sprint (min)
	}
	if sec<10 {
		s = fmt.Sprintf ("0%d", sec)
	} else {
		s = fmt.Sprint (sec)
	}
	return fmt.Sprint(h,":",m,":",s)
}
