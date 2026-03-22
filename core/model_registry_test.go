package core

import "testing"

func TestEmbeddingModelRegistry(t *testing.T) {
	for name, definition := range embeddingModelRegistry {
		if definition.EmbeddingDim <= 0 {
			t.Fatalf("%s embedding dim must be positive", name)
		}
		if definition.ModelFile == "" || definition.TokenizerFile == "" {
			t.Fatalf("%s must define model and tokenizer files", name)
		}
		if len(definition.Assets) == 0 {
			t.Fatalf("%s must define downloadable assets", name)
		}
	}
}

func TestSummaryModelRegistry(t *testing.T) {
	for name, definition := range summaryModelRegistry {
		if definition.EncoderFile == "" || definition.DecoderFile == "" || definition.TokenizerFile == "" {
			t.Fatalf("%s must define encoder, decoder, and tokenizer files", name)
		}
		if definition.MaxInputTokens <= 0 || definition.MaxOutputTokens <= 0 {
			t.Fatalf("%s must define token limits", name)
		}
		if len(definition.Assets) == 0 {
			t.Fatalf("%s must define downloadable assets", name)
		}
	}
}
