package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseDamageThenOptionalXBoundedHandFreeCast(t *testing.T) {
	t.Parallel()
	const oracle = "Electrodominance deals X damage to any target. You may cast a spell with mana value X or less from your hand without paying its mana cost."
	document, diagnostics := Parse(oracle, Context{InstantOrSorcery: true, CardName: "Electrodominance"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("document shape = %#v", document.Abilities)
	}
	damage := document.Abilities[0].Sentences[0].Effects[0]
	if damage.Kind != EffectDealDamage || !damage.Amount.VariableX || len(damage.Targets) != 1 {
		t.Fatalf("damage effect = %#v", damage)
	}
	cast := document.Abilities[0].Sentences[1].Effects[0]
	if cast.Kind != EffectCast ||
		!cast.Optional ||
		!cast.CastWithoutPayingManaCost ||
		cast.FromZone != zone.Hand ||
		cast.Selection.Kind != SelectionSpell ||
		!cast.Selection.MatchManaValue ||
		!cast.Selection.ManaValueX ||
		cast.Selection.ManaValue.Op != compare.LessOrEqual {
		t.Fatalf("cast effect = %#v", cast)
	}
}
