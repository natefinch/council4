package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestSourceLinkedExileCountTracksCardsRemainingInExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	const link = "exiled-with-source"
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Unlicensed Hearse",
		Types:     []types.Card{types.Artifact},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
		DynamicPower: opt.Val(game.DynamicValue{
			Kind:               game.DynamicValueSourceLinkedExileCount,
			LinkedKey:          link,
			LinkedObjectScoped: true,
		}),
		DynamicToughness: opt.Val(game.DynamicValue{
			Kind:               game.DynamicValueSourceLinkedExileCount,
			LinkedKey:          link,
			LinkedObjectScoped: true,
		}),
	}})
	cardA := addCfzGraveyardCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	cardB := g.IDGen.Next()
	g.CardInstances[cardB] = &game.CardInstance{
		ID:    cardB,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "B"}},
		Owner: game.Player2,
	}
	g.Players[game.Player2].Exile.Add(cardB)
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{
			Kind:               game.TargetCard,
			CardID:             cardA,
			CardZoneVersion:    g.CardInstances[cardA].ZoneVersion,
			CardZoneVersionSet: true,
		}},
	}
	resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	if !handleMoveCard(resolver, game.MoveCard{
		Card:                      game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:                  zone.Graveyard,
		Destination:               zone.Exile,
		PublishLinked:             link,
		PublishLinkedObjectScoped: true,
	}).succeeded {
		t.Fatal("targeted exile failed")
	}
	key := linkedObjectByObjectKey(g, obj, link)
	rememberLinkedObject(g, key, game.LinkedObjectRef{
		CardID:          cardB,
		CardZoneVersion: g.CardInstances[cardB].ZoneVersion,
	})
	if power := effectivePower(g, source); power != 2 {
		t.Fatalf("power = %d, want 2 linked exiled cards", power)
	}
	moveCardBetweenZones(g, game.Player2, cardA, zone.Exile, zone.Hand)
	if power := effectivePower(g, source); power != 1 {
		t.Fatalf("power = %d after card left exile, want 1", power)
	}
	moveCardBetweenZones(g, game.Player2, cardA, zone.Hand, zone.Exile)
	if power := effectivePower(g, source); power != 1 {
		t.Fatalf("power = %d after card re-entered exile, want stale link excluded", power)
	}
	source.ObjectID = g.IDGen.Next()
	if power := effectivePower(g, source); power != 0 {
		t.Fatalf("power = %d after source re-entered, want fresh object-scoped pool", power)
	}
}
