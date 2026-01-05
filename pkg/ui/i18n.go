package ui

import "fmt"

// Language represents supported languages
type Language string

const (
	LangEN Language = "en"
	LangVI Language = "vi"
)

// currentLang is the current active language
// Default: English (changed from Vietnamese per plan validation)
var currentLang = LangEN

// SetLanguage sets the current language
func SetLanguage(lang Language) {
	currentLang = lang
}

// GetLanguage returns the current language
func GetLanguage() Language {
	return currentLang
}

// Msg returns the localized message for the given key
func Msg(key string) string {
	var messages map[string]string
	switch currentLang {
	case LangEN:
		messages = messagesEN
	case LangVI:
		messages = messagesVI
	default:
		messages = messagesEN
	}

	if msg, ok := messages[key]; ok {
		return msg
	}
	// Fallback to English if key not found
	if msg, ok := messagesEN[key]; ok {
		return msg
	}
	return key // Return key itself as last resort
}

// MsgF returns the localized message with format arguments
func MsgF(key string, args ...interface{}) string {
	return fmt.Sprintf(Msg(key), args...)
}
