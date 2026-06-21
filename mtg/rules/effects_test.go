package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addEffectSpellToStack(g *game.Game, controller game.PlayerID, primitive game.Primitive, targets []game.Target) id.ID {
	return addInstructionSpellToStackForController(g, controller, []game.Instruction{{Primitive: primitive}}, targets)
}

func addInstructionSpellToStack(g *game.Game, instructions []game.Instruction) id.ID {
	return addInstructionSpellToStackForController(g, game.Player1, instructions, nil)
}

func addInstructionSpellToStackForController(g *game.Game, controller game.PlayerID, instructions []game.Instruction, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Effect Spell",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: append([]game.Instruction(nil), instructions...),
			}.Ability())},
		},
		Owner: controller,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Targets:    targets,
	})
	return sourceID
}

func resolveInstruction(engine *Engine, g *game.Game, obj *game.StackObject, primitive game.Primitive, log *TurnLog) {
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: primitive}, [game.NumPlayers]PlayerAgent{}, log)
}

func TestSacrificePermanentsEffectAsksPlayerToChooseWhenExcessEligible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player2)
	creature2 := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		Player:    game.TargetPlayerReference(0),
		Amount:    game.Fixed(1),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment {
		t.Fatalf("choices = %+v, want one ChoicePayment", log.Choices)
	}
	if len(log.Choices[0].Request.Options) != 2 {
		t.Fatalf("options = %d, want 2 (one per candidate)", len(log.Choices[0].Request.Options))
	}
	// Agent chose index 1 (creature2).
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("chosen creature remained on battlefield")
	}
	if !g.Players[game.Player2].Graveyard.Contains(creature2.CardInstanceID) {
		t.Fatal("chosen creature was not moved to graveyard")
	}
	if _, ok := permanentByObjectID(g, creature1.ObjectID); !ok {
		t.Fatal("unchosen creature was removed from battlefield")
	}
}

func TestSacrificePermanentsEffectSacrificesAllWhenEligibleCountLEAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		Player:    game.TargetPlayerReference(0),
		Amount:    game.Fixed(2),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no choice when eligible <= amount", log.Choices)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature was not sacrificed when eligible count <= amount")
	}
	if !g.Players[game.Player2].Graveyard.Contains(creature.CardInstanceID) {
		t.Fatal("sacrificed creature was not moved to graveyard")
	}
}

// TestSacrificePermanentsFallbackDiscardWhenPlayerCantSacrifice covers the
// "Each player who can't discards a card." rider: a player controlling no
// eligible permanent discards instead of sacrificing.
func TestSacrificePermanentsFallbackDiscardWhenPlayerCantSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Player1 controls a creature (can sacrifice); Player2 controls none but
	// holds a card in hand (must discard via the fallback).
	creature := addCreaturePermanent(g, game.Player1)
	handCard := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Spare Card",
		Types: []types.Card{types.Sorcery},
	}})
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		PlayerGroup: game.AllPlayersReference(),
		Amount:      game.Fixed(1),
		Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Fallback: game.SacrificeFallback{
			Kind:   game.SacrificeFallbackDiscard,
			Amount: game.Fixed(1),
		},
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("Player1's creature was not sacrificed")
	}
	if g.Players[game.Player2].Hand.Contains(handCard) {
		t.Fatal("Player2 who couldn't sacrifice did not discard via the fallback")
	}
	if !g.Players[game.Player2].Graveyard.Contains(handCard) {
		t.Fatal("Player2's discarded card was not moved to graveyard")
	}
}

type sacrificeChoiceAgent struct {
	t                   *testing.T
	g                   *game.Game
	mustRemainForChoice id.ID
	choice              []int
	order               *[]game.PlayerID
}

func (*sacrificeChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *sacrificeChoiceAgent) ChooseChoice(obs PlayerObservation, _ game.ChoiceRequest) []int {
	a.t.Helper()
	if a.mustRemainForChoice != 0 {
		if _, ok := permanentByObjectID(a.g, a.mustRemainForChoice); !ok {
			a.t.Fatal("earlier player's chosen permanent left the battlefield before all players chose")
		}
	}
	if a.order != nil {
		*a.order = append(*a.order, obs.Player)
	}
	return append([]int(nil), a.choice...)
}

func TestSacrificePermanentsEffectGroupUsesAPNAPChoicesBeforeSimultaneousMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player3
	engine := NewEngine(nil)
	p2creature1 := addCreaturePermanent(g, game.Player2)
	p2creature2 := addCreaturePermanent(g, game.Player2)
	p3creature1 := addCreaturePermanent(g, game.Player3)
	p3creature2 := addCreaturePermanent(g, game.Player3)
	p4creature1 := addCreaturePermanent(g, game.Player4)
	p4creature2 := addCreaturePermanent(g, game.Player4)
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		PlayerGroup: game.OpponentsReference(),
		Amount:      game.Fixed(1),
		Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, nil)
	var choiceOrder []game.PlayerID
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &sacrificeChoiceAgent{
			t:                   t,
			g:                   g,
			mustRemainForChoice: p3creature1.ObjectID,
			choice:              []int{0},
			order:               &choiceOrder,
		},
		game.Player3: &sacrificeChoiceAgent{
			t:      t,
			g:      g,
			choice: []int{0},
			order:  &choiceOrder,
		},
		game.Player4: &sacrificeChoiceAgent{
			t:                   t,
			g:                   g,
			mustRemainForChoice: p3creature1.ObjectID,
			choice:              []int{0},
			order:               &choiceOrder,
		},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if want := []game.PlayerID{game.Player3, game.Player4, game.Player2}; !slices.Equal(choiceOrder, want) {
		t.Fatalf("choice order = %v, want APNAP order %v", choiceOrder, want)
	}
	for _, chosen := range []*game.Permanent{p2creature1, p3creature1, p4creature1} {
		if _, ok := permanentByObjectID(g, chosen.ObjectID); ok {
			t.Fatalf("chosen permanent %v remained on battlefield", chosen.ObjectID)
		}
	}
	for _, unchosen := range []*game.Permanent{p2creature2, p3creature2, p4creature2} {
		if _, ok := permanentByObjectID(g, unchosen.ObjectID); !ok {
			t.Fatalf("unchosen permanent %v was removed from battlefield", unchosen.ObjectID)
		}
	}
}

func TestSacrificePermanentsEffectSinglePlayerChoosesBeforeMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player2)
	creature2 := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		Player:    game.TargetPlayerReference(0),
		Amount:    game.Fixed(1),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &sacrificeChoiceAgent{
			t:      t,
			g:      g,
			choice: []int{1},
		},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, creature1.ObjectID); !ok {
		t.Fatal("unchosen creature was removed from battlefield")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("chosen creature remained on battlefield")
	}
}

func TestSacrificePermanentsEffectGroupChoosesBeforeSacrificingSimultaneously(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	p2creature1 := addCreaturePermanent(g, game.Player2)
	p2creature2 := addCreaturePermanent(g, game.Player2)
	p3creature1 := addCreaturePermanent(g, game.Player3)
	p3creature2 := addCreaturePermanent(g, game.Player3)
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		PlayerGroup: game.OpponentsReference(),
		Amount:      game.Fixed(1),
		Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &sacrificeChoiceAgent{
			t:                   t,
			g:                   g,
			mustRemainForChoice: p2creature1.ObjectID,
			choice:              []int{0},
		},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, p2creature1.ObjectID); ok {
		t.Fatal("chosen player2 creature remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, p2creature2.ObjectID); !ok {
		t.Fatal("unchosen player2 creature was removed from battlefield")
	}
	if _, ok := permanentByObjectID(g, p3creature1.ObjectID); ok {
		t.Fatal("chosen player3 creature remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, p3creature2.ObjectID); !ok {
		t.Fatal("unchosen player3 creature was removed from battlefield")
	}
}

func TestSacrificePermanentsEffectSelectsOnlyMatchingType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player2)
	// Add a land (non-creature) controlled by the same player.
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}},
		Owner: game.Player2,
	}
	g.Battlefield = append(g.Battlefield, &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player2,
		Controller:     game.Player2,
	})
	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		Player:    game.TargetPlayerReference(0),
		Amount:    game.Fixed(1),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	// The land is not a creature so it should not be sacrificed.
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature was not sacrificed")
	}
	// The land should remain.
	foundLand := false
	for _, p := range g.Battlefield {
		if p.CardInstanceID == cardID {
			foundLand = true
		}
	}
	if !foundLand {
		t.Fatal("land was incorrectly removed when only creatures should have been sacrificed")
	}
}
