package common

type Captcha struct {
	Token       string   `json:"token"`
	Type        string   `json:"type"`
	OriImage    string   `json:"ori_image"`
	JigsawImage string   `json:"jigsaw_image"`
	ClickWords  []string `json:"click_words"`
}

type RespMsg struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
