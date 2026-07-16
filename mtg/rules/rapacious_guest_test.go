package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// resolveLeaveTriggerLoseLife resolves the Rapacious Guest leaves-the-battlefield
// body — "target opponent loses life equal to its power" — against a triggering
// zone-change event that names the departed permanent, so the amount reads the
// permanent's last-known power (CR 603.10, CR 608.2h).
func resolveLeaveTriggerLoseLife(
	t *testing.T,
	g *game.Game,
	engine *Engine,
	sourceID, leftPermanentID id.ID,
	target game.PlayerID,
) {
	t.Helper()
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        sourceID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventZoneChanged,
			PermanentID: leftPermanentID,
			FromZone:    zone.Battlefield,
		},
		Targets: []game.Target{game.PlayerTarget(target)},
	}
	resolveInstruction(engine, g, obj, game.LoseLife{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.EventPermanentReference(),
		}),
		Player: game.TargetPlayerReference(0),
	}, &TurnLog{})
}

// TestRapaciousGuestLeaveTriggerLoseLifeByDestination proves the target opponent
// loses life equal to the departed creature's last-known power regardless of
// which zone the creature left to — bounce (hand), exile, or death (graveyard).
// The source is a real battlefield permanent moved off the battlefield through
// the ordinary zone-change flow, which snapshots its last-known information.
func TestRapaciousGuestLeaveTriggerLoseLifeByDestination(t *testing.T) {
	for _, tc := range []struct {
		name        string
		destination zone.Type
	}{
		{"bounce to hand", zone.Hand},
		{"exile", zone.Exile},
		{"die to graveyard", zone.Graveyard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			guest := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			before := g.Players[game.Player2].Life

			if !movePermanentToZone(g, guest, tc.destination) {
				t.Fatalf("failed to move guest to %v", tc.destination)
			}

			resolveLeaveTriggerLoseLife(t, g, engine, guest.ObjectID, guest.ObjectID, game.Player2)

			if got := before - g.Players[game.Player2].Life; got != 2 {
				t.Fatalf("opponent life lost = %d, want 2 (last-known power)", got)
			}
		})
	}
}

// TestRapaciousGuestLeaveTriggerReadsModifiedPower proves the amount reads the
// departed creature's last-known power including counters and continuous effects
// present when it left (CR 608.2h). A 2/2 with a +1/+1 counter under a +1/+1
// anthem is a 4/4 at departure, so the target opponent loses 4 life.
func TestRapaciousGuestLeaveTriggerReadsModifiedPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	guest := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	guest.Counters.Add(counter.PlusOnePlusOne, 1)

	anthem := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Anthem Source",
		Types: []types.Card{types.Enchantment},
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             g.IDGen.Next(),
		Controller:     game.Player1,
		SourceObjectID: anthem.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, guest); got != 4 {
		t.Fatalf("guest power before departure = %d, want 4 (2 base +1 counter +1 anthem)", got)
	}

	before := g.Players[game.Player2].Life
	if !movePermanentToZone(g, guest, zone.Graveyard) {
		t.Fatal("failed to move guest to graveyard")
	}

	resolveLeaveTriggerLoseLife(t, g, engine, guest.ObjectID, guest.ObjectID, game.Player2)

	if got := before - g.Players[game.Player2].Life; got != 4 {
		t.Fatalf("opponent life lost = %d, want 4 (modified last-known power)", got)
	}
}

// TestRapaciousGuestLeaveTriggerTokenSource proves a token source works the same
// as a card source: a token creature that leaves the battlefield still has its
// last-known power recorded, so the target opponent loses life equal to it even
// though the token ceases to exist.
func TestRapaciousGuestLeaveTriggerTokenSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player1,
		Controller: game.Player1,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:      "Guest Token",
			Types:     []types.Card{types.Creature},
			Colors:    []color.Color{color.Black},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		}},
	}
	g.Battlefield = append(g.Battlefield, token)
	before := g.Players[game.Player2].Life

	if !movePermanentToZone(g, token, zone.Graveyard) {
		t.Fatal("failed to move token off the battlefield")
	}

	resolveLeaveTriggerLoseLife(t, g, engine, token.ObjectID, token.ObjectID, game.Player2)

	if got := before - g.Players[game.Player2].Life; got != 3 {
		t.Fatalf("opponent life lost = %d, want 3 (token last-known power)", got)
	}
}

// TestRapaciousGuestLeaveTriggerAffectsOnlyChosenTarget proves the life loss
// lands only on the chosen target opponent, not on the controller or any other
// opponent. Only Player3 was targeted, so only Player3 loses life.
func TestRapaciousGuestLeaveTriggerAffectsOnlyChosenTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	guest := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	p1Before := g.Players[game.Player1].Life
	p2Before := g.Players[game.Player2].Life
	p3Before := g.Players[game.Player3].Life

	if !movePermanentToZone(g, guest, zone.Graveyard) {
		t.Fatal("failed to move guest to graveyard")
	}

	resolveLeaveTriggerLoseLife(t, g, engine, guest.ObjectID, guest.ObjectID, game.Player3)

	if got := p3Before - g.Players[game.Player3].Life; got != 2 {
		t.Fatalf("targeted opponent life lost = %d, want 2", got)
	}
	if g.Players[game.Player1].Life != p1Before {
		t.Fatalf("controller life = %d, want %d (untouched)", g.Players[game.Player1].Life, p1Before)
	}
	if g.Players[game.Player2].Life != p2Before {
		t.Fatalf("untargeted opponent life = %d, want %d (untouched)", g.Players[game.Player2].Life, p2Before)
	}
}

// TestRapaciousGuestLeaveTriggerNoLastKnownLosesNoLife proves the amount fails
// closed when the departed permanent has no recorded last-known information: with
// no snapshot to read, the power resolves to zero and no life is lost.
func TestRapaciousGuestLeaveTriggerNoLastKnownLosesNoLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Rapacious Guest"}})
	before := g.Players[game.Player2].Life

	resolveLeaveTriggerLoseLife(t, g, engine, sourceID, g.IDGen.Next(), game.Player2)

	if g.Players[game.Player2].Life != before {
		t.Fatalf("opponent life = %d, want %d (no last-known power)", g.Players[game.Player2].Life, before)
	}
}
