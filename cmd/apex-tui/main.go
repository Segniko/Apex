package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- STYLE DEFINITIONS ---
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFB800")).
			Padding(0, 1).
			MarginBottom(1)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#FFB800")).
			Width(35)

	detailStyle = lipgloss.NewStyle().
			Padding(1, 2)

	hazardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB800"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
)

// --- DATA MODELS ---

type ReportItem struct {
	ID        string `json:"error_id"`
	Msg       string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Context   struct {
		OS  string `json:"os"`
		Mem int64  `json:"total_memory"`
	} `json:"context"`
	StackTrace string `json:"stack_trace"`
	AIInsight  string `json:"ai_insight"`
}

func (i ReportItem) Title() string { return i.Msg }
func (i ReportItem) Description() string {
	return fmt.Sprintf("ID: %s... | OS: %s", i.ID[:8], i.Context.OS)
}
func (i ReportItem) FilterValue() string { return i.Msg }

// --- BUBBLE TEA MODEL ---

type model struct {
	list          list.Model
	viewport      viewport.Model
	reports       []ReportItem
	selected      *ReportItem
	loading       bool
	err           error
	chatMode      bool
	chatInput     string
	aiResponse    string
	width, height int
	program       *tea.Program

	configPath string
	apiKey     string
	setupInput string
	showSetup  bool
}

type Config struct {
	APIKey string `json:"api_key"`
}

type startChatMsg struct {
	reportID string
	message  string
}

type msgStartWorker struct {
	msg startChatMsg
}

func initialModel() model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "APEX_FORENSICS_TERMINAL"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".apex_config.json")

	m := model{
		list:       l,
		viewport:   viewport.New(0, 0),
		configPath: configPath,
		loading:    true,
	}

	// Check for local config
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil {
			m.apiKey = cfg.APIKey
		}
	}

	if m.apiKey == "" {
		m.showSetup = true
		m.loading = false
	}

	return m
}

func (m model) Init() tea.Cmd {
	if m.showSetup {
		return nil
	}
	return func() tea.Msg {
		return fetchReports(m.apiKey)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showSetup {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if m.setupInput != "" {
					m.apiKey = m.setupInput
					m.showSetup = false
					m.loading = true
					// Save config
					cfg := Config{APIKey: m.apiKey}
					data, _ := json.Marshal(cfg)
					os.WriteFile(m.configPath, data, 0644)
					return m, func() tea.Msg {
						return fetchReports(m.apiKey)
					}
				}
			case "backspace":
				if len(m.setupInput) > 0 {
					m.setupInput = m.setupInput[:len(m.setupInput)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.setupInput += msg.String()
				}
			}
			return m, nil
		}

		if m.chatMode {
			switch msg.String() {
			case "esc":
				m.chatMode = false
				m.chatInput = ""
				return m, nil
			case "enter":
				if m.chatInput != "" && m.selected != nil {
					m.aiResponse = "ANALYZING_TELEMETRY..."
					// We pass the program pointer to the chat function for async updates
					return m, tea.Batch(func() tea.Msg {
						return startChatMsg{m.selected.ID, m.chatInput}
					})
				}
			case "backspace":
				if len(m.chatInput) > 0 {
					m.chatInput = m.chatInput[:len(m.chatInput)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.chatInput += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, func() tea.Msg {
				return fetchReports(m.apiKey)
			}
		case "enter":
			if i, ok := m.list.SelectedItem().(ReportItem); ok {
				m.selected = &i
				m.aiResponse = ""
				m.updateViewport()
			}
		case "c":
			if m.selected != nil {
				m.chatMode = true
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(35, msg.Height-v-4)
		m.viewport.Width = msg.Width - 40 - h
		m.viewport.Height = msg.Height - v - 6
		m.updateViewport()

	case reportsMsg:
		m.loading = false
		m.reports = msg
		var items []list.Item
		for _, r := range m.reports {
			items = append(items, r)
		}
		m.list.SetItems(items)

	case startChatMsg:
		m.aiResponse = ""
		m.chatInput = ""
		// Start the async worker
		return m, func() tea.Msg {
			startChat(m.program, m.apiKey, msg.reportID, msg.message)
			return nil
		}

	case aiChunkMsg:
		if m.aiResponse == "ANALYZING_TELEMETRY..." {
			m.aiResponse = ""
		}
		m.aiResponse += string(msg)
		m.updateViewport()

	case errorMsg:
		m.err = msg
		m.loading = false
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateViewport() {
	if m.selected == nil {
		m.viewport.SetContent("SELECT A REPORT FROM THE LEFT TO ANALYZE.")
		return
	}

	content := fmt.Sprintf(
		"%s\n\n%s %s\n%s %s\n\n%s\n%s\n\n%s\n%s",
		hazardStyle.Render("=== TACTICAL_REPORT_DNA ==="),
		hazardStyle.Render("ID:       "), m.selected.ID,
		hazardStyle.Render("TIMESTAMP:"), time.Unix(m.selected.Timestamp, 0).Format(time.RFC1123),
		hazardStyle.Render("ERROR_MESSAGE:"),
		m.selected.Msg,
		hazardStyle.Render("STACK_TRACE:"),
		m.selected.StackTrace,
	)

	if m.selected.AIInsight != "" {
		content += fmt.Sprintf("\n\n%s\n%s", hazardStyle.Render("AI_FORENSIC_INSIGHT:"), m.selected.AIInsight)
	}

	if m.aiResponse != "" {
		content += fmt.Sprintf("\n\n%s\n%s", hazardStyle.Render("AI_CHAT_STREAM:"), m.aiResponse)
	}

	m.viewport.SetContent(content)
}

func (m model) View() string {
	if m.showSetup {
		return docStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				headerStyle.Render(" INITIAL_DEPLOYMENT_CONFIG_REQUIRED "),
				"\nAPEX detected no project identity. Please provide your Ingest Key.",
				infoStyle.Render("Visit https://apex-addis.vercel.app to get your key.\n"),
				hazardStyle.Render("INGEST_KEY: ")+m.setupInput+"█",
				"\n [Enter] Confirm  |  [Ctrl+C] Abort",
			),
		)
	}

	if m.err != nil {
		return fmt.Sprintf("\n  ❌ SIGNAL_LOSS: %v\n\n  Press 'q' to quit.", m.err)
	}

	if m.loading {
		return "\n  🛰 LOADING_TELEMETRY_FROM_NODE...\n"
	}

	sidebar := listStyle.Render(m.list.View())

	detail := detailStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.viewport.View(),
		),
	)

	help := infoStyle.Render("\n  [↑/↓]: Scroll List | [Enter]: View Details | [c]: Chat with AI | [r]: Refresh | [q]: Quit")
	if m.chatMode {
		help = titleStyle.Render(" CHAT_MODE ") + " Enter query: " + m.chatInput + "█" + infoStyle.Render("  [Enter]: Send | [Esc]: Cancel")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(" APEX_TACTICAL_FORENSICS_UNIT // VERSION_1.0_TUI "),
		lipgloss.JoinHorizontal(lipgloss.Top, sidebar, detail),
		help,
	)
}

// --- MESSAGES & COMMANDS ---

type reportsMsg []ReportItem
type aiChunkMsg string
type errorMsg error

func fetchReports(apiKey string) tea.Msg {
	apiBase := os.Getenv("APEX_API_URL")
	if apiBase == "" {
		apiBase = "http://localhost:8081"
	}

	req, _ := http.NewRequest("GET", apiBase+"/api/reports", nil)
	req.Header.Set("X-Apex-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errorMsg(err)
	}
	defer resp.Body.Close()

	var reports []ReportItem
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return errorMsg(err)
	}
	return reportsMsg(reports)
}

func startChat(p *tea.Program, apiKey, reportID, message string) {
	go func() {
		apiBase := os.Getenv("APEX_API_URL")
		if apiBase == "" {
			apiBase = "http://localhost:8081"
		}

		body, _ := json.Marshal(map[string]string{
			"report_id": reportID,
			"message":   message,
		})

		req, _ := http.NewRequest("POST", apiBase+"/api/chat", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Apex-API-Key", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			p.Send(errorMsg(err))
			return
		}
		defer resp.Body.Close()

		// Consume SSE stream
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				content := strings.TrimPrefix(line, "data: ")
				if content == "[DONE]" {
					break
				}
				// Send chunk to the UI
				p.Send(aiChunkMsg(content))
			}
		}

		if err := scanner.Err(); err != nil {
			p.Send(errorMsg(err))
		}
	}()
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p // Store the program instance for async updates

	if _, err := p.Run(); err != nil {
		fmt.Printf("ALARM_ERROR: %v", err)
		os.Exit(1)
	}
}
