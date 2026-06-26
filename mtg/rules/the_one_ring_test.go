package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCastByControllerInterveningCondition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := game.TriggerCondition{
		InterveningIfEventPermanentWasCastByController: true,
	}
	tests := []struct {
		name       string
		controller game.PlayerID
		event      game.Event
		want       bool
	}{
		{
			name:       "controller cast",
			controller: game.Player1,
			event: game.Event{
				EnterWasCast:           true,
				EnterCastController:    game.Player1,
				EnterHasCastController: true,
			},
			want: true,
		},
		{
			name:       "opponent cast",
			controller: game.Player1,
			event: game.Event{
				EnterWasCast:           true,
				EnterCastController:    game.Player2,
				EnterHasCastController: true,
			},
		},
		{
			name:       "put onto battlefield",
			controller: game.Player1,
			event:      game.Event{},
		},
		{
			name:       "copied spell",
			controller: game.Player1,
			event: game.Event{
				EnterCastController:    game.Player1,
				EnterHasCastController: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := triggerInterveningIf(g, nil, test.controller, &trigger, &test.event); got != test.want {
				t.Fatalf("triggerInterveningIf() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestObjectCounterAmountUsesLiveStateAndLKI(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "The One Ring",
		Types: []types.Card{types.Artifact},
	}})
	source.Counters.Add(counter.Burden, 3)
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
	dynamic := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.Burden,
	}
	if got := dynamicAmountValueBeforeLayer(g, opt.Val(obj), game.Player1, dynamic, 0); got != 3 {
		t.Fatalf("live counter amount = %d, want 3", got)
	}

	snapshot := snapshotPermanent(g, source, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	g.Battlefield = nil
	if got := dynamicAmountValueBeforeLayer(g, opt.Val(obj), game.Player1, dynamic, 0); got != 3 {
		t.Fatalf("LKI counter amount = %d, want 3", got)
	}
}

func TestBurdenCounterIsAddedBeforeDrawCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "The One Ring",
		Types: []types.Card{types.Artifact},
	}})
	source.Counters.Add(counter.Burden, 1)
	for range 4 {
		cardID := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
		g.Players[game.Player1].Library.Add(cardID)
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	resolveRingDrawAbility(engine, g, obj)
	if got := source.Counters.Get(counter.Burden); got != 2 {
		t.Fatalf("burden counters = %d, want 2", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("cards drawn = %d, want 2", got)
	}
}

func TestRingDrawAbilityUsesOriginalSourceLKIThroughBlink(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "The One Ring",
		Types: []types.Card{types.Artifact},
	}})
	source.Counters.Add(counter.Burden, 3)
	for range 8 {
		cardID := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
		g.Players[game.Player1].Library.Add(cardID)
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("exiling source failed")
	}
	g.Players[game.Player1].Exile.Remove(source.CardInstanceID)
	reentered := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: source.CardInstanceID,
		Owner:          game.Player1,
		Controller:     game.Player1,
	}
	reentered.Counters.Add(counter.Burden, 7)
	g.Battlefield = append(g.Battlefield, reentered)

	resolveRingDrawAbility(engine, g, obj)
	if got := reentered.Counters.Get(counter.Burden); got != 7 {
		t.Fatalf("re-entered Ring burden counters = %d, want 7", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("cards drawn = %d, want 3 from original Ring LKI", got)
	}
}

func TestRingDrawAbilityUsesOriginalSourceSnapshotWhilePhasedOut(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "The One Ring",
		Types: []types.Card{types.Artifact},
	}})
	source.Counters.Add(counter.Burden, 3)
	for range 4 {
		cardID := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
		g.Players[game.Player1].Library.Add(cardID)
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	source.PhasedOut = true

	resolveRingDrawAbility(engine, g, obj)
	if got := source.Counters.Get(counter.Burden); got != 3 {
		t.Fatalf("phased-out Ring burden counters = %d, want 3", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("cards drawn = %d, want 3 from phased source snapshot", got)
	}
}

func resolveRingDrawAbility(engine *Engine, g *game.Game, obj *game.StackObject) {
	instructions := []game.Instruction{
		{Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.Burden,
		}},
		{Primitive: game.Draw{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:        game.DynamicAmountObjectCounters,
				Object:      game.SourcePermanentReference(),
				CounterKind: counter.Burden,
			}),
			Player: game.ControllerReference(),
		}},
	}
	log := TurnLog{}
	for i := range instructions {
		engine.resolveInstructionWithChoices(g, obj, &instructions[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
}

func TestPlayerProtectionPreventsTargetingAndDamage(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectPlayerProtection,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Protection:     game.ProtectionKeyword{Everything: true},
	})
	source := &game.CardDef{CardFace: game.CardFace{
		Name:   "Threat",
		Colors: []color.Color{color.Red},
	}}
	if !targetProtectedFromSource(g, game.Player2, source, 0, game.PlayerTarget(game.Player1)) {
		t.Fatal("protected player remained a legal target")
	}
	if targetProtectedFromSource(g, game.Player2, source, 0, game.PlayerTarget(game.Player2)) {
		t.Fatal("protection affected the wrong player")
	}
	if got := applyDamageModifications(g, damageEvent{player: game.Player1, amount: 7}); got != 0 {
		t.Fatalf("damage after protection = %d, want 0", got)
	}
}

func TestPlayerProtectionSourceUsesLKI(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Source",
		Colors: []color.Color{color.Blue},
		Types:  []types.Card{types.Creature},
	}})
	snapshot := snapshotPermanent(g, source, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	g.Battlefield = nil
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectPlayerProtection,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Protection: game.ProtectionKeyword{
			FromColors: []color.Color{color.Blue},
		},
	})
	if !playerProtectedFromSource(g, game.Player1, 0, source.ObjectID, nil) {
		t.Fatal("player protection did not use source last-known information")
	}
}

func TestPlayerProtectionExpiresAtControllersNextTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 5
	g.Turn.ActivePlayer = game.Player2
	g.RuleEffects = append(g.RuleEffects,
		game.RuleEffect{
			Kind:        game.RuleEffectPlayerProtection,
			Duration:    game.DurationUntilYourNextTurn,
			ExpiresFor:  game.Player1,
			CreatedTurn: 5,
		},
		game.RuleEffect{
			Kind:        game.RuleEffectPlayerProtection,
			Duration:    game.DurationUntilYourNextTurn,
			ExpiresFor:  game.Player2,
			CreatedTurn: 5,
		},
	)
	expireTurnStartDurations(g)
	if len(g.RuleEffects) != 2 {
		t.Fatalf("effects expired early: %#v", g.RuleEffects)
	}

	g.Turn.TurnNumber = 6
	g.Turn.ActivePlayer = game.Player1
	expireTurnStartDurations(g)
	if len(g.RuleEffects) != 1 || g.RuleEffects[0].ExpiresFor != game.Player2 {
		t.Fatalf("effects after player 1 turn start = %#v, want only player 2 effect", g.RuleEffects)
	}
}
