package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// dannyPinkGrantedAbility is the quoted triggered ability Danny Pink grants to
// each creature its controller controls: "Whenever one or more counters are put
// on this creature for the first time each turn, draw a card." It matches any
// counter kind (MatchCounterKind unset), coalesces simultaneous placements
// (OneOrMore), fires at most once per turn per creature (MaxTriggersPerTurn),
// and draws for that creature's controller (ControllerReference).
func dannyPinkGrantedAbility() *game.TriggeredAbility {
	return &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:     game.EventCountersAdded,
				Source:    game.TriggerSourceSelf,
				OneOrMore: true,
			},
		},
		MaxTriggersPerTurn: 1,
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}},
		}.Ability(),
	}
}

// dannyPinkCardDef mirrors the generated Danny Pink card: a continuous
// LayerAbility effect that grants dannyPinkGrantedAbility to every creature the
// card's controller controls.
func dannyPinkCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Danny Pink",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Human, types.Soldier, types.Advisor},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				AddAbilities: []game.Ability{dannyPinkGrantedAbility()},
			}},
		}},
	}}
}

func addDannyPink(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, dannyPinkCardDef())
}

// TestDannyPinkGrantsCounterDrawToControlledCreatures proves the group grant
// reaches every creature its controller controls but no others.
func TestDannyPinkGrantsCounterDrawToControlledCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addDannyPink(g, game.Player1)
	creatureA := addCombatCreaturePermanent(g, game.Player1)
	creatureB := addCombatCreaturePermanent(g, game.Player1)
	opponent := addCombatCreaturePermanent(g, game.Player2)

	if got := countGrantedTriggeredAbilities(g, creatureA); got != 1 {
		t.Fatalf("controlled creature A granted abilities = %d, want 1", got)
	}
	if got := countGrantedTriggeredAbilities(g, creatureB); got != 1 {
		t.Fatalf("controlled creature B granted abilities = %d, want 1", got)
	}
	if got := countGrantedTriggeredAbilities(g, opponent); got != 0 {
		t.Fatalf("opponent creature granted abilities = %d, want 0", got)
	}
}

// TestDannyPinkControlChangeUpdatesGrant proves the continuous grant follows
// control: a creature gains the ability when it comes under the controller's
// control and loses it when control changes away.
func TestDannyPinkControlChangeUpdatesGrant(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addDannyPink(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player2)

	if got := countGrantedTriggeredAbilities(g, creature); got != 0 {
		t.Fatalf("opponent creature granted abilities = %d, want 0", got)
	}

	creature.Controller = game.Player1
	if got := countGrantedTriggeredAbilities(g, creature); got != 1 {
		t.Fatalf("creature under controller's control granted abilities = %d, want 1", got)
	}

	creature.Controller = game.Player2
	if got := countGrantedTriggeredAbilities(g, creature); got != 0 {
		t.Fatalf("creature after control leaves granted abilities = %d, want 0", got)
	}
}

// TestDannyPinkCounterDrawFiresIndependentlyPerCreature proves the granted
// per-turn limit is keyed per creature: placing a counter on two different
// controlled creatures the same turn draws once for each.
func TestDannyPinkCounterDrawFiresIndependentlyPerCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addDannyPink(g, game.Player1)
	creatureA := addCombatCreaturePermanent(g, game.Player1)
	creatureB := addCombatCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn B"}})

	addCountersToPermanent(g, creatureA, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter on creature A did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after creature A counter = %d, want 1", got)
	}

	addCountersToPermanent(g, creatureB, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter on creature B did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size after creature B counter = %d, want 2", got)
	}
}

// TestDannyPinkCounterDrawCoalescesSimultaneousPlacements proves the granted
// OneOrMore trigger draws once when several counters are placed simultaneously.
func TestDannyPinkCounterDrawCoalescesSimultaneousPlacements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addDannyPink(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	simultaneousID := g.IDGen.Next()
	for range 3 {
		emitEvent(g, game.Event{
			Kind:           game.EventCountersAdded,
			SourceObjectID: creature.ObjectID,
			CardID:         creature.CardInstanceID,
			PermanentID:    creature.ObjectID,
			Controller:     game.Player1,
			CounterKind:    counter.PlusOnePlusOne,
			Amount:         1,
			SimultaneousID: simultaneousID,
		})
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("simultaneous counters did not put the granted trigger on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after simultaneous counters = %d, want 1", got)
	}
}

// TestDannyPinkSecondPlacementSameTurnDoesNotDraw proves the granted trigger
// fires only for the first counter placement each turn on a given creature.
func TestDannyPinkSecondPlacementSameTurnDoesNotDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addDannyPink(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unused"}})

	addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first counter did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after first counter = %d, want 1", got)
	}

	addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second counter the same turn triggered the granted ability again")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after second counter same turn = %d, want 1", got)
	}
}

// TestDannyPinkCounterDrawResetsNextTurn proves the per-turn limit clears at the
// turn boundary, so a new turn's first placement draws again.
func TestDannyPinkCounterDrawResetsNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addDannyPink(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 1"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 2"}})

	addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first-turn counter did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after first turn = %d, want 1", got)
	}

	// The turn boundary resets the per-turn trigger ledger.
	g.TriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)

	addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("next-turn counter did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size after next turn = %d, want 2", got)
	}
}

// TestDannyPinkCounterDrawMatchesAnyCounterKind proves the granted trigger is
// not restricted to +1/+1 counters: a charge counter placement also draws.
func TestDannyPinkCounterDrawMatchesAnyCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addDannyPink(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	addCountersToPermanent(g, creature, counter.Charge, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("charge counter did not put the granted trigger on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after charge counter = %d, want 1", got)
	}
}
