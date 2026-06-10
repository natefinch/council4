package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludesMutateCast(t *testing.T) {
	fixture := newMutateFixture(t)
	want := action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)

	if !actionsContain(fixture.engine.legalActions(fixture.game, game.Player1), want) {
		t.Fatalf("legal actions do not include Mutate action %+v", want)
	}
	if !fixture.engine.applyAction(fixture.game, game.Player1, want) {
		t.Fatal("applyAction() = false, want true for Mutate")
	}

	obj, ok := fixture.game.Stack.Peek()
	if !ok || !obj.Mutate || obj.MutateTargetID != fixture.target.ObjectID ||
		obj.SourceZone != zone.Hand || len(obj.Targets) != 1 ||
		obj.Targets[0] != game.PermanentTarget(fixture.target.ObjectID) {
		t.Fatalf("Mutate stack object = %+v", obj)
	}
	if !fixture.forest.Tapped {
		t.Fatal("Mutate cost was not paid")
	}
	if fixture.game.Players[game.Player1].Hand.Contains(fixture.mutatorID) {
		t.Fatal("Mutate spell remained in hand after casting")
	}
	assertEvent(t, fixture.game.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == fixture.mutatorID &&
			len(event.Colors) == 1 &&
			event.Colors[0] == color.Green
	})
}

func TestMutateRequiresOwnedNonHumanCreatureTarget(t *testing.T) {
	tests := []struct {
		name   string
		change func(*game.Game, *game.Permanent)
	}{
		{
			name: "Human",
			change: func(g *game.Game, target *game.Permanent) {
				g.CardInstances[target.CardInstanceID].Def.Subtypes = []types.Sub{types.Human}
			},
		},
		{
			name: "not owned",
			change: func(_ *game.Game, target *game.Permanent) {
				target.Owner = game.Player2
			},
		},
		{
			name: "not a creature",
			change: func(g *game.Game, target *game.Permanent) {
				g.CardInstances[target.CardInstanceID].Def.Types = []types.Card{types.Artifact}
			},
		},
		{
			name: "phased out",
			change: func(_ *game.Game, target *game.Permanent) {
				target.PhasedOut = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newMutateFixture(t)
			tt.change(fixture.game, fixture.target)
			mutate := action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)

			if actionsContain(fixture.engine.legalActions(fixture.game, game.Player1), mutate) {
				t.Fatal("illegal Mutate target produced a legal action")
			}
			if fixture.engine.applyAction(fixture.game, game.Player1, mutate) {
				t.Fatal("applyAction() = true for illegal Mutate target")
			}
		})
	}
}

func TestMutateResolutionChoosesOverOrUnder(t *testing.T) {
	tests := []struct {
		name         string
		choice       []int
		wantTopName  string
		wantTopIsNew bool
	}{
		{name: "over", choice: []int{0}, wantTopName: "Mutating Beast", wantTopIsNew: true},
		{name: "under", choice: []int{1}, wantTopName: "Target Beast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newMutateFixture(t)
			objectID := fixture.target.ObjectID
			if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, objectID)) {
				t.Fatal("Mutate cast failed")
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{tt.choice}},
			}

			fixture.engine.resolveTopOfStackWithChoices(fixture.game, agents, &TurnLog{})

			merged, ok := permanentByObjectID(fixture.game, objectID)
			if !ok {
				t.Fatal("Mutate replaced the target permanent object")
			}
			if got := permanentEffectiveName(fixture.game, merged); got != tt.wantTopName {
				t.Fatalf("effective name = %q, want %q", got, tt.wantTopName)
			}
			if got := merged.CardInstanceID == fixture.mutatorID; got != tt.wantTopIsNew {
				t.Fatalf("mutating card is top = %v, want %v", got, tt.wantTopIsNew)
			}
			if len(merged.MergedCards) != 1 {
				t.Fatalf("merged components = %+v, want one lower card", merged.MergedCards)
			}
			if len(permanentEffectiveAbilities(fixture.game, merged)) != 3 {
				t.Fatalf("effective abilities = %d, want abilities from both cards", len(permanentEffectiveAbilities(fixture.game, merged)))
			}
			assertEvent(t, fixture.game.Events, game.EventPermanentMutated, func(event game.Event) bool {
				return event.PermanentID == objectID && event.CardID == fixture.mutatorID
			})
			assertNoEvent(t, fixture.game.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
				return event.CardID == fixture.mutatorID
			})
		})
	}
}

func TestMutateWithIllegalTargetResolvesAsCreature(t *testing.T) {
	fixture := newMutateFixture(t)
	if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("Mutate cast failed")
	}
	if !movePermanentToZone(fixture.game, fixture.target, zone.Graveyard) {
		t.Fatal("failed to remove Mutate target")
	}

	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})

	mutator := permanentWithCardID(fixture.game, fixture.mutatorID)
	if mutator == nil || len(mutator.MergedCards) != 0 {
		t.Fatalf("resolved Mutate spell = %+v, want ordinary creature permanent", mutator)
	}
	assertEvent(t, fixture.game.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == fixture.mutatorID
	})
	assertNoEvent(t, fixture.game.Events, game.EventPermanentMutated, func(event game.Event) bool {
		return event.CardID == fixture.mutatorID
	})
}

func TestMutateTriggerUsesMergedPermanentAbilities(t *testing.T) {
	fixture := newMutateFixture(t)
	addCardToLibrary(fixture.game, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("Mutate cast failed")
	}

	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})
	if !fixture.engine.putTriggeredAbilitiesOnStack(fixture.game) {
		t.Fatal("Mutate trigger was not put on the stack")
	}
	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})

	if got := fixture.game.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want Mutate trigger to draw one card", got)
	}
}

func TestMutatedPermanentCanActivateLowerCardAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	top := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Top Card",
		Types: []types.Card{types.Creature},
	}})
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Lower Card",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}
	g.Turn.PriorityPlayer = game.Player1

	activate := action.ActivateAbility(top.ObjectID, 0, nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("lower-card activated ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatal("lower-card activated ability failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want lower-card ability to draw one card", got)
	}
}

func TestMutatedPermanentMovesEveryCardWithCommanderReplacement(t *testing.T) {
	fixture := newMutateFixture(t)
	commanderID := fixture.target.CardInstanceID
	trackCommanderID(fixture.game, game.Player1, commanderID)
	if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("Mutate cast failed")
	}
	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})
	merged, ok := permanentByObjectID(fixture.game, fixture.target.ObjectID)
	if !ok {
		t.Fatal("merged permanent missing")
	}

	if !movePermanentToZone(fixture.game, merged, zone.Graveyard) {
		t.Fatal("moving merged permanent failed")
	}

	if !fixture.game.Players[game.Player1].Graveyard.Contains(fixture.mutatorID) {
		t.Fatal("top card did not move to graveyard")
	}
	if !fixture.game.Players[game.Player1].CommandZone.Contains(commanderID) {
		t.Fatal("lower commander card did not move to command zone")
	}
	if fixture.game.Players[game.Player1].Graveyard.Contains(commanderID) {
		t.Fatal("lower commander card also moved to graveyard")
	}
}

func TestMutateCanBeCastFromCommandZone(t *testing.T) {
	fixture := newMutateFixture(t)
	player := fixture.game.Players[game.Player1]
	player.Hand.Remove(fixture.mutatorID)
	player.CommandZone.Add(fixture.mutatorID)
	trackCommanderID(fixture.game, game.Player1, fixture.mutatorID)
	mutate := action.CastMutateSpellFromZone(fixture.mutatorID, zone.Command, fixture.target.ObjectID)

	if !actionsContain(fixture.engine.legalActions(fixture.game, game.Player1), mutate) {
		t.Fatal("legal actions do not include command-zone Mutate")
	}
	if !fixture.engine.applyAction(fixture.game, game.Player1, mutate) {
		t.Fatal("command-zone Mutate cast failed")
	}
	obj, ok := fixture.game.Stack.Peek()
	if !ok || obj.SourceZone != zone.Command {
		t.Fatalf("Mutate stack source zone = %v, want command", obj.SourceZone)
	}
	if player.CommanderCastCount != 1 {
		t.Fatalf("commander cast count = %d, want 1", player.CommanderCastCount)
	}
}

func TestMutateTargetOwnershipUsesSpellOwner(t *testing.T) {
	fixture := newMutateFixture(t)
	fixture.game.Players[game.Player1].Hand.Remove(fixture.mutatorID)
	fixture.game.Players[game.Player2].Hand.Add(fixture.mutatorID)
	addBasicLandPermanent(fixture.game, game.Player2, types.Forest)
	fixture.target.Controller = game.Player2
	fixture.game.Turn.ActivePlayer = game.Player2
	fixture.game.Turn.PriorityPlayer = game.Player2
	mutate := action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)

	if !actionsContain(fixture.engine.legalActions(fixture.game, game.Player2), mutate) {
		t.Fatal("caster could not mutate onto a creature owned by the spell owner")
	}
	if !fixture.engine.applyAction(fixture.game, game.Player2, mutate) {
		t.Fatal("Mutate cast by nonowner failed")
	}
}

func TestMutatedPermanentUsesLowerStaticEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Lower Static",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerAbility,
				AffectedSource: true,
				AddKeywords:    []game.Keyword{game.Haste},
			}},
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectCostModifier,
				CostModifier: game.CostModifier{
					Kind:            game.CostModifierSpell,
					GenericIncrease: 1,
				},
			}},
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}

	if !hasKeyword(g, top, game.Haste) {
		t.Fatal("lower component continuous effect did not grant Haste")
	}
	effects := staticRuleEffects(g)
	if len(effects) != 1 || effects[0].SourceCardID != lowerID {
		t.Fatalf("lower component rule effects = %+v, want one sourced by %d", effects, lowerID)
	}
}

func TestMutatedPermanentLowerDiesTriggerUsesLastKnownInformation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Lower Dies",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}

	if !movePermanentToZone(g, top, zone.Graveyard) {
		t.Fatal("moving merged permanent failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("lower component dies trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want lower dies trigger to draw one", got)
	}
	permanentMoves := 0
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.PermanentID == top.ObjectID {
			permanentMoves++
		}
	}
	if permanentMoves != 1 {
		t.Fatalf("permanent zone-change events = %d, want 1", permanentMoves)
	}
}

func TestMergedCommanderComponentDealsCommanderDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	commanderID := addCardToHand(g, game.Player1, mutateCard())
	g.Players[game.Player1].Hand.Remove(commanderID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: commanderID, Owner: game.Player1}}
	trackCommanderID(g, game.Player1, commanderID)

	markPlayerCombatDamage(g, top, game.Player2, 3, &TurnLog{})

	if got := g.Players[game.Player2].CommanderDamage[commanderID]; got != 3 {
		t.Fatalf("commander damage = %d, want 3", got)
	}
	if got := g.Players[game.Player2].CommanderDamage[top.CardInstanceID]; got != 0 {
		t.Fatalf("top noncommander damage = %d, want 0", got)
	}
}

func TestMutateOverFaceDownDisguiseBecomesFaceUp(t *testing.T) {
	fixture := newMutateFixture(t)
	fixture.target.FaceDown = true
	fixture.target.FaceDownFace = game.FaceFront
	fixture.target.FaceDownKind = game.FaceDownDisguise
	if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("Mutate cast failed")
	}

	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})

	if fixture.target.FaceDown {
		t.Fatal("Mutate placed over a face-down creature remained face down")
	}
	if len(fixture.target.MergedCards) != 1 || !fixture.target.MergedCards[0].FaceDown {
		t.Fatalf("lower component face-down state = %+v", fixture.target.MergedCards)
	}
	if got := permanentEffectiveName(fixture.game, fixture.target); got != "Mutating Beast" {
		t.Fatalf("effective name = %q, want Mutating Beast", got)
	}
	if !hasKeyword(fixture.game, fixture.target, game.Ward) {
		t.Fatal("face-down Disguise lower component did not retain visible Ward")
	}
}

func TestCopiedMutateSpellMergesAsTokenComponent(t *testing.T) {
	fixture := newMutateFixture(t)
	card := fixture.game.CardInstances[fixture.mutatorID]
	obj := &game.StackObject{
		ID:             fixture.game.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       fixture.mutatorID,
		Face:           game.FaceFront,
		Controller:     game.Player1,
		Mutate:         true,
		MutateTargetID: fixture.target.ObjectID,
		Copy:           true,
	}

	if got := fixture.engine.resolveMutateSpell(fixture.game, obj, card, card.Def, [game.NumPlayers]PlayerAgent{}, &TurnLog{}); got != "mutated" {
		t.Fatalf("copied Mutate resolution = %q, want mutated", got)
	}
	if !fixture.target.Token || fixture.target.TokenDef == nil || fixture.target.CardInstanceID != 0 {
		t.Fatalf("copied Mutate top component = %+v, want token", fixture.target)
	}
	if !fixture.game.Players[game.Player1].Hand.Contains(fixture.mutatorID) {
		t.Fatal("resolving copied Mutate moved the original card")
	}
}

func TestCopiedMutateTopPreservesLowerDiesTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dies Target",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}})
	mutatorID := addCardToHand(g, game.Player1, mutateCard())
	card := g.CardInstances[mutatorID]
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       mutatorID,
		Face:           game.FaceFront,
		Controller:     game.Player1,
		Mutate:         true,
		MutateTargetID: target.ObjectID,
		Copy:           true,
	}
	if got := engine.resolveMutateSpell(g, obj, card, card.Def, [game.NumPlayers]PlayerAgent{}, &TurnLog{}); got != "mutated" {
		t.Fatalf("copied Mutate resolution = %q, want mutated", got)
	}

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("moving copied merged permanent failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("lower dies trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want original Mutate card plus drawn card", got)
	}
}

func TestLowerTokenComponentDoesNotDuplicatePermanentLeaveEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Lower Token", Types: []types.Card{types.Creature}}}
	top.MergedCards = []game.MergedCard{{TokenDef: token, Owner: game.Player1}}

	if !movePermanentToZone(g, top, zone.Graveyard) {
		t.Fatal("moving merged permanent failed")
	}
	permanentMoves := 0
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.PermanentID == top.ObjectID {
			permanentMoves++
		}
	}
	if permanentMoves != 1 {
		t.Fatalf("permanent zone-change events = %d, want 1", permanentMoves)
	}
}

func TestCopyOfMutatedPermanentIncludesLowerAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Lower Ability",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}

	copied, ok := permanentCopyDef(g, top)
	if !ok {
		t.Fatal("permanentCopyDef() = false")
	}
	if len(copied.ActivatedAbilities) != 1 {
		t.Fatalf("copied activated abilities = %d, want 1", len(copied.ActivatedAbilities))
	}
	if len(copied.StaticAbilities) != 1 {
		t.Fatalf("copied static abilities = %d, want top card's Flying", len(copied.StaticAbilities))
	}
}

func TestAutomaticPaymentUsesLowerManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	top.SummoningSick = false
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Lower Mana",
		Types: []types.Card{types.Creature},
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{ManaColor: mana.G, Amount: game.Fixed(1)},
			}}}.Ability(),
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Green Spell",
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Creature},
	}}

	if !paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   game.Player1,
		SourceZone: zone.Hand,
		Card:       spell,
	}) {
		t.Fatal("automatic payment could not use lower component mana ability")
	}
}

func TestFaceDownLowerComponentIsRevealedWhenLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCombatPermanent(g, game.Player1, mutateTargetCard())
	lowerID := addCardToHand(g, game.Player1, mutateCard())
	g.Players[game.Player1].Hand.Remove(lowerID)
	top.MergedCards = []game.MergedCard{{
		CardInstanceID: lowerID,
		FaceDown:       true,
		FaceDownFace:   game.FaceFront,
		FaceDownKind:   game.FaceDownMorph,
		Owner:          game.Player1,
	}}

	if !movePermanentToZone(g, top, zone.Graveyard) {
		t.Fatal("moving merged permanent failed")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == lowerID && event.Face == game.FaceFront
	})
}

func TestCopiedMutateSpellUsesCopyControllerAsOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, mutateTargetCard())
	mutatorID := addCardToHand(g, game.Player2, mutateCard())
	card := g.CardInstances[mutatorID]
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       mutatorID,
		Face:           game.FaceFront,
		Controller:     game.Player1,
		Mutate:         true,
		MutateTargetID: target.ObjectID,
		Copy:           true,
	}

	if got := engine.resolveMutateSpell(g, obj, card, card.Def, [game.NumPlayers]PlayerAgent{}, &TurnLog{}); got != "mutated" {
		t.Fatalf("copied Mutate resolution = %q, want mutated", got)
	}
	if target.Owner != game.Player1 {
		t.Fatalf("copied Mutate component owner = %d, want copy controller", target.Owner)
	}
}

func TestCopyOfFaceDownMergedPermanentCopiesVisibleCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, mutateTargetCard())
	permanent.FaceDown = true
	permanent.FaceDownKind = game.FaceDownMorph
	lowerID := addCardToHand(g, game.Player1, mutateCard())
	g.Players[game.Player1].Hand.Remove(lowerID)
	permanent.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}

	copied, ok := permanentCopyDef(g, permanent)
	if !ok {
		t.Fatal("permanentCopyDef() = false")
	}
	if copied.Name != "" || copied.AbilityCount() != 0 || !copied.HasType(types.Creature) {
		t.Fatalf("face-down copy = %+v, want nameless creature with no abilities", copied)
	}
	if !copied.Power.Exists || copied.Power.Val.Value != 2 || !copied.Toughness.Exists || copied.Toughness.Val.Value != 2 {
		t.Fatalf("face-down copy power/toughness = %+v/%+v, want 2/2", copied.Power, copied.Toughness)
	}

	permanent.FaceDownKind = game.FaceDownDisguise
	copied, ok = permanentCopyDef(g, permanent)
	if !ok || copied.AbilityCount() != 1 || !copied.HasKeyword(game.Ward) {
		t.Fatalf("face-down Disguise copy = %+v, want visible Ward ability", copied)
	}

	permanent.FaceDown = false
	permanent.MergedCards[0].FaceDown = true
	permanent.MergedCards[0].FaceDownKind = game.FaceDownDisguise
	copied, ok = permanentCopyDef(g, permanent)
	if !ok || !copied.HasKeyword(game.Ward) {
		t.Fatalf("copy with lower face-down Disguise component = %+v, want visible Ward ability", copied)
	}
}

func TestMutateOverPreservesPerTurnAbilityUseIndexes(t *testing.T) {
	fixture := newMutateFixture(t)
	oldActivated := game.ActivatedAbilityUse{SourceID: fixture.target.ObjectID, AbilityIndex: 0}
	oldTriggered := game.TriggeredAbilityUse{SourceID: fixture.target.ObjectID, AbilityIndex: 0}
	fixture.game.ActivatedAbilitiesThisTurn[oldActivated] = true
	fixture.game.TriggeredAbilitiesThisTurn[oldTriggered] = 1
	if !fixture.engine.applyAction(fixture.game, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("Mutate cast failed")
	}

	fixture.engine.resolveTopOfStack(fixture.game, &TurnLog{})

	offset := mutateCard().AbilityCount()
	newActivated := game.ActivatedAbilityUse{SourceID: fixture.target.ObjectID, AbilityIndex: offset}
	newTriggered := game.TriggeredAbilityUse{SourceID: fixture.target.ObjectID, AbilityIndex: offset}
	if fixture.game.ActivatedAbilitiesThisTurn[oldActivated] || !fixture.game.ActivatedAbilitiesThisTurn[newActivated] {
		t.Fatalf("activated ability uses = %+v, want old ability shifted by %d", fixture.game.ActivatedAbilitiesThisTurn, offset)
	}
	if fixture.game.TriggeredAbilitiesThisTurn[oldTriggered] != 0 || fixture.game.TriggeredAbilitiesThisTurn[newTriggered] != 1 {
		t.Fatalf("triggered ability uses = %+v, want old ability shifted by %d", fixture.game.TriggeredAbilitiesThisTurn, offset)
	}
}

type mutateFixture struct {
	game      *game.Game
	engine    *Engine
	mutatorID game.ObjectID
	target    *game.Permanent
	forest    *game.Permanent
}

func newMutateFixture(t *testing.T) mutateFixture {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, mutateTargetCard())
	mutatorID := addCardToHand(g, game.Player1, mutateCard())
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	return mutateFixture{
		game:      g,
		engine:    engine,
		mutatorID: mutatorID,
		target:    target,
		forest:    forest,
	}
}

func mutateTargetCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Beast",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
	}}
}

func mutateCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Mutating Beast",
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		Colors:   []color.Color{color.Green},
		Types:    []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.MutateStaticAbility(cost.Mana{cost.G}),
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentMutated,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}
