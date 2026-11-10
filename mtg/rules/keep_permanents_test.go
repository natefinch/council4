package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned-corpus Oracle text for Razia's Purification (RAV
// 73bfefd3-bddd-47bb-92f3-9356a7bca637, "Each player chooses three permanents
// they control, then sacrifices the rest."), Planetary Annihilation (EOC
// 8695796f-b6ed-46db-b1b8-07ed0cc5a5e5, "Each player chooses six lands they
// control, then sacrifices the rest. Planetary Annihilation deals 6 damage to
// each creature."), Archfiend of Depravity (E01
// 0ec85db7-c8ec-4730-b692-c140977436aa, "At the beginning of each opponent's
// end step, that player chooses up to two creatures they control, then
// sacrifices the rest.") and Covetous Elegy (CLU
// f8f1f5e8-c1bf-405f-b8e3-655d5558e18d, "Each player chooses up to two
// creatures they control, then sacrifices the rest. Then you create a tapped
// Treasure token for each creature your opponents control."). Scryfall
// /cards/<id>/rulings: Razia 2005-10-01 (a player with fewer than three keeps
// everything; choices in turn order, then one simultaneous sacrifice);
// Archfiend 2014-11-24 (may spare one or none; not targeted); Covetous Elegy
// 2024-02-02 (APNAP choices, later players know earlier choices, one
// simultaneous sacrifice; may choose fewer than two); Planetary Annihilation
// has none. CR 101.4/101.4b (APNAP choices), 608.2c (instructions in order)
// and 701.21a (sacrifice ignores indestructible) apply.
// These fixtures prove the shared KeepPermanents operation; no card is admitted.

func castKeepFixture(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, generic int) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("%s cast failed", def.Name)
	}
}

func keepEventIndex(g *game.Game, kind game.EventKind, permanent *game.Permanent) int {
	for i, event := range g.Events {
		if event.Kind == kind && event.PermanentID == permanent.ObjectID {
			return i
		}
	}
	return -1
}

func TestRaziasPurificationEachPlayerKeepsThreeThenSacrificesTogether(t *testing.T) {
	razia := &game.CardDef{CardFace: game.CardFace{
		Name:     "Razia's Purification",
		ManaCost: opt.Val(cost.Mana{cost.O(4), cost.R, cost.W}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.KeepPermanents{
			PlayerGroup: game.AllPlayersReference(),
			Keep:        game.Fixed(3),
		}}}}.Ability()),
	}}
	deathWatcher := &game.CardDef{CardFace: game.CardFace{
		Name:      "Death Watcher",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			}}}.Ability(),
		}},
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	p1Land := addLandPermanent(g, game.Player1, "Caster Land")
	p1Artifact := addNamedTypedPermanent(g, game.Player1, "Caster Artifact", types.Artifact)
	p1Creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	p1Indestructible := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Indestructible)
	p1Extra := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	watcher := addCombatPermanent(g, game.Player2, deathWatcher)
	p2Land := addLandPermanent(g, game.Player2, "Small Board Land")
	p2Enchantment := addNamedTypedPermanent(g, game.Player2, "Small Board Enchantment", types.Enchantment)
	p3Sacrificed := addCombatCreaturePermanentWithPower(g, game.Player3, 4)
	p3Kept := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player3, 1),
		addLandPermanent(g, game.Player3, "Wide Board Land"),
		addNamedTypedPermanent(g, game.Player3, "Wide Board Artifact", types.Artifact),
	}

	castKeepFixture(t, g, engine, razia, 4)
	p1 := &edictChoiceRecorder{g: g, t: t, choice: []int{0, 1, 2}}
	p2 := &edictChoiceRecorder{g: g, t: t}
	// Player3 chooses only after every earlier choice, while Player1's unkept
	// permanents are still on the battlefield.
	p3 := &edictChoiceRecorder{g: g, t: t, choice: []int{1, 2, 3}, mustRemain: p1Indestructible.ObjectID}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player3] = p1, p2, p3
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, kept := range append([]*game.Permanent{p1Land, p1Artifact, p1Creature, watcher, p2Land, p2Enchantment}, p3Kept...) {
		if !permanentStillPresent(g, kept.ObjectID) {
			t.Fatalf("kept permanent %v was sacrificed", kept.ObjectID)
		}
	}
	for _, unkept := range []*game.Permanent{p1Indestructible, p1Extra, p3Sacrificed} {
		if permanentStillPresent(g, unkept.ObjectID) {
			t.Fatalf("unkept permanent %v survived", unkept.ObjectID)
		}
	}
	if len(p1.requests) != 1 || p1.requests[0].MinChoices != 3 || p1.requests[0].MaxChoices != 3 ||
		len(p1.requests[0].Options) != 5 {
		t.Fatalf("Player1 requests = %#v, want one exactly-three choice among five", p1.requests)
	}
	if len(p2.requests) != 0 {
		t.Fatal("a player with exactly three permanents was asked to choose")
	}
	if len(p3.requests) != 1 {
		t.Fatalf("Player3 requests = %d, want 1", len(p3.requests))
	}
	wantEarlier := []game.KeepDecision{
		{Player: game.Player1, Kept: []id.ID{p1Land.ObjectID, p1Artifact.ObjectID, p1Creature.ObjectID}},
		{Player: game.Player2, Kept: []id.ID{watcher.ObjectID, p2Land.ObjectID, p2Enchantment.ObjectID}},
	}
	if got := p3.requests[0].KeepDecisions; !reflect.DeepEqual(got, wantEarlier) {
		t.Fatalf("Player3 saw earlier decisions %#v, want %#v", got, wantEarlier)
	}
	batches := sacrificeBatchIDs(g, p1Indestructible, p1Extra, p3Sacrificed)
	if len(batches) != 3 || batches[0] == 0 || batches[0] != batches[1] || batches[1] != batches[2] {
		t.Fatalf("sacrifice batches = %v, want one simultaneous batch", batches)
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the creature-dies observer did not trigger")
	}
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("observer triggers = %d, want one per creature that died together (3)", got)
	}
	for range 3 {
		engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	}
	if g.Players[game.Player2].Life != 43 {
		t.Fatalf("observer controller life = %d, want 43", g.Players[game.Player2].Life)
	}
}

func TestPlanetaryAnnihilationKeepsSixLandsThenDamagesCreatures(t *testing.T) {
	planetary := &game.CardDef{CardFace: game.CardFace{
		Name:     "Planetary Annihilation",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R, cost.R}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			{Primitive: game.KeepPermanents{
				PlayerGroup:       game.AllPlayersReference(),
				AffectedSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
				Keep:              game.Fixed(6),
			}},
			{Primitive: game.Damage{
				Amount:    game.Fixed(6),
				Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
			}},
		}}.Ability()),
	}}
	landCreature := &game.CardDef{CardFace: game.CardFace{
		Name:      "Land Creature",
		Types:     []types.Card{types.Land, types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	var p1Lands []*game.Permanent
	for range 8 {
		p1Lands = append(p1Lands, addLandPermanent(g, game.Player1, "Caster Land"))
	}
	p1Creature := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	p1Artifact := addNamedTypedPermanent(g, game.Player1, "Caster Artifact", types.Artifact)
	var p2Lands []*game.Permanent
	for range 6 {
		p2Lands = append(p2Lands, addLandPermanent(g, game.Player2, "Six Lands"))
	}
	p2Creature := addCombatCreaturePermanentWithPower(g, game.Player2, 7)
	p3LandCreature := addCombatPermanent(g, game.Player3, landCreature)
	for range 6 {
		addLandPermanent(g, game.Player3, "Seven Lands")
	}

	// The helper supplies one {R}; add the second.
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	castKeepFixture(t, g, engine, planetary, 1)
	p1 := &edictChoiceRecorder{g: g, t: t, choice: []int{0, 1, 2, 3, 4, 5}}
	p2 := &edictChoiceRecorder{g: g, t: t}
	p3 := &edictChoiceRecorder{g: g, t: t, choice: []int{1, 2, 3, 4, 5, 6}}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player3] = p1, p2, p3
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for i, land := range p1Lands {
		if permanentStillPresent(g, land.ObjectID) != (i < 6) {
			t.Fatalf("Player1 land %d presence = %v, want kept only for the first six", i, !(i < 6))
		}
	}
	for _, land := range p2Lands {
		if !permanentStillPresent(g, land.ObjectID) {
			t.Fatal("a player with exactly six lands lost a land")
		}
	}
	if len(p2.requests) != 0 || len(p1.requests) != 1 || len(p1.requests[0].Options) != 8 {
		t.Fatalf("keep choices: Player1 %#v, Player2 %#v; want only Player1's eight-land choice", p1.requests, p2.requests)
	}
	if !permanentStillPresent(g, p1Artifact.ObjectID) || !permanentStillPresent(g, p1Creature.ObjectID) {
		t.Fatal("a nonland permanent was sacrificed")
	}
	if p1Creature.MarkedDamage != 6 || p2Creature.MarkedDamage != 6 {
		t.Fatalf("creature damage = %d/%d, want 6 each", p1Creature.MarkedDamage, p2Creature.MarkedDamage)
	}
	sacrificed := keepEventIndex(g, game.EventPermanentSacrificed, p3LandCreature)
	if sacrificed < 0 || permanentStillPresent(g, p3LandCreature.ObjectID) {
		t.Fatal("the unkept land creature was not sacrificed")
	}
	if keepEventIndex(g, game.EventDamageDealt, p3LandCreature) >= 0 {
		t.Fatal("the sacrificed land creature was dealt damage, so damage preceded the sacrifice")
	}
	if damaged := keepEventIndex(g, game.EventDamageDealt, p1Creature); damaged < sacrificed {
		t.Fatalf("damage event %d precedes the sacrifice event %d", damaged, sacrificed)
	}
}

func TestArchfiendOfDepravityOpponentKeepsUpToTwoCreatures(t *testing.T) {
	archfiend := &game.CardDef{CardFace: game.CardFace{
		Name:      "Archfiend of Depravity",
		ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Demon},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerAt, Pattern: game.TriggerPattern{
				Event:  game.EventBeginningOfStep,
				Step:   game.StepEnd,
				Player: game.TriggerPlayerOpponent,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.KeepPermanents{
				Player:            game.EventPlayerReference(),
				AffectedSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				Keep:              game.Fixed(2),
				UpTo:              true,
			}}}}.Ability(),
		}},
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, archfiend)
	ownCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	p2Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player2, 1),
		addCombatCreaturePermanentWithPower(g, game.Player2, 2),
		addCombatCreaturePermanentWithPower(g, game.Player2, 3),
	}
	p2Land := addLandPermanent(g, game.Player2, "Opponent Land")
	p3Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player3, 1),
		addCombatCreaturePermanentWithPower(g, game.Player3, 2),
		addCombatCreaturePermanentWithPower(g, game.Player3, 3),
	}
	p4Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player4, 1),
		addCombatCreaturePermanentWithPower(g, game.Player4, 2),
	}

	// Player2 spares only one of three creatures.
	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Archfiend did not trigger on an opponent's end step")
	}
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{1}}
	agents := simPassPolicies()
	agents[game.Player2] = p2
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(p2.requests) != 1 || p2.requests[0].MinChoices != 0 || p2.requests[0].MaxChoices != 2 {
		t.Fatalf("Player2 requests = %#v, want one up-to-two choice", p2.requests)
	}
	if permanentStillPresent(g, p2Creatures[0].ObjectID) || !permanentStillPresent(g, p2Creatures[1].ObjectID) ||
		permanentStillPresent(g, p2Creatures[2].ObjectID) || !permanentStillPresent(g, p2Land.ObjectID) {
		t.Fatal("Player2 did not keep only the chosen creature and their land")
	}
	for _, creature := range append(slicesConcat(p3Creatures, p4Creatures), ownCreature) {
		if !permanentStillPresent(g, creature.ObjectID) {
			t.Fatal("a player other than the end-step opponent lost a creature")
		}
	}

	// Player4 controls only two creatures but may still keep none.
	g.Turn.ActivePlayer = game.Player4
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Archfiend did not trigger on Player4's end step")
	}
	p4 := &edictChoiceRecorder{g: g, t: t, choice: []int{}}
	agents[game.Player4] = p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(p4.requests) != 1 || p4.requests[0].MaxChoices != 2 {
		t.Fatalf("Player4 requests = %#v, want an up-to choice despite controlling only two", p4.requests)
	}
	for _, creature := range p4Creatures {
		if permanentStillPresent(g, creature.ObjectID) {
			t.Fatal("Player4 kept a creature after choosing none")
		}
	}

	g.Turn.ActivePlayer = game.Player1
	emitBeginningOfStepEvent(g, game.StepEnd)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Archfiend triggered on its controller's own end step")
	}
}

func slicesConcat(a, b []*game.Permanent) []*game.Permanent {
	return append(append([]*game.Permanent(nil), a...), b...)
}

func TestCovetousElegyProgramCountsOpponentsCreaturesAfterTheSacrifice(t *testing.T) {
	creatures := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	var keepOutcome cardprogram.Outcome
	reference := game.ProgramReference{ID: "test/covetous-elegy", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			outcome, err := ctx.Execute(game.KeepPermanents{
				PlayerGroup:       game.AllPlayersReference(),
				AffectedSelection: creatures,
				Keep:              game.Fixed(2),
				UpTo:              true,
			})
			if err != nil {
				return err
			}
			keepOutcome = outcome
			opponentCreatures := creatures
			opponentCreatures.Controller = game.ControllerOpponent
			remaining, err := ctx.View().Permanents(game.BattlefieldGroup(opponentCreatures))
			if err != nil || len(remaining) == 0 {
				return err
			}
			_, err = ctx.Execute(game.CreateToken{
				Amount:      game.Fixed(len(remaining)),
				Source:      game.TokenDef(treasureTokenCardDef()),
				EntryTapped: true,
			})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	elegy := &game.CardDef{CardFace: game.CardFace{
		Name:         "Covetous Elegy",
		ManaCost:     opt.Val(cost.Mana{cost.O(4), cost.W, cost.B}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: reference}.Ability()),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, catalog)
	p1Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player1, 1),
		addCombatCreaturePermanentWithPower(g, game.Player1, 2),
		addCombatCreaturePermanentWithPower(g, game.Player1, 3),
	}
	p2Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player2, 1),
		addCombatCreaturePermanentWithPower(g, game.Player2, 2),
		addCombatCreaturePermanentWithPower(g, game.Player2, 3),
	}
	p3Creatures := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player3, 1),
		addCombatCreaturePermanentWithPower(g, game.Player3, 2),
	}
	p4Creature := addCombatCreaturePermanentWithPower(g, game.Player4, 4)
	// Player4's borrowed creature can't be sacrificed, so it survives unkept.
	cantSacrificeControlNotOwnEnchantment(g, game.Player4)
	borrowed := makeCreaturePermanent(g, game.Player1, "Borrowed Beast")
	borrowed.Controller = game.Player4

	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	castKeepFixture(t, g, engine, elegy, 2)
	p1 := &edictChoiceRecorder{g: g, t: t, choice: []int{0, 2}}
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}}
	p3 := &edictChoiceRecorder{g: g, t: t, choice: []int{0, 1}}
	p4 := &edictChoiceRecorder{g: g, t: t, choice: []int{}}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player3], agents[game.Player4] = p1, p2, p3, p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	sacrificed := []*game.Permanent{p1Creatures[1], p2Creatures[1], p2Creatures[2], p4Creature}
	for _, creature := range sacrificed {
		if permanentStillPresent(g, creature.ObjectID) {
			t.Fatal("an unkept creature survived")
		}
	}
	for _, creature := range []*game.Permanent{p1Creatures[0], p1Creatures[2], p2Creatures[0], p3Creatures[0], p3Creatures[1]} {
		if !permanentStillPresent(g, creature.ObjectID) {
			t.Fatal("a kept creature was sacrificed")
		}
	}
	if !permanentStillPresent(g, borrowed.ObjectID) {
		t.Fatal("an unkept creature that can't be sacrificed left the battlefield")
	}
	if len(p4.requests) != 1 || len(p4.requests[0].Options) != 2 {
		t.Fatalf("Player4 requests = %#v, want the protected creature offered as a keep option", p4.requests)
	}
	if len(p3.requests) != 1 || p3.requests[0].MinChoices != 0 {
		t.Fatalf("Player3 requests = %#v, want one up-to choice", p3.requests)
	}
	wantEarlier := []game.KeepDecision{
		{Player: game.Player1, Kept: []id.ID{p1Creatures[0].ObjectID, p1Creatures[2].ObjectID}},
		{Player: game.Player2, Kept: []id.ID{p2Creatures[0].ObjectID}},
	}
	if got := p3.requests[0].KeepDecisions; !reflect.DeepEqual(got, wantEarlier) {
		t.Fatalf("Player3 saw earlier decisions %#v, want %#v", got, wantEarlier)
	}
	if !keepOutcome.Available || !keepOutcome.Succeeded || len(keepOutcome.Objects) != len(sacrificed) {
		t.Fatalf("keep outcome = %#v, want %d sacrificed result objects", keepOutcome, len(sacrificed))
	}
	treasures := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == game.Player1 && permanentHasSubtype(g, permanent, types.Treasure) {
			if !permanent.Tapped {
				t.Fatal("a Treasure entered untapped")
			}
			treasures++
		}
	}
	if treasures != 4 {
		t.Fatalf("Treasures = %d, want 4 for the opponents' creatures remaining after the sacrifice", treasures)
	}
}
