package mysql

import "testing"

func TestSanitizeDSN(t *testing.T) {
	got := sanitizeDSN("user:secret@tcp(127.0.0.1:3306)/app")
	if got == "" || got == "user:secret@tcp(127.0.0.1:3306)/app" {
		t.Fatalf("dsn not sanitized: %q", got)
	}
}

func TestKind(t *testing.T) {
	if New("x").Kind() != "mysql" {
		t.Fatal("kind")
	}
}
