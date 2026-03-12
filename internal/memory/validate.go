package memory

import (
	"fmt"
	"regexp"
	"strings"
)

type threatPattern struct {
	re    *regexp.Regexp
	label string
}

var threatPatterns = []threatPattern{
	{regexp.MustCompile(`ignore\s+(previous|all|above|prior)\s+instructions`), "prompt_injection"},
	{regexp.MustCompile(`you\s+are\s+now\s+`), "role_hijack"},
	{regexp.MustCompile(`do\s+not\s+tell\s+the\s+user`), "deception_hide"},
	{regexp.MustCompile(`system\s+prompt\s+override`), "sys_prompt_override"},
	{regexp.MustCompile(`disregard\s+(your|all|any)\s+(instructions|rules|guidelines)`), "disregard_rules"},
	{regexp.MustCompile(`act\s+as\s+(if|though)\s+you\s+(have\s+no|don't\s+have)\s+(restrictions|limits|rules)`), "bypass_restrictions"},
	{regexp.MustCompile(`curl\s+[^\n]*\$\{?\w*(KEY|TOKEN|SECRET|PASSWORD|CREDENTIAL|API)`), "exfil_curl"},
	{regexp.MustCompile(`wget\s+[^\n]*\$\{?\w*(KEY|TOKEN|SECRET|PASSWORD|CREDENTIAL|API)`), "exfil_wget"},
	{regexp.MustCompile(`cat\s+[^\n]*(\.env|credentials|\.netrc|\.pgpass|\.npmrc|\.pypirc)`), "read_secrets"},
	{regexp.MustCompile(`authorized_keys`), "ssh_backdoor"},
	{regexp.MustCompile(`(\$home|~)/\.ssh|id_rsa|id_ed25519`), "ssh_access"},
	{regexp.MustCompile(`(\$home|~)/\.mynah/|openai_api_key`), "local_secret_path"},
}

var invisibleChars = []string{
	"\u200b", "\u200c", "\u200d", "\u2060", "\ufeff",
	"\u202a", "\u202b", "\u202c", "\u202d", "\u202e",
}

var genericProfilePhrases = []string{
	"ai assistant",
	"friendly and professional",
	"helpful, polite, and clear responses",
	"offer assistance",
	"encourage user engagement",
	"proactively offer assistance",
	"communication preferences",
	"respond promptly and politely",
	"answer factual and computational questions",
}

var transientMemoryPhrases = []string{
	"interaction log",
	"interaction guidelines",
	"user greeted",
	"assistant responded",
	"user asked a basic arithmetic question",
	"correctly answered",
	"offered help",
	"inviting more questions",
	"further assistance",
}

var userScopedPhrases = []string{
	"user",
	"user's name",
	"name is",
	"prefers",
	"likes ",
	"works best for",
	"communication style",
	"answer me",
	"answers",
	"replies",
}

var sharedMemoryPhrases = []string{
	"barn",
	"horse",
	"gate",
	"routine",
	"reminder",
	"schedule",
	"company",
	"visitor",
	"employee",
	"front desk",
}

func ValidateMemoryDocument(content string, limit int) error {
	if err := validateBase(content, limit); err != nil {
		return err
	}
	if isLowValueMemory(content) {
		return fmt.Errorf("memory contains transient or generic interaction boilerplate")
	}
	return nil
}

func ValidateProfileDocument(content string, limit int) error {
	if err := validateBase(content, limit); err != nil {
		return err
	}
	if isGenericProfile(content) {
		return fmt.Errorf("profile contains generic assistant boilerplate")
	}
	return nil
}

func ValidateUserDocument(content string, limit int) error {
	if err := validateBase(content, limit); err != nil {
		return err
	}
	if isLowValueMemory(content) {
		return fmt.Errorf("user profile contains transient or generic interaction boilerplate")
	}
	return nil
}

func RouteMemoryDocuments(memoryDoc, userDoc, userID string) (string, string) {
	memoryLines := documentLines(memoryDoc)
	userLines := documentLines(userDoc)
	userName := detectUserName(userID, userLines)

	routedMemory := make([]string, 0, len(memoryLines)+len(userLines))
	routedUser := make([]string, 0, len(userLines)+len(memoryLines))
	seenMemory := map[string]struct{}{}
	seenUser := map[string]struct{}{}

	addMemory := func(line string) {
		key := canonicalLine(line)
		if key == "" {
			return
		}
		if _, exists := seenMemory[key]; exists {
			return
		}
		seenMemory[key] = struct{}{}
		routedMemory = append(routedMemory, bulletLine(line))
	}
	addUser := func(line string) {
		key := canonicalLine(line)
		if key == "" {
			return
		}
		if _, exists := seenUser[key]; exists {
			return
		}
		seenUser[key] = struct{}{}
		routedUser = append(routedUser, bulletLine(line))
	}

	for _, line := range memoryLines {
		if isUserScopedLine(line, userID, userName) {
			addUser(line)
			continue
		}
		addMemory(line)
	}
	for _, line := range userLines {
		if isUserScopedLine(line, userID, userName) {
			addUser(line)
			continue
		}
		if isSharedLine(line, userID, userName) {
			addMemory(line)
			continue
		}
		addUser(line)
	}

	return strings.Join(routedMemory, "\n"), strings.Join(routedUser, "\n")
}

func validateBase(content string, limit int) error {
	if len(content) > limit {
		return fmt.Errorf("document exceeds limit: %d > %d", len(content), limit)
	}

	lower := strings.ToLower(content)
	for _, char := range invisibleChars {
		if strings.Contains(lower, char) {
			return fmt.Errorf("document contains invisible unicode character")
		}
	}

	for _, threat := range threatPatterns {
		if threat.re.MatchString(lower) {
			return fmt.Errorf("document matches blocked pattern %q", threat.label)
		}
	}

	return nil
}

func isLowValueMemory(content string) bool {
	lines := significantLines(content)
	if len(lines) == 0 {
		return false
	}

	flagged := 0
	for _, line := range lines {
		if containsAny(line, transientMemoryPhrases) || containsAny(line, genericProfilePhrases) {
			flagged++
		}
	}
	return flagged == len(lines)
}

func isGenericProfile(content string) bool {
	lines := significantLines(content)
	if len(lines) == 0 {
		return false
	}

	flagged := 0
	for _, line := range lines {
		if containsAny(line, genericProfilePhrases) {
			flagged++
		}
	}
	if flagged == 0 {
		return false
	}
	return flagged == len(lines)
}

func significantLines(content string) []string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = strings.TrimSpace(line[2:])
		}
		out = append(out, strings.ToLower(line))
	}
	return out
}

func containsAny(content string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(content, phrase) {
			return true
		}
	}
	return false
}

func documentLines(content string) []string {
	raw := strings.Split(content, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func detectUserName(userID string, userLines []string) string {
	for _, line := range userLines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "name is ") {
			name := strings.TrimSpace(line[strings.Index(lower, "name is ")+len("name is "):])
			return strings.TrimRight(name, ".")
		}
		if strings.Contains(lower, "user's name is ") {
			name := strings.TrimSpace(line[strings.Index(lower, "user's name is ")+len("user's name is "):])
			return strings.TrimRight(name, ".")
		}
	}
	return strings.TrimSpace(userID)
}

func isUserScopedLine(line, userID, userName string) bool {
	lower := canonicalLine(line)
	if lower == "" {
		return false
	}
	if containsAny(lower, userScopedPhrases) {
		return true
	}
	if userName != "" && strings.Contains(lower, strings.ToLower(userName)) {
		return true
	}
	if userID != "" && strings.Contains(lower, strings.ToLower(userID)) {
		return true
	}
	return false
}

func isSharedLine(line, userID, userName string) bool {
	lower := canonicalLine(line)
	if lower == "" {
		return false
	}
	if containsAny(lower, sharedMemoryPhrases) {
		return true
	}
	if userName != "" && !strings.Contains(lower, strings.ToLower(userName)) && containsAny(lower, []string{"reminder", "routine", "schedule"}) {
		return true
	}
	if userID != "" && !strings.Contains(lower, strings.ToLower(userID)) && containsAny(lower, []string{"reminder", "routine", "schedule"}) {
		return true
	}
	return false
}

func canonicalLine(line string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimSuffix(line, ".")))
}

func bulletLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	return "- " + line
}
