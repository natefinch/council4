package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSpellCastTriggerFiltersCardTypesAndController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerOpponent,
		RequireCardTypes: []types.Card{types.Instant},
		ExcludeCardTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent instant cast trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want spell-cast trigger draw", got)
	}
}

func TestSpellCastTriggerFiltersSubtypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			SubtypesAny: []types.Sub{types.Spirit, types.Arcane},
		},
	}

	event := game.Event{
		Kind:         game.EventSpellCast,
		Controller:   game.Player1,
		CardSubtypes: []types.Sub{types.Spirit},
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("Spirit spell did not match Spirit or Arcane cast trigger")
	}
	event.CardSubtypes = []types.Sub{types.Arcane}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("Arcane spell did not match Spirit or Arcane cast trigger")
	}
	event.CardSubtypes = []types.Sub{types.Wizard}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("Wizard spell matched Spirit or Arcane cast trigger")
	}
}

func TestSpellCastTriggerFiltersHistoric(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:           game.EventSpellCast,
		Controller:      game.TriggerControllerYou,
		RequireHistoric: true,
	}
	tests := []struct {
		name  string
		event game.Event
		want  bool
	}{
		{
			name:  "artifact",
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardTypes: []types.Card{types.Artifact}},
			want:  true,
		},
		{
			name:  "legendary",
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardSupertypes: []types.Super{types.Legendary}},
			want:  true,
		},
		{
			name:  "Saga",
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardSubtypes: []types.Sub{types.Saga}},
			want:  true,
		},
		{
			name:  "nonhistoric",
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardTypes: []types.Card{types.Instant}},
			want:  false,
		},
		{
			name: "missing event types fails closed",
			event: game.Event{
				Kind:       game.EventSpellCast,
				Controller: game.Player1,
				CardID:     addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Artifact", Types: []types.Card{types.Artifact}}}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := triggerMatchesEvent(g, source, pattern, tt.event); got != tt.want {
				t.Fatalf("triggerMatchesEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockedAttackerSubjectMatchesAttachedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	blocker := addCombatCreaturePermanent(g, game.Player2)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment}},
	})
	if !attachPermanent(g, equipment, attacker) {
		t.Fatal("attachPermanent failed")
	}
	event := game.Event{
		Kind:              game.EventBlockerDeclared,
		Controller:        game.Player2,
		PermanentID:       blocker.ObjectID,
		BlockedAttackerID: attacker.ObjectID,
	}
	pattern := &game.TriggerPattern{
		Event:                 game.EventBlockerDeclared,
		Controller:            game.TriggerControllerYou,
		Source:                game.TriggerSourceAttachedPermanent,
		Subject:               game.TriggerSubjectBlockedAttacker,
		RequirePermanentTypes: []types.Card{types.Creature},
	}
	if !triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached equipment did not match blocked attacker subject")
	}
	pattern.Subject = game.TriggerSubjectDefault
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached equipment matched blocker as default subject")
	}
	pattern.Subject = game.TriggerSubjectBlockedAttacker
	nonCreatureBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})
	event.PermanentID = nonCreatureBlocker.ObjectID
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("creature type filter matched blocked attacker instead of blocker")
	}
}

func TestSpellTargetTriggerPredicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	ownCreature := addCombatCreaturePermanent(g, game.Player1)
	opponentCreature := addCombatCreaturePermanent(g, game.Player2)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(source.ObjectID),
			game.PermanentTarget(opponentCreature.ObjectID),
		},
	}
	g.Stack.Push(obj)
	event := game.Event{Kind: game.EventSpellCast, StackObjectID: obj.ID, Controller: game.Player1}
	if !triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:              game.EventSpellCast,
		Controller:         game.TriggerControllerYou,
		SpellTargetsSource: true,
	}, event) {
		t.Fatal("spell-targets-source trigger did not match source target")
	}
	if !triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerYou,
		SpellTargetAllow: game.TargetAllowPermanent,
		SpellTargetPattern: opt.Val(game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
			Controller:     game.ControllerNotYou,
		}),
	}, event) {
		t.Fatal("spell target predicate did not match opponent creature target")
	}
	obj.Targets = []game.Target{game.PermanentTarget(ownCreature.ObjectID)}
	if triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerYou,
		SpellTargetAllow: game.TargetAllowPermanent,
		SpellTargetPattern: opt.Val(game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
			Controller:     game.ControllerNotYou,
		}),
	}, event) {
		t.Fatal("spell target predicate matched own creature target")
	}
}

func TestTriggeredAbilityMaxTriggersPerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].MaxTriggersPerTurn = 1

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size after two same-turn events = %d, want 1", got)
	}
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger exceeded max triggers per turn")
	}
	engine.advanceToNextTurn(g)
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not reset next turn")
	}
}

func TestSpellCastTriggerMatchesColorSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Green},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("green-spell cast trigger did not fire for green instant")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want green-spell trigger to draw one card", got)
	}
}

func TestSpellCastTriggerColorSelectionExcludesWrongColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Blue},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("blue-spell trigger incorrectly fired for green instant")
	}
}

func TestSpellCastTriggerMatchesColorCardinalitySelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Spell Watcher",
		Types: []types.Card{types.Creature},
	}})
	tests := []struct {
		name      string
		selection game.Selection
		colors    []color.Color
		want      bool
	}{
		{
			name:      "colorless matches no colors",
			selection: game.Selection{Colorless: true},
			want:      true,
		},
		{
			name:      "colorless rejects colored",
			selection: game.Selection{Colorless: true},
			colors:    []color.Color{color.Green},
		},
		{
			name:      "multicolored matches two colors",
			selection: game.Selection{Multicolored: true},
			colors:    []color.Color{color.Blue, color.Red},
			want:      true,
		},
		{
			name:      "multicolored rejects monocolored",
			selection: game.Selection{Multicolored: true},
			colors:    []color.Color{color.Blue},
		},
		{
			name:      "multicolored counts distinct colors",
			selection: game.Selection{Multicolored: true},
			colors:    []color.Color{color.Blue, color.Blue},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := triggerMatchesEvent(g, source, &game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				CardSelection: tt.selection,
			}, game.Event{
				Kind:       game.EventSpellCast,
				Controller: game.Player1,
				Colors:     tt.colors,
			})
			if got != tt.want {
				t.Fatalf("triggerMatchesEvent = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpellCastTriggerMatchesManaValueKickerAndSourceZone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Spell Watcher",
		Types: []types.Card{types.Creature},
	}})
	tests := []struct {
		name    string
		pattern game.TriggerPattern
		event   game.Event
		want    bool
	}{
		{
			name: "mana value matches",
			pattern: game.TriggerPattern{
				Event: game.EventSpellCast,
				CardSelection: game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				},
			},
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, ManaValue: opt.Val(5)},
			want:  true,
		},
		{
			name: "mana value rejects lower value",
			pattern: game.TriggerPattern{
				Event: game.EventSpellCast,
				CardSelection: game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				},
			},
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1, ManaValue: opt.Val(4)},
		},
		{
			name: "mana value rejects missing event data",
			pattern: game.TriggerPattern{
				Event: game.EventSpellCast,
				CardSelection: game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				},
			},
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player1},
		},
		{
			name:    "kicked matches",
			pattern: game.TriggerPattern{Event: game.EventSpellCast, RequireKickerPaid: true},
			event:   game.Event{Kind: game.EventSpellCast, Controller: game.Player1, KickerPaid: true},
			want:    true,
		},
		{
			name:    "kicked rejects unkicked",
			pattern: game.TriggerPattern{Event: game.EventSpellCast, RequireKickerPaid: true},
			event:   game.Event{Kind: game.EventSpellCast, Controller: game.Player1},
		},
		{
			name:    "from graveyard matches",
			pattern: game.TriggerPattern{Event: game.EventSpellCast, MatchFromZone: true, FromZone: zone.Graveyard},
			event:   game.Event{Kind: game.EventSpellCast, Controller: game.Player1, FromZone: zone.Graveyard},
			want:    true,
		},
		{
			name:    "from graveyard rejects hand",
			pattern: game.TriggerPattern{Event: game.EventSpellCast, MatchFromZone: true, FromZone: zone.Graveyard},
			event:   game.Event{Kind: game.EventSpellCast, Controller: game.Player1, FromZone: zone.Hand},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := triggerMatchesEvent(g, source, &tt.pattern, tt.event)
			if got != tt.want {
				t.Fatalf("triggerMatchesEvent = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpellCastEventPopulatesColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	var castEvent *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventSpellCast {
			castEvent = &g.Events[i]
			break
		}
	}
	if castEvent == nil {
		t.Fatal("no EventSpellCast found")
	}
	if len(castEvent.Colors) != 1 || castEvent.Colors[0] != color.Green {
		t.Fatalf("EventSpellCast.Colors = %v, want [Green]", castEvent.Colors)
	}
	if !castEvent.ManaValue.Exists || castEvent.ManaValue.Val != 1 {
		t.Fatalf("EventSpellCast.ManaValue = %+v, want 1", castEvent.ManaValue)
	}
	if castEvent.FromZone != zone.Hand || castEvent.ToZone != zone.Stack {
		t.Fatalf("EventSpellCast zones = %v -> %v, want Hand -> Stack", castEvent.FromZone, castEvent.ToZone)
	}
}

func TestSpellCastEventManaValueIncludesChosenX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "X Growth",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.G}),
		Types:    []types.Card{types.Sorcery},
	}})
	for range 5 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 4, nil)) {
		t.Fatal("cast X spell failed")
	}
	for i := range g.Events {
		if g.Events[i].Kind == game.EventSpellCast {
			if !g.Events[i].ManaValue.Exists || g.Events[i].ManaValue.Val != 5 {
				t.Fatalf("EventSpellCast.ManaValue = %+v, want 5", g.Events[i].ManaValue)
			}
			return
		}
	}
	t.Fatal("missing EventSpellCast")
}

func TestKickedSpellCastEventRecordsKickerPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Kicked Spell",
		Types: []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: greenCost().Val}},
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked spell cast failed")
	}
	for i := range g.Events {
		if g.Events[i].Kind == game.EventSpellCast {
			if !g.Events[i].KickerPaid {
				t.Fatal("EventSpellCast.KickerPaid = false, want true")
			}
			return
		}
	}
	t.Fatal("missing EventSpellCast")
}
