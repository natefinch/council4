package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

const veilOfSummerOracle = "Draw a card if an opponent has cast a blue or black spell this turn. " +
	"Spells you control can't be countered this turn. " +
	"You and permanents you control gain hexproof from blue and from black until end of turn. " +
	"(You and they can't be the targets of blue or black spells or abilities your opponents control.)"

func veilOfSummerCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Veil of Summer",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{G}",
		OracleText: veilOfSummerOracle,
	}
}

// TestLowerVeilOfSummer proves Veil of Summer's three clauses compose without a
// card-name special case: a conditional draw gated on an opponent's blue/black
// spell-cast this turn, a "spells you control can't be countered" rule effect,
// and a "you and permanents you control gain hexproof from blue and black"
// subject that splits into a player-scoped hexproof-from rule effect plus a
// group ability grant over the permanents you control.
func TestLowerVeilOfSummer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, veilOfSummerCard())

	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 4 {
		t.Fatalf("sequence length = %d, want 4: %+v", len(sequence), sequence)
	}

	// [0] draw a card gated on an opponent casting a blue/black spell this turn.
	draw, ok := sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("sequence[0] = %+v, want a draw", sequence[0])
	}
	_ = draw
	if !sequence[0].Condition.Exists {
		t.Fatal("draw must be gated on the event-history condition")
	}
	cond := sequence[0].Condition.Val.Condition
	if !cond.Exists || !cond.Val.EventHistory.Exists {
		t.Fatalf("draw condition = %+v, want an event-history condition", sequence[0].Condition)
	}
	history := cond.Val.EventHistory.Val
	if history.Window != game.EventHistoryCurrentTurn ||
		history.Pattern.Event != game.EventSpellCast ||
		history.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("event history = %+v, want opponent spell-cast this turn", history)
	}
	if colors := history.Pattern.CardSelection.ColorsAny; len(colors) != 2 ||
		colors[0] != color.Blue || colors[1] != color.Black {
		t.Fatalf("event colors = %+v, want [blue black]", history.Pattern.CardSelection.ColorsAny)
	}

	// [1] spells you control can't be countered this turn.
	cantCounter, ok := sequence[1].Primitive.(game.ApplyRule)
	if !ok || cantCounter.Duration != game.DurationThisTurn ||
		len(cantCounter.RuleEffects) != 1 ||
		cantCounter.RuleEffects[0].Kind != game.RuleEffectCantBeCountered ||
		cantCounter.RuleEffects[0].AffectedController != game.ControllerYou {
		t.Fatalf("sequence[1] = %+v, want you-control can't-be-countered this turn", sequence[1])
	}

	// [2] player hexproof-from blue and black for you, until end of turn.
	rule, ok := sequence[2].Primitive.(game.ApplyRule)
	if !ok || rule.Duration != game.DurationUntilEndOfTurn ||
		len(rule.RuleEffects) != 1 ||
		rule.RuleEffects[0].Kind != game.RuleEffectPlayerHexproof ||
		rule.RuleEffects[0].AffectedPlayer != game.PlayerYou {
		t.Fatalf("sequence[2] = %+v, want until-end-of-turn player hexproof for you", sequence[2])
	}
	if from := rule.RuleEffects[0].Protection.FromColors; len(from) != 2 ||
		from[0] != color.Blue || from[1] != color.Black {
		t.Fatalf("player hexproof-from colors = %+v, want [blue black]", rule.RuleEffects[0].Protection.FromColors)
	}

	// [3] group hexproof-from grant over the permanents you control.
	cont, ok := sequence[3].Primitive.(game.ApplyContinuous)
	if !ok || cont.Duration != game.DurationUntilEndOfTurn ||
		len(cont.ContinuousEffects) != 1 {
		t.Fatalf("sequence[3] = %+v, want until-end-of-turn continuous grant", sequence[3])
	}
	grant := cont.ContinuousEffects[0]
	if selection := grant.Group.Selection(); selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 0 {
		t.Fatalf("grant selection = %+v, want all permanents you control", grant.Group.Selection())
	}
	if len(grant.AddAbilities) != 1 {
		t.Fatalf("grant abilities = %+v, want one hexproof-from ability", grant.AddAbilities)
	}
	body, ok := grant.AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("granted ability = %T, want *game.StaticAbility", grant.AddAbilities[0])
	}
	hexproof, ok := game.StaticBodyHexproofFromKeyword(body)
	if !ok {
		t.Fatalf("granted ability = %#v, want a hexproof-from keyword", body)
	}
	if len(hexproof.FromColors) != 2 || hexproof.FromColors[0] != color.Blue ||
		hexproof.FromColors[1] != color.Black {
		t.Fatalf("granted hexproof-from colors = %+v, want [blue black]", hexproof.FromColors)
	}
}
