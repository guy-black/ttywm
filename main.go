package main

import (
	"fmt"
	"os"
	// "time"
	"strings"
	// "slices"

	// "github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/bubbles/textinput"
)

type model struct {
	windows []window
	active  int
	visWS   byte
	width   int
	height  int
	bg      string
	ready   bool
}

type window struct {
	id    int
	onWS  byte
	top   int
	bot   int
	left  int
	right int
}

func initialModel() model {
	return model {
		windows: make([]window, 0),
		active : 0,
		visWS  : 0,
		bg     : "/|/ \\|\\ ",
	}
}

/*
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
*/

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		tea.EnterAltScreen,
		tea.SetWindowTitle("ttywm"),
		// doTick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "still loading"
	}
	finStrs := []string{}
	for len(finStrs) < m.height {
		str := strings.Repeat(m.bg, m.width/len(m.bg)) + m.bg[0:m.width%len(m.bg)]
		finStrs = append(finStrs, str)
	}
	wd := fmt.Sprint (m.width, " x ", m.height)
	finStrs[0] = wd + finStrs[0][len(wd):]
	return strings.Join(finStrs, "\n")
}



func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}
