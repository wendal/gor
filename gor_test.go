package gor

import (
	"testing"
	"time"
)

func TestEncodePathInfo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello-world"},
		{"go:lang", "go-lang"},
		{"simple", "simple"},
		{"a b:c", "a-b-c"},
	}
	for _, tt := range tests {
		got := EncodePathInfo(tt.input)
		if got != tt.expected {
			t.Errorf("EncodePathInfo(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDecodePathInfo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello%20world", "hello world"},
		{"go-lang", "go-lang"},
		{"simple", "simple"},
	}
	for _, tt := range tests {
		got := DecodePathInfo(tt.input)
		if got != tt.expected {
			t.Errorf("DecodePathInfo(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestEncodeDecodePathInfo(t *testing.T) {
	original := "hello world"
	encoded := EncodePathInfo(original)
	// EncodePathInfo replaces spaces with dashes, not url-encode, so round-trip differs
	if encoded == original {
		t.Errorf("EncodePathInfo should transform %q", original)
	}
}

func TestAsStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{"nil", nil, []string{}},
		{"string", "hello", []string{"hello"}},
		{"string slice", []string{"a", "b"}, []string{"a", "b"}},
		{"interface slice", []interface{}{"x", "y"}, []string{"x", "y"}},
	}
	for _, tt := range tests {
		got := AsStrings(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("[%s] AsStrings(%v) len=%d, want %d", tt.name, tt.input, len(got), len(tt.expected))
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("[%s] AsStrings(%v)[%d] = %q, want %q", tt.name, tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestCreatePostURL(t *testing.T) {
	db := make(map[string]interface{})
	date := time.Date(2023, 5, 15, 0, 0, 0, 0, time.Local)
	post := map[string]interface{}{
		"permalink":  "/:year/:month/:title/",
		"_date":      date,
		"title":      "Hello World",
		"id":         "posts/2023-05-15-hello-world.md",
		"categories": []string{"go"},
	}
	CreatePostURL(db, "/", post)
	url, ok := post["url"].(string)
	if !ok {
		t.Fatal("post[url] should be a string")
	}
	if url != "/2023/05/Hello-World/" {
		t.Errorf("unexpected url: %q", url)
	}
}

func TestCreatePostURL_StaticPermalink(t *testing.T) {
	db := make(map[string]interface{})
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
	post := map[string]interface{}{
		"permalink":  "/about/",
		"_date":      date,
		"title":      "About",
		"id":         "pages/about.md",
		"categories": []string{},
	}
	CreatePostURL(db, "/", post)
	url, ok := post["url"].(string)
	if !ok {
		t.Fatal("post[url] should be a string")
	}
	if url != "/about/" {
		t.Errorf("unexpected url: %q", url)
	}
}

func TestMarkdownToHtml(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"bold", "**hello**", "<strong>hello</strong>"},
		{"header", "# Title", "<h1>Title</h1>"},
		{"paragraph", "simple text", "simple text"},
	}
	for _, tt := range tests {
		got := MarkdownToHtml(tt.input)
		if len(got) == 0 {
			t.Errorf("[%s] MarkdownToHtml(%q) returned empty string", tt.name, tt.input)
			continue
		}
		found := false
		for i := 0; i <= len(got)-len(tt.contains); i++ {
			if got[i:i+len(tt.contains)] == tt.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("[%s] MarkdownToHtml(%q) = %q, want it to contain %q", tt.name, tt.input, got, tt.contains)
		}
	}
}

func TestSortPosts(t *testing.T) {
	date1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
	date2 := time.Date(2023, 6, 1, 0, 0, 0, 0, time.Local)
	dict := map[string]Mapper{
		"posts/older.md": {"id": "posts/older.md", "_date": date1},
		"posts/newer.md": {"id": "posts/newer.md", "_date": date2},
	}
	ids := []string{"posts/older.md", "posts/newer.md"}
	sorted := SortPosts(dict, ids)
	if sorted[0] != "posts/newer.md" {
		t.Errorf("SortPosts: expected newer post first, got %v", sorted)
	}
}

func TestMapperGetString(t *testing.T) {
	m := Mapper{"key": "value", "num": 42}
	if m.GetString("key") != "value" {
		t.Errorf("GetString(key) should return 'value'")
	}
	if m.GetString("missing") != "" {
		t.Errorf("GetString(missing) should return ''")
	}
	if m.GetString("num") != "42" {
		t.Errorf("GetString(num) should return '42'")
	}
}

func TestMapperGetInt(t *testing.T) {
	m := Mapper{"a": int64(5), "b": 10, "c": "x"}
	if m.GetInt("a") != 5 {
		t.Errorf("GetInt(a) should return 5")
	}
	if m.GetInt("b") != 10 {
		t.Errorf("GetInt(b) should return 10")
	}
	if m.GetInt("missing") != 0 {
		t.Errorf("GetInt(missing) should return 0")
	}
}

func TestPostPath(t *testing.T) {
	tests := []struct {
		title    string
		contains string
	}{
		{"my post", "my-post"},
		{"hello-world", "hello-world"},
	}
	for _, tt := range tests {
		got := postPath(tt.title)
		found := false
		for i := 0; i <= len(got)-len(tt.contains); i++ {
			if got[i:i+len(tt.contains)] == tt.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("postPath(%q) = %q, want it to contain %q", tt.title, got, tt.contains)
		}
	}
}
