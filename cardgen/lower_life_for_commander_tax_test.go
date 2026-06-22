package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerLifeForCommanderTaxStatic proves that Liesa's command-zone tax
// alternative ("Rather than pay {2} for each previous time you've cast this
// spell from the command zone this game, pay 2 life that many times.") lowers to
// a command-zone-scoped static ability carrying the controller-scoped,
// self-targeting RuleEffectPayLifeForCommanderTax.
func TestLowerLifeForCommanderTaxStatic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Liesa",
		Layout:     "normal",
		ManaCost:   "{2}{W}{W}{B}",
		TypeLine:   "Legendary Creature — Angel",
		OracleText: "Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times.",
		Power:      new("5"),
		Toughness:  new("5"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	body := face.StaticAbilities[0].Body
	if body.ZoneOfFunction != zone.Command {
		t.Fatalf("zone of function = %v, want zone.Command", body.ZoneOfFunction)
	}
	if len(body.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", body.RuleEffects)
	}
	effect := body.RuleEffects[0]
	if effect.Kind != game.RuleEffectPayLifeForCommanderTax {
		t.Fatalf("kind = %v, want RuleEffectPayLifeForCommanderTax", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if !effect.AffectedSource {
		t.Fatal("affected source = false, want true")
	}
}

// TestLowerLiesaFullCard proves the whole Liesa, Shroud of Dusk card lowers with
// zero diagnostics: the command-zone tax static, the Flying and lifelink keyword
// statics, and the symmetric spell-cast life-loss trigger.
func TestLowerLiesaFullCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Liesa, Shroud of Dusk",
		Layout:     "normal",
		ManaCost:   "{2}{W}{W}{B}",
		TypeLine:   "Legendary Creature — Angel",
		OracleText: "Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times.\nFlying, lifelink\nWhenever a player casts a spell, they lose 2 life.",
		Power:      new("5"),
		Toughness:  new("5"),
	})
	if len(face.StaticAbilities) != 3 {
		t.Fatalf("static abilities = %d, want three (tax, flying, lifelink)", len(face.StaticAbilities))
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventSpellCast {
		t.Fatalf("trigger event = %v, want EventSpellCast", trigger.Trigger.Pattern.Event)
	}
	loseLife, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", trigger.Content.Modes[0].Sequence[0].Primitive)
	}
	if loseLife.Player != game.EventPlayerReference() {
		t.Fatalf("lose-life player = %#v, want EventPlayerReference", loseLife.Player)
	}
}
