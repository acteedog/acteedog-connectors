package fetch

import "time"

// filterEventsByDate filters events by the target date range
func filterEventsByDate(events []map[string]any, startTime, endTime time.Time) ([]map[string]any, bool) {
	filtered := []map[string]any{}
	shouldStop := false

	for _, event := range events {
		createdAtStr, ok := event["created_at"].(string)
		if !ok {
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			continue
		}

		if createdAt.After(startTime) && createdAt.Before(endTime) {
			filtered = append(filtered, event)
		}

		if createdAt.Before(startTime) {
			shouldStop = true
			break
		}
	}

	return filtered, shouldStop
}

// filterEventsByRepository filters events by repository patterns
func filterEventsByRepository(events []map[string]any, patterns []string) []map[string]any {
	if len(patterns) == 0 {
		return events
	}

	filtered := []map[string]any{}
	for _, event := range events {
		repo, ok := event["repo"].(map[string]any)
		if !ok {
			continue
		}

		repoName, ok := repo["name"].(string)
		if !ok {
			continue
		}

		if matchesAnyPattern(repoName, patterns) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// matchesAnyPattern checks if a repository name matches any of the patterns
func matchesAnyPattern(repoName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(repoName, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if a string matches a pattern with * wildcards
// Supports patterns like:
// - "owner/*" (matches all repos in owner)
// - "owner/repo" (exact match)
// - "owner/prefix-*" (matches repos starting with prefix-)
// - "*/*" (matches all repos)
func matchPattern(str, pattern string) bool {
	// Convert pattern to regex-like matching
	// Split both by '/' to handle owner/repo structure
	strParts := splitRepo(str)
	patternParts := splitRepo(pattern)

	// Must have same structure (owner/repo)
	if len(strParts) != len(patternParts) {
		return false
	}

	// Check each part
	for i := range strParts {
		if !matchWildcard(strParts[i], patternParts[i]) {
			return false
		}
	}

	return true
}

// splitRepo splits repository name by '/'
func splitRepo(repo string) []string {
	parts := []string{}
	current := ""
	for _, ch := range repo {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// matchWildcard checks if a string matches a pattern with * wildcard
// * matches one or more characters (not empty string)
func matchWildcard(str, pattern string) bool {
	// Exact match
	if pattern == str {
		return true
	}

	// No wildcard
	if !containsWildcard(pattern) {
		return false
	}

	// Handle wildcard matching
	if pattern == "*" {
		return len(str) > 0
	}

	// Pattern with wildcard at the end (e.g., "prefix-*")
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(str) >= len(prefix) && str[:len(prefix)] == prefix
	}

	// Pattern with wildcard at the start (e.g., "*-suffix")
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix
	}

	// Pattern with wildcard in the middle (e.g., "pre*fix")
	// Find the position of '*'
	starPos := -1
	for i, ch := range pattern {
		if ch == '*' {
			starPos = i
			break
		}
	}

	if starPos >= 0 {
		prefix := pattern[:starPos]
		suffix := pattern[starPos+1:]

		if len(str) < len(prefix)+len(suffix) {
			return false
		}

		return str[:len(prefix)] == prefix && str[len(str)-len(suffix):] == suffix
	}

	return false
}

// containsWildcard checks if a string contains '*'
func containsWildcard(s string) bool {
	for _, ch := range s {
		if ch == '*' {
			return true
		}
	}
	return false
}
