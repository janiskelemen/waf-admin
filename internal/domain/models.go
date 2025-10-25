package domain

type SiteInfo struct {
	Name        string   `json:"name"`
	SnippetPath string   `json:"snippetPath"`
	RulesPath   string   `json:"rulesPath"`
	RuleFiles   []string `json:"ruleFiles,omitempty"`
	HasSnippet  bool     `json:"hasSnippet"`
	HasRulesDir bool     `json:"hasRulesDir"`
}
