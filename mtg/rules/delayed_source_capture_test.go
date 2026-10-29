package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	ocards "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Oracle text from the pinned corpus for Hobbling Zombie
// (68284193-a38e-40d2-8d15-92d223756751), Mardu Blazebringer
// (cbf7a797-f32a-4ed2-b835-a356120f5817), Dark Maze
// (74fe0349-a114-4bb8-a9ae-97d0f8f46ddb), Zektar Shrine Expedition
// (f12d2c01-0ef2-40e6-8c7d-8512e2c8c074) and Optimus Prime, Hero // Optimus
// Prime, Autobot Leader (f04ed2cc-aae1-400b-9cc5-5d4e3e3dd714).
// https://api.scryfall.com/cards/<id>/rulings returned four 2021-09-24 decayed
// rulings for Hobbling Zombie ("Once a creature with decayed attacks, it will be
// sacrificed at end of combat, even if it no longer has decayed"), none for
// Mardu Blazebringer or Dark Maze, three 2024-11-08 landfall rulings for Zektar
// and the 2022-10-14 Optimus rulings (convert works like transform; bolster
// chooses on resolution). CR 2026-09-25 from the official TXT
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 603.7a/603.7c (a delayed trigger acts on the particular object it refers to;
// an object that left its zone and returned is a new object it won't affect),
// 400.7 (zone change makes a new object), 511.2 (end-of-combat triggers),
// 702.147a (decayed), 701.21a (sacrifice), 701.27a/701.28a (transform/convert)
// and 111.1 (tokens). Each delayed trigger below binds the exact permanent it
// was created for, keeps that binding observable after the permanent leaves,
// and never acts on a later object for the same card.

func delayedViewBoundTo(views []DelayedTriggerView, objectID id.ID) int {
	count := 0
	for _, view := range views {
		if slices.Contains(view.BoundObjectIDs, objectID) {
			count++
		}
	}
	return count
}

func registryCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup(name)
	if def == nil {
		t.Fatalf("%s is not in the default registry", name)
	}
	return def
}

// hobblingZombieDecayedToken lets the real Hobbling Zombie die so its dies
// trigger creates the real decayed Zombie token, then attacks with that token
// so its decayed attack trigger schedules the end-of-combat sacrifice.
func hobblingZombieDecayedToken(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	zombie := addCombatPermanent(g, game.Player1, registryCard(t, "Hobbling Zombie"))
	if !movePermanentToZone(g, zombie, zone.Graveyard) {
		t.Fatal("could not put Hobbling Zombie into the graveyard")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Hobbling Zombie's dies trigger did not trigger")
	}
	resolveStackWithTriggers(engine, g, agents)
	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			token = permanent
		}
	}
	if token == nil {
		t.Fatal("Hobbling Zombie did not create its decayed Zombie token")
	}
	token.SummoningSick = false
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{}
	declareAttack(t, g, engine, token, game.Player2)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the token's decayed attack trigger did not trigger")
	}
	resolveStackWithTriggers(engine, g, agents)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want the decayed end-of-combat sacrifice", len(g.DelayedTriggers))
	}
	return g, engine, token
}

func fireEndOfCombat(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepEndOfCombat
	emitBeginningOfStepEvent(g, game.StepEndOfCombat)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end-of-combat delayed trigger did not trigger")
	}
}

func fireEndStep(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the next-end-step delayed trigger did not trigger")
	}
}

// The decayed token's end-of-combat sacrifice (CR 702.147a) is bound to that
// token while pending, on the token's own view, and on the stack, and it
// sacrifices that token.
func TestDecayedTokenSacrificeBindsAndSacrificesToken(t *testing.T) {
	g, engine, token := hobblingZombieDecayedToken(t)
	obs := observe(g, game.Player2)
	if delayedViewBoundTo(obs.DelayedTriggers(), token.ObjectID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the sacrifice bound to token %d", obs.DelayedTriggers(), token.ObjectID)
	}
	if delayedViewBoundTo(findPermanentView(t, obs, token.ObjectID).PendingDelayedTriggers, token.ObjectID) != 1 {
		t.Fatal("the token's view does not list its bound decayed sacrifice")
	}
	fireEndOfCombat(t, g, engine)
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, token.ObjectID) {
		t.Fatalf("stack = %+v, want the decayed sacrifice bound to token %d", stack, token.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(token.ObjectID); ok {
		t.Fatal("the decayed token was not sacrificed at end of combat")
	}
}

// After the decayed token leaves the battlefield the pending sacrifice still
// names it (CR 603.7c), so an observer can tell what it was for even though its
// source is gone; the fired trigger keeps that binding and does nothing.
func TestDecayedTokenSacrificeKeepsBindingAfterTokenLeaves(t *testing.T) {
	g, engine, token := hobblingZombieDecayedToken(t)
	tokenID := token.ObjectID
	if !movePermanentToZone(g, token, zone.Graveyard) {
		t.Fatal("could not remove the decayed token")
	}
	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || views[0].SourceObjectID != 0 || !slices.Contains(views[0].BoundObjectIDs, tokenID) {
		t.Fatalf("pending delayed triggers = %+v, want the sacrifice still bound to departed token %d", views, tokenID)
	}
	fireEndOfCombat(t, g, engine)
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, tokenID) {
		t.Fatalf("stack = %+v, want the fired sacrifice bound to departed token %d", stack, tokenID)
	}
	before := len(g.Battlefield)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != before {
		t.Fatal("the departed token's sacrifice affected another permanent")
	}
}

// Two Mardu Blazebringers attack; each "sacrifice it at end of combat" trigger
// is bound to its own Blazebringer, so observers can tell the same-name triggers
// apart and each sacrifices only its own creature.
func TestMarduBlazebringerSacrificesAreBoundPerPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	def := registryCard(t, "Mardu Blazebringer")
	first := addCombatPermanent(g, game.Player1, def)
	second := addCombatPermanent(g, game.Player1, def)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("could not attack with both Mardu Blazebringers")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the Blazebringers' attack triggers did not trigger")
	}
	resolveStackWithTriggers(engine, g, agents)
	obs := observe(g, game.Player3)
	views := obs.DelayedTriggers()
	if len(views) != 2 || delayedViewBoundTo(views, first.ObjectID) != 1 || delayedViewBoundTo(views, second.ObjectID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want one sacrifice bound to each Blazebringer", views)
	}
	for _, permanent := range []*game.Permanent{first, second} {
		pending := findPermanentView(t, obs, permanent.ObjectID).PendingDelayedTriggers
		if len(pending) != 1 || !slices.Contains(pending[0].BoundObjectIDs, permanent.ObjectID) {
			t.Fatalf("Blazebringer %d pending = %+v, want only its own sacrifice", permanent.ObjectID, pending)
		}
	}
	fireEndOfCombat(t, g, engine)
	stack := observe(g, game.Player3).Stack()
	if len(stack) != 2 {
		t.Fatalf("stack = %+v, want both sacrifices", stack)
	}
	topObject, _ := g.Stack.Peek()
	var top StackObjectView
	for _, view := range stack {
		if view.ID == topObject.ID {
			top = view
		}
	}
	if len(top.BoundObjectIDs) != 1 {
		t.Fatalf("top sacrifice bindings = %v, want exactly one Blazebringer", top.BoundObjectIDs)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	survivor := first
	if top.BoundObjectIDs[0] == first.ObjectID {
		survivor = second
	}
	if _, ok := g.PermanentByID(top.BoundObjectIDs[0]); ok {
		t.Fatal("the top sacrifice did not sacrifice its bound Blazebringer")
	}
	if _, ok := g.PermanentByID(survivor.ObjectID); !ok {
		t.Fatal("the top sacrifice sacrificed the other Blazebringer")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(survivor.ObjectID); ok {
		t.Fatal("the second sacrifice did not sacrifice its Blazebringer")
	}
}

func activateDarkMaze(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	maze := addCombatPermanent(g, game.Player1, registryCard(t, "Dark Maze"))
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(maze.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
		t.Fatal("could not activate Dark Maze")
	}
	resolveStackWithTriggers(engine, g, agents)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want Dark Maze's exile", len(g.DelayedTriggers))
	}
	return g, engine, maze
}

// Dark Maze's "Exile it at the beginning of the next end step." is bound to
// the activated Dark Maze and exiles it.
func TestDarkMazeExilesItselfAtNextEndStep(t *testing.T) {
	g, engine, maze := activateDarkMaze(t)
	obs := observe(g, game.Player2)
	if delayedViewBoundTo(findPermanentView(t, obs, maze.ObjectID).PendingDelayedTriggers, maze.ObjectID) != 1 {
		t.Fatalf("Dark Maze pending = %+v, want its exile bound to it", obs.DelayedTriggers())
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(maze.ObjectID); ok {
		t.Fatal("Dark Maze was not exiled at the next end step")
	}
	if !g.Players[game.Player1].Exile.Contains(maze.CardInstanceID) {
		t.Fatal("Dark Maze is not in exile")
	}
}

// If Dark Maze leaves and returns before the end step it is a new object
// (CR 400.7), and the delayed exile created for the old object does not exile
// it (CR 603.7c), even though both are the same card.
func TestDarkMazeExileDoesNotFollowCardToNewObject(t *testing.T) {
	g, engine, maze := activateDarkMaze(t)
	oldID := maze.ObjectID
	if !movePermanentToZone(g, maze, zone.Exile) {
		t.Fatal("could not exile Dark Maze")
	}
	card, ok := g.GetCardInstance(maze.CardInstanceID)
	if !ok {
		t.Fatal("Dark Maze card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == oldID {
		t.Fatal("Dark Maze did not return as a new object")
	}
	obs := observe(g, game.Player2)
	if delayedViewBoundTo(obs.DelayedTriggers(), oldID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the exile still bound to the old object %d", obs.DelayedTriggers(), oldID)
	}
	if pending := findPermanentView(t, obs, returned.ObjectID).PendingDelayedTriggers; len(pending) != 0 {
		t.Fatalf("returned Dark Maze pending = %+v, want none", pending)
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(returned.ObjectID); !ok {
		t.Fatal("the old object's delayed exile exiled the returned Dark Maze")
	}
}

// Zektar Shrine Expedition's "Create a 7/1 ... token ... Exile it at the
// beginning of the next end step." exiles the token it created, though the
// enchantment was sacrificed to activate the ability.
func TestZektarShrineExpeditionExilesItsElementalToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	expedition := addCombatPermanent(g, game.Player1, registryCard(t, "Zektar Shrine Expedition"))
	expedition.Counters.Add(counter.Quest, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(expedition.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
		t.Fatal("could not activate Zektar Shrine Expedition")
	}
	if _, ok := g.PermanentByID(expedition.ObjectID); ok {
		t.Fatal("Zektar Shrine Expedition was not sacrificed as a cost")
	}
	resolveStackWithTriggers(engine, g, agents)
	var elemental *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			elemental = permanent
		}
	}
	if elemental == nil {
		t.Fatal("Zektar Shrine Expedition did not create its Elemental token")
	}
	obs := observe(g, game.Player2)
	if delayedViewBoundTo(findPermanentView(t, obs, elemental.ObjectID).PendingDelayedTriggers, elemental.ObjectID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the exile bound to Elemental %d", obs.DelayedTriggers(), elemental.ObjectID)
	}
	fireEndStep(t, g, engine)
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, elemental.ObjectID) {
		t.Fatalf("stack = %+v, want the exile bound to Elemental %d", stack, elemental.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(elemental.ObjectID); ok {
		t.Fatal("the Elemental token was not exiled at the next end step")
	}
}

// scheduleOptimusConvert resolves the real back face's attack trigger so it
// bolsters the lone low-toughness creature and schedules "When that creature
// deals combat damage to a player this turn, convert Optimus Prime."
func scheduleOptimusConvert(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	optimus := newOptimusBack(g, game.Player1)
	chosen := addCreatureWithPowerToughness(g, game.Player1, 2, 1)
	content := ocards.OptimusPrimeHero().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, optimusObject(g, optimus), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want the combat-damage convert", len(g.DelayedTriggers))
	}
	return g, engine, optimus, chosen
}

func emitOptimusChosenCombatDamage(g *game.Game, source *game.Permanent) {
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  source.ObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
}

// The event-based convert is bound to both the bolstered creature whose damage
// it waits for and the Optimus Prime it converts, while pending, on Optimus's
// view and on the stack.
func TestOptimusConvertBindsChosenCreatureAndOptimus(t *testing.T) {
	g, engine, optimus, chosen := scheduleOptimusConvert(t)
	obs := observe(g, game.Player2)
	views := obs.DelayedTriggers()
	if len(views) != 1 || !views[0].EventBased || !slices.Contains(views[0].BoundObjectIDs, chosen.ObjectID) || !slices.Contains(views[0].BoundObjectIDs, optimus.ObjectID) {
		t.Fatalf("pending delayed triggers = %+v, want the convert bound to creature %d and Optimus %d", views, chosen.ObjectID, optimus.ObjectID)
	}
	if delayedViewBoundTo(findPermanentView(t, obs, optimus.ObjectID).PendingDelayedTriggers, optimus.ObjectID) != 1 {
		t.Fatal("Optimus's view does not list its pending convert")
	}
	emitOptimusChosenCombatDamage(g, chosen)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the bound creature's combat damage did not fire the convert")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, optimus.ObjectID) {
		t.Fatalf("stack = %+v, want the convert bound to Optimus %d", stack, optimus.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if optimus.Face != game.FaceFront || optimus.Transformed {
		t.Fatalf("Optimus face/transformed = %v/%v, want converted to its front face", optimus.Face, optimus.Transformed)
	}
}

// If Optimus Prime leaves and returns before the bolstered creature deals
// combat damage, the returned Optimus is a new object (CR 400.7) and the
// convert created for the old object does not convert it (CR 603.7c).
func TestOptimusConvertDoesNotFollowCardToNewObject(t *testing.T) {
	g, engine, optimus, chosen := scheduleOptimusConvert(t)
	oldID := optimus.ObjectID
	if !movePermanentToZone(g, optimus, zone.Exile) {
		t.Fatal("could not exile Optimus Prime")
	}
	card, ok := g.GetCardInstance(optimus.CardInstanceID)
	if !ok {
		t.Fatal("Optimus Prime card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == oldID {
		t.Fatal("Optimus Prime did not return as a new object")
	}
	face, transformed := returned.Face, returned.Transformed
	if delayedViewBoundTo(observe(g, game.Player2).DelayedTriggers(), oldID) != 1 {
		t.Fatal("the pending convert lost its binding to the departed Optimus")
	}
	emitOptimusChosenCombatDamage(g, chosen)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the bound creature's combat damage did not fire the convert")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if returned.Face != face || returned.Transformed != transformed {
		t.Fatalf("returned Optimus face/transformed = %v/%v, want unchanged %v/%v", returned.Face, returned.Transformed, face, transformed)
	}
}
