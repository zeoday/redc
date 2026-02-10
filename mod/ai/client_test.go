package ai

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestChatStreamOpenAI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping test")
	}

	client := NewClient("openai", apiKey, "https://api.openai.com/v1", "gpt-3.5-turbo")

	messages := []Message{
		{Role: "user", Content: "Say hello in one word"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result string
	err := client.ChatStream(ctx, messages, func(chunk string) error {
		result += chunk
		t.Logf("Chunk: %s", chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	if result == "" {
		t.Fatal("No response received")
	}

	t.Logf("Full response: %s", result)
}

func TestChatStreamAnthropic(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	client := NewClient("anthropic", apiKey, "https://api.anthropic.com", "claude-3-haiku-20240307")

	messages := []Message{
		{Role: "user", Content: "Say hello in one word"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result string
	err := client.ChatStream(ctx, messages, func(chunk string) error {
		result += chunk
		t.Logf("Chunk: %s", chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	if result == "" {
		t.Fatal("No response received")
	}

	t.Logf("Full response: %s", result)
}
