package match

import (
	"github-connector/internal/core"
	"regexp"
)

var (
	rePullRequest = regexp.MustCompile(core.ContextPatternPullRequest)
	reIssue       = regexp.MustCompile(core.ContextPatternIssue)
	reRepository  = regexp.MustCompile(core.ContextPatternRepository)
	reExcludeRepo = regexp.MustCompile(core.ContextExcludePatternRepository)
)

// MatchURL returns the context hierarchy for a single URL, or an empty slice if no pattern matches.
// No external API calls are made; context hierarchy is constructed from URL captures alone.
func MatchURL(gen *core.ContextGenerator, url string) []*core.Context {
	// Pull Request pattern (checked before Repository to avoid partial match)
	if m := namedCaptures(rePullRequest, url); m != nil {
		repoName := m["owner"] + "/" + m["repo"]
		prNum := parseInt(m["number"])
		return []*core.Context{
			gen.CreateSourceContext(),
			gen.CreateRepositoryContext(repoName),
			gen.CreatePRContext(repoName, prNum),
		}
	}

	// Issue pattern (checked before Repository to avoid partial match)
	if m := namedCaptures(reIssue, url); m != nil {
		repoName := m["owner"] + "/" + m["repo"]
		issueNum := parseInt(m["number"])
		return []*core.Context{
			gen.CreateSourceContext(),
			gen.CreateRepositoryContext(repoName),
			gen.CreateIssueContext(repoName, issueNum),
		}
	}

	// Repository pattern (with exclusion check)
	if reExcludeRepo.MatchString(url) {
		return []*core.Context{}
	}
	if m := namedCaptures(reRepository, url); m != nil {
		repoName := m["owner"] + "/" + m["repo"]
		return []*core.Context{
			gen.CreateSourceContext(),
			gen.CreateRepositoryContext(repoName),
		}
	}

	return []*core.Context{}
}

// namedCaptures returns a map of named capture groups for the first match, or nil if no match.
func namedCaptures(re *regexp.Regexp, s string) map[string]string {
	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil
	}
	result := make(map[string]string, len(re.SubexpNames()))
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// parseInt converts a decimal string to int, returning 0 on any non-digit character.
func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
