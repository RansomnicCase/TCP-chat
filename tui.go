package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- CONFIG ---
// Change this if you are using a different port or ngrok
const DEFAULT_ADDR = "localhost:9001"

// --- STYLES ---
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// --- DATA MODEL ---
type model struct {
	viewport viewport.Model // The chat history window
	textarea textarea.Model // The input box
	messages []string       // List of messages
	conn     net.Conn       // The connection to server
	addr     string         // Server address
	err      error
}

// Custom message types for the event loop
type serverMsg string
type errMsg error

// 1. INITIALIZE (Setup the UI elements)
func initialModel(conn net.Conn, addr string) model {
	// Setup the Input Box
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Hit Enter to send, not new line

	// Setup the Chat History Window
	vp := viewport.New(30, 5) // Dimensions will be overwritten by window resize
	vp.SetContent("Connected to " + addr + "\nType a message and hit Enter.\n---------------------------")

	return model{
		textarea: ta,
		messages: []string{},
		viewport: vp,
		conn:     conn,
		addr:     addr,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// 2. UPDATE (Handle Logic & Keypresses)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {

	// Resize window automatically
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - 2
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	// Receive message from server
	case serverMsg:
		m.messages = append(m.messages, string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	// Handle Keyboard
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			input := m.textarea.Value()
			if input != "" {
				// Send to server
				fmt.Fprintf(m.conn, "%s\n", input)
				m.textarea.Reset()
			}
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, tiCmd = m.textarea.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

// 3. VIEW (Render the UI)
func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

// --- MAIN EXECUTION ---
func main1() {
	reader := bufio.NewReader(os.Stdin)

	// A. Ask for Server Address
	fmt.Printf("Enter Server Address (default: %s): ", DEFAULT_ADDR)
	addr, _ := reader.ReadString('\n')
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = DEFAULT_ADDR
	}

	// B. Connect to Server
	fmt.Printf("Connecting to %s...\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ Connection failed:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// C. Handshake (Handle the "Enter Name" step)
	// We do this BEFORE starting the UI so the text doesn't mess up the graphics
	connReader := bufio.NewReader(conn)

	// Read "Enter Name:" prompt from server
	_, _ = connReader.ReadString('\n') // Ignore the prompt text

	fmt.Print("Enter your Alias: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Send name to server
	fmt.Fprintf(conn, "%s\n", name)

	// D. Start the Bubble Tea UI
	p := tea.NewProgram(initialModel(conn, addr))

	// Start a background listener to read incoming messages
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			text := scanner.Text()
			// Send message to the UI loop
			p.Send(serverMsg(text))
		}
		// If connection dies, quit
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
