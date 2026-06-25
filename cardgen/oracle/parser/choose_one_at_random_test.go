package parser

import "testing"

// TestParseChapterChooseOneAtRandom verifies a Final Fantasy "Summon:" Saga
// chapter grouped as "I, II, III — Choose one at random —" parses as a chapter
// ability whose modal body uses the at-random choice kind with a one/one range
// and three flavor-named options.
func TestParseChapterChooseOneAtRandom(t *testing.T) {
	t.Parallel()
	source := "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
		"I, II, III — Choose one at random —\n" +
		"• Combine Powers! — Put three +1/+1 counters on target creature.\n" +
		"• Defense! — Put a shield counter on target creature. You gain 3 life.\n" +
		"• Fight! — This creature fights up to one target creature an opponent controls."
	document, diagnostics := Parse(source, Context{CardName: "Summon: Magus Sisters", Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}

	var chapter *Ability
	for i := range document.Abilities {
		if document.Abilities[i].Modal != nil {
			chapter = &document.Abilities[i]
			break
		}
	}
	if chapter == nil {
		t.Fatal("no modal chapter ability parsed")
	}
	if len(chapter.Chapters) != 3 {
		t.Fatalf("chapter numbers = %v, want three", chapter.Chapters)
	}
	if !chapter.Modal.ChoiceKnown || chapter.Modal.ChoiceKind != ModalChoiceKindOneAtRandom {
		t.Fatalf("choice = (known %v, kind %v), want known one-at-random",
			chapter.Modal.ChoiceKnown, chapter.Modal.ChoiceKind)
	}
	if chapter.Modal.MinModes != 1 || chapter.Modal.MaxModes != 1 {
		t.Fatalf("modes range = %d/%d, want 1/1", chapter.Modal.MinModes, chapter.Modal.MaxModes)
	}
	if len(chapter.Modal.Options) != 3 {
		t.Fatalf("options = %d, want 3", len(chapter.Modal.Options))
	}
}

// TestParseChooseOneAtRandomRequiresSagaChapter verifies the at-random header is
// only grouped onto a Saga chapter. A plain instant/sorcery modal that opens
// with "Choose one at random —" outside a chapter context still recognizes the
// at-random kind so the lowering layer can decide whether it is representable.
func TestParseChooseOneAtRandomKind(t *testing.T) {
	t.Parallel()
	source := "Choose one at random —\n" +
		"• Draw a card.\n" +
		"• You gain 3 life."
	document, _ := Parse(source, Context{CardName: "Test Random Modal", InstantOrSorcery: true})
	if len(document.Abilities) == 0 || document.Abilities[0].Modal == nil {
		t.Fatalf("abilities = %#v, want a modal ability", document.Abilities)
	}
	modal := document.Abilities[0].Modal
	if !modal.ChoiceKnown || modal.ChoiceKind != ModalChoiceKindOneAtRandom ||
		modal.MinModes != 1 || modal.MaxModes != 1 {
		t.Fatalf("choice = %d..%d kind=%v known=%v, want one-at-random 1..1",
			modal.MinModes, modal.MaxModes, modal.ChoiceKind, modal.ChoiceKnown)
	}
}
