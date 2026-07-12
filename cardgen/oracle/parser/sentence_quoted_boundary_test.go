package parser

import "testing"

func TestParseSentencesSplitsAfterTerminalQuotedAbility(t *testing.T) {
	t.Parallel()
	source := "Create X 1/1 black Rat creature tokens with \"This token can't block.\" Creatures you control gain haste until end of turn."
	tokens, diagnostics := lexAll(source)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	sentences := ParseSentences(source, tokens[:len(tokens)-1])
	if len(sentences) != 2 {
		t.Fatalf("sentences = %#v, want 2", sentences)
	}
	if sentences[0].Text != "Create X 1/1 black Rat creature tokens with \"This token can't block.\"" ||
		sentences[1].Text != "Creatures you control gain haste until end of turn." {
		t.Fatalf("sentence texts = %q, %q", sentences[0].Text, sentences[1].Text)
	}
}

func TestParseSentencesKeepsLowercaseQuotedContinuation(t *testing.T) {
	t.Parallel()
	source := "You may say \"Ach! Hans, run! It's the . . .\" and the name of a creature card. Draw a card."
	tokens, diagnostics := lexAll(source)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	sentences := ParseSentences(source, tokens[:len(tokens)-1])
	if len(sentences) != 2 {
		t.Fatalf("sentences = %#v, want 2", sentences)
	}
	if sentences[0].Text != "You may say \"Ach! Hans, run! It's the . . .\" and the name of a creature card." {
		t.Fatalf("first sentence = %q", sentences[0].Text)
	}
}
