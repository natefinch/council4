package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCaptureReviewDirectTargetRetainsOwnSlot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Direct Capture", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Tap target creature. Destroy target creature at end of combat.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	delayed := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger).Trigger
	if len(mode.Targets) != 2 || delayed.CapturedObject.Val != game.TargetPermanentReference(1) {
		t.Fatal("direct delayed target lost its own selection or captured an earlier target")
	}
}

func TestCaptureReviewUnavailableQuantityAndPluralOwnership(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Tap target creature. Put X +1/+1 counters on it at the beginning of the next end step.",
		"Create a 1/1 green Insect creature token. Create a Treasure token. Sacrifice them at the beginning of the next end step.",
		"Create a 1/1 green Insect creature token. Create a Treasure token. Sacrifice the tokens at the beginning of the next end step.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Unsupported Capture", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{X}", OracleText: text})
		})
	}
}

func TestCaptureReviewReturnedCardPreservesEnchantment(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Enchanting Return", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Exile target creature. Return that card to the battlefield at the beginning of the next end step. It's an enchantment. (It's not a creature.)",
	})
	delayed := face.SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.CreateDelayedTrigger).Trigger
	put := delayed.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if len(put.ContinuousEffects) != 1 || put.ContinuousEffects[0].Layer != game.LayerType ||
		len(put.ContinuousEffects[0].SetTypes) != 1 || put.ContinuousEffects[0].SetTypes[0] != types.Enchantment {
		t.Fatal("returned card silently discarded its type-setting entry rider")
	}
}

func TestCaptureReviewDepartedCardUsesCapturedOperand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Departed Capture", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever another creature you control leaves the battlefield, return that card to the battlefield at the beginning of the next end step.",
	})
	delayed := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger).Trigger
	put := delayed.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	card, ok := put.Source.CardRef()
	if !delayed.CapturedCard.Exists || !ok || card != game.CapturedCardReference() {
		t.Fatal("departed card body bypassed its captured card incarnation")
	}
}
