package factoryaidroid

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entireio/cli/cmd/entire/cli/agent"
)

func TestNewFactoryAIDroidAgent(t *testing.T) {
	t.Parallel()
	ag := NewFactoryAIDroidAgent()
	if ag == nil {
		t.Fatal("NewFactoryAIDroidAgent() returned nil")
	}
	if _, ok := ag.(*FactoryAIDroidAgent); !ok {
		t.Fatal("NewFactoryAIDroidAgent() didn't return *FactoryAIDroidAgent")
	}
}

// TestDetectPresence uses t.Chdir so it cannot be parallel.
func TestDetectPresence(t *testing.T) {
	t.Run("factory directory exists", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		if err := os.Mkdir(".factory", 0o755); err != nil {
			t.Fatalf("failed to create .factory: %v", err)
		}

		ag := &FactoryAIDroidAgent{}
		present, err := ag.DetectPresence()
		if err != nil {
			t.Fatalf("DetectPresence() error = %v", err)
		}
		if !present {
			t.Error("DetectPresence() = false, want true")
		}
	})

	t.Run("no factory directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		ag := &FactoryAIDroidAgent{}
		present, err := ag.DetectPresence()
		if err != nil {
			t.Fatalf("DetectPresence() error = %v", err)
		}
		if present {
			t.Error("DetectPresence() = true, want false")
		}
	})
}

// --- Transcript tests ---

func TestReadTranscript(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "transcript.jsonl")
	content := `{"role":"user","content":"hello"}
{"role":"assistant","content":"hi"}`
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ag := &FactoryAIDroidAgent{}
	data, err := ag.ReadTranscript(file)
	if err != nil {
		t.Fatalf("ReadTranscript() error = %v", err)
	}
	if string(data) != content {
		t.Errorf("ReadTranscript() = %q, want %q", string(data), content)
	}
}

func TestReadTranscript_MissingFile(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}
	_, err := ag.ReadTranscript("/nonexistent/path/transcript.jsonl")
	if err == nil {
		t.Error("ReadTranscript() should error on missing file")
	}
}

func TestChunkTranscript_LargeContent(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}

	// Build multi-line JSONL that exceeds a small maxSize
	var lines []string
	for i := range 50 {
		lines = append(lines, fmt.Sprintf(`{"role":"user","content":"message %d %s"}`, i, strings.Repeat("x", 200)))
	}
	content := []byte(strings.Join(lines, "\n"))

	maxSize := 2000
	chunks, err := ag.ChunkTranscript(content, maxSize)
	if err != nil {
		t.Fatalf("ChunkTranscript() error = %v", err)
	}
	if len(chunks) < 2 {
		t.Errorf("Expected at least 2 chunks for large content, got %d", len(chunks))
	}

	// Verify each chunk is valid JSONL (each line is valid JSON)
	for i, chunk := range chunks {
		chunkLines := strings.Split(string(chunk), "\n")
		for j, line := range chunkLines {
			if line == "" {
				continue
			}
			if line[0] != '{' {
				t.Errorf("Chunk %d, line %d doesn't look like JSON: %q", i, j, line[:min(len(line), 40)])
			}
		}
	}
}

func TestChunkTranscript_RoundTrip(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}

	original := `{"role":"user","content":"hello"}
{"role":"assistant","content":"hi there"}
{"role":"user","content":"thanks"}`

	chunks, err := ag.ChunkTranscript([]byte(original), 60)
	if err != nil {
		t.Fatalf("ChunkTranscript() error = %v", err)
	}

	reassembled, err := ag.ReassembleTranscript(chunks)
	if err != nil {
		t.Fatalf("ReassembleTranscript() error = %v", err)
	}

	if string(reassembled) != original {
		t.Errorf("Round-trip mismatch:\n got: %q\nwant: %q", string(reassembled), original)
	}
}

// --- ParseHookInput tests ---

func TestParseHookInput_Valid(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}
	input := `{"session_id":"sess-abc","transcript_path":"/tmp/transcript.jsonl"}`

	result, err := ag.ParseHookInput(agent.HookSessionStart, strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}
	if result.SessionID != "sess-abc" {
		t.Errorf("SessionID = %q, want %q", result.SessionID, "sess-abc")
	}
	if result.SessionRef != "/tmp/transcript.jsonl" {
		t.Errorf("SessionRef = %q, want %q", result.SessionRef, "/tmp/transcript.jsonl")
	}
}

func TestParseHookInput_Empty(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}
	_, err := ag.ParseHookInput(agent.HookSessionStart, strings.NewReader(""))
	if err == nil {
		t.Error("ParseHookInput() should error on empty input")
	}
}

func TestParseHookInput_InvalidJSON(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}
	_, err := ag.ParseHookInput(agent.HookSessionStart, strings.NewReader("not json"))
	if err == nil {
		t.Error("ParseHookInput() should error on invalid JSON")
	}
}

func TestGetSessionDir(t *testing.T) {
	t.Parallel()
	ag := &FactoryAIDroidAgent{}

	dir, err := ag.GetSessionDir("/Users/alisha/Projects/test-repos/factoryai-droid")
	if err != nil {
		t.Fatalf("GetSessionDir() error = %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	expected := filepath.Join(homeDir, ".factory", "sessions", "-Users-alisha-Projects-test-repos-factoryai-droid")
	if dir != expected {
		t.Errorf("GetSessionDir() = %q, want %q", dir, expected)
	}
}

// TestGetSessionDir_EnvOverride cannot use t.Parallel() due to t.Setenv.
func TestGetSessionDir_EnvOverride(t *testing.T) {
	ag := &FactoryAIDroidAgent{}
	override := "/tmp/test-droid-sessions"
	t.Setenv("ENTIRE_TEST_DROID_PROJECT_DIR", override)

	dir, err := ag.GetSessionDir("/any/repo/path")
	if err != nil {
		t.Fatalf("GetSessionDir() error = %v", err)
	}
	if dir != override {
		t.Errorf("GetSessionDir() = %q, want %q (env override)", dir, override)
	}
}
