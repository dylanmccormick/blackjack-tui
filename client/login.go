package client

import (
	tea "github.com/charmbracelet/bubbletea"
)

type LoginModel struct{}

func (lm *LoginModel) Init() tea.Cmd {
	return nil
}

func (lm *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	// var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// this could be a command?
			cmds = append(cmds, ChangeRootPage(menuPage))
		}
	}
	return lm, tea.Batch(cmds...)
}

func (lm *LoginModel) View() string {
	return blackjackSplash
}

const blackjackSplash = `
╔════════════════════════════════════════════════════════════════════════════╗
║   ┌─────────┐  ┌─────────┐                                                 ║
║   │ A       │  │ K       │                                                 ║
║   │         │  │         │                                                 ║
║   │    ♠    │  │    ♥    │                                                 ║
║   │         │  │         │                                                 ║
║   │       A │  │       K │                                                 ║
║   └─────────┘  └─────────┘                                                 ║
║                                                                            ║
║  ██████╗ ██╗      █████╗  ██████╗██╗  ██╗     ██╗ █████╗  ██████╗██╗  ██╗  ║
║  ██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝     ██║██╔══██╗██╔════╝██║ ██╔╝  ║
║  ██████╔╝██║     ███████║██║     █████╔╝      ██║███████║██║     █████╔╝   ║
║  ██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██   ██║██╔══██║██║     ██╔═██╗   ║
║  ██████╔╝███████╗██║  ██║╚██████╗██║  ██╗╚█████╔╝██║  ██║╚██████╗██║  ██╗  ║
║  ╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝  ║
║                                                                            ║
║                               ♠  ♥  ♣  ♦                                   ║
║                                                                            ║
║                   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━                             ║
║                        GET AS CLOSE TO 21                                  ║
║                       WITHOUT GOING OVER                                   ║
║                   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━                             ║
║                                                                            ║
║                                                                            ║
║                       [H] HIT    [S] STAND                                 ║
║                       [D] DOUBLE [Q] QUIT                                  ║
║                                                                            ║
║                                                                            ║
║                   ╔═══════════════════════════╗                            ║
║                   ║   PRESS ENTER TO START    ║                            ║
║                   ╚═══════════════════════════╝                            ║
║                                                                            ║
║                        v1.0  ·  Terminal Edition                           ║
╚════════════════════════════════════════════════════════════════════════════╝
`
