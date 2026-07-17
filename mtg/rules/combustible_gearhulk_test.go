package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const combustibleMilledKey = game.LinkedKey("combustible-milled")

func combustibleGearhulkSequence() []game.Instruction {
	declined := opt.Val(game.InstructionResultGate{Key: "offer", Accepted: game.TriFalse})
	return []game.Instruction{
		{
			Primitive: game.Draw{Amount: game.Fixed(3), Player: game.ControllerReference()},
			Optional:  true, OptionalActor: opt.Val(game.TargetPlayerReference(0)), PublishResult: "offer",
		},
		{
			Primitive: game.Mill{
				Amount: game.Fixed(3), Player: game.ControllerReference(), PublishLinked: combustibleMilledKey,
			},
			ResultGate: declined,
		},
		{
			Primitive: game.Damage{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind: game.DynamicAmountReferencedCardsTotalManaValue, LinkedKey: combustibleMilledKey,
				}),
				Recipient:    game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			},
			ResultGate: declined,
		},
	}
}

func combustibleCard(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}}
}

func combustibleResolver(t *testing.T, target game.PlayerID) (*game.Game, *Engine, *game.StackObject) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Combustible Source", Types: []types.Card{types.Artifact, types.Creature}, Colors: []color.Color{color.Red},
	}})
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{game.PlayerTarget(target)},
	}
	return g, engine, obj
}

func resolveCombustibleSequence(engine *Engine, g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent) {
	log := TurnLog{}
	for i := range combustibleGearhulkSequence() {
		sequence := combustibleGearhulkSequence()
		engine.resolveInstructionWithChoices(g, obj, &sequence[i], agents, &log)
	}
}

func TestCombustibleGearhulkAcceptOnlyDraws(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	for i := 1; i <= 3; i++ {
		addCardToLibrary(g, game.Player1, combustibleCard("Card", i))
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}}
	resolveCombustibleSequence(engine, g, obj, agents)
	if g.Players[game.Player1].Hand.Size() != 3 ||
		g.Players[game.Player1].Graveyard.Size() != 0 ||
		g.Players[game.Player2].Life != 40 {
		t.Fatalf("hand/graveyard/life = %d/%d/%d", g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Graveyard.Size(), g.Players[game.Player2].Life)
	}
}

func TestCombustibleGearhulkAcceptedEmptyDrawDoesNotTakeDeclineBranch(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}}
	resolveCombustibleSequence(engine, g, obj, agents)
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("opponent life/graveyard = %d/%d", g.Players[game.Player2].Life, g.Players[game.Player1].Graveyard.Size())
	}
}

func TestCombustibleGearhulkDeclineMillsAndDamagesTarget(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player3)
	for i := 1; i <= 3; i++ {
		addCardToLibrary(g, game.Player1, combustibleCard("Card", i))
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}}}
	resolveCombustibleSequence(engine, g, obj, agents)
	if g.Players[game.Player1].Graveyard.Size() != 3 ||
		g.Players[game.Player3].Life != 34 ||
		g.Players[game.Player2].Life != 40 {
		t.Fatalf("graveyard/target/other life = %d/%d/%d", g.Players[game.Player1].Graveyard.Size(), g.Players[game.Player3].Life, g.Players[game.Player2].Life)
	}
}

func TestCombustibleGearhulkShortAndEmptyLibraries(t *testing.T) {
	for _, count := range []int{0, 2} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			g, engine, obj := combustibleResolver(t, game.Player2)
			for range count {
				addCardToLibrary(g, game.Player1, combustibleCard("Two", 2))
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
			resolveCombustibleSequence(engine, g, obj, agents)
			if g.Players[game.Player1].Graveyard.Size() != count ||
				g.Players[game.Player2].Life != 40-2*count {
				t.Fatalf("count %d: graveyard/life = %d/%d", count, g.Players[game.Player1].Graveyard.Size(), g.Players[game.Player2].Life)
			}
		})
	}
}

func TestCombustibleGearhulkUsesNonStackManaValues(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace:  game.CardFace{Name: "Split", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(2)})},
		Layout:    game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{Name: "Other", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(4)})}),
	})
	addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "DFC", Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(3)})},
		Layout:   game.LayoutModalDFC,
		Back:     opt.Val(game.CardFace{Name: "Back", Types: []types.Card{types.Land}, ManaCost: opt.Val(cost.Mana{cost.O(7)})}),
	})
	addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace:  game.CardFace{Name: "Adventure", Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(5)})},
		Layout:    game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{Name: "Journey", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(1)})}),
	})
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	resolveCombustibleSequence(engine, g, obj, agents)
	if got := g.Players[game.Player2].Life; got != 26 {
		t.Fatalf("opponent life = %d, want 26 from 6+3+5 mana value", got)
	}
}

func TestCombustibleGearhulkDamageTracksExactMilledBatch(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	addCardToGraveyard(g, game.Player1, combustibleCard("Unrelated", 10))
	milled := []id.ID{
		addCardToLibrary(g, game.Player1, combustibleCard("One", 1)),
		addCardToLibrary(g, game.Player1, combustibleCard("Two", 2)),
		addCardToLibrary(g, game.Player1, combustibleCard("Three", 3)),
	}
	sequence := combustibleGearhulkSequence()
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	log := TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, &sequence[0], agents, &log)
	engine.resolveInstructionWithChoices(g, obj, &sequence[1], agents, &log)
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, string(combustibleMilledKey)))
	if len(refs) != 3 {
		t.Fatalf("linked mill batch size = %d, want 3", len(refs))
	}
	wantIDs := make(map[id.ID]bool, len(milled))
	for _, cardID := range milled {
		wantIDs[cardID] = true
	}
	for _, ref := range refs {
		if !wantIDs[ref.CardID] || ref.CardZoneVersion == 0 {
			t.Fatalf("linked mill ref = %#v, want a milled card identity and zone version", ref)
		}
	}
	var millBatch id.ID
	millEvents := 0
	for _, event := range g.Events {
		if event.Kind != game.EventZoneChanged ||
			event.FromZone != zone.Library ||
			event.ToZone != zone.Graveyard ||
			!wantIDs[event.CardID] {
			continue
		}
		millEvents++
		if millBatch == 0 {
			millBatch = event.SimultaneousID
		}
		if event.SimultaneousID == 0 || event.SimultaneousID != millBatch {
			t.Fatalf("milled cards did not share one simultaneous batch: %#v", g.Events)
		}
	}
	if millEvents != 3 {
		t.Fatalf("mill zone-change events = %d, want 3", millEvents)
	}
	if !moveCardBetweenZones(g, game.Player1, milled[0], zone.Graveyard, zone.Hand) {
		t.Fatal("moving a milled card after the batch failed")
	}
	engine.resolveInstructionWithChoices(g, obj, &sequence[2], agents, &log)
	if got := g.Players[game.Player2].Life; got != 34 {
		t.Fatalf("opponent life = %d, want 34 from the exact 1+2+3 batch", got)
	}
}

func TestCombustibleGearhulkSourceCanLeaveBeforeDamage(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	addCardToLibrary(g, game.Player1, combustibleCard("Four", 4))
	resolveInstruction(engine, g, obj, game.Sacrifice{Object: game.SourcePermanentReference()}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	resolveCombustibleSequence(engine, g, obj, agents)
	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("opponent life = %d, want 36", got)
	}
}

func TestCombustibleGearhulkDamageUsesPreventionAndReplacement(t *testing.T) {
	t.Run("prevention", func(t *testing.T) {
		g, engine, obj := combustibleResolver(t, game.Player2)
		addCardToLibrary(g, game.Player1, combustibleCard("Four", 4))
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player2}, game.PreventDamage{
			Player: game.ControllerReference(), All: true, OneShot: true,
		}, nil)
		agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
		resolveCombustibleSequence(engine, g, obj, agents)
		if got := g.Players[game.Player2].Life; got != 40 {
			t.Fatalf("opponent life = %d, want 40", got)
		}
	})
	t.Run("replacement", func(t *testing.T) {
		g, engine, obj := combustibleResolver(t, game.Player2)
		addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
		addCardToLibrary(g, game.Player1, combustibleCard("Four", 4))
		agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
		resolveCombustibleSequence(engine, g, obj, agents)
		if got := g.Players[game.Player2].Life; got != 32 {
			t.Fatalf("opponent life = %d, want 32 after doubling", got)
		}
	})
}

func TestCombustibleGearhulkTargetLossFizzles(t *testing.T) {
	g, engine, obj := combustibleResolver(t, game.Player2)
	addCardToLibrary(g, game.Player1, combustibleCard("Four", 4))
	trigger := &game.TriggeredAbility{Content: game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Constraint: "target opponent",
			Allow: game.TargetAllowPlayer, Selection: opt.Val(game.Selection{Player: game.PlayerOpponent}),
		}},
		Sequence: combustibleGearhulkSequence(),
	}.Ability()}
	obj.InlineTrigger = trigger
	g.Stack.Push(obj)
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Library.Size() != 1 ||
		g.Players[game.Player1].Hand.Size() != 0 ||
		g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("library/hand/graveyard = %d/%d/%d", g.Players[game.Player1].Library.Size(), g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Graveyard.Size())
	}
}
