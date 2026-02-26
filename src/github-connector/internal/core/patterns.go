package core

const (
	ContextPatternPullRequest = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/pull/(?P<number>\d+)`
	ContextPatternIssue       = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/issues/(?P<number>\d+)`
	ContextPatternRepository  = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/|>)\]"'?]+)/?`

	// ContextExcludePatternRepository excludes GitHub special paths that are not repositories
	// (e.g., user-attachments asset URLs like https://github.com/user-attachments/assets/...)
	ContextExcludePatternRepository = `https://github\.com/user-attachments/`
)
