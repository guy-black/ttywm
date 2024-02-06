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
	active  int
	visWS   byte
	width   int
	height  int
	bg      int
	ready   bool
	dt      time.Time
}

type window struct {
	id    int
	name  string
	cont  [] string
	onWS  byte
	top   int // index of first line
	lines int // how many lines to take
	left  int // index of leftmost char
	cols  int // length of lines
}

// ~~~~~~~~~~~~~~
// initial setup
// ~~~~~~~~~~~~~~

func initialModel() model {
	w := window {
		id    : 0,
		name  : "test window",
		cont  : make([]string, 0),
		onWS  : 0,
		top   : 5,
		lines : 10,
		left  : 5,
		cols  : 20,
	}
	return model {
		windows: []window {w},
		active : 0,
		visWS  : 0,
		bg     : 0,
		dt     : time.Now(),
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
			}
	}
	return m, nil
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
		if i >= w.top && i <= w.top - 1 + w.lines {
			nstr:=""
			for ii, _ := range v {
				if ii >= w.left && ii <= w.left - 1 + w.cols {
					nstr+=" "
				} else {
					nstr+=fmt.Sprintf("%c", v[ii])
				}
			}
			strs[i]=nstr
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
