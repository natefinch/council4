package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// aesirExileLink is the source-keyed link a Saga's first chapter publishes for
// the permanent card it exiles from its controller's graveyard, mirroring the
// constant the cardgen lowering emits for The Aesir Escape Valhalla.
const aesirExileLink = "exile-graveyard-card"

// TestExileFromGraveyardPublishesLinkedManaValue verifies the full cross-chapter
// linked-exile path issue #1486's Aesir chapters rely on: an ExileFromGraveyard
// that publishes the chosen card under a source-keyed link, followed by a
// GainLife whose dynamic amount reads that exiled card's mana value through a
// LinkedObjectReference. The exiled card never touched the battlefield, so its
// mana value must resolve from its card identity rather than a battlefield
// object, which is the card-only linked-object path.
func TestExileFromGraveyardPublishesLinkedManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Aesir Saga",
		Types: []types.Card{types.Enchantment},
	}})
	obj := triggeredObjFor(source)

	gyCard := g.IDGen.Next()
	g.CardInstances[gyCard] = &game.CardInstance{
		ID:    gyCard,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "Buried Relic",
			Types:    []types.Card{types.Artifact},
			ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		}},
	}
	g.Players[game.Player1].Graveyard.AddToBottom(gyCard)

	startingLife := g.Players[game.Player1].Life
	resolver := newEffectResolver(engine, g, obj,
		[game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}, &TurnLog{})

	resolver.resolveInstruction(&game.Instruction{Primitive: game.ExileFromGraveyardChoice(
		game.ControllerReference(),
		game.Selection{},
		game.Fixed(1),
		false,
		aesirExileLink,
	)})

	if !g.Players[game.Player1].Exile.Contains(gyCard) {
		t.Fatalf("graveyard card was not exiled: exile=%v", g.Players[game.Player1].Exile.All())
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, aesirExileLink))
	if len(refs) != 1 || refs[0].CardID != gyCard {
		t.Fatalf("linked objects = %v, want one ref to %v", refs, gyCard)
	}

	resolver.resolveInstruction(&game.Instruction{Primitive: game.GainLife{
		Player: game.ControllerReference(),
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectManaValue,
			Multiplier: 1,
			Object:     game.LinkedObjectReference(aesirExileLink),
		}),
	}})

	if got, want := g.Players[game.Player1].Life, startingLife+4; got != want {
		t.Fatalf("life = %d, want %d (gained the exiled card's mana value)", got, want)
	}
}
