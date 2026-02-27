package compiler

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// regexes compiled once at package init.
var (
	// [Speaker] (key: val, key2: val2) "text"
	// Speaker and params are optional; text is required.
	reDialog = regexp.MustCompile(
		`^\[([^\]]*)\]\s*(?:\(([^)]*)\)\s*)?"((?:[^"\\]|\\.)*)"$`,
	)

	// @command: value --key1 val1 --key2 val2
	// value part is optional (e.g. @wait)
	reSystem = regexp.MustCompile(
		`^@([\w-]+)(?::\s*([^-\s][^\s]*))?(.*)$`,
	)

	// --key value  (value may be multi-word until next --)
	reOption = regexp.MustCompile(`--(\w+)\s+([^\s-][^\s]*)`)

	// - text -> Label  or  - text
	reChoiceOption = regexp.MustCompile(`^-\s+(.+?)(?:\s+->\s*(\S+))?$`)

	// # LabelName
	reLabel = regexp.MustCompile(`^#\s*(\S+)`)
)

// Parse converts a .nvn source string into an ordered slice of Nodes.
// Malformed lines are logged and skipped (Syntax Fallback Rule from SKILL.md).
func Parse(src string) []Node {
	tokens := Tokenize(src)
	nodes := make([]Node, 0, len(tokens))

	i := 0
	for i < len(tokens) {
		tok := tokens[i]

		switch tok.Kind {
		case TokBlank:
			i++

		case TokLabel:
			n := parseLabel(tok)
			if n != nil {
				nodes = append(nodes, n)
			}
			i++

		case TokSystem:
			n := parseSystem(tok)
			if n != nil {
				nodes = append(nodes, n)
			}
			i++

		case TokDialog:
			n := parseDialog(tok)
			if n != nil {
				nodes = append(nodes, n)
			}
			i++

		case TokChoice:
			n, consumed := parseChoice(tokens, i)
			nodes = append(nodes, n)
			i += consumed

		case TokOption:
			// Orphan option outside a choice block — log and skip
			log.Printf("[WARN] compiler: line %d: orphan option %q, skipping", tok.Line, tok.Raw)
			i++

		default:
			i++
		}
	}

	return nodes
}

// parseLabel parses "# LabelName".
func parseLabel(tok Token) Node {
	m := reLabel.FindStringSubmatch(tok.Raw)
	if m == nil {
		log.Printf("[WARN] compiler: line %d: invalid label %q, skipping", tok.Line, tok.Raw)
		return nil
	}
	return &LabelNode{Name: m[1]}
}

// parseSystem parses "@command: value --key val ...".
func parseSystem(tok Token) Node {
	m := reSystem.FindStringSubmatch(tok.Raw)
	if m == nil {
		log.Printf("[WARN] compiler: line %d: invalid system directive %q, skipping", tok.Line, tok.Raw)
		return nil
	}
	cmd := m[1]
	value := strings.TrimSpace(m[2])
	optStr := m[3]

	opts := make(map[string]string)
	for _, om := range reOption.FindAllStringSubmatch(optStr, -1) {
		opts[om[1]] = om[2]
	}

	return &SystemNode{Command: cmd, Value: value, Options: opts}
}

// parseDialog parses [Speaker] (params) "text".
func parseDialog(tok Token) Node {
	m := reDialog.FindStringSubmatch(tok.Raw)
	if m == nil {
		// Syntax fallback: treat entire line as anonymous dialogue
		log.Printf("[WARN] compiler: line %d: cannot parse dialogue %q, treating as raw text", tok.Line, tok.Raw)
		return &DialogNode{Text: tok.Raw}
	}

	speaker := strings.TrimSpace(m[1])
	paramStr := m[2]
	text := m[3]

	params := parseInlineParams(paramStr)
	return &DialogNode{Speaker: speaker, Params: params, Text: text}
}

// parseInlineParams converts "key: val, key2: val2" into a map.
func parseInlineParams(s string) map[string]string {
	params := make(map[string]string)
	if s == "" {
		return params
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if k != "" {
				params[k] = v
			}
		}
	}
	return params
}

// parseChoice parses a ">>" block followed by "- option" lines.
// Returns the ChoiceNode and the number of tokens consumed.
func parseChoice(tokens []Token, start int) (Node, int) {
	tok := tokens[start]
	prompt := strings.TrimSpace(strings.TrimPrefix(tok.Raw, ">>"))
	prompt = strings.TrimSuffix(prompt, "：") // trim CJK colon
	prompt = strings.TrimSuffix(prompt, ":")
	prompt = strings.TrimSpace(prompt)

	node := &ChoiceNode{Prompt: prompt}
	consumed := 1

	for start+consumed < len(tokens) {
		next := tokens[start+consumed]
		if next.Kind != TokOption {
			break
		}
		opt := parseChoiceOption(next)
		if opt != nil {
			node.Options = append(node.Options, *opt)
		}
		consumed++
	}

	if len(node.Options) == 0 {
		log.Printf("[WARN] compiler: line %d: choice block has no options", tok.Line)
	}

	return node, consumed
}

// parseChoiceOption parses "- text -> Label" or "- text".
func parseChoiceOption(tok Token) *ChoiceOption {
	m := reChoiceOption.FindStringSubmatch(tok.Raw)
	if m == nil {
		log.Printf("[WARN] compiler: line %d: invalid option %q, skipping", tok.Line, tok.Raw)
		return nil
	}
	return &ChoiceOption{Text: strings.TrimSpace(m[1]), Label: m[2]}
}

// Compile is a convenience wrapper: source string → (nodes, error).
// It always returns nodes (partial on error); err is non-nil only for
// completely empty input.
func Compile(src string) ([]Node, error) {
	if strings.TrimSpace(src) == "" {
		return nil, fmt.Errorf("compiler: empty source")
	}
	return Parse(src), nil
}
