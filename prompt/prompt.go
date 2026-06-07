// Package prompt provides tiny stdin helpers for interactive CLI questions.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Prompter reads lines from r and writes questions to w.
type Prompter struct {
	r *bufio.Reader
	w io.Writer
}

// New returns a Prompter backed by stdin/stderr (stderr so prompts survive
// stdout redirection).
func New() *Prompter {
	return &Prompter{r: bufio.NewReader(os.Stdin), w: os.Stderr}
}

// String asks the user and returns their reply, or def if the reply is empty.
func (p *Prompter) String(question, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(p.w, "%s [%s]: ", question, def)
	} else {
		fmt.Fprintf(p.w, "%s: ", question)
	}
	line, err := p.r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

// Choice prompts until the user picks one of the allowed options.
func (p *Prompter) Choice(question string, options []string, def string) (string, error) {
	for {
		ans, err := p.String(question, def)
		if err != nil {
			return "", err
		}
		for _, o := range options {
			if ans == o {
				return ans, nil
			}
		}
		fmt.Fprintf(p.w, "  invalid — must be one of %s\n", strings.Join(options, ", "))
	}
}

// Int asks for an integer, re-prompting on invalid input.
func (p *Prompter) Int(question string, def int) (int, error) {
	for {
		ans, err := p.String(question, strconv.Itoa(def))
		if err != nil {
			return 0, err
		}
		n, err := strconv.Atoi(ans)
		if err != nil {
			fmt.Fprintln(p.w, "  invalid — must be an integer")
			continue
		}
		return n, nil
	}
}

// Confirm asks a yes/no question.
func (p *Prompter) Confirm(question string, def bool) (bool, error) {
	d := "y/N"
	if def {
		d = "Y/n"
	}
	ans, err := p.String(question+" ["+d+"]", "")
	if err != nil {
		return false, err
	}
	if ans == "" {
		return def, nil
	}
	switch strings.ToLower(ans[:1]) {
	case "y":
		return true, nil
	case "n":
		return false, nil
	}
	return def, nil
}

// Info prints an informational line to the prompt output stream.
func (p *Prompter) Info(format string, args ...any) {
	fmt.Fprintf(p.w, format+"\n", args...)
}
