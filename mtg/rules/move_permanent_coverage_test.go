package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestMovePermanentCoversUnwrittenDestinationRiderCombinations is the payoff
// test for collapsing Bounce, Exile, and PutPermanentOnLibrary into one
// destination-parameterized primitive.
//
// Each of the three old primitives owned a rider the others could not express:
// only Bounce could let the controller choose which permanents move, only Exile
// could publish a linked key, only PutPermanentOnLibrary could reach the bottom
// of the library. Every combination below crosses one of those riders with a
// destination that did not own it, so each one previously required a new
// primitive, a new handler, a new validator, and a new lowering matcher. They now
// resolve with no new runtime code at all.
//
// If this test ever fails, the family has been re-specialized and the tax is
// back.
func TestMovePermanentCoversUnwrittenDestinationRiderCombinations(t *testing.T) {
	creatures := game.Selection{RequiredTypes: []types.Card{types.Creature}}

	t.Run("controlled choice to exile", func(t *testing.T) {
		// "Exile a creature you control." Controlled choice was a bounce-only
		// rider; the chooser is skipped because the pool holds exactly one.
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		mine := addCombatCreaturePermanent(g, game.Player1)
		theirs := addCombatCreaturePermanent(g, game.Player2)
		source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())

		engine.resolveInstructionWithChoices(g, linkedSourceObject(source), &game.Instruction{
			Primitive: game.MovePermanent{
				ControlledChoice: true,
				Amount:           game.Fixed(1),
				Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: creatures.RequiredTypes, Controller: game.ControllerYou}),
				Destination:      zone.Exile,
			},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if _, ok := permanentByObjectID(g, mine.ObjectID); ok {
			t.Fatal("chosen creature did not leave the battlefield")
		}
		if _, ok := permanentByObjectID(g, theirs.ObjectID); !ok {
			t.Fatal("an opponent's creature was exiled by a you-control choice")
		}
		if got, ok := cardZone(g, mine.CardInstanceID); !ok || got != zone.Exile {
			t.Fatalf("chosen creature zone = %v (ok=%v), want exile", got, ok)
		}
	})

	t.Run("controlled choice to library bottom", func(t *testing.T) {
		// "Put a creature you control on the bottom of its owner's library."
		// Controlled choice was bounce-only and bottom placement was
		// library-only, so this wording had no expressible primitive.
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		mine := addCombatCreaturePermanent(g, game.Player1)
		source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
		player, ok := playerByID(g, game.Player1)
		if !ok {
			t.Fatal("player 1 missing")
		}
		before := player.Library.Size()

		engine.resolveInstructionWithChoices(g, linkedSourceObject(source), &game.Instruction{
			Primitive: game.MovePermanent{
				ControlledChoice: true,
				Amount:           game.Fixed(1),
				Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: creatures.RequiredTypes, Controller: game.ControllerYou}),
				Destination:      zone.Library,
				LibraryBottom:    true,
			},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if _, ok := permanentByObjectID(g, mine.ObjectID); ok {
			t.Fatal("chosen creature did not leave the battlefield")
		}
		if got := player.Library.Size(); got != before+1 {
			t.Fatalf("library size = %d, want %d", got, before+1)
		}
		if bottom, ok := player.Library.Bottom(); !ok || bottom != mine.CardInstanceID {
			t.Fatalf("library bottom = %v (ok=%v), want the moved card %v", bottom, ok, mine.CardInstanceID)
		}
	})

	t.Run("linked publish on a return to hand", func(t *testing.T) {
		// "Return target creature to its owner's hand" while remembering the
		// card, so a sibling instruction can act on exactly that card.
		// Publishing a link was an exile-only rider.
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		theirs := addCombatCreaturePermanent(g, game.Player2)
		source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
		obj := linkedSourceObject(source)

		engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
			Primitive: game.MovePermanent{
				Group:         game.BattlefieldGroup(creatures),
				Destination:   zone.Hand,
				PublishLinked: game.LinkedKey("returned-cards"),
			},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if _, ok := permanentByObjectID(g, theirs.ObjectID); ok {
			t.Fatal("creature was not returned to hand")
		}
		linked := linkedObjects(g, linkedObjectSourceKey(g, obj, "returned-cards"))
		if len(linked) != 1 {
			t.Fatalf("linked objects = %d, want the one returned card", len(linked))
		}
	})
}
