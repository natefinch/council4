package rules

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const mobilizeSacrificeText = "Sacrifice them at the beginning of the next end step."

// Two same-name Mobilize sources each schedule a next-end-step sacrifice of
// their own token (CR 702.181a, 603.7c). Every observer can tell the pending
// delayed triggers apart by source permanent and bound token, both from the
// global list and from each permanent's view, and after the end step begins the
// fired triggers on the stack keep the same text and bindings.
func TestObservationDistinguishesSameNameDelayedTriggersByBinding(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	amount := game.MobilizeAmount{Fixed: 1}
	captainA := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(amount))
	captainB := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(amount))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: captainA.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: captainB.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	tokenA, delayedA := attackWithMobilize(t, g, engine, captainA)
	tokenB, delayedB := attackWithMobilize(t, g, engine, captainB)

	obs := observe(g, game.Player2)
	pending := obs.DelayedTriggers()
	if len(pending) != 2 {
		t.Fatalf("DelayedTriggers() = %d, want 2", len(pending))
	}
	for i, want := range []struct {
		trigger game.DelayedTrigger
		source  id.ID
		token   id.ID
	}{
		{delayedA, captainA.ObjectID, tokenA.ObjectID},
		{delayedB, captainB.ObjectID, tokenB.ObjectID},
	} {
		view := pending[i]
		if view.ID != want.trigger.ID || view.Controller != game.Player1 {
			t.Errorf("trigger %d = id %d controller %d, want id %d controller Player1", i, view.ID, view.Controller, want.trigger.ID)
		}
		if view.SourceName != "Mobilize Captain" || view.SourceObjectID != want.source {
			t.Errorf("trigger %d source = %q/%d, want Mobilize Captain/%d", i, view.SourceName, view.SourceObjectID, want.source)
		}
		if view.Text != mobilizeSacrificeText {
			t.Errorf("trigger %d text = %q, want %q", i, view.Text, mobilizeSacrificeText)
		}
		if view.Timing != game.DelayedAtBeginningOfNextEndStep || view.EventBased {
			t.Errorf("trigger %d timing = %v eventBased=%v, want next end step", i, view.Timing, view.EventBased)
		}
		if !slices.Equal(view.BoundObjectIDs, []id.ID{want.token}) {
			t.Errorf("trigger %d bound = %v, want [%d]", i, view.BoundObjectIDs, want.token)
		}
	}

	for _, check := range []struct {
		name      string
		permanent *game.Permanent
		want      id.ID
	}{
		{"token A", tokenA, delayedA.ID},
		{"token B", tokenB, delayedB.ID},
		{"captain A", captainA, delayedA.ID},
		{"captain B", captainB, delayedB.ID},
	} {
		view := findPermanentView(t, obs, check.permanent.ObjectID)
		if len(view.PendingDelayedTriggers) != 1 || view.PendingDelayedTriggers[0].ID != check.want {
			t.Errorf("%s pending delayed triggers = %+v, want only %d", check.name, view.PendingDelayedTriggers, check.want)
		}
	}

	// Mutating a returned view must not reach the game or a later observation.
	pending[0].BoundObjectIDs[0] = 0
	if got := obs.DelayedTriggers()[0].BoundObjectIDs; !slices.Equal(got, []id.ID{tokenA.ObjectID}) {
		t.Fatalf("observation aliased the delayed trigger's bound objects: %v", got)
	}

	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEnd, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end step did not put the mobilize sacrifices on the stack")
	}
	obs = observe(g, game.Player2)
	if left := obs.DelayedTriggers(); len(left) != 0 {
		t.Fatalf("pending delayed triggers after firing = %+v, want none", left)
	}
	stack := obs.Stack()
	if len(stack) != 2 {
		t.Fatalf("Stack() = %d objects, want the 2 fired sacrifices", len(stack))
	}
	seen := map[id.ID]id.ID{}
	for _, view := range stack {
		if view.Kind != game.StackTriggeredAbility || view.Name != "Mobilize Captain" || view.Text != mobilizeSacrificeText {
			t.Errorf("fired trigger = kind %v name %q text %q", view.Kind, view.Name, view.Text)
		}
		if len(view.BoundObjectIDs) != 1 {
			t.Fatalf("fired trigger bound = %v, want one token", view.BoundObjectIDs)
		}
		seen[view.SourceObjectID] = view.BoundObjectIDs[0]
	}
	if seen[captainA.ObjectID] != tokenA.ObjectID || seen[captainB.ObjectID] != tokenB.ObjectID {
		t.Fatalf("fired triggers source->bound = %v, want captain A->%d and captain B->%d", seen, tokenA.ObjectID, tokenB.ObjectID)
	}
}

// attackWithMobilize declares captain's attack, puts its real Mobilize 1 trigger
// on the stack and resolves it, returning the created token and the scheduled
// delayed sacrifice.
func attackWithMobilize(t *testing.T, g *game.Game, engine *Engine, captain *game.Permanent) (*game.Permanent, game.DelayedTrigger) {
	t.Helper()
	before := len(g.Battlefield)
	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		SourceObjectID: captain.ObjectID,
		PermanentID:    captain.ObjectID,
		Controller:     game.Player1,
		Player:         game.Player2,
	})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("attacking did not trigger mobilize")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != before+1 || len(g.DelayedTriggers) == 0 {
		t.Fatalf("mobilize created %d permanents and %d delayed triggers, want 1 token and a sacrifice", len(g.Battlefield)-before, len(g.DelayedTriggers))
	}
	return g.Battlefield[len(g.Battlefield)-1], g.DelayedTriggers[len(g.DelayedTriggers)-1]
}

// Unearth's "exile it at the beginning of the next end step" is public text on
// the returned permanent's view (CR 702.84a).
func TestObservationPermanentListsUnearthDelayedTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)
	if permanent == nil {
		t.Fatal("unearth did not return the card")
	}

	view := findPermanentView(t, observe(g, game.Player3), permanent.ObjectID)
	if len(view.PendingDelayedTriggers) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the unearth exile", view.PendingDelayedTriggers)
	}
	trigger := view.PendingDelayedTriggers[0]
	if trigger.Text != "Exile it at the beginning of the next end step." || trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("unearth delayed trigger = %+v", trigger)
	}
	if trigger.Controller != game.Player1 || trigger.SourceName != "Unearth Source" {
		t.Fatalf("unearth delayed trigger controller/source = %d/%q", trigger.Controller, trigger.SourceName)
	}
}

// A triggered ability of a face-up permanent shows its printed text and source
// permanent to every observer (CR 603.3).
func TestObservationStackShowsTriggeredAbilityText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Step:       game.StepUpkeep,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	const text = "At the beginning of your upkeep, you gain 1 life."
	card.Def.TriggeredAbilities[0].Text = text
	g.Turn.ActivePlayer = game.Player1
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player1, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("upkeep trigger was not put on the stack")
	}

	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 {
		t.Fatalf("Stack() = %d objects, want 1", len(stack))
	}
	view := stack[0]
	if view.Name != "Triggered Creature" || view.Text != text || view.SourceObjectID != source.ObjectID {
		t.Fatalf("trigger view = name %q text %q source %d, want Triggered Creature/%q/%d", view.Name, view.Text, view.SourceObjectID, text, source.ObjectID)
	}
	if len(view.BoundObjectIDs) != 0 || view.BoundCardID != 0 {
		t.Fatalf("ordinary trigger bound = %v/%d, want none", view.BoundObjectIDs, view.BoundCardID)
	}
}

// A face-down disguised permanent has ward {2} (CR 702.168b). Its ward trigger
// is public, but an opponent learns neither the hidden card's name nor its
// printed text (CR 708.5); the permanent's controller may still see both.
func TestObservationFaceDownWardTriggerHidesIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addFaceDownPermanent(g, game.Player2, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("face-down ward trigger was not put on the stack")
	}

	opponent := observe(g, game.Player1)
	if reachableNames(opponent)["Veiled Guard"] {
		t.Fatal("opponent observation exposed the face-down permanent's name")
	}
	stack := opponent.Stack()
	ward := stack[len(stack)-1]
	if ward.Kind != game.StackTriggeredAbility || ward.Name != "" || ward.Text != "Ward" || ward.SourceObjectID != warded.ObjectID {
		t.Fatalf("opponent ward view = kind %v name %q text %q source %d", ward.Kind, ward.Name, ward.Text, ward.SourceObjectID)
	}
	if own := observe(g, game.Player2).Stack()[len(stack)-1]; own.Name != "Veiled Guard" || own.Text != "Ward" {
		t.Fatalf("controller ward view = name %q text %q, want Veiled Guard/Ward", own.Name, own.Text)
	}
}

// An ability that left a permanent before it was turned face down keeps its
// printed text secret from opponents once the source is face down, while the
// controller can still read it (CR 708.2, 708.5).
func TestObservationFaceDownSourceHidesPrintedAbilityText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := triggeredCreature(&game.TriggerPattern{Event: game.EventBeginningOfStep, Step: game.StepUpkeep}, nil, nil)
	def.Name = "Secret Schemer"
	def.TriggeredAbilities[0].Text = "Secret upkeep text."
	source := addFaceDownPermanent(g, game.Player2, def, game.FaceDownMorph)
	index := -1
	for i := 0; def.BodyAt(i) != nil; i++ {
		if _, ok := def.BodyAt(i).(*game.TriggeredAbility); ok {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("printed trigger not found")
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player2,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		AbilityIndex: index,
	})

	opponent := observe(g, game.Player1).Stack()[0]
	if opponent.Name != "" || opponent.Text != "" || opponent.SourceObjectID != source.ObjectID {
		t.Fatalf("opponent view = name %q text %q source %d, want hidden identity", opponent.Name, opponent.Text, opponent.SourceObjectID)
	}
	if own := observe(g, game.Player2).Stack()[0]; own.Name != "Secret Schemer" || own.Text != "Secret upkeep text." {
		t.Fatalf("controller view = name %q text %q", own.Name, own.Text)
	}
}

// Necropotence's delayed return is bound to the card it exiled face down (CR
// 603.7c). That exile position is public, so every observer sees the bound card
// on the pending trigger and on the fired trigger, without its name.
func TestObservationNecropotenceDelayedTriggerBindsExiledCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Hidden Exile", Types: []types.Card{types.Creature}}})
	exiled := activateNecropotencePayLife(t, g, engine, source)

	obs := observe(g, game.Player2)
	pending := obs.DelayedTriggers()
	if len(pending) != 1 || pending[0].BoundCardID != exiled || pending[0].SourceName != "Necropotence" {
		t.Fatalf("pending = %+v, want Necropotence bound to %d", pending, exiled)
	}
	if reachableNames(obs)["Hidden Exile"] {
		t.Fatal("face-down exiled card's name leaked")
	}

	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEnd, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end step did not fire Necropotence's delayed return")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || stack[0].BoundCardID != exiled || stack[0].Name != "Necropotence" {
		t.Fatalf("fired trigger = %+v, want Necropotence bound to %d", stack, exiled)
	}
}

// A captured card is shown only while it remains in the zone incarnation the
// trigger captured and that position is visible: never in a library, only to
// players who may see the hand it is in, and not after it changed zones.
func TestObservationDelayedTriggerBoundCardRespectsHiddenZones(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	handCard := addCardToHand(g, game.Player1, vanillaCreature("Hand Secret", 1, 1))
	libraryCard := addLibraryCard(g, game.Player1, vanillaCreature("Library Secret", 1, 1))
	movedCard := addCardToGraveyard(g, game.Player1, vanillaCreature("Moved", 1, 1))
	graveyardCard := addCardToGraveyard(g, game.Player1, vanillaCreature("Public", 1, 1))
	moved, _ := g.GetCardInstance(movedCard)
	staleVersion := moved.ZoneVersion
	moved.ZoneVersion++
	for _, cardID := range []id.ID{handCard, libraryCard, movedCard, graveyardCard} {
		card, _ := g.GetCardInstance(cardID)
		version := card.ZoneVersion
		if cardID == movedCard {
			version = staleVersion
		}
		g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
			ID:                      g.IDGen.Next(),
			Controller:              game.Player1,
			Timing:                  game.DelayedAtBeginningOfNextEndStep,
			CapturedCardID:          cardID,
			CapturedCardZoneVersion: version,
		})
	}

	for _, check := range []struct {
		viewer game.PlayerID
		want   []id.ID
	}{
		{game.Player1, []id.ID{handCard, 0, 0, graveyardCard}},
		{game.Player2, []id.ID{0, 0, 0, graveyardCard}},
	} {
		var got []id.ID
		for _, view := range observe(g, check.viewer).DelayedTriggers() {
			got = append(got, view.BoundCardID)
		}
		if !slices.Equal(got, check.want) {
			t.Errorf("Player%d bound cards = %v, want %v", check.viewer, got, check.want)
		}
	}
}

// An event-based delayed trigger reports the permanent and player whose events
// it waits for (CR 603.7c).
func TestObservationEventDelayedTriggerReportsBindings(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	watched := addCombatPermanent(g, game.Player2, vanillaCreature("Watched", 2, 2))
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		ID:                    g.IDGen.Next(),
		Controller:            game.Player1,
		EventPattern:          opt.Val(game.TriggerPattern{Event: game.EventAttackerDeclared}),
		OneShot:               true,
		Window:                game.DelayedWindowThisTurn,
		BoundAttackerObjectID: watched.ObjectID,
		BoundEventPlayer:      opt.Val(game.Player2),
		Ability:               game.TriggeredAbility{Text: "When it attacks this turn, draw a card."},
	})

	view := findPermanentView(t, observe(g, game.Player3), watched.ObjectID)
	if len(view.PendingDelayedTriggers) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the attack watcher", view.PendingDelayedTriggers)
	}
	trigger := view.PendingDelayedTriggers[0]
	if !trigger.EventBased || !trigger.OneShot || trigger.Window != game.DelayedWindowThisTurn {
		t.Errorf("event trigger shape = %+v", trigger)
	}
	if !slices.Equal(trigger.BoundObjectIDs, []id.ID{watched.ObjectID}) || trigger.BoundPlayer != opt.Val(game.Player2) {
		t.Errorf("event trigger bindings = %v/%v", trigger.BoundObjectIDs, trigger.BoundPlayer)
	}
	if trigger.Text != "When it attacks this turn, draw a card." {
		t.Errorf("event trigger text = %q", trigger.Text)
	}
}

const splinterTwinGrantedText = "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step."

// Splinter Twin's delayed exile records no wording of its own, so it carries the
// public text of the granted ability whose resolution created it (CR 603.7a)
// while pending, on the token's view, and on the stack once it fires.
func TestObservationSplinterTwinDelayedExileCarriesCreatingAbilityText(t *testing.T) {
	twin := cards.NewDefaultRegistry().Lookup("Splinter Twin")
	if twin == nil || twin.Name != "Splinter Twin" {
		t.Fatal("Splinter Twin is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	host := addCombatPermanent(g, game.Player1, vanillaCreature("Twin Host", 2, 2))
	aura := addCombatPermanent(g, game.Player1, twin)
	if !attachPermanent(g, aura, host) {
		t.Fatal("could not attach Splinter Twin")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(host.ObjectID, 1, nil, 0)) {
		t.Fatal("activating the granted Splinter Twin ability failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	obs := observe(g, game.Player2)
	pending := obs.DelayedTriggers()
	if len(pending) != 1 || len(pending[0].BoundObjectIDs) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want one exile bound to the token", pending)
	}
	if pending[0].Text != splinterTwinGrantedText || pending[0].SourceName != "Twin Host" {
		t.Fatalf("pending exile = source %q text %q, want the granted ability's host Twin Host/%q", pending[0].SourceName, pending[0].Text, splinterTwinGrantedText)
	}
	token := findPermanentView(t, obs, pending[0].BoundObjectIDs[0])
	if len(token.PendingDelayedTriggers) != 1 || token.PendingDelayedTriggers[0].Text != splinterTwinGrantedText {
		t.Fatalf("token pending delayed triggers = %+v", token.PendingDelayedTriggers)
	}

	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEnd, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end step did not fire Splinter Twin's delayed exile")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || stack[0].Text != splinterTwinGrantedText {
		t.Fatalf("fired exile = %+v, want text %q", stack, splinterTwinGrantedText)
	}
	if !slices.Equal(stack[0].BoundObjectIDs, []id.ID{token.ObjectID}) {
		t.Fatalf("fired exile bound = %v, want token %d", stack[0].BoundObjectIDs, token.ObjectID)
	}
}

// A delayed trigger created by a resolving spell carries that spell's public
// Oracle text, as rules-text lines (CR 603.7a).
func TestObservationSpellDelayedTriggerCarriesSpellText(t *testing.T) {
	wine := cards.NewDefaultRegistry().Lookup("Blessed Wine")
	if wine == nil || wine.Name != "Blessed Wine" {
		t.Fatal("Blessed Wine is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, wine)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Blessed Wine failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	pending := observe(g, game.Player3).DelayedTriggers()
	const want = "You gain 1 life.\nDraw a card at the beginning of the next turn's upkeep."
	if len(pending) != 1 || pending[0].Text != want || pending[0].SourceName != "Blessed Wine" {
		t.Fatalf("pending delayed triggers = %+v, want Blessed Wine with text %q", pending, want)
	}
}

// A printed triggered ability whose definition omits per-ability text shows its
// source's Oracle text on the stack, and the delayed trigger it creates carries
// the same text while pending and after it fires.
func TestObservationPrintedTriggerDelayedTriggerFallsBackToOracleText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Step:       game.StepUpkeep,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}}}}, nil)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	const text = "At the beginning of your upkeep, you gain 1 life at the beginning of the next end step."
	card.Def.OracleText = "\n\t\t" + text + "\n\t"
	g.Turn.ActivePlayer = game.Player1
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player1, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("upkeep trigger was not put on the stack")
	}
	if stack := observe(g, game.Player2).Stack(); len(stack) != 1 || stack[0].Text != text {
		t.Fatalf("upkeep trigger = %+v, want text %q", stack, text)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	view := findPermanentView(t, observe(g, game.Player2), source.ObjectID)
	if len(view.PendingDelayedTriggers) != 1 || view.PendingDelayedTriggers[0].Text != text {
		t.Fatalf("pending delayed triggers = %+v, want text %q", view.PendingDelayedTriggers, text)
	}
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEnd, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end step did not fire the delayed trigger")
	}
	if stack := observe(g, game.Player2).Stack(); len(stack) != 1 || stack[0].Text != text || stack[0].SourceObjectID != source.ObjectID {
		t.Fatalf("fired delayed trigger = %+v, want text %q from %d", stack, text, source.ObjectID)
	}
}

// A delayed trigger created by a face-down permanent's ability never captures
// text from the hidden card (CR 708.2, 708.5), for any observer.
func TestObservationFaceDownSourceDelayedTriggerCapturesNoHiddenText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := triggeredCreature(&game.TriggerPattern{Event: game.EventBeginningOfStep, Step: game.StepUpkeep}, []game.Instruction{{
		Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}, nil)
	def.Name = "Secret Schemer"
	def.OracleText = "Secret oracle text."
	def.TriggeredAbilities[0].Text = "Secret upkeep text."
	source := addFaceDownPermanent(g, game.Player2, def, game.FaceDownMorph)
	index := -1
	for i := 0; def.BodyAt(i) != nil; i++ {
		if _, ok := def.BodyAt(i).(*game.TriggeredAbility); ok {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("printed trigger not found")
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player2,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		AbilityIndex: index,
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	if text := g.DelayedTriggers[0].Ability.Text; text != "" {
		t.Fatalf("delayed trigger captured hidden text %q", text)
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		pending := observe(g, viewer).DelayedTriggers()
		if len(pending) != 1 || pending[0].Text != "" {
			t.Fatalf("Player%d pending = %+v, want no text", viewer, pending)
		}
	}
}

// An untexted ability that is not one of its source's printed abilities, such
// as one granted by another permanent, shows no text rather than the source's
// unrelated Oracle text, and a delayed trigger it creates captures none.
func TestObservationUntextedGrantedTriggerGetsNoOracleFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	host := addCombatPermanent(g, game.Player1, vanillaCreature("Plain Host", 2, 2))
	card, _ := g.GetCardInstance(host.CardInstanceID)
	card.Def.OracleText = "Unrelated host text."
	granted := &game.TriggeredAbility{Content: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}}.Ability()}
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		Controller:    game.Player1,
		SourceID:      host.ObjectID,
		SourceCardID:  host.CardInstanceID,
		InlineTrigger: granted,
	})
	if stack := observe(g, game.Player2).Stack(); len(stack) != 1 || stack[0].Text != "" {
		t.Fatalf("granted trigger view = %+v, want no text", stack)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if pending := observe(g, game.Player2).DelayedTriggers(); len(pending) != 1 || pending[0].Text != "" {
		t.Fatalf("pending = %+v, want one delayed trigger with no text", pending)
	}
}

// A delayed trigger's text comes from the public ability that created it, so it
// stays visible to every observer even when its source is a face-down permanent
// whose identity is hidden (CR 603.7a, 708.5). Splinter Twin grants its ability
// to a face-down Willbender: opponents see the granted ability's delayed exile
// text while it is pending and once it fires, but never the hidden source's name.
func TestObservationFaceDownHostDelayedTriggerKeepsPublicText(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	twin := registry.Lookup("Splinter Twin")
	secret := registry.Lookup("Willbender")
	if twin == nil || secret == nil {
		t.Fatal("Splinter Twin or Willbender is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	host := addFaceDownPermanent(g, game.Player1, secret, game.FaceDownMorph)
	host.SummoningSick = false
	aura := addCombatPermanent(g, game.Player1, twin)
	if !attachPermanent(g, aura, host) {
		t.Fatal("could not attach Splinter Twin")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(host.ObjectID, 0, nil, 0)) {
		t.Fatal("activating the granted Splinter Twin ability on the face-down host failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		pending := observe(g, viewer).DelayedTriggers()
		if len(pending) != 1 || pending[0].Text != splinterTwinGrantedText {
			t.Fatalf("Player%d pending = %+v, want text %q", viewer, pending, splinterTwinGrantedText)
		}
	}
	if pending := observe(g, game.Player2).DelayedTriggers()[0]; pending.SourceName != "" {
		t.Fatalf("opponent sees hidden source name %q", pending.SourceName)
	}
	if pending := observe(g, game.Player1).DelayedTriggers()[0]; pending.SourceName != secret.Name {
		t.Fatalf("controller source name = %q, want %q", pending.SourceName, secret.Name)
	}

	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEnd, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the end step did not fire the delayed exile")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || stack[0].Text != splinterTwinGrantedText || stack[0].Name != "" {
		t.Fatalf("fired exile seen by opponent = %+v, want text %q and no source name", stack, splinterTwinGrantedText)
	}
}

// A fired event-based delayed trigger keeps identifying the object whose event
// it waited for (CR 603.7c) on the stack, as it did while pending: Okoye's
// "whenever that creature attacks the monarch" trigger names its attacker, and
// a "when that creature dies this turn" trigger names the creature that died.
// Copying the stack object keeps an independent binding list.
func TestObservationFiredEventDelayedTriggersKeepBindings(t *testing.T) {
	t.Run("attacker", func(t *testing.T) {
		okoye := cards.NewDefaultRegistry().Lookup("Okoye, Mighty and Adored")
		if okoye == nil {
			t.Fatal("Okoye, Mighty and Adored is not in the default registry")
		}
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, okoye)
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepBeginningOfCombat, game.Player1
		g.Combat = &game.CombatState{}
		fireOkoyeBeginningOfCombat(t, g, engine, creature)
		pending := observe(g, game.Player3).DelayedTriggers()
		if len(pending) != 1 || !slices.Equal(pending[0].BoundObjectIDs, []id.ID{creature.ObjectID}) {
			t.Fatalf("pending = %+v, want one trigger bound to %d", pending, creature.ObjectID)
		}
		if !setMonarch(g, game.Player2) {
			t.Fatal("could not make Player2 the monarch")
		}
		declareAttack(t, g, engine, creature, game.Player2)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
		if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
			t.Fatal("attacking the monarch did not fire Okoye's delayed trigger")
		}
		stack := observe(g, game.Player3).Stack()
		if len(stack) != 1 || !slices.Equal(stack[0].BoundObjectIDs, []id.ID{creature.ObjectID}) || stack[0].Text != pending[0].Text {
			t.Fatalf("fired trigger = %+v, want binding %d and text %q", stack, creature.ObjectID, pending[0].Text)
		}
		original, _ := g.Stack.Peek()
		copied := game.NewStackObjectCopy(original, g.IDGen.Next())
		original.BoundEventObjectIDs[0] = 0
		if !slices.Equal(copied.BoundEventObjectIDs, []id.ID{creature.ObjectID}) {
			t.Fatalf("copied bindings = %v, want independent %d", copied.BoundEventObjectIDs, creature.ObjectID)
		}
	})
	t.Run("dying creature", func(t *testing.T) {
		g := mainPhaseGame(game.Player1)
		engine := NewEngine(nil)
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		castDiesThisTurnSpell(t, g, engine, searingBloodShape(), victim)
		engine.applyStateBasedActions(g)
		if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}, &TurnLog{}) {
			t.Fatal("the creature's death did not fire the delayed trigger")
		}
		stack := observe(g, game.Player4).Stack()
		if len(stack) != 1 || !slices.Equal(stack[0].BoundObjectIDs, []id.ID{victim.ObjectID}) {
			t.Fatalf("fired dies trigger = %+v, want binding %d", stack, victim.ObjectID)
		}
	})
}

const (
	okoyeDelayedTriggerText = "Whenever that creature attacks the monarch this turn"
	okoyeMonarchText        = "When Okoye enters, you become the monarch."
	okoyeCombatText         = "At the beginning of combat on your turn, put a +1/+1 counter on target creature. " + okoyeDelayedOwnText
	okoyeDelayedOwnText     = okoyeDelayedTriggerText + ", it gains double strike and trample until end of turn."
)

// Source review: Okoye, Mighty and Adored (oracle_id
// 8fbfc5b1-2146-4d07-a5fe-5f025da31bd6) and Clone (oracle_id
// 42226b87-0746-4ebf-9fd0-108d508462af) Oracle text from the run's pinned
// corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Scryfall lists no rulings for Okoye. Comprehensive Rules effective
// 2026-09-25 (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt),
// read 113.2c, 207.2c, 602.1, 603.1, 603.3, 603.7, 603.7a, 603.7e, 700.2,
// 707.2, 707.10 and 708.2: each paragraph break in a card's text marks a
// separate ability (113.2c); a triggered ability on the stack has the text of
// the ability that created it and no other characteristics (603.3); a delayed
// triggered ability is created by a resolving ability, contains "when",
// "whenever" or "at" (603.7, 603.7a) and shares its creator's source (603.7e);
// a copy acquires the original's rules text as a copiable value (707.2), but a
// face-down permanent copied has none (707.2, 708.2).
//
// Clone entering as a copy of Okoye has Okoye's untexted beginning-of-combat
// trigger. Every observer sees exactly that ability's paragraph on the stack,
// not Okoye's enters ability or Clone's own text, and the delayed trigger it
// creates carries only its own "Whenever that creature attacks the monarch"
// sentence. The text and copied-source identity stay after the copied Okoye
// and then the Clone leave the battlefield.
func TestObservationCopiedUntextedTriggerKeepsCopiedText(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	okoye := registry.Lookup("Okoye, Mighty and Adored")
	clone := registry.Lookup("Clone")
	if okoye == nil || clone == nil {
		t.Fatal("Okoye or Clone is not in the default registry")
	}
	for i := range okoye.TriggeredAbilities {
		if okoye.TriggeredAbilities[i].Text != "" {
			t.Fatalf("Okoye trigger %d has its own text %q; this scenario needs untexted triggers", i, okoye.TriggeredAbilities[i].Text)
		}
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blueprint := addCombatPermanent(g, game.Player2, okoye)
	addImplementationSpellToStack(g, game.Player1, clone, nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	if copied.Controller != game.Player1 || permanentEffectiveName(g, copied) != okoye.Name {
		t.Fatalf("Clone entered as %q, want a copy of %q", permanentEffectiveName(g, copied), okoye.Name)
	}
	g.TriggerEventCursor = len(g.Events)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepBeginningOfCombat, game.Player1
	g.Combat = &game.CombatState{}
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{targetPick: target.ObjectID}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the copied beginning-of-combat trigger did not trigger")
	}
	wantText := func(where, text, want string) {
		t.Helper()
		if text != want {
			t.Errorf("%s text = %q, want exactly %q", where, text, want)
		}
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
		stack := observe(g, viewer).Stack()
		if len(stack) != 1 || stack[0].SourceObjectID != copied.ObjectID {
			t.Fatalf("player %d stack = %+v, want the copied trigger from %d", viewer, stack, copied.ObjectID)
		}
		wantText("stack", stack[0].Text, okoyeCombatText)
	}
	if !movePermanentToZone(g, blueprint, zone.Graveyard) {
		t.Fatal("could not remove the copied Okoye")
	}
	if stack := observe(g, game.Player3).Stack(); len(stack) == 1 {
		wantText("stack after the copied Okoye left", stack[0].Text, okoyeCombatText)
	} else {
		t.Fatalf("stack = %+v, want the copied trigger", stack)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if target.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatal("the copied trigger did not put a +1/+1 counter on its target")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
		pending := observe(g, viewer).DelayedTriggers()
		if len(pending) != 1 || pending[0].SourceObjectID != copied.ObjectID ||
			!slices.Contains(pending[0].BoundObjectIDs, target.ObjectID) {
			t.Fatalf("player %d pending = %+v, want the copied Okoye's delayed trigger bound to %d", viewer, pending, target.ObjectID)
		}
		wantText("pending delayed trigger", pending[0].Text, okoyeDelayedOwnText)
	}
	if !movePermanentToZone(g, copied, zone.Graveyard) {
		t.Fatal("could not remove the Clone")
	}
	pending := observe(g, game.Player3).DelayedTriggers()
	if len(pending) != 1 || pending[0].SourceObjectID != 0 {
		t.Fatalf("pending after the Clone left = %+v, want one sourceless delayed trigger", pending)
	}
	wantText("delayed trigger after the Clone left", pending[0].Text, okoyeDelayedOwnText)
}

// A permanent that becomes a copy of a card in a graveyard (Shifting Woodland's
// path, CR 707.2) gains the card's untexted activated abilities with that
// card's Oracle text, so its activation is labeled with the copied text rather
// than the copying permanent's own Oracle text.
func TestCopiedUntextedActivatedAbilityCarriesCopiedText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pinger := vanillaCreature("Untexted Pinger", 1, 1)
	pinger.OracleText = `
		{T}: Untexted Pinger deals 1 damage to any target.
	`
	pinger.ActivatedAbilities = []game.ActivatedAbility{{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
		ZoneOfFunction:  zone.Battlefield,
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "any target",
				Allow: game.TargetAllowPermanent | game.TargetAllowPlayer,
			}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0),
			}}},
		}.Ability(),
	}}
	shared := pinger.ActivatedAbilities
	cardID := addCardToGraveyard(g, game.Player1, pinger)
	copying := addCombatPermanent(g, game.Player1, vanillaCreature("Copying Host", 2, 2))
	r := &effectResolver{engine: NewEngine(nil), game: g, log: &TurnLog{}, obj: &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility,
		SourceID: copying.ObjectID, Controller: game.Player1,
		Targets: []game.Target{currentCardTarget(t, g, cardID)},
	}}
	if !handleBecomeCopy(r, game.BecomeCopy{Card: game.CardReference{Kind: game.CardReferenceTarget}}).succeeded {
		t.Fatal("BecomeCopy of the graveyard card failed")
	}
	if shared[0].Text != "" {
		t.Fatalf("copying wrote text %q into the shared card definition", shared[0].Text)
	}
	text := activatedAbilityText(g, game.Player1, action.ActivateAbilityAction{SourceID: copying.ObjectID, AbilityIndex: 0})
	if text != "{T}: Untexted Pinger deals 1 damage to any target." {
		t.Fatalf("copied activated ability text = %q, want the copied card's Oracle text", text)
	}
}

// Copying a face-down permanent copies only its face-down characteristics
// (CR 707.2, 708.2): Clone copying a face-down Okoye gains none of Okoye's
// abilities or text, so nothing about the hidden card reaches observation.
func TestObservationCopyOfFaceDownPermanentCarriesNoHiddenText(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	okoye := registry.Lookup("Okoye, Mighty and Adored")
	clone := registry.Lookup("Clone")
	if okoye == nil || clone == nil {
		t.Fatal("Okoye or Clone is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hidden := addFaceDownPermanent(g, game.Player1, okoye, game.FaceDownMorph)
	addImplementationSpellToStack(g, game.Player1, clone, nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	if copied == hidden || copied.FaceDown {
		t.Fatal("Clone did not enter")
	}
	def, ok := permanentCopyDef(g, copied)
	if !ok || def.Name != "" || def.AbilityCount() != 0 || def.OracleText != "" {
		t.Fatalf("copy of a face-down permanent = %+v, want a nameless, abilityless 2/2", def)
	}
	g.TriggerEventCursor = len(g.Events)
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepBeginningOfCombat, game.Player1
	g.Combat = &game.CombatState{}
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, object := range observe(g, game.Player2).Stack() {
		if strings.Contains(object.Text, okoyeDelayedTriggerText) || object.Name == okoye.Name {
			t.Fatalf("stack object %+v reveals the face-down Okoye", object)
		}
	}
}

// Dash's delayed return acts on the specific permanent that was dashed (CR
// 702.109a, 603.7c), so that permanent is the trigger's bound object while the
// return is pending, after the permanent has left, and once the trigger fires.
func TestObservationDashReturnBindsDashedPermanent(t *testing.T) {
	ragavan := cards.NewDefaultRegistry().Lookup("Ragavan, Nimble Pilferer")
	if ragavan == nil {
		t.Fatal("Ragavan, Nimble Pilferer is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addCardToHand(g, game.Player1, ragavan)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Dash"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("could not cast Ragavan for its dash cost")
	}
	resolveStackWithTriggers(engine, g, agents)
	dashed := permanentForCard(g, card)
	if dashed == nil {
		t.Fatal("Ragavan did not enter")
	}
	wantBound := func(where string, views []DelayedTriggerView) {
		t.Helper()
		if len(views) != 1 || !slices.Contains(views[0].BoundObjectIDs, dashed.ObjectID) ||
			views[0].Text != "Return it to its owner's hand at the beginning of the next end step." {
			t.Fatalf("%s = %+v, want the dash return bound to %d", where, views, dashed.ObjectID)
		}
	}
	wantBound("pending return", observe(g, game.Player2).DelayedTriggers())
	wantBound("Ragavan's pending triggers", observe(g, game.Player2).Battlefield()[0].PendingDelayedTriggers)
	if !movePermanentToZone(g, dashed, zone.Graveyard) {
		t.Fatal("could not remove Ragavan")
	}
	wantBound("pending return after Ragavan left", observe(g, game.Player2).DelayedTriggers())
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the dash return did not trigger")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, dashed.ObjectID) {
		t.Fatalf("stack = %+v, want the dash return bound to %d", stack, dashed.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Contains(card) {
		t.Fatal("the dash return moved the card from the graveyard to its owner's hand")
	}
}

// Okoye's own untexted abilities are observed with their own Oracle paragraphs
// (CR 113.2c, 603.3): its enters trigger shows only the monarch ability, its
// beginning-of-combat trigger only that paragraph, and the delayed trigger it
// creates only its own "Whenever that creature attacks the monarch" sentence
// (CR 603.7, 603.7a), from creation until it fires from Okoye (CR 603.7e).
// See the source review on TestObservationCopiedUntextedTriggerKeepsCopiedText.
func TestObservationPrintedOkoyeAbilitiesShowOwnParagraphs(t *testing.T) {
	okoye := cards.NewDefaultRegistry().Lookup("Okoye, Mighty and Adored")
	if okoye == nil {
		t.Fatal("Okoye is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	wantStack := func(where string, source id.ID, want string) {
		t.Helper()
		for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
			stack := observe(g, viewer).Stack()
			if len(stack) != 1 || stack[0].SourceObjectID != source || stack[0].Text != want {
				t.Fatalf("player %d %s stack = %+v, want text %q from %d", viewer, where, stack, want, source)
			}
		}
	}
	addImplementationSpellToStack(g, game.Player1, okoye, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	printed := g.Battlefield[len(g.Battlefield)-1]
	if permanentEffectiveName(g, printed) != okoye.Name {
		t.Fatalf("resolved permanent = %q, want Okoye", permanentEffectiveName(g, printed))
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("Okoye's enters trigger did not trigger")
	}
	wantStack("enters trigger", printed.ObjectID, okoyeMonarchText)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("Okoye's enters trigger did not make Player1 the monarch")
	}

	g.TriggerEventCursor = len(g.Events)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepBeginningOfCombat, game.Player1
	g.Combat = &game.CombatState{}
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{targetPick: target.ObjectID}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Okoye's beginning-of-combat trigger did not trigger")
	}
	wantStack("beginning-of-combat trigger", printed.ObjectID, okoyeCombatText)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		pending := observe(g, viewer).DelayedTriggers()
		if len(pending) != 1 || pending[0].SourceObjectID != printed.ObjectID || pending[0].Text != okoyeDelayedOwnText {
			t.Fatalf("player %d pending = %+v, want Okoye's delayed trigger with text %q", viewer, pending, okoyeDelayedOwnText)
		}
	}

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, target, game.Player2)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the delayed attacks-the-monarch trigger did not fire")
	}
	wantStack("fired delayed trigger", printed.ObjectID, okoyeDelayedOwnText)
}

const paragraphBeastOracle = `
	Flying
	{1}: You gain 1 life.
	{T}: Add {G}.
	Landfall — Whenever a land you control enters, you gain 1 life.
	At the beginning of your upkeep, you gain 1 life.
`

// paragraphBeast is a creature whose untexted activated and triggered
// abilities are each written as their own Oracle paragraph (CR 113.2c), among a
// keyword paragraph and a mana ability.
func paragraphBeast() *game.CardDef {
	gain := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}}.Ability()
	def := vanillaCreature("Paragraph Beast", 2, 2, game.Flying)
	def.OracleText = paragraphBeastOracle
	def.ActivatedAbilities = []game.ActivatedAbility{{
		ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
		ZoneOfFunction: zone.Battlefield,
		Content:        gain,
	}}
	def.ManaAbilities = []game.ManaAbility{game.TapManaAbility(mana.G)}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event:            game.EventPermanentEnteredBattlefield,
			Controller:       game.TriggerControllerYou,
			SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
		}},
		Content: gain,
	}, {
		Trigger: game.TriggerCondition{Type: game.TriggerAt, Pattern: game.TriggerPattern{
			Event:      game.EventBeginningOfStep,
			Step:       game.StepUpkeep,
			Controller: game.TriggerControllerYou,
		}},
		Content: gain,
	}}
	return def
}

// Clone copying a creature with several untexted abilities acquires each
// ability with only its own Oracle paragraph (CR 113.2c, 707.2), so the copied
// activated ability's label and stack text and the copied upkeep trigger's
// stack text name only that ability, never the keyword line, the mana ability or
// the other trigger, while the shared definition stays untexted.
func TestObservationCopiedAbilitiesCarryOnlyTheirOwnParagraphs(t *testing.T) {
	clone := cards.NewDefaultRegistry().Lookup("Clone")
	if clone == nil {
		t.Fatal("Clone is not in the default registry")
	}
	beast := paragraphBeast()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, beast)
	addImplementationSpellToStack(g, game.Player1, clone, nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	if copied.Controller != game.Player1 || permanentEffectiveName(g, copied) != beast.Name {
		t.Fatalf("Clone entered as %q, want a copy of %q", permanentEffectiveName(g, copied), beast.Name)
	}
	if beast.ActivatedAbilities[0].Text != "" || beast.TriggeredAbilities[0].Text != "" || beast.TriggeredAbilities[1].Text != "" {
		t.Fatal("copying wrote text into the shared card definition")
	}

	const activatedText = "{1}: You gain 1 life."
	activate := action.ActivateAbilityAction{SourceID: copied.ObjectID, AbilityIndex: 0}
	if text := activatedAbilityText(g, game.Player1, activate); text != activatedText {
		t.Fatalf("copied activated ability label = %q, want %q", text, activatedText)
	}
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyActivateAbilityWithChoices(g, game.Player1, activate, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the copied activated ability could not be activated")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		stack := observe(g, viewer).Stack()
		if len(stack) != 1 || stack[0].SourceObjectID != copied.ObjectID || stack[0].Text != activatedText {
			t.Fatalf("player %d stack = %+v, want the copied activated ability with text %q", viewer, stack, activatedText)
		}
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	const upkeepText = "At the beginning of your upkeep, you gain 1 life."
	g.TriggerEventCursor = len(g.Events)
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseBeginning, game.StepUpkeep, game.Player1
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the copied upkeep trigger did not trigger")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		stack := observe(g, viewer).Stack()
		if len(stack) != 1 || stack[0].SourceObjectID != copied.ObjectID || stack[0].Text != upkeepText {
			t.Fatalf("player %d stack = %+v, want the copied upkeep trigger with text %q", viewer, stack, upkeepText)
		}
	}
}

// When an untexted ability's own paragraph cannot be identified, as when one
// "enters or dies" paragraph is authored as two triggered abilities so the
// face's triggered abilities and triggered paragraphs no longer pair up, the
// ability is observed with no printed text rather than with the card's
// unrelated abilities, both on the original and on a copy.
func TestObservationUnidentifiedAbilityParagraphShowsNoUnrelatedText(t *testing.T) {
	gain := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}}.Ability()
	def := vanillaCreature("Split Trigger Beast", 2, 2, game.Flying)
	def.OracleText = `
		Flying
		When this creature enters or dies, you gain 1 life.
		At the beginning of your upkeep, you gain 1 life.
	`
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
		}},
		Content: gain,
	}, {
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentDied, Source: game.TriggerSourceSelf,
		}},
		Content: gain,
	}, {
		Trigger: game.TriggerCondition{Type: game.TriggerAt, Pattern: game.TriggerPattern{
			Event: game.EventBeginningOfStep, Step: game.StepUpkeep, Controller: game.TriggerControllerYou,
		}},
		Content: gain,
	}}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCombatPermanent(g, game.Player1, def)
	copyDef, ok := permanentCopyDef(g, original)
	if !ok {
		t.Fatal("permanentCopyDef failed")
	}
	for i := range copyDef.TriggeredAbilities {
		if text := copyDef.TriggeredAbilities[i].Text; text != "" {
			t.Fatalf("copied trigger %d text = %q, want none", i, text)
		}
	}
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseBeginning, game.StepUpkeep, game.Player1
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the upkeep trigger did not trigger")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || stack[0].SourceObjectID != original.ObjectID || stack[0].Text != "" {
		t.Fatalf("stack = %+v, want the upkeep trigger from %d with no unrelated text", stack, original.ObjectID)
	}
}

// Oracle paragraphs pair with untexted abilities of their own kind (CR 113.2c):
// ability words (CR 207.2c), "•" mode lines (CR 700.2), loyalty and chapter
// abilities are recognized, and an "At" paragraph never pairs with a When or
// Whenever trigger.
func TestPrintedAbilityParagraphsPairByKind(t *testing.T) {
	when := game.TriggeredAbility{Trigger: game.TriggerCondition{Type: game.TriggerWhen}}
	at := game.TriggeredAbility{Trigger: game.TriggerCondition{Type: game.TriggerAt}}
	face := game.CardFace{
		LoyaltyAbilities:   []game.LoyaltyAbility{{LoyaltyCost: 1}, {LoyaltyCost: -2}},
		TriggeredAbilities: []game.TriggeredAbility{when, at},
	}
	const oracle = `
		Eerie — When this enters, choose one —
		• You gain 1 life.
		• Scry 1.
		At the beginning of your end step, draw a card.
		+1: Scry 1.
		−2: Draw a card.
	`
	got := printedAbilityParagraphs(&face, oracle)
	want := map[int]string{
		face.LoyaltyAbilityIndex(0):   "+1: Scry 1.",
		face.LoyaltyAbilityIndex(1):   "−2: Draw a card.",
		face.TriggeredAbilityIndex(0): "Eerie — When this enters, choose one —\n• You gain 1 life.\n• Scry 1.",
		face.TriggeredAbilityIndex(1): "At the beginning of your end step, draw a card.",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("paragraphs = %v, want %v", got, want)
	}

	face.TriggeredAbilities = []game.TriggeredAbility{at, when}
	got = printedAbilityParagraphs(&face, oracle)
	if text, ok := got[face.TriggeredAbilityIndex(0)]; ok {
		t.Fatalf("an At trigger paired with %q", text)
	}

	saga := game.CardFace{ChapterAbilities: []game.ChapterAbility{{Chapters: []int{1, 2}}, {Chapters: []int{3}}}}
	got = printedAbilityParagraphs(&saga, `
		(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
		I, II — Scry 1.
		III — Draw a card.
	`)
	want = map[int]string{
		saga.TriggeredAbilityIndex(0):     "I, II — Scry 1.",
		saga.TriggeredAbilityIndex(0) + 1: "III — Draw a card.",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("chapter paragraphs = %v, want %v", got, want)
	}
}
