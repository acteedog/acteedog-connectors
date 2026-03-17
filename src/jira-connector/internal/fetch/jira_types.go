package fetch

// JiraSearchResponse represents the response from the Jira search/jql API
type JiraSearchResponse struct {
	Issues        []JiraIssue `json:"issues"`
	IsLast        bool        `json:"isLast"`
	NextPageToken string      `json:"nextPageToken"`
}

// JiraIssue represents a single Jira issue
type JiraIssue struct {
	ID        string        `json:"id"`
	Key       string        `json:"key"`
	Fields    JiraFields    `json:"fields"`
	Changelog JiraChangelog `json:"changelog"`
}

// JiraFields represents the fields of a Jira issue
type JiraFields struct {
	Summary   string            `json:"summary"`
	Created   string            `json:"created"`
	Updated   string            `json:"updated"`
	Creator   *JiraUser         `json:"creator"`
	Comment   *JiraCommentField `json:"comment"`
	Project   *JiraProjectRef   `json:"project"`
	Parent    *JiraIssueRef     `json:"parent"`
	IssueType *JiraIssueType    `json:"issuetype"`
	Status    *JiraStatus       `json:"status"`
	Priority  *JiraPriority     `json:"priority"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	EmailAddress string `json:"emailAddress"`
}

// JiraCommentField represents the comment field in a Jira issue
type JiraCommentField struct {
	Comments []JiraComment `json:"comments"`
}

// JiraComment represents a single comment on a Jira issue
type JiraComment struct {
	ID      string    `json:"id"`
	Author  *JiraUser `json:"author"`
	Body    *ADFNode  `json:"body"`
	Created string    `json:"created"`
	Updated string    `json:"updated"`
}

// JiraProjectRef represents a project reference in an issue's fields
type JiraProjectRef struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraIssueRef represents a parent issue reference in an issue's fields
type JiraIssueRef struct {
	ID     string               `json:"id"`
	Key    string               `json:"key"`
	Fields *JiraIssueRef_Fields `json:"fields"`
}

// JiraIssueRef_Fields contains minimal fields for a referenced issue
type JiraIssueRef_Fields struct {
	Summary   string         `json:"summary"`
	IssueType *JiraIssueType `json:"issuetype"`
}

// JiraIssueType represents the issue type
type JiraIssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraStatus represents the status of an issue
type JiraStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraPriority represents the priority of an issue
type JiraPriority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraChangelog represents the changelog of a Jira issue
type JiraChangelog struct {
	Histories []JiraHistory `json:"histories"`
}

// JiraHistory represents a single changelog history entry
type JiraHistory struct {
	ID      string            `json:"id"`
	Author  *JiraUser         `json:"author"`
	Created string            `json:"created"`
	Items   []JiraHistoryItem `json:"items"`
}

// JiraHistoryItem represents a single item in a changelog history entry
type JiraHistoryItem struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}

// ADFNode represents a node in Atlassian Document Format
type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []ADFNode `json:"content,omitempty"`
}

// PlainText recursively extracts plain text from an ADF node
func (n *ADFNode) PlainText() string {
	if n == nil {
		return ""
	}
	if n.Type == "text" {
		return n.Text
	}
	result := ""
	for i := range n.Content {
		text := n.Content[i].PlainText()
		if text != "" {
			if result != "" {
				result += " "
			}
			result += text
		}
	}
	return result
}
