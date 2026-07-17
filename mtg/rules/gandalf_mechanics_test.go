package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func gandalfRuleEffect() game.RuleEffect {
	return game.RuleEffect{
		Kind: game.RuleEffectAdditionalTriggerForControlledPermanent,
		TriggerCausePermanentFilters: []game.CharacteristicFilter{
			{Supertypes: []types.Super{types.Legendary}},
			{Types: []types.Card{types.Artifact}},
		},
		TriggerCausePermanentEnters: true,
		TriggerCausePermanentLeaves: true,
	}
}

func gandalfTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Gandalf the White",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Avatar, types.Wizard},
		StaticAbilities: []game.StaticAbility{
			{RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer: game.PlayerYou,
				SpellCharacteristicFilters: []game.CharacteristicFilter{
					{Supertypes: []types.Super{types.Legendary}},
					{Types: []types.Card{types.Artifact}},
				},
			}}},
			{RuleEffects: []game.RuleEffect{gandalfRuleEffect()}},
		},
	}}
}

func gandalfWatcher(event game.EventKind, from, to zone.Type, required []types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Watcher",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:                 event,
				MatchFromZone:         from != zone.None,
				FromZone:              from,
				MatchToZone:           to != zone.None,
				ToZone:                to,
				RequirePermanentTypes: required,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func TestGandalfFlashPermissionIsLiveAndComposable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	gandalf := addCombatPermanent(g, game.Player1, gandalfTestDef())
	legendarySorcery := &game.CardDef{CardFace: game.CardFace{
		Name: "Legendary Sorcery", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Sorcery},
	}}
	artifactCreature := &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact Creature", Types: []types.Card{types.Artifact, types.Creature},
	}}
	ordinarySorcery := &game.CardDef{CardFace: game.CardFace{Name: "Ordinary", Types: []types.Card{types.Sorcery}}}

	if !playerCanCastAsThoughFlash(g, game.Player1, legendarySorcery) ||
		!playerCanCastAsThoughFlash(g, game.Player1, artifactCreature) ||
		playerCanCastAsThoughFlash(g, game.Player1, ordinarySorcery) ||
		playerCanCastAsThoughFlash(g, game.Player2, legendarySorcery) {
		t.Fatal("Gandalf flash filter or controller scope was incorrect")
	}
	gandalf.Controller = game.Player2
	if playerCanCastAsThoughFlash(g, game.Player1, artifactCreature) ||
		!playerCanCastAsThoughFlash(g, game.Player2, artifactCreature) {
		t.Fatal("flash permission did not follow live control")
	}
}

func TestGandalfFlashPermissionAppliesFromOtherwiseCastableZone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, gandalfTestDef())
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Grave Relic", Types: []types.Card{types.Artifact},
	}})
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Players[game.Player1].Graveyard.Add(spellID)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		AffectedCardID: spellID,
		CastFromZone:   zone.Graveyard,
	})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1

	act := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("artifact was not castable from its permitted graveyard zone at instant timing")
	}
}

func TestGandalfDoublesQualifyingEnterAndIgnoresOtherEnter(t *testing.T) {
	for name, tc := range map[string]struct {
		supertypes []types.Super
		cardTypes  []types.Card
		want       int
	}{
		"legendary permanent": {supertypes: []types.Super{types.Legendary}, cardTypes: []types.Card{types.Creature}, want: 2},
		"artifact permanent":  {cardTypes: []types.Card{types.Artifact}, want: 2},
		"ordinary permanent":  {cardTypes: []types.Card{types.Creature}, want: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, gandalfTestDef())
			addCombatPermanent(g, game.Player1, gandalfWatcher(game.EventZoneChanged, zone.None, zone.Battlefield, nil))
			cause := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Cause", Supertypes: tc.supertypes, Types: tc.cardTypes,
			}})
			emitSelfEnter(g, cause)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("enter trigger was not put on the stack")
			}

			if got := g.Stack.Size(); got != tc.want {
				t.Fatalf("stack size = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGandalfUsesEventTimeCharacteristicsForTokenEnter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, gandalfTestDef())
	addCombatPermanent(g, game.Player1, gandalfWatcher(game.EventZoneChanged, zone.None, zone.Battlefield, nil))
	token := &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}}
	if !createTokenPermanentsWithChoices(engine, g, game.Player2, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("artifact token was not created")
	}
	token.Types = []types.Card{types.Creature}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token enter trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 from event-time artifact type", got)
	}
}

func TestGandalfDoublesLeaveAndDieUsingLKI(t *testing.T) {
	for name, tc := range map[string]struct {
		event       game.EventKind
		destination zone.Type
	}{
		"leave": {event: game.EventZoneChanged, destination: zone.Exile},
		"die":   {event: game.EventPermanentDied, destination: zone.Graveyard},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, gandalfTestDef())
			addCombatPermanent(g, game.Player1, gandalfWatcher(tc.event, zone.Battlefield, tc.destination, nil))
			cause := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Departing Relic", Types: []types.Card{types.Artifact},
			}})
			if !movePermanentToZone(g, cause, tc.destination) {
				t.Fatal("qualifying permanent did not leave")
			}

			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("leave trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != 2 {
				t.Fatalf("stack size = %d, want 2", got)
			}
		})
	}
}

func TestGandalfStacksAndRequiresControlledPermanentTriggerSource(t *testing.T) {
	for name, tc := range map[string]struct {
		doublers         int
		sourceController game.PlayerID
		want             int
	}{
		"two Gandalfs stack":                    {doublers: 2, sourceController: game.Player1, want: 3},
		"opponent permanent source not doubled": {doublers: 2, sourceController: game.Player2, want: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for range tc.doublers {
				addCombatPermanent(g, game.Player1, gandalfTestDef())
			}
			addCombatPermanent(g, tc.sourceController, gandalfWatcher(game.EventZoneChanged, zone.None, zone.Battlefield, nil))
			cause := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Legend", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Creature},
			}})
			emitSelfEnter(g, cause)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.want {
				t.Fatalf("stack size = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGandalfUsesTriggerSourceControllerAtEventTime(t *testing.T) {
	for name, tc := range map[string]struct {
		eventController game.PlayerID
		laterController game.PlayerID
		want            int
	}{
		"controlled when triggered":     {eventController: game.Player1, laterController: game.Player2, want: 2},
		"not controlled when triggered": {eventController: game.Player2, laterController: game.Player1, want: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, gandalfTestDef())
			watcher := addCombatPermanent(g, tc.eventController, gandalfWatcher(game.EventZoneChanged, zone.None, zone.Battlefield, nil))
			cause := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Legend", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Creature},
			}})
			emitSelfEnter(g, cause)
			watcher.Controller = tc.laterController

			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.want {
				t.Fatalf("stack size = %d, want %d", got, tc.want)
			}
			top, ok := g.Stack.Peek()
			if !ok {
				t.Fatal("trigger missing from stack")
			}
			if top.Controller != tc.eventController {
				t.Fatalf("trigger controller = %v, want event-time controller %v", top.Controller, tc.eventController)
			}
		})
	}
}

func TestGandalfUsesEventTimeControllerWhenMatchingTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, gandalfTestDef())
	watcherDef := gandalfWatcher(game.EventZoneChanged, zone.None, zone.Battlefield, nil)
	watcherDef.TriggeredAbilities[0].Trigger.Pattern.Controller = game.TriggerControllerYou
	watcher := addCombatPermanent(g, game.Player1, watcherDef)
	cause := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Legend", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Creature},
	}})
	emitSelfEnter(g, cause)
	watcher.Controller = game.Player2

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controller-relative trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2", got)
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Controller != game.Player1 {
		t.Fatalf("trigger controller = %v, want event-time controller Player1", top.Controller)
	}
}

func TestGandalfSBADeathsUseOnePreBatchDoublerSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	one := game.PT{Value: 1}
	two := game.PT{Value: 2}

	gandalfDef := gandalfTestDef()
	gandalfDef.Power = opt.Val(one)
	gandalfDef.Toughness = opt.Val(one)
	gandalf := addCombatPermanent(g, game.Player1, gandalfDef)

	watcherDef := gandalfWatcher(game.EventPermanentDied, zone.None, zone.None, []types.Card{types.Artifact})
	watcherDef.Power = opt.Val(two)
	watcherDef.Toughness = opt.Val(two)
	addCombatPermanent(g, game.Player1, watcherDef)

	artifactDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Relic Creature", Types: []types.Card{types.Artifact, types.Creature},
		Power: opt.Val(one), Toughness: opt.Val(one),
	}}
	artifact := addCombatPermanent(g, game.Player1, artifactDef)
	gandalf.MarkedDamage = 1
	artifact.MarkedDamage = 1

	engine.applyStateBasedActions(g)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("artifact SBA death trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 when Gandalf is first in the SBA death batch", got)
	}
}

func TestGandalfOneOrMoreTriggerUsesAnyQualifyingCauseInBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, gandalfTestDef())
	watcherDef := gandalfWatcher(game.EventPermanentDied, zone.None, zone.None, nil)
	watcherDef.TriggeredAbilities[0].Trigger.Pattern.OneOrMore = true
	addCombatPermanent(g, game.Player1, watcherDef)
	ordinary := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Ordinary Creature", Types: []types.Card{types.Creature},
	}})
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact Creature", Types: []types.Card{types.Artifact, types.Creature},
	}})

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{ordinary, artifact}, zone.Graveyard) {
		t.Fatal("simultaneous deaths failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more death trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger plus one Gandalf occurrence", got)
	}
}

func TestGandalfLeavingSimultaneouslyStillDoublesArtifactDiesTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gandalf := addCombatPermanent(g, game.Player1, gandalfTestDef())
	addCombatPermanent(g, game.Player1, gandalfWatcher(game.EventPermanentDied, zone.None, zone.None, []types.Card{types.Artifact}))
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Relic Creature", Types: []types.Card{types.Artifact, types.Creature},
	}})
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{gandalf, artifact}, zone.Graveyard) {
		t.Fatal("simultaneous deaths failed")
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("artifact dies trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 with departed Gandalf snapshot", got)
	}
}

func TestGandalfDoublesOwnEnterTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := gandalfTestDef()
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event:       game.EventZoneChanged,
			Source:      game.TriggerSourceSelf,
			MatchToZone: true,
			ToZone:      zone.Battlefield,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}}
	gandalf := addCombatPermanent(g, game.Player1, def)
	emitSelfEnter(g, gandalf)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Gandalf's own enter trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 for Gandalf's own trigger", got)
	}
}

func TestGandalfUsesActualDestinationAfterLeaveReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, gandalfTestDef())
	addCombatPermanent(g, game.Player1, gandalfWatcher(game.EventZoneChanged, zone.Battlefield, zone.Exile, nil))
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, true)
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Redirected Relic", Types: []types.Card{types.Artifact},
	}})
	if !movePermanentToZone(g, artifact, zone.Graveyard) {
		t.Fatal("artifact did not leave through replacement")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("replaced leave trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 for artifact leaving to replacement destination", got)
	}
}

func TestGandalfDoublesSelfDiesTriggerWhenSourceLeavesSimultaneously(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gandalf := addCombatPermanent(g, game.Player1, gandalfTestDef())
	artifactDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Mourning Relic",
		Types: []types.Card{types.Artifact, types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	artifact := addCombatPermanent(g, game.Player1, artifactDef)
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{gandalf, artifact}, zone.Graveyard) {
		t.Fatal("simultaneous deaths failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("artifact self-dies trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 for departed controlled source", got)
	}
}
