package ui

import "testing"

func TestSetLanguage(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	if GetLanguage() != LangEN {
		t.Errorf("Expected EN, got %s", GetLanguage())
	}

	SetLanguage(LangVI)
	if GetLanguage() != LangVI {
		t.Errorf("Expected VI, got %s", GetLanguage())
	}
}

func TestMsgEN(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	msg := Msg("checking_docker")
	expected := "Checking Docker..."
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}
}

func TestMsgVI(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangVI)
	msg := Msg("checking_docker")
	expected := "Đang kiểm tra Docker..."
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}
}

func TestMsgF(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	msg := MsgF("created", "test.yml")
	expected := "Created: test.yml"
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}

	SetLanguage(LangVI)
	msg = MsgF("created", "test.yml")
	expected = "Đã tạo: test.yml"
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}
}

func TestMsgFallback(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	// If a key doesn't exist, it should fallback to the key itself
	msg := Msg("nonexistent_key")
	expected := "nonexistent_key"
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}
}

func TestAllKeysMatch(t *testing.T) {
	// Verify messagesEN and messagesVI have the same keys
	for key := range messagesVI {
		if _, ok := messagesEN[key]; !ok {
			t.Errorf("Key %q missing in EN", key)
		}
	}
	for key := range messagesEN {
		if _, ok := messagesVI[key]; !ok {
			t.Errorf("Key %q missing in VI", key)
		}
	}
}

func TestDefaultLanguage(t *testing.T) {
	// Verify default language is English
	if currentLang != LangEN {
		t.Errorf("Expected default language to be EN, got %s", currentLang)
	}
}

func TestLanguageConstants(t *testing.T) {
	if LangEN != "en" {
		t.Errorf("Expected LangEN to be 'en', got %q", LangEN)
	}
	if LangVI != "vi" {
		t.Errorf("Expected LangVI to be 'vi', got %q", LangVI)
	}
}
