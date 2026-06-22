package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
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

func TestSpellCastTriggerFiltersChosenType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Zombie},
	}
	pattern := &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypeChoice: game.SubtypeChoiceSourceEntry,
		},
	}

	event := game.Event{
		Kind:         game.EventSpellCast,
		Controller:   game.Player1,
		CardTypes:    []types.Card{types.Creature},
		CardSubtypes: []types.Sub{types.Zombie},
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("Zombie creature spell did not match chosen-type cast trigger")
	}
	event.CardSubtypes = []types.Sub{types.Elf}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("Elf creature spell matched Zombie chosen-type cast trigger")
	}
	event.CardTypes = []types.Card{types.Instant}
	event.CardSubtypes = []types.Sub{types.Zombie}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("noncreature spell matched chosen-type creature cast trigger")
	}
}

func TestSpellCastTriggerChosenTypeFailsWithoutEntryChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			SubtypeChoice: game.SubtypeChoiceSourceEntry,
		},
	}
	event := game.Event{
		Kind:         game.EventSpellCast,
		Controller:   game.Player1,
		CardTypes:    []types.Card{types.Creature},
		CardSubtypes: []types.Sub{types.Zombie},
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("chosen-type trigger matched when the source recorded no entry choice")
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

func TestSpellCastTriggerColorUnionMatchesEitherColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Blue, color.Black},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	blackSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Dark Ritual",
		ManaCost: opt.Val(cost.Mana{cost.B}),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Black},
	}}
	spellID := addCardToHand(g, game.Player1, blackSpell)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast black instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("blue-or-black union trigger did not fire for black instant")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want union trigger to draw one card", got)
	}
}

func TestSpellCastTriggerColorUnionExcludesUnlistedColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Blue, color.Black},
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
		t.Fatal("blue-or-black union trigger incorrectly fired for green instant")
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

// instantSorceryCopyTrigger is a magecraft-style "Whenever you cast or copy an
// instant or sorcery spell, draw a card" pattern.
func instantSorceryCopyTrigger() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:          game.EventSpellCast,
		Controller:     game.TriggerControllerYou,
		MatchSpellCopy: true,
		CardSelection:  game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
	}
}

// instantSorceryCastOnlyTrigger is an ordinary "Whenever you cast an instant or
// sorcery spell, draw a card" pattern that must ignore spell copies.
func instantSorceryCastOnlyTrigger() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:         game.EventSpellCast,
		Controller:    game.TriggerControllerYou,
		CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
	}
}

func TestSpellCopyTriggerMatchesCastOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	castEvent := game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardTypes: []types.Card{types.Instant}}
	copyEvent := game.Event{Kind: game.EventSpellCopied, Controller: game.Player1, CardTypes: []types.Card{types.Instant}}

	if !triggerMatchesEvent(g, source, instantSorceryCopyTrigger(), castEvent) {
		t.Fatal("cast-or-copy trigger did not match EventSpellCast")
	}
	if !triggerMatchesEvent(g, source, instantSorceryCopyTrigger(), copyEvent) {
		t.Fatal("cast-or-copy trigger did not match EventSpellCopied")
	}
	if !triggerMatchesEvent(g, source, instantSorceryCastOnlyTrigger(), castEvent) {
		t.Fatal("cast-only trigger did not match EventSpellCast")
	}
	if triggerMatchesEvent(g, source, instantSorceryCastOnlyTrigger(), copyEvent) {
		t.Fatal("cast-only trigger matched EventSpellCopied, want fail-closed")
	}
}

func TestSpellCopyTriggerFiresForStormCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	}
	magecraft := addTriggeredPermanent(g, game.Player1, instantSorceryCopyTrigger(),
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	castOnly := addTriggeredPermanent(g, game.Player1, instantSorceryCastOnlyTrigger(),
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	stormID := addCardToHand(g, game.Player1, stormGainLifeInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
		t.Fatal("first spell cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	// Ignore triggers from the first cast; only inspect the storm cast and its
	// single copy.
	g.TriggerEventCursor = len(g.Events)
	castEventsBefore := spellCastEventCount(g)

	if !engine.applyAction(g, game.Player1, action.CastSpell(stormID, nil, 0, nil)) {
		t.Fatal("storm spell cast failed")
	}

	if got := spellCopiedEventCount(g); got != 1 {
		t.Fatalf("EventSpellCopied count = %d, want 1 storm copy", got)
	}
	if got := spellCastEventCount(g) - castEventsBefore; got != 1 {
		t.Fatalf("new EventSpellCast count = %d, want only the storm cast (copies excluded)", got)
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("storm cast did not put any triggered abilities on stack")
	}

	magecraftTriggers, castOnlyTriggers := 0, 0
	for _, obj := range g.Stack.Objects() {
		if obj.Kind != game.StackTriggeredAbility {
			continue
		}
		switch obj.SourceID {
		case magecraft.ObjectID:
			magecraftTriggers++
		case castOnly.ObjectID:
			castOnlyTriggers++
		default:
		}
	}
	if magecraftTriggers != 2 {
		t.Fatalf("magecraft triggers = %d, want 2 (one cast, one copy)", magecraftTriggers)
	}
	if castOnlyTriggers != 1 {
		t.Fatalf("cast-only triggers = %d, want 1 (cast only, ignoring copy)", castOnlyTriggers)
	}
}

func spellCopiedEventCount(g *game.Game) int {
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCopied {
			count++
		}
	}
	return count
}

// TestSpellCastOrdinalTriggerFiresOnNthSpell verifies an ordinal cast trigger
// ("Whenever you cast your second spell each turn") fires only on the
// controller's second spell of the turn, and that EventSpellCast carries the
// per-turn ordinal.
func TestSpellCastOrdinalTriggerFiresOnNthSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                      game.EventSpellCast,
		Controller:                 game.TriggerControllerYou,
		PlayerEventOrdinalThisTurn: 2,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	for range 5 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	}
	spellIDs := make([]id.ID, 3)
	for i := range spellIDs {
		spellIDs[i] = addCardToHand(g, game.Player1, greenInstant())
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	drawsAfter := make([]int, 3)
	for i, spellID := range spellIDs {
		if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
			t.Fatalf("cast spell %d failed", i+1)
		}
		// Put any triggered abilities on the stack, then resolve everything so
		// the ordinal trigger's draw is reflected before the next cast.
		engine.putTriggeredAbilitiesOnStack(g)
		for !g.Stack.IsEmpty() {
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.putTriggeredAbilitiesOnStack(g)
		}
		drawsAfter[i] = countSpellCastOrdinalDraws(g)
	}

	if drawsAfter[0] != 0 {
		t.Fatalf("after first spell: draws = %d, want 0", drawsAfter[0])
	}
	if drawsAfter[1] != 1 {
		t.Fatalf("after second spell: draws = %d, want 1 (ordinal trigger fires)", drawsAfter[1])
	}
	if drawsAfter[2] != 1 {
		t.Fatalf("after third spell: draws = %d, want 1 (ordinal trigger does not refire)", drawsAfter[2])
	}

	var ordinals []int
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.Controller == game.Player1 {
			ordinals = append(ordinals, event.PlayerEventOrdinalThisTurn)
		}
	}
	if want := []int{1, 2, 3}; !slices.Equal(ordinals, want) {
		t.Fatalf("spell-cast ordinals = %v, want %v", ordinals, want)
	}
}

// TestSpellCastAnyPlayerOrdinalTriggerFiresOnOpponentNthSpell verifies a
// non-controller ordinal cast trigger ("Whenever a player casts their second
// spell each turn", TriggerControllerAny) fires on an opponent's second spell of
// the turn, exercising the broadened actor scope for per-turn spell ordinals.
func TestSpellCastAnyPlayerOrdinalTriggerFiresOnOpponentNthSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                      game.EventSpellCast,
		Controller:                 game.TriggerControllerAny,
		PlayerEventOrdinalThisTurn: 2,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	spellIDs := make([]id.ID, 2)
	for i := range spellIDs {
		spellIDs[i] = addCardToHand(g, game.Player2, greenInstant())
		addBasicLandPermanent(g, game.Player2, types.Forest)
	}
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	drawsAfter := make([]int, 2)
	for i, spellID := range spellIDs {
		if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
			t.Fatalf("cast spell %d failed", i+1)
		}
		engine.putTriggeredAbilitiesOnStack(g)
		for !g.Stack.IsEmpty() {
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.putTriggeredAbilitiesOnStack(g)
		}
		drawsAfter[i] = g.Players[game.Player1].Hand.Size()
	}

	if drawsAfter[0] != 0 {
		t.Fatalf("after opponent's first spell: Player1 hand = %d, want 0", drawsAfter[0])
	}
	if drawsAfter[1] != 1 {
		t.Fatalf("after opponent's second spell: Player1 hand = %d, want 1 (ordinal trigger fires)", drawsAfter[1])
	}
}

func countSpellCastOrdinalDraws(g *game.Game) int {
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventCardDrawn && event.Player == game.Player1 {
			count++
		}
	}
	return count
}
