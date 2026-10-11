package cardgen

import "testing"

// TestGenerateExecutableCardSourceSelfInflictedWound covers the affirmative
// non-controller edict → event-player consequence gate combined with a
// color-disjunction sacrifice selection. The edicted opponent sacrifices "a
// green or white creature of their choice" (a Selection.ColorsAny filter), the
// edict publishes whether it happened, and "they lose 2 life" resolves for the
// targeted player only when the sacrifice succeeded (TriTrue). It exercises two
// pieces working together: the color-disjunction sacrifice noun and the
// resolving-gate consequence whose "they" reference must survive alongside the
// gate's own "that player" anaphor without failing closed.
func TestGenerateExecutableCardSourceSelfInflictedWound(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Self-Inflicted Wound",
		Layout:     "normal",
		ManaCost:   "{B}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target opponent sacrifices a green or white creature of their choice. If that player does, they lose 2 life.",
	}
	assertCardPaths(t, card,
		"Sequence[0].Primitive.(game.SacrificePermanents).Amount.fixed = 1",
		"Sequence[0].Primitive.(game.SacrificePermanents).Player.kind = game.PlayerReferenceTargetPlayer",
		"Sequence[0].Primitive.(game.SacrificePermanents).Selection.RequiredTypes[0] = types.Creature",
		"Sequence[0].Primitive.(game.SacrificePermanents).Selection.ColorsAny[0] = color.Green",
		"Sequence[0].Primitive.(game.SacrificePermanents).Selection.ColorsAny[1] = color.White",
		"Sequence[0].PublishResult = \"if-you-do\"",
		"Sequence[1].Primitive.(game.LoseLife).Amount.fixed = 2",
		"Sequence[1].Primitive.(game.LoseLife).Player.kind = game.PlayerReferenceTargetPlayer",
		"Sequence[1].ResultGate.Val.Key = \"if-you-do\"",
		"Sequence[1].ResultGate.Val.Succeeded = game.TriTrue")
	assertCardPathsAbsent(t, card, "Sequence[1].Primitive.(game.LoseLife).Player.kind = game.PlayerReferenceEventPlayer")
}
