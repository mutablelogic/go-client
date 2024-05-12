package schema

// A chat completion message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// One choice of chat completion messages
type MessageChoice struct {
	Message      `json:"message"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type Content struct {
	Type string `json:"type,width:4"`
	Text string `json:"text,wrap,width:60"`
}
