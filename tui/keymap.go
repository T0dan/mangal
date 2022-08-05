package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type statefulKeymap struct {
	state state

	quit, forceQuit,
	selectOne, selectAll,
	confirm,
	openURL,
	read,
	back,
	filter,
	up, down, left, right,
	top, bottom,
	showHelp key.Binding
}

func (k *statefulKeymap) setState(newState state) {
	k.state = newState
}

func newStatefulKeymap() *statefulKeymap {
	k := key.NewBinding
	keys := key.WithKeys
	help := key.WithHelp

	return &statefulKeymap{
		state: idle,

		quit: k(
			keys("q"),
			help("q", "quit"),
		),
		forceQuit: k(
			keys("ctrl+c", "ctrl+d"),
			help("ctrl+c", "force quit"),
		),
		selectOne: k(
			keys(" "),
			help("space", "select one"),
		),
		selectAll: k(
			keys("ctrl+a", "tab", "*"),
			help("tab", "select all"),
		),
		confirm: k(
			keys("enter"),
			help("enter", "confirm"),
		),
		openURL: k(
			keys("o"),
			help("o", "open url"),
		),
		read: k(
			keys("r"),
			help("r", "read"),
		),
		back: k(
			keys("esc"),
			help("esc", "back"),
		),
		filter: k(
			keys("/"),
			help("/", "filter"),
		),
		up: k(
			keys("up", "k"),
			help("↑", "up"),
		),
		down: k(
			keys("down", "j"),
			help("↓", "down"),
		),
		left: k(
			keys("left", "h"),
			help("←", "left"),
		),
		right: k(
			keys("right", "l"),
			help("→", "right"),
		),
		top: k(
			keys("g"),
			help("g", "top"),
		),
		bottom: k(
			keys("G"),
			help("G", "bottom"),
		),
		showHelp: k(
			keys("?", "h"),
			help("?", "help"),
		),
	}
}

// help returns short and full help for the state
// TODO: add more information for full help
func (k *statefulKeymap) help() ([]key.Binding, []key.Binding) {
	h := func(bindings ...key.Binding) []key.Binding {
		return bindings
	}

	to2 := func(a []key.Binding) ([]key.Binding, []key.Binding) {
		return a, a
	}

	switch k.state {
	case idle:
		return to2(h(k.forceQuit))
	case loadingState:
		return to2(h(k.forceQuit, k.back))
	case historyState:
		return to2(h(k.selectOne, k.back, k.openURL, k.filter))
	case sourcesState:
		return to2(h(k.selectOne, k.back, k.filter))
	case searchState:
		return to2(h(k.confirm, k.forceQuit))
	case mangasState:
		return to2(h(k.selectOne, k.back, k.filter))
	case chaptersState:
		return to2(h(k.selectOne, k.selectAll, k.back, k.filter))
	case confirmState:
		return to2(h(k.confirm, k.back, k.forceQuit))
	case readDownloadState:
		return to2(h(k.back, k.forceQuit))
	case readDownloadDoneState:
		return to2(h(k.back, k.forceQuit, k.confirm))
	case downloadState:
		return to2(h(k.back, k.forceQuit))
	case downloadDoneState:
		return to2(h(k.back, k.forceQuit, k.confirm))
	case exitState:
		return to2(h(k.quit))
	default:
		// unreachable
		panic("unknown state")
	}
}

func (k *statefulKeymap) shortHelp() []key.Binding {
	short, _ := k.help()
	return short
}

func (k *statefulKeymap) fullHelp() []key.Binding {
	_, full := k.help()
	return full
}

func (k *statefulKeymap) forList() list.KeyMap {
	return list.KeyMap{
		CursorUp:             k.up,
		CursorDown:           k.down,
		NextPage:             k.right,
		PrevPage:             k.left,
		GoToStart:            k.top,
		GoToEnd:              k.bottom,
		Filter:               k.filter,
		ClearFilter:          key.Binding{},
		CancelWhileFiltering: k.back,
		AcceptWhileFiltering: k.confirm,
		ShowFullHelp:         k.showHelp,
		CloseFullHelp:        k.showHelp,
		Quit:                 k.quit,
		ForceQuit:            k.forceQuit,
	}
}