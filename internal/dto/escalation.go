package dto

type (
	EscalationRuleRecipientDto struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	}

	EscalationRuleDto struct {
		Condition  string                     `json:"condition"`
		NotifyType string                     `json:"notifyType"`
		Delay      int64                      `json:"delay"`
		Recipient  EscalationRuleRecipientDto `json:"recipient"`
	}

	EscalationRepeatDto struct {
		WaitInterval         int32 `json:"waitInterval"`
		Count                int32 `json:"count"`
		ResetRecipientStates bool  `json:"resetRecipientStates"`
		CloseAlertAfterAll   bool  `json:"closeAlertAfterAll"`
	}

	EscalationDto struct {
		Id          string               `json:"id"`
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Rules       []EscalationRuleDto  `json:"rules"`
		Enabled     bool                 `json:"enabled"`
		Repeat      *EscalationRepeatDto `json:"repeat"`
	}

	ListEscalationDto struct {
		Values []EscalationDto `json:"values"`
		Links  LinksDto        `json:"links"`
	}
)
