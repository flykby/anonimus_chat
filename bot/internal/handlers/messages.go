package handlers

const StartMessage = "Привет! Я echo-бот — пришли текст, и я верну его обратно."

// EchoText returns the same text for echo replies.
func EchoText(text string) string {
	return text
}
