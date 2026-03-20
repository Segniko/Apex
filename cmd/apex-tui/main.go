package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type Message struct {
	Role    string // "user" or "ai"
	Content string
}

func (i ReportItem) Title() string { return i.Msg }
func (i ReportItem) Description() string {
	return fmt.Sprintf("ID: %s... | OS: %s", i.ID[:8], i.Context.OS)
}
func (i ReportItem) FilterValue() string { return i.Msg }

// --- BUBBLE TEA MODEL ---

type model struct {
	list          list.Model
	viewport      viewport.Model // Chat History
	infoViewport  viewport.Model // Selected Error Details/Stacktrace
	reports       []ReportItem
	selected      ReportItem
	hasSelected   bool
	loading       bool
	err           error
	chatMode      bool
	chatInput     string
	messages      []Message
	width, height int
	program       *tea.Program

	configPath string
	apiKey     string
	apiBase    string
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

func initialModel() *model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "APEX_FORENSICS_TERMINAL"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".apex_config.json")

	m := &model{
		list:         l,
		viewport:     viewport.New(0, 0),
		infoViewport: viewport.New(0, 0),
		configPath:   configPath,
		loading:      true,
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

	// Use NEXT_PUBLIC_API_URL for consistency with the web dashboard
	envFiles := []string{"dashboard/.env.local", ".env"}
	found := false
	for _, envFile := range envFiles {
		if content, err := os.ReadFile(envFile); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "NEXT_PUBLIC_API_URL=") {
					m.apiBase = strings.TrimPrefix(line, "NEXT_PUBLIC_API_URL=")
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	return m
}

func (m *model) Init() tea.Cmd {
	if m.showSetup {
		return nil
	}
	return func() tea.Msg {
		return fetchReports(m.apiBase, m.apiKey)
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys (work in both modes)
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "pgup", "ctrl+u":
			m.viewport.LineUp(6)
			return m, nil
		case "pgdown", "ctrl+d":
			m.viewport.LineDown(6)
			return m, nil
		}

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
						return fetchReports(m.apiBase, m.apiKey)
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
				if m.chatInput != "" && m.hasSelected {
					// Setup local state immediately before spawning the command
					msgStr := m.chatInput
					return m, tea.Batch(func() tea.Msg {
						return startChatMsg{m.selected.ID, msgStr}
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
		case "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, func() tea.Msg {
				return fetchReports(m.apiBase, m.apiKey)
			}
		case "enter":
			if i, ok := m.list.SelectedItem().(ReportItem); ok {
				m.selected = i
				m.hasSelected = true
				m.messages = nil
				m.updateViewport()
			}
		case "c":
			if m.hasSelected {
				m.chatMode = true
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := docStyle.GetFrameSize()

		// Layout: Top Row (Errors + Info) | Bottom Row (Chat)
		// Total usable height: msg.Height - v - 6
		usableHeight := msg.Height - v - 8
		if usableHeight < 10 {
			usableHeight = 10
		}

		topHeight := usableHeight / 3
		if topHeight < 5 {
			topHeight = 5
		}
		chatHeight := usableHeight - topHeight
		if chatHeight < 5 {
			chatHeight = 5
		}

		m.list.SetSize(35, topHeight-1)

		infoWidth := msg.Width - 40 - h
		if infoWidth < 10 {
			infoWidth = 10
		}
		m.infoViewport.Width = infoWidth
		m.infoViewport.Height = topHeight - 1

		m.viewport.Width = msg.Width - h
		m.viewport.Height = chatHeight
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
		m.messages = append(m.messages, Message{Role: "user", Content: msg.message})
		m.messages = append(m.messages, Message{Role: "ai", Content: "ANALYZING_TELEMETRY..."})
		m.chatInput = ""
		m.updateViewport()
		return m, func() tea.Msg {
			startChat(m.program, m.apiBase, m.apiKey, msg.reportID, msg.message)
			return nil
		}

	case aiChunkMsg:
		if len(m.messages) > 0 {
			last := &m.messages[len(m.messages)-1]
			if last.Role == "ai" {
				if last.Content == "ANALYZING_TELEMETRY..." {
					last.Content = ""
				}
				last.Content += string(msg)
			}
		}
		m.updateViewport()
		m.viewport.GotoBottom()
		return m, nil

	case errorMsg:
		m.err = msg
		m.loading = false
	}

	var cmd tea.Cmd
	var vpCmd tea.Cmd
	var infoCmd tea.Cmd

	m.list, cmd = m.list.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.infoViewport, infoCmd = m.infoViewport.Update(msg)

	cmds = append(cmds, cmd, vpCmd, infoCmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateViewport() {
	if !m.hasSelected {
		m.infoViewport.SetContent("SELECT A REPORT.")
		m.viewport.SetContent("CHAT_OFFLINE")
		return
	}

	// Update Info Viewport (Top Right)
	info := fmt.Sprintf(
		"%s\n%s %s\n%s %s\n%s\n%s\n\n%s\n%s",
		hazardStyle.Render("=== TACTICAL_REPORT_DNA ==="),
		hazardStyle.Render("ID:       "), m.selected.ID,
		hazardStyle.Render("TIME:     "), time.Unix(m.selected.Timestamp, 0).Format(time.Kitchen),
		hazardStyle.Render("ERROR:    "), m.selected.Msg,
		hazardStyle.Render("TRACE:    "), m.selected.StackTrace,
	)
	m.infoViewport.SetContent(info)

	// Update Chat Viewport (Bottom)
	var chatContent strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "user" {
			chatContent.WriteString(hazardStyle.Render("USER_QUERY: ") + msg.Content + "\n\n")
		} else {
			chatContent.WriteString(hazardStyle.Render("AI_RESPONSE:\n") + msg.Content + "\n\n")
		}
	}

	if len(m.messages) == 0 && m.selected.AIInsight != "" {
		chatContent.WriteString(hazardStyle.Render("PREVIOUS_INSIGHT:\n") + m.selected.AIInsight + "\n")
	}

	wrapped := lipgloss.NewStyle().Width(m.viewport.Width - 2).Render(chatContent.String())
	m.viewport.SetContent(wrapped)
}

func (m *model) View() string {
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
	info := detailStyle.Render(m.infoViewport.View())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, info)

	chat := docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			hazardStyle.Render("=== OPERATIONAL_CHAT_HISTORY ==="),
			m.viewport.View(),
		),
	)

	help := infoStyle.Render("\n  [↑/↓]: Scroll List | [Enter]: Select | [c]: Chat | [r]: Refresh | [q]: Quit")
	if m.chatMode {
		help = titleStyle.Render(" CHAT_MODE ") + " Next Query: " + m.chatInput + "█" + infoStyle.Render("  [Enter]: Post | [Esc]: Back")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(fmt.Sprintf(" APEX_TACTICAL_FORENSICS_UNIT // VERSION_1.0_TUI // DATA_FEED: %d_ITEMS ", len(m.reports))),
		topRow,
		chat,
		help,
	)
}

// --- MESSAGES & COMMANDS ---

type reportsMsg []ReportItem
type aiChunkMsg string
type errorMsg error

func fetchReports(apiBase, apiKey string) tea.Msg {
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

func startChat(p *tea.Program, apiBase, apiKey, reportID, message string) {
	go func() {

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
			p.Send(aiChunkMsg(fmt.Sprintf("ERROR_SIGNAL: SIGNAL_LOSS - %v", err)))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			p.Send(aiChunkMsg(fmt.Sprintf("ERROR_SIGNAL: HTTP %d - %s", resp.StatusCode, string(bodyBytes))))
			return
		}

		// Handle SSE or JSON response
		contentType := resp.Header.Get("Content-Type")

		if strings.Contains(contentType, "application/json") {
			var result struct {
				Response string `json:"response"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				p.Send(aiChunkMsg(result.Response))
			} else {
				p.Send(aiChunkMsg("ERROR_SIGNAL: Failed to decode AI JSON response."))
			}
			return
		}

		// Consume SSE stream
		scanner := bufio.NewScanner(resp.Body)
		var receivedLines int
		for scanner.Scan() {
			line := scanner.Text()
			receivedLines++

			if strings.HasPrefix(line, "data: ") {
				content := strings.TrimPrefix(line, "data: ")
				if content == "[DONE]" {
					break
				}
				receivedLines++
				p.Send(aiChunkMsg(content))
			} else if strings.HasPrefix(line, "{") {
				// Try parsing as JSON even without 'data: ' prefix if the server is inconsistent
				var result struct {
					Response string `json:"response"`
				}
				if err := json.Unmarshal([]byte(line), &result); err == nil {
					p.Send(aiChunkMsg(result.Response))
				}
			}
		}

		if receivedLines == 0 {
			p.Send(aiChunkMsg("\nERROR_SIGNAL: Stream Closed with 0 Chunks. Valid Request, Empty AI Response."))
		}

		if err := scanner.Err(); err != nil {
			p.Send(aiChunkMsg(fmt.Sprintf("\nERROR_SIGNAL: Stream Error - %v", err)))
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
