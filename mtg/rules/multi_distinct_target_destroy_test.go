package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestMultiDistinctTargetDestroyDestroysEachTypedTarget proves the
// multi-distinct-typed-target destroy sequence the cardgen backend emits for
// "Destroy target artifact, target creature, target enchantment, and target
// land." (Decimate) destroys every one of the four chosen targets. Each slot's
// Destroy addresses its own flat target index, so the artifact, creature,
// enchantment, and land the controller chose all move to their owner's
// graveyard.
func TestMultiDistinctTargetDestroyDestroysEachTypedTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	artifact := addArtifactPermanent(g, game.Player2)
	creature := addCreaturePermanent(g, game.Player2)
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Scrap Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
	land := addLandPermanent(g, game.Player2, "Scrap Land")

	destroyed := []*game.Permanent{artifact, creature, enchantment, land}
	sequence := make([]game.Instruction, 0, len(destroyed))
	for i := range destroyed {
		sequence = append(sequence, game.Instruction{
			Primitive: game.Destroy{Object: game.TargetPermanentReference(i)},
		})
	}
	addInstructionSpellToStackForController(g, game.Player1, sequence, []game.Target{
		game.PermanentTarget(artifact.ObjectID),
		game.PermanentTarget(creature.ObjectID),
		game.PermanentTarget(enchantment.ObjectID),
		game.PermanentTarget(land.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, target := range destroyed {
		if _, ok := permanentByObjectID(g, target.ObjectID); ok {
			t.Fatalf("target %s remained on battlefield", g.CardInstances[target.CardInstanceID].Def.Name)
		}
		if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
			t.Fatalf("target %s was not put into owner's graveyard", g.CardInstances[target.CardInstanceID].Def.Name)
		}
	}
}
