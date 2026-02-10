//go:build !simplewizard

// Package tui provides terminal user interface components for clawdepl.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// NewInstanceResult holds the result of the new instance wizard.
type NewInstanceResult struct {
	Name        string
	AuthChoice  string
	ClaudeToken string
	Purpose     string
	Cancelled   bool
	Error       error
}

// Step represents a step in the wizard.
type step int

const (
	stepName step = iota
	stepAuthChoice
	stepClaudeToken
	stepPurpose
	stepProvisioning
	stepDone
)

// NewInstanceModel is the Bubble Tea model for creating a new instance.
type NewInstanceModel struct {
	step         step
	inputs       []textinput.Model
	tokenInput   textarea.Model
	purposeInput textarea.Model
	spinner      spinner.Model
	result       NewInstanceResult
	initialName  string
	width        int
	provisionErr error
	authChoices  []string
	authIndex    int
	tokenHint    string
	// Callback for provisioning
	onProvision func(name, token, authChoice, purpose string) error
}

// NewInstanceOption is a functional option for NewInstanceModel.
type NewInstanceOption func(*NewInstanceModel)

// WithName sets the initial name (skips name step).
func WithName(name string) NewInstanceOption {
	return func(m *NewInstanceModel) {
		m.initialName = name
	}
}

// WithProvisionCallback sets the callback for provisioning.
func WithProvisionCallback(fn func(name, token, authChoice, purpose string) error) NewInstanceOption {
	return func(m *NewInstanceModel) {
		m.onProvision = fn
	}
}

// NewNewInstanceModel creates a new model for the instance creation wizard.
func NewNewInstanceModel(opts ...NewInstanceOption) NewInstanceModel {
	m := NewInstanceModel{
		step:        stepName,
		inputs:      make([]textinput.Model, 1),
		authChoices: []string{"anthropic-api-key", "anthropic-setup-token"},
	}

	// Apply options
	for _, opt := range opts {
		opt(&m)
	}

	// Name input
	nameInput := textinput.New()
	nameInput.Placeholder = "my-agent"
	nameInput.CharLimit = 64
	nameInput.Width = 40
	if m.initialName != "" {
		nameInput.SetValue(m.initialName)
	}
	m.inputs[0] = nameInput

	// Claude credential input (API key or setup-token).
	tokenInput := textarea.New()
	tokenInput.Placeholder = "sk-ant-api03-... (or setup-token)"
	tokenInput.CharLimit = 512
	tokenInput.SetWidth(70)
	tokenInput.SetHeight(3)
	tokenInput.ShowLineNumbers = false
	m.tokenInput = tokenInput

	// Purpose input (textarea for multiline)
	purposeInput := textarea.New()
	purposeInput.Placeholder = "# IDENTITY\n\nYou are...\n"
	purposeInput.CharLimit = 1024
	purposeInput.SetWidth(60)
	purposeInput.SetHeight(4)
	purposeInput.ShowLineNumbers = false
	m.purposeInput = purposeInput

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	m.spinner = s

	// Skip name step if name was provided
	if m.initialName != "" {
		m.result.Name = m.initialName
		m.step = stepAuthChoice
	} else {
		m.inputs[0].Focus()
	}

	return m
}

// Init initializes the model.
func (m NewInstanceModel) Init() tea.Cmd {
	return textinput.Blink
}

// provisionMsg is sent when provisioning is complete.
type provisionMsg struct {
	err error
}

// Update handles messages.
func (m NewInstanceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.result.Cancelled = true
			return m, tea.Quit

		case "enter":
			// Enter always submits
			return m.handleEnter()

		case "ctrl+enter":
			// Ctrl+Enter adds newline only for purpose step (not token)
			if m.step == stepPurpose {
				m.purposeInput.InsertString("\n")
				return m, nil
			}

		case "tab", "shift+tab":
			// Don't allow tab navigation, just use enter to proceed
			return m, nil
		}

		if m.step == stepAuthChoice {
			switch msg.String() {
			case "up", "left":
				if m.authIndex > 0 {
					m.authIndex--
				}
				return m, nil
			case "down", "right":
				if m.authIndex < len(m.authChoices)-1 {
					m.authIndex++
				}
				return m, nil
			case "1":
				m.authIndex = 0
				return m, nil
			case "2":
				if len(m.authChoices) > 1 {
					m.authIndex = 1
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case spinner.TickMsg:
		if m.step == stepProvisioning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case provisionMsg:
		if msg.err != nil {
			m.provisionErr = msg.err
			m.result.Error = msg.err
			return m, tea.Quit
		}
		m.step = stepDone
		return m, tea.Quit
	}

	// Update current input
	var cmd tea.Cmd
	switch m.step {
	case stepName:
		m.inputs[0], cmd = m.inputs[0].Update(msg)
	case stepClaudeToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	case stepPurpose:
		m.purposeInput, cmd = m.purposeInput.Update(msg)
	}

	return m, cmd
}

func (m NewInstanceModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepName:
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			return m, nil
		}
		m.result.Name = name
		m.step = stepAuthChoice
		m.inputs[0].Blur()
		return m, nil

	case stepAuthChoice:
		m.result.AuthChoice = m.authChoices[m.authIndex]
		m.step = stepClaudeToken
		m.tokenInput.Focus()
		if m.result.AuthChoice == "anthropic-setup-token" {
			// Manual flow: do not run `claude setup-token` or attempt auto-import.
			m.tokenInput.Placeholder = "sk-ant-oat01-..."
			m.tokenHint = "Run: claude setup-token. Then paste the raw token (sk-ant-oat...) and press Enter."
		} else {
			m.tokenInput.Placeholder = "sk-ant-api03-..."
			m.tokenHint = ""
		}
		return m, textarea.Blink

	case stepClaudeToken:
		// Manual flow: accept the raw token only (no blob parsing).
		token := strings.TrimSpace(m.tokenInput.Value())
		token = strings.ReplaceAll(token, "\n", "")
		token = strings.ReplaceAll(token, "\r", "")
		if token == "" {
			if m.result.AuthChoice == "anthropic-setup-token" {
				m.tokenHint = "Setup-token is required. Paste the raw sk-ant-oat... token and press Enter."
			} else {
				m.tokenHint = "An API key is required."
			}
			return m, nil
		}
		// Token validity is checked before provisioning in cmd/new.go.
		m.result.ClaudeToken = token
		m.step = stepPurpose
		m.tokenInput.Blur()
		m.purposeInput.Focus()
		return m, textarea.Blink

	case stepPurpose:
		purpose := strings.TrimSpace(m.purposeInput.Value())
		if purpose == "" {
			return m, nil
		}
		m.result.Purpose = purpose
		m.step = stepProvisioning
		m.purposeInput.Blur()

		// Start provisioning
		return m, tea.Batch(
			m.spinner.Tick,
			m.startProvisioning(),
		)
	}

	return m, nil
}

func (m NewInstanceModel) startProvisioning() tea.Cmd {
	return func() tea.Msg {
		if m.onProvision != nil {
			err := m.onProvision(m.result.Name, m.result.ClaudeToken, m.result.AuthChoice, m.result.Purpose)
			return provisionMsg{err: err}
		}
		return provisionMsg{}
	}
}

// View renders the UI.
func (m NewInstanceModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Create a new OpenClaw instance"))
	b.WriteString("\n\n")

	switch m.step {
	case stepName:
		b.WriteString(promptStyle.Render("What would you like to name your instance?"))
		b.WriteString("\n\n")
		b.WriteString(m.inputs[0].View())
		b.WriteString("\n\n")
		b.WriteString(blurredStyle.Render("Press Enter to continue, Esc to cancel"))

	case stepAuthChoice:
		b.WriteString(m.renderCompletedStep("Name", m.result.Name))
		b.WriteString("\n")
		b.WriteString(promptStyle.Render("How do you want to authenticate Anthropic?"))
		b.WriteString("\n\n")
		b.WriteString(m.renderAuthChoice())
		b.WriteString("\n\n")
		b.WriteString(blurredStyle.Render("Use ↑/↓ (or 1/2), then Enter to continue."))

	case stepClaudeToken:
		b.WriteString(m.renderCompletedStep("Name", m.result.Name))
		b.WriteString(m.renderCompletedStep("Auth", humanAuthChoice(m.result.AuthChoice)))
		b.WriteString("\n")
		b.WriteString(promptStyle.Render("Enter your Claude credential:"))
		b.WriteString("\n\n")
		b.WriteString(m.tokenInput.View())
		b.WriteString("\n\n")
		if m.result.AuthChoice == "anthropic-setup-token" {
			b.WriteString(blurredStyle.Render("Paste the raw token from: claude setup-token (starts with sk-ant-oat...)."))
		} else {
			b.WriteString(blurredStyle.Render("Use an Anthropic API key (starts with sk-ant-...)."))
		}
		b.WriteString("\n")
		b.WriteString(blurredStyle.Render("Your token is stored securely. Press Enter to continue."))
		if m.tokenHint != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.tokenHint))
		}

	case stepPurpose:
		b.WriteString(m.renderCompletedStep("Name", m.result.Name))
		b.WriteString(m.renderCompletedStep("Auth", humanAuthChoice(m.result.AuthChoice)))
		b.WriteString(m.renderCompletedStep("Token", "••••••••"))
		b.WriteString("\n")
		b.WriteString(promptStyle.Render("Write IDENTITY.md (who am I?)"))
		b.WriteString("\n\n")
		b.WriteString(m.purposeInput.View())
		b.WriteString("\n\n")
		b.WriteString(blurredStyle.Render("Enter to submit, Ctrl+Enter for new line."))

	case stepProvisioning:
		b.WriteString(m.renderCompletedStep("Name", m.result.Name))
		b.WriteString(m.renderCompletedStep("Auth", humanAuthChoice(m.result.AuthChoice)))
		b.WriteString(m.renderCompletedStep("Token", "••••••••"))
		b.WriteString(m.renderCompletedStep("Purpose", m.result.Purpose))
		b.WriteString("\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" Provisioning your instance...")

	case stepDone:
		b.WriteString(m.renderCompletedStep("Name", m.result.Name))
		b.WriteString(m.renderCompletedStep("Auth", humanAuthChoice(m.result.AuthChoice)))
		b.WriteString(m.renderCompletedStep("Token", "••••••••"))
		b.WriteString(m.renderCompletedStep("Purpose", m.result.Purpose))
		b.WriteString("\n")
		b.WriteString(successStyle.Render("✓ Done!"))
		b.WriteString(" Your instance is ready.\n\n")
		b.WriteString("  Next: run ")
		b.WriteString(focusedStyle.Render("clawdepl list"))
		b.WriteString(" to find the instance name/ID, then use ")
		b.WriteString(focusedStyle.Render("clawdepl status <name-or-id>"))
		b.WriteString(".\n")
	}

	b.WriteString("\n")
	return b.String()
}

func (m NewInstanceModel) renderCompletedStep(label, value string) string {
	return fmt.Sprintf("%s %s\n",
		successStyle.Render("✓"),
		fmt.Sprintf("%s: %s", promptStyle.Render(label), value))
}

func (m NewInstanceModel) renderAuthChoice() string {
	var b strings.Builder
	for i, choice := range m.authChoices {
		prefix := "  "
		label := humanAuthChoice(choice)
		if i == m.authIndex {
			prefix = focusedStyle.Render("➤ ")
			label = focusedStyle.Render(label)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, label))
	}
	return b.String()
}

func humanAuthChoice(choice string) string {
	switch choice {
	case "anthropic-setup-token":
		return "Setup-token"
	default:
		return "API key"
	}
}

// Result returns the result of the wizard.
func (m NewInstanceModel) Result() NewInstanceResult {
	return m.result
}

// RunNewInstanceWizard runs the new instance wizard and returns the result.
func RunNewInstanceWizard(name string, onProvision func(name, token, authChoice, purpose string) error) (*NewInstanceResult, error) {
	opts := []NewInstanceOption{
		WithProvisionCallback(onProvision),
	}
	if name != "" {
		opts = append(opts, WithName(name))
	}

	model := NewNewInstanceModel(opts...)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := finalModel.(NewInstanceModel).Result()
	if result.AuthChoice == "" {
		result.AuthChoice = "anthropic-api-key"
	}
	return &result, nil
}
