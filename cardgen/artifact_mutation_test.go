package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerArtifactMutationManaValueTokens(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Artifact Mutation",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}{G}",
		OracleText: "Destroy target artifact. It can't be regenerated. Create X 1/1 green Saproling creature tokens, where X is that artifact's mana value.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok || !destroy.PreventRegeneration {
		t.Fatalf("destroy = %#v", mode.Sequence[0].Primitive)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	dynamic := create.Amount.DynamicAmount()
	if !ok ||
		!dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountObjectManaValue ||
		dynamic.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("create = %#v", mode.Sequence[1].Primitive)
	}
}
