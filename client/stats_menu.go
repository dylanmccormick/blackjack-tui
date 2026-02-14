package client

import (
	"fmt"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type StatsMenuModel struct {
	Commands map[string]string
	Stats    protocol.StatsDTO
}

func NewStatsMenu() *StatsMenuModel {
	return &StatsMenuModel{
		Stats: protocol.StatsDTO{},

		Commands: map[string]string{
			"j":     "down",
			"k":     "up",
			"r":     "refresh_stats",
			"enter": "select",
			"esc":   "back",
		},
	}
}

func (sm *StatsMenuModel) Init() tea.Cmd {
	return ReloadStatsCmd()
}

func (sm *StatsMenuModel) View() string {
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "Lifetime Bet: %d\n", sm.Stats.LifetimeBet)
	fmt.Fprintf(&sb, "Lifetime Won: %d\n", sm.Stats.LifetimeWon)
	fmt.Fprintf(&sb, "Lifetime Lost: %d\n", sm.Stats.LifetimeLoss)
	fmt.Fprintf(&sb, "Hands Played: %d\n", sm.Stats.HandsPlayed)
	fmt.Fprintf(&sb, "Hands Won: %d\n", sm.Stats.HandsWon)
	fmt.Fprintf(&sb, "Hands Lost: %d\n", sm.Stats.HandsLost)
	fmt.Fprintf(&sb, "Win Percentage: %d%%\n", sm.Stats.WinPercentage)
	fmt.Fprintf(&sb, "Total Blackjacks: %d\n", sm.Stats.Blackjacks)
	return sb.String()
}

func (sm *StatsMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case protocol.StatsDTO:
		slog.Info("Updating stats in stats menu")
		sm.Stats = msg
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
		case tea.KeyEnter:
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "r":
				slog.Info("Got reload commands")
				cmd = ReloadStatsCmd()
				cmds = append(cmds, cmd)
			}
		}
	}
	return sm, tea.Batch(cmds...)
}
