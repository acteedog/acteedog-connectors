package enrich

// JiraProjectResponse represents the response from the Jira project API
type JiraProjectResponse struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ProjectTypeKey string `json:"projectTypeKey"`
}

// JiraIssueResponse represents the response from the Jira issue API
type JiraIssueResponse struct {
	ID     string          `json:"id"`
	Key    string          `json:"key"`
	Fields JiraIssueFields `json:"fields"`
}

// JiraIssueFields represents the fields of a Jira issue in enrichment context
type JiraIssueFields struct {
	Summary     string        `json:"summary"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
	Description *ADFNode      `json:"description"`
	IssueType   *JiraNamedRef `json:"issuetype"`
	Status      *JiraNamedRef `json:"status"`
	Priority    *JiraNamedRef `json:"priority"`
	Creator     *JiraUser     `json:"creator"`
}

// JiraNamedRef represents a Jira resource with an ID and name
type JiraNamedRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
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
