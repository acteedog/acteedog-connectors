package core

const (
	ContextPatternPullRequest = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/pull/(?P<number>\d+)`
	ContextPatternIssue       = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/issues/(?P<number>\d+)`
	ContextPatternRepository  = `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/|>)\]"'?]+)/?`
)
