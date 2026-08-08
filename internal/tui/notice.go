package tui

type noticeKind int

const (
	noticeInfo noticeKind = iota
	noticeError
)

type statusNotice struct {
	text string
	kind noticeKind
}

func infoNotice(text string) statusNotice {
	return statusNotice{text: text, kind: noticeInfo}
}

func errorNotice(text string) statusNotice {
	return statusNotice{text: text, kind: noticeError}
}

func (n statusNotice) empty() bool {
	return n.text == ""
}

func (n statusNotice) render(theme Theme) string {
	if n.empty() {
		return ""
	}
	if n.kind == noticeError {
		return theme.Incorrect.Render(n.text)
	}
	return theme.Help.Render(n.text)
}
