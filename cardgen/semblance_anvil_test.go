package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerSemblanceAnvilImprintReduction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Semblance Anvil",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: "Imprint — When this artifact enters, you may exile a nonland card from your hand.\nSpells you cast that share a card type with the exiled card cost {2} less to cast.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	instruction := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	choose, ok := instruction.Primitive.(game.ChooseFromZone)
	if !ok || !instruction.Optional ||
		choose.SourceZone != zone.Hand ||
		len(choose.Filter.ExcludedTypes) != 1 ||
		choose.Filter.ExcludedTypes[0] != types.Land {
		t.Fatalf("imprint instruction = %#v, want optional nonland hand exile", instruction)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.SharedExiledCardTypeReduction != 2 ||
		!modifier.SharedExiledCardTypeReductionOnce ||
		!modifier.ExiledLinkObjectScoped {
		t.Fatalf("modifier = %#v, want flat shared-type reduction", modifier)
	}
}
