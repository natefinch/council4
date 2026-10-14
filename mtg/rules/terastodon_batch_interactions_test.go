package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestTerastodonCreatesOneSimultaneousTokenEventAcrossRecipients(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 3)
	targets[2].Controller = game.Player3
	staffDef := cards.NewDefaultRegistry().Lookup("Staff of the Storyteller")
	if staffDef == nil {
		t.Fatal("production Staff of the Storyteller is missing")
	}
	staff := addCombatPermanent(g, game.Player2, staffDef)
	engine.resolveTopOfStack(g, &TurnLog{})
	var batchID game.ObjectID
	for _, event := range g.Events {
		if event.Kind == game.EventTokenCreated {
			if batchID == 0 {
				batchID = event.SimultaneousID
			}
			if batchID == 0 || event.SimultaneousID != batchID {
				t.Fatal("token creation was not one event across all recipients")
			}
		}
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) || g.Stack.Size() != 1 {
		t.Fatal("Staff of the Storyteller did not see one creature-token creation event")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if staff.Counters.Get(counter.Story) != 1 ||
		namedTokenCount(g, game.Player2, "Elephant") != 2 || namedTokenCount(g, game.Player3, "Elephant") != 1 {
		t.Fatal("token batch quantity or one-or-more trigger coalescing was wrong")
	}
}

func TestTerastodonOverlappingUmbraArmorDestruction(t *testing.T) {
	t.Parallel()
	for _, protected := range []bool{false, true} {
		t.Run(map[bool]string{false: "Aura departs", true: "Aura regenerates"}[protected], func(t *testing.T) {
			t.Parallel()
			g, engine, _ := terastodonStack(t, 0)
			land := addBasicLandPermanent(g, game.Player2, types.Forest)
			aura := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Fertile Ground with granted umbra armor", Types: []types.Card{types.Enchantment},
				Subtypes: []types.Sub{types.Aura},
				StaticAbilities: []game.StaticAbility{{
					KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
					}}},
				}, game.UmbraArmorStaticBody},
			}})
			if !attachPermanent(g, aura, land) {
				t.Fatal("land Aura did not attach")
			}
			if protected {
				aura.RegenerationShields = 1
			}
			obj, _ := g.Stack.Peek()
			obj.Targets = []game.Target{game.PermanentTarget(land.ObjectID), game.PermanentTarget(aura.ObjectID)}
			obj.TargetCounts = []int{2}
			engine.resolveTopOfStack(g, &TurnLog{})
			if permanentByCardID(g, land.CardInstanceID) == nil {
				t.Fatal("umbra armor did not protect the land")
			}
			want := 1
			if protected {
				want = 0
				if aura.RegenerationShields != 0 || !aura.Tapped {
					t.Fatal("overlapping destruction did not consume one regeneration replacement")
				}
			}
			if namedTokenCount(g, game.Player2, "Elephant") != want ||
				g.Players[game.Player2].Graveyard.Contains(aura.CardInstanceID) == protected {
				t.Fatal("targeted Aura's real departure was lost or its replacement applied twice")
			}
		})
	}
}

func TestAuthoredTokenCreationForDepartedPlayerIsKnownEmpty(t *testing.T) {
	t.Parallel()
	g, engine, _ := terastodonStack(t, 0)
	g.Players[game.Player3].Eliminated = true
	obj, _ := g.Stack.Peek()
	ctx := &programContext{engine: engine, game: g, obj: obj, active: true, log: &TurnLog{}}
	created, err := ctx.CreateTokens(vanillaCreature("Token", 1, 1), cardprogram.TokenRecipient{Player: game.Player3, Count: 1})
	if err != nil || created == nil || len(created) != 0 || namedTokenCount(g, game.Player3, "Token") != 0 {
		t.Fatal("a valid effect for a departed recipient did not produce a known empty result")
	}
}

func TestTerastodonCommanderRedirectedToExileDoesNotQualify(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 1)
	g.CommanderIDs[targets[0].CardInstanceID] = true
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), Description: "exile instead", MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard,
		ReplaceToZone: zone.Exile,
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player2].CommandZone.Contains(targets[0].CardInstanceID) ||
		namedTokenCount(g, game.Player2, "Elephant") != 0 {
		t.Fatal("exile arrival plus command SBA was mistaken for a graveyard arrival")
	}
}
