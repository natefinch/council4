package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const sacrificeReanimationResultKey = game.ResultKey("sacrifice-succeeded")

func sacrificeConditionedReanimationInstructions() []game.Instruction {
	gate := opt.Val(game.InstructionResultGate{
		Key:       sacrificeReanimationResultKey,
		Succeeded: game.TriTrue,
	})
	return []game.Instruction{
		{
			Primitive: game.SacrificePermanents{
				Player:    game.ControllerReference(),
				Amount:    game.Fixed(1),
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			PublishResult: sacrificeReanimationResultKey,
		},
		{
			Primitive: game.PutOnBattlefield{
				Sources: []game.BattlefieldSource{
					game.CardBattlefieldSource(game.CardReference{
						Kind:        game.CardReferenceTarget,
						TargetIndex: 0,
					}),
					game.CardBattlefieldSource(game.CardReference{
						Kind:        game.CardReferenceTarget,
						TargetIndex: 1,
					}),
				},
				EntryTapped: true,
			},
			ResultGate: gate,
		},
	}
}

func addSacrificeConditionedReanimationSpell(
	g *game.Game,
	targets []game.Target,
) id.ID {
	sourceID := addInstructionSpellToStackForController(
		g,
		game.Player1,
		sacrificeConditionedReanimationInstructions(),
		targets,
	)
	card, _ := g.GetCardInstance(sourceID)
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 2,
		MaxTargets: 2,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			Controller:    game.ControllerYou,
			RequiredTypes: []types.Card{types.Creature},
		}),
	}}
	return sourceID
}

func TestSacrificeConditionedReanimationReturnsBothTargetsTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sacrifice := addCreaturePermanent(g, game.Player1)
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "First Target",
		Types: []types.Card{types.Creature},
	}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Second Target",
		Types: []types.Card{types.Creature},
	}})
	addSacrificeConditionedReanimationSpell(g, []game.Target{
		currentCardTarget(t, g, first),
		currentCardTarget(t, g, second),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, sacrifice.ObjectID); ok {
		t.Fatal("chosen creature remained on the battlefield")
	}
	for _, cardID := range []id.ID{first, second} {
		permanent, ok := reanimatedPermanent(g, cardID)
		if !ok {
			t.Fatalf("target %d was not returned", cardID)
		}
		if !permanent.Tapped {
			t.Fatalf("target %d entered untapped", cardID)
		}
	}
	var entryEvents []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged &&
			event.FromZone == zone.Graveyard &&
			event.ToZone == zone.Battlefield &&
			(event.CardID == first || event.CardID == second) {
			entryEvents = append(entryEvents, event)
		}
	}
	if len(entryEvents) != 2 ||
		entryEvents[0].SimultaneousID == 0 ||
		entryEvents[0].SimultaneousID != entryEvents[1].SimultaneousID {
		t.Fatalf("entry events = %+v; want one simultaneous zone change", entryEvents)
	}
}

func TestSacrificeConditionedReanimationReturnsNothingWhenSacrificeFails(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "First Target",
		Types: []types.Card{types.Creature},
	}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Second Target",
		Types: []types.Card{types.Creature},
	}})
	addSacrificeConditionedReanimationSpell(g, []game.Target{
		currentCardTarget(t, g, first),
		currentCardTarget(t, g, second),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, cardID := range []id.ID{first, second} {
		if _, ok := reanimatedPermanent(g, cardID); ok {
			t.Fatalf("target %d returned without a successful sacrifice", cardID)
		}
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("target %d left the graveyard", cardID)
		}
	}
}

func TestSacrificeConditionedReanimationHandlesTargetsIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	illegal := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Illegal Target",
		Types: []types.Card{types.Creature},
	}})
	legal := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Legal Target",
		Types: []types.Card{types.Creature},
	}})
	addSacrificeConditionedReanimationSpell(g, []game.Target{
		currentCardTarget(t, g, illegal),
		currentCardTarget(t, g, legal),
	})
	if !moveCardBetweenZones(g, game.Player1, illegal, zone.Graveyard, zone.Exile) {
		t.Fatal("moving first target before resolution")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Exile.Contains(illegal) {
		t.Fatal("illegal target left exile")
	}
	if _, ok := reanimatedPermanent(g, illegal); ok {
		t.Fatal("illegal target was returned")
	}
	permanent, ok := reanimatedPermanent(g, legal)
	if !ok || !permanent.Tapped {
		t.Fatalf("legal target = %#v; want returned tapped", permanent)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result == "countered by rules" {
		t.Fatalf("resolve log = %+v; want spell to resolve with one legal target", log.Resolves)
	}
}

func TestSacrificeConditionedReanimationPreparesSimultaneousEntriesTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	doubler := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Doubler",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{{
			Replacement: game.ReplacementEffect{CounterMultiplier: 2},
		}},
	}})
	entersWithCounter := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Entrant",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement(
				"Counter Entrant enters with a +1/+1 counter on it.",
				game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
			),
		},
	}})
	addSacrificeConditionedReanimationSpell(g, []game.Target{
		currentCardTarget(t, g, doubler),
		currentCardTarget(t, g, entersWithCounter),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := reanimatedPermanent(g, entersWithCounter)
	if !ok {
		t.Fatal("counter entrant was not returned")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d; want 1 from pre-entry replacement state", got)
	}
}
