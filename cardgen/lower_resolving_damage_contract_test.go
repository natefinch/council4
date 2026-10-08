package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestResolvingSpellDamageRidersKeepSpellAttribution(t *testing.T) {
	t.Parallel()
	for _, card := range []ScryfallCard{
		{Name: "Burn the Accursed", Layout: "normal", TypeLine: "Instant", OracleText: "Burn the Accursed deals 5 damage to target creature and 2 damage to that creature's controller. If that creature would die this turn, exile it instead."},
		{Name: "Lava Coil", Layout: "normal", TypeLine: "Sorcery", OracleText: "Lava Coil deals 4 damage to target creature. If that creature would die this turn, exile it instead."},
	} {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			c, _ := compileTestOracle(card.OracleText, parser.Context{CardName: card.Name, InstantOrSorcery: true}, compiler.Context{})
			for _, ability := range c.Abilities {
				for _, ref := range ability.Content.References {
					if ref.Kind == compiler.ReferenceSelfName && ref.DamageAttribution() != compiler.DamageAttributionResolvingSource {
						t.Fatalf("printed spell subject lacks resolving-source proof: %+v", ref)
					}
				}
			}
			face := lowerSingleFace(t, &card)
			for _, instruction := range face.SpellAbility.Val.Modes[0].Sequence {
				if damage, ok := instruction.Primitive.(game.Damage); ok && damage.DamageSource.Exists {
					t.Fatalf("resolving spell acquired object attribution: %+v", damage)
				}
			}
		})
	}
}

func TestDestroyedTargetDamageAmountKeepsSeparateSpellSource(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{Name: "Agonizing Demise", Layout: "normal", TypeLine: "Instant",
		OracleText: "Kicker {1}{R}\nDestroy target nonblack creature. It can't be regenerated. If this spell was kicked, Agonizing Demise deals damage equal to that creature's power to the creature's controller."}
	face := lowerSingleFace(t, card)
	damage, ok := face.SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.Damage)
	if !ok || damage.DamageSource.Exists {
		t.Fatalf("damage must retain independent resolving spell attribution: %+v", damage)
	}
	assertCardPaths(t, card, "Kind = game.DynamicAmountObjectPower", "Object.kind = game.ObjectReferenceTargetPermanent",
		"Recipient.player.object.Val.kind = game.ObjectReferenceTargetPermanent")
}
