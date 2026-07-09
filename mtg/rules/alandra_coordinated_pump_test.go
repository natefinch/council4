package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestAlandraCoordinatedSelfGroupPumpResolves proves the coordinated "Alandra and
// Drakes you control each get +X/+X until end of turn, where X is the number of
// cards in your hand." ability pumps the source permanent and every OTHER Drake
// the controller controls by the hand-size amount, without double-pumping the
// source and without touching opponent Drakes or non-Drake creatures. The lowered
// ability is a ModifyPT on the source followed by an ApplyContinuous over the
// source-excluding Drake group; resolving both in order reproduces the card.
func TestAlandraCoordinatedSelfGroupPumpResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	alandra := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Alandra, Sky Dreamer",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}})
	drakeDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Drake",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Drake},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	myDrake := addReplacementPermanent(t, g, game.Player1, drakeDef)
	opponentDrake := addReplacementPermanent(t, g, game.Player2, drakeDef)
	myNonDrake := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Merfolk Ally",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	// Three cards in the controller's hand => +3/+3.
	for i := range 3 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}

	handCount := game.DynamicAmount{
		Kind:       game.DynamicAmountCountCardsInZone,
		Multiplier: 1,
		Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
		CardZone:   zone.Hand,
		Selection:  &game.Selection{},
	}
	obj := &game.StackObject{Controller: game.Player1, SourceID: alandra.ObjectID}
	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.SourcePermanentReference(),
		PowerDelta:     game.Dynamic(handCount),
		ToughnessDelta: game.Dynamic(handCount),
		Duration:       game.DurationUntilEndOfTurn,
	}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroupExcluding(
				game.Selection{SubtypesAny: []types.Sub{types.Drake}, Controller: game.ControllerYou},
				game.SourcePermanentReference(),
			),
			PowerDeltaDynamic:     opt.Val(handCount),
			ToughnessDeltaDynamic: opt.Val(handCount),
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	assertPT := func(name string, permanent *game.Permanent, wantPower, wantToughness int) {
		t.Helper()
		if got := effectivePower(g, permanent); got != wantPower {
			t.Fatalf("%s power = %d, want %d", name, got, wantPower)
		}
		got, ok := effectiveToughness(g, permanent)
		if !ok || got != wantToughness {
			t.Fatalf("%s toughness = %d (ok=%v), want %d", name, got, ok, wantToughness)
		}
	}

	// Source pumped exactly once (+3/+3), never double-counted by the group.
	assertPT("Alandra", alandra, 5, 7)
	// Other Drake the controller controls pumped by the same amount.
	assertPT("my Drake", myDrake, 5, 5)
	// Opponent's Drake and the controller's non-Drake stay at base.
	assertPT("opponent Drake", opponentDrake, 2, 2)
	assertPT("my non-Drake", myNonDrake, 2, 2)
}
