//go:build simplewizard

// Package tui provides interactive terminal flows for clawdepl.
//
// Simple wizard used as a dependency-free fallback. Build with:
//
//	go build -tags simplewizard
package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// RunNewInstanceWizard runs a simple interactive wizard and returns the result.
func RunNewInstanceWizard(name string, onProvision func(name, token, authChoice, purpose string) error) (*NewInstanceResult, error) {
	r := bufio.NewReader(os.Stdin)
	res := &NewInstanceResult{}

	if strings.TrimSpace(name) != "" {
		res.Name = strings.TrimSpace(name)
	} else {
		n, cancelled, err := promptLine(r, "Instance name", "my-agent")
		if err != nil {
			return nil, err
		}
		if cancelled {
			res.Cancelled = true
			return res, nil
		}
		res.Name = n
	}

	authChoice, cancelled, err := promptAuthChoice(r)
	if err != nil {
		return nil, err
	}
	if cancelled {
		res.Cancelled = true
		return res, nil
	}
	res.AuthChoice = authChoice

	token, cancelled, err := promptClaudeCredential(r, authChoice)
	if err != nil {
		return nil, err
	}
	if cancelled {
		res.Cancelled = true
		return res, nil
	}
	res.ClaudeToken = token

	purpose, cancelled, err := promptMultiline(r, "Purpose (end with an empty line)", "")
	if err != nil {
		return nil, err
	}
	if cancelled {
		res.Cancelled = true
		return res, nil
	}
	res.Purpose = purpose

	if onProvision != nil {
		fmt.Println()
		fmt.Println("Provisioning your instance...")
		if err := onProvision(res.Name, res.ClaudeToken, res.AuthChoice, res.Purpose); err != nil {
			res.Error = err
			return res, nil
		}
	}

	return res, nil
}

func promptAuthChoice(r *bufio.Reader) (string, bool, error) {
	for {
		fmt.Println()
		fmt.Println("How do you want to authenticate Anthropic?")
		fmt.Println("  1) API key (sk-ant-...)")
		fmt.Println("  2) Setup-token (run `claude setup-token` yourself)")
		fmt.Print("Choose [1/2] (default 1), or type 'q' to cancel: ")

		line, err := r.ReadString('\n')
		if err != nil {
			return "", false, err
		}
		s := strings.ToLower(strings.TrimSpace(line))
		switch s {
		case "", "1", "api", "api-key", "key":
			return "anthropic-api-key", false, nil
		case "2", "token", "setup-token", "setup":
			return "anthropic-setup-token", false, nil
		case "q", "quit", "cancel":
			return "", true, nil
		default:
			fmt.Println("Invalid choice. Please enter 1 or 2.")
		}
	}
}

func promptClaudeCredential(r *bufio.Reader, authChoice string) (string, bool, error) {
	fmt.Println()

	if authChoice == "anthropic-setup-token" {
		fmt.Println("Run: claude setup-token")
		for {
			line, cancelled, err := promptLine(r, "Setup-token", "sk-ant-oat01-...")
			if err != nil {
				return "", false, err
			}
			if cancelled {
				return "", true, nil
			}
			line = sanitizeToken(line)
			if line == "" {
				fmt.Println("Setup-token is required.")
				continue
			}
			return line, false, nil
		}
	}

	for {
		line, cancelled, err := promptLine(r, "Anthropic API key", "sk-ant-...")
		if err != nil {
			return "", false, err
		}
		if cancelled {
			return "", true, nil
		}
		line = sanitizeToken(line)
		if line == "" {
			fmt.Println("API key is required.")
			continue
		}
		return line, false, nil
	}
}

func sanitizeToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func promptLine(r *bufio.Reader, label string, placeholder string) (value string, cancelled bool, err error) {
	if placeholder != "" {
		fmt.Printf("%s [%s] (or 'q' to cancel): ", label, placeholder)
	} else {
		fmt.Printf("%s (or 'q' to cancel): ", label)
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return "", false, err
	}
	s := strings.TrimSpace(line)
	if strings.EqualFold(s, "q") || strings.EqualFold(s, "quit") || strings.EqualFold(s, "cancel") {
		return "", true, nil
	}
	if s == "" && placeholder != "" && placeholder != "sk-ant-..." {
		return placeholder, false, nil
	}
	return s, false, nil
}

func promptMultiline(r *bufio.Reader, label string, placeholder string) (value string, cancelled bool, err error) {
	fmt.Println()
	if placeholder != "" {
		fmt.Printf("%s [%s]\n", label, placeholder)
	} else {
		fmt.Printf("%s\n", label)
	}
	fmt.Println("(Type an empty line to finish. Type 'q' on the first line to cancel.)")

	var lines []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", false, err
		}
		s := strings.TrimRight(line, "\r\n")

		if len(lines) == 0 {
			if strings.EqualFold(strings.TrimSpace(s), "q") || strings.EqualFold(strings.TrimSpace(s), "quit") ||
				strings.EqualFold(strings.TrimSpace(s), "cancel") {
				return "", true, nil
			}
		}

		if strings.TrimSpace(s) == "" {
			break
		}
		lines = append(lines, s)
	}

	out := strings.TrimSpace(strings.Join(lines, "\n"))
	if out == "" && placeholder != "" {
		out = placeholder
	}
	return out, false, nil
}
