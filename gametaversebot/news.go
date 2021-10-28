package gametaversebot

type News struct {
	ID          int           `json:"id"`
	ProjectID   string        `json:"project_id"`
	ProjectName string        `json:"project_name"`
	Title       string        `json:"title"`
	TitleCn     string        `json:"title_cn"`
	Content     string        `json:"content"`
	ContentCn   string        `json:"content_cn"`
	Photourl    string        `json:"photoUrl"`
	Sourceurl   string        `json:"sourceUrl"`
	Languange   interface{}   `json:"languange"`
	StartAt     string        `json:"start_at"`
	StartAtMd   string        `json:"start_at_md"`
	StartAtHm   string        `json:"start_at_hm"`
	Tags        []interface{} `json:"tags"`
	Publishdate string        `json:"publishDate"`
}
