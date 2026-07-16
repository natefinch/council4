package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// pollywogPattern is the Pollywog Prodigy spell-cast trigger: an opponent casts
// a noncreature spell whose mana value is strictly less than this creature's
// live power.
func pollywogPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerOpponent,
		CardSelection: game.Selection{
			ExcludedTypes:                []types.Card{types.Creature},
			ManaValueLessThanSourcePower: true,
		},
	}
}

// pollywogSource builds a creature source with a specific live power that
// carries the Pollywog Prodigy trigger drawing a card.
func pollywogSource(power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Pollywog Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: *pollywogPattern()},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}}.Ability(),
		}},
	}}
}

func TestPollywogSourcePowerManaValueMatching(t *testing.T) {
	pattern := pollywogPattern()
	instant := []types.Card{types.Instant}
	creature := []types.Card{types.Creature}

	cases := []struct {
		name       string
		power      int
		manaValue  int
		controller game.PlayerID
		cardTypes  []types.Card
		want       bool
	}{
		{name: "less than power fires", power: 3, manaValue: 2, controller: game.Player2, cardTypes: instant, want: true},
		{name: "equal to power does not fire", power: 2, manaValue: 2, controller: game.Player2, cardTypes: instant, want: false},
		{name: "greater than power does not fire", power: 2, manaValue: 3, controller: game.Player2, cardTypes: instant, want: false},
		{name: "zero mana value fires below positive power", power: 1, manaValue: 0, controller: game.Player2, cardTypes: instant, want: true},
		{name: "creature spell excluded", power: 3, manaValue: 2, controller: game.Player2, cardTypes: creature, want: false},
		{name: "own cast does not fire", power: 3, manaValue: 2, controller: game.Player1, cardTypes: instant, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatCreaturePermanentWithPower(g, game.Player1, tc.power)
			event := game.Event{
				Kind:       game.EventSpellCast,
				Controller: tc.controller,
				ManaValue:  opt.Val(tc.manaValue),
				CardTypes:  tc.cardTypes,
			}
			if got := triggerMatchesEvent(g, source, pattern, event); got != tc.want {
				t.Fatalf("triggerMatchesEvent = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPollywogSourcePowerUsesLivePower confirms the bound tracks the source's
// current power, not its printed power: a source pumped above the spell's mana
// value fires, and the same source shrunk below it does not.
func TestPollywogSourcePowerUsesLivePower(t *testing.T) {
	pattern := pollywogPattern()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	event := game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		ManaValue:  opt.Val(2),
		CardTypes:  []types.Card{types.Instant},
	}
	// Printed/base power 2 equals the mana value, so the strict comparison fails.
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("expected no match while source power equals spell mana value")
	}
	// Two +1/+1 counters raise live power to 4, above the spell's value.
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("expected match once source live power exceeds spell mana value")
	}
}

// TestPollywogSourcePowerSourceGoneFailsClosed confirms the source-relative
// bound fails closed when the trigger source is no longer a battlefield
// permanent, so a spell cast after the source left does not fire.
func TestPollywogSourcePowerSourceGoneFailsClosed(t *testing.T) {
	pattern := pollywogPattern()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove source from battlefield")
	}
	event := game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		ManaValue:  opt.Val(2),
		CardTypes:  []types.Card{types.Instant},
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("expected no match once trigger source has left the battlefield")
	}
}

// TestPollywogSourcePowerEndToEndDraws casts a real noncreature spell whose mana
// value is below the source's power and confirms the controller draws.
func TestPollywogSourcePowerEndToEndDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, pollywogSource(3))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Cheap Instant",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Types:    []types.Card{types.Instant},
	}}
	spellID := addCardToHand(g, game.Player2, spell)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("casting the mana value 2 instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent noncreature cast did not fire the source-power trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 from source-power trigger draw", got)
	}
}

// TestPollywogSourcePowerCheckedAtCastNotResolution confirms the trigger
// condition is evaluated when the spell is cast, not re-checked at resolution:
// the source leaving the battlefield after the trigger is on the stack does not
// stop the draw (the trigger carries no intervening-if).
func TestPollywogSourcePowerCheckedAtCastNotResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, pollywogSource(3))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Cheap Instant",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Types:    []types.Card{types.Instant},
	}}
	spellID := addCardToHand(g, game.Player2, spell)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("casting the mana value 2 instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent noncreature cast did not fire the source-power trigger")
	}
	// The source leaves after the trigger is already on the stack; the draw
	// resolves regardless because the condition was met at cast time.
	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove source from battlefield")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 draw despite source leaving before resolution", got)
	}
}
