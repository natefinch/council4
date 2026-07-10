package rules

import (
	"testing"

	cardsb "github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func artifactCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// findActivateAction locates a legal activate-ability action for a given source
// permanent and chosen X, independent of the ability's index within the face.
func findActivateAction(actions []action.Action, source game.ObjectID, x int) (action.Action, bool) {
	for _, a := range actions {
		payload, ok := a.ActivateAbilityPayload()
		if ok && payload.SourceID == source && payload.XValue == x {
			return a, true
		}
	}
	return action.Action{}, false
}

// TestBlasterGrantsGroupModularEntersWithCounter proves the front face's
// "Other nontoken artifact creatures and Vehicles you control have modular 1."
// grant: the reusable group-modular expansion lowers to a single
// EntersWithCountersGroupReplacement over the union selection {nontoken artifact
// creature} OR {nontoken Vehicle}, both scoped to you and excluding the source.
// A nontoken artifact creature and a Vehicle you control each enter with one
// additional +1/+1 counter, while a plain creature, an opponent's artifact
// creature, and Blaster itself (ExcludeSource) get none.
func TestBlasterGrantsGroupModularEntersWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	blaster := addReplacementPermanent(t, g, game.Player1, cardsb.BlasterCombatDJ())
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("registered replacement effects = %d, want 1 (group modular EWC)", len(g.ReplacementEffects))
	}
	if got := blaster.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("Blaster +1/+1 counters = %d, want 0 (ExcludeSource)", got)
	}

	artCreature := addReplacementPermanent(t, g, game.Player1, artifactCreatureDef("Ornithopter"))
	if got := artCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("controlled artifact creature +1/+1 counters = %d, want 1 (granted modular)", got)
	}

	vehicle := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Smuggler's Copter",
		Types:     []types.Card{types.Artifact},
		Subtypes:  []types.Sub{types.Sub("Vehicle")},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	if got := vehicle.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("controlled Vehicle +1/+1 counters = %d, want 1 (granted modular via union)", got)
	}

	plainCreature := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	if got := plainCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("controlled non-artifact creature +1/+1 counters = %d, want 0 (not modular)", got)
	}

	opponentArtCreature := addReplacementPermanent(t, g, game.Player2, artifactCreatureDef("Steel Overseer"))
	if got := opponentArtCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent artifact creature +1/+1 counters = %d, want 0 (you-only grant)", got)
	}
}

// modularDiesTriggerCreatureDef builds an artifact creature carrying the modular
// keyword's lowered dies-trigger payload: "When this creature dies, you may move
// all +1/+1 counters from this creature onto target artifact creature." The move
// is the reusable kind-specific mass form — only +1/+1 counters move (CR 702.44),
// their count read from the dying source's last-known information through
// DynamicAmountObjectCounters so a source that resolves from the graveyard still
// donates every +1/+1 counter it had. It mirrors the generated Modular payload
// shared by Power Depot and Blaster, Morale Booster.
func modularDiesTriggerCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:            game.EventPermanentDied,
					Source:           game.TriggerSourceSelf,
					SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				},
			},
			Optional: true,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.MoveCounters{
						Object:      game.TargetPermanentReference(0),
						CounterKind: counter.PlusOnePlusOne,
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:        game.DynamicAmountObjectCounters,
							CounterKind: counter.PlusOnePlusOne,
							Object:      game.SourcePermanentReference(),
						}),
						Source: game.CounterSourceSpec{Kind: game.CounterSourceSelf},
					},
				}},
			}.Ability(),
		}},
	}}
}

// TestModularDeathMovesCountersToTargetArtifactCreature proves the modular
// dies-trigger through the real death pipeline: a modular artifact creature
// carrying two +1/+1 counters and a stun counter receives lethal damage, dies via
// state-based actions (leaving the battlefield so its counters are only reachable
// through last-known information), and its dies-trigger goes on the stack. On
// resolution the targeted artifact creature gains exactly the source's two +1/+1
// counters, while the stun counter stays behind (CR 702.44 moves only +1/+1
// counters). This fails on the pre-fix code, where CounterSourceSelf read the live
// battlefield only (moving zero from a dead source) and AllKinds would have moved
// the stun counter too.
func TestModularDeathMovesCountersToTargetArtifactCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, modularDiesTriggerCreatureDef("Arcbound Worker"))
	sourceObjectID := source.ObjectID
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	source.Counters.Add(counter.Stun, 1)

	target := addCombatPermanent(g, game.Player1, artifactCreatureDef("Steel Overseer"))

	// Deal lethal damage (effective toughness 1 base + 2 counters = 3) and run
	// state-based actions so the source dies and records last-known information.
	source.MarkedDamage = 3
	g.AppendEvent(game.Event{
		Kind:            game.EventDamageDealt,
		PermanentID:     source.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		Amount:          3,
	})
	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, sourceObjectID); ok {
		t.Fatal("modular source still on battlefield after lethal damage + SBAs")
	}
	snapshot, ok := lastKnownObject(g, sourceObjectID)
	if !ok {
		t.Fatal("no last-known information for dead modular source")
	}
	if got := snapshot.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("LKI +1/+1 counters = %d, want 2", got)
	}

	// The lone legal target (the artifact creature) is auto-selected at stacking;
	// the only choice consumed is the optional "may" decision (yes) at resolution.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("modular dies-trigger was not put on the stack after death")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (modular dies-trigger)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("target artifact creature +1/+1 counters = %d, want 2 (received modular counters via LKI)", got)
	}
	if got := target.Counters.Get(counter.Stun); got != 0 {
		t.Fatalf("target stun counters = %d, want 0 (Modular moves only +1/+1 counters)", got)
	}
}

// TestNonDeathCounterSourceSelfMoveIgnoresDepartedSource proves the modular
// last-known-information fallback is scoped to a self-dies trigger and does not
// regress the other CounterSourceSelf "move a counter from this permanent"
// families. A source carrying a +1/+1 counter is killed (recording last-known
// information) before a NON-dies move resolves; because the counters ceased to
// exist when the source left (CR 121.5), the destination gains nothing. Both the
// graft ETB trigger ("Whenever another creature enters, you may move a +1/+1
// counter from this creature onto it.") and the "{T}: move a counter from this
// permanent" activated abilities (Explorer's Cache, Diamond City) are covered.
// This fails on the unconditional fallback, which reads the departed source's
// last-known counter and wrongly moves it.
func TestNonDeathCounterSourceSelfMoveIgnoresDepartedSource(t *testing.T) {
	move := game.MoveCounters{
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.PlusOnePlusOne,
		Amount:      game.Fixed(1),
		Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
	}
	cases := []struct {
		name     string
		stackObj func(source, target game.ObjectID) *game.StackObject
	}{
		{
			name: "graft ETB trigger",
			stackObj: func(source, target game.ObjectID) *game.StackObject {
				return &game.StackObject{
					Controller:      game.Player1,
					SourceID:        source,
					Targets:         []game.Target{game.PermanentTarget(target)},
					HasTriggerEvent: true,
					TriggerEvent:    game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: target},
				}
			},
		},
		{
			name: "activated ability",
			stackObj: func(source, target game.ObjectID) *game.StackObject {
				return &game.StackObject{
					Controller: game.Player1,
					SourceID:   source,
					Targets:    []game.Target{game.PermanentTarget(target)},
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)

			source := addCombatPermanent(g, game.Player1, artifactCreatureDef("Aquastrand Spider"))
			sourceObjectID := source.ObjectID
			source.Counters.Add(counter.PlusOnePlusOne, 1)

			target := addCombatPermanent(g, game.Player1, artifactCreatureDef("Steel Overseer"))

			// Kill the source (effective toughness 1 base + 1 counter = 2) so it
			// leaves the battlefield and records last-known information before the
			// move resolves.
			source.MarkedDamage = 2
			g.AppendEvent(game.Event{
				Kind:            game.EventDamageDealt,
				PermanentID:     source.ObjectID,
				DamageRecipient: game.DamageRecipientPermanent,
				Amount:          2,
			})
			engine.applyStateBasedActions(g)

			if _, ok := permanentByObjectID(g, sourceObjectID); ok {
				t.Fatal("source still on battlefield after lethal damage + SBAs")
			}
			snapshot, ok := lastKnownObject(g, sourceObjectID)
			if !ok || snapshot.Counters.Get(counter.PlusOnePlusOne) != 1 {
				t.Fatal("expected last-known information with one +1/+1 counter for the dead source")
			}

			resolveInstruction(engine, g, tc.stackObj(sourceObjectID, target.ObjectID), move, nil)

			if got := target.Counters.Get(counter.PlusOnePlusOne); got != 0 {
				t.Fatalf("target +1/+1 counters = %d, want 0 (a non-dies move from a departed source moves nothing)", got)
			}
		})
	}
}

// TestBlasterFrontConvertsWhenPlusOneCounterPlaced proves the front face trigger
// "Whenever you put one or more +1/+1 counters on Blaster, convert it.": placing
// a +1/+1 counter on Blaster fires the self-sourced counters-added trigger and
// transforms it, while an unrelated counter kind (Charge) never triggers.
func TestBlasterFrontConvertsWhenPlusOneCounterPlaced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	blaster := addCombatPermanent(g, game.Player1, cardsb.BlasterCombatDJ())
	blaster.Face = game.FaceFront

	addCountersToPermanent(g, blaster, counter.Charge, 1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("a Charge counter fired the +1/+1 convert trigger")
	}
	if blaster.Transformed {
		t.Fatal("Blaster converted from a Charge counter")
	}

	addCountersToPermanent(g, blaster, counter.PlusOnePlusOne, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("placing a +1/+1 counter did not fire the convert trigger")
	}
	engine.resolveTopOfStack(g, nil)
	if blaster.Face != game.FaceBack || !blaster.Transformed {
		t.Fatalf("Blaster face/transformed = %v/%v, want back/true (convert it)", blaster.Face, blaster.Transformed)
	}
}

// TestBlasterBackMovesXCountersGrantsHasteAndConverts proves the back face
// activation "{X}, {T}: Move X +1/+1 counters from Blaster onto another target
// artifact. That artifact gains haste until end of turn. If Blaster has no
// +1/+1 counters on it, convert it. Activate only as a sorcery.": moving all of
// Blaster's counters transfers exactly X, grants the target haste, and — because
// Blaster is left with no +1/+1 counters — converts it.
func TestBlasterBackMovesXCountersGrantsHasteAndConverts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	blaster := addCombatPermanent(g, game.Player1, cardsb.BlasterCombatDJ())
	blaster.Face = game.FaceBack
	blaster.Transformed = true
	blaster.Counters.Add(counter.PlusOnePlusOne, 3)

	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ornithopter",
		Types: []types.Card{types.Artifact},
	}})
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act, ok := findActivateAction(engine.legalActions(g, game.Player1), blaster.ObjectID, 3)
	if !ok {
		t.Fatal("moving X=3 counters was not a legal sorcery-speed activation")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate Blaster back) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("target artifact +1/+1 counters = %d, want 3 (X moved)", got)
	}
	if got := blaster.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("Blaster +1/+1 counters after move = %d, want 0", got)
	}
	if !hasKeyword(g, target, game.Haste) {
		t.Fatal("target artifact did not gain haste until end of turn")
	}
	if blaster.Face != game.FaceFront || blaster.Transformed {
		t.Fatalf("Blaster face/transformed = %v/%v, want front/false (convert on empty)", blaster.Face, blaster.Transformed)
	}
}

// TestBlasterBackKeepsCountersDoesNotConvert proves the conditional convert only
// fires when Blaster is left empty: moving fewer than all of its +1/+1 counters
// (X=2 of 3) still grants haste but leaves a counter behind, so Blaster does not
// convert.
func TestBlasterBackKeepsCountersDoesNotConvert(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	blaster := addCombatPermanent(g, game.Player1, cardsb.BlasterCombatDJ())
	blaster.Face = game.FaceBack
	blaster.Transformed = true
	blaster.Counters.Add(counter.PlusOnePlusOne, 3)

	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ornithopter",
		Types: []types.Card{types.Artifact},
	}})
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act, ok := findActivateAction(engine.legalActions(g, game.Player1), blaster.ObjectID, 2)
	if !ok {
		t.Fatal("moving X=2 counters was not a legal activation")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate Blaster back) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("target artifact +1/+1 counters = %d, want 2 (X=2 moved)", got)
	}
	if got := blaster.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Blaster +1/+1 counters = %d, want 1 (one kept)", got)
	}
	if !hasKeyword(g, target, game.Haste) {
		t.Fatal("target artifact did not gain haste until end of turn")
	}
	if blaster.Face != game.FaceBack || !blaster.Transformed {
		t.Fatalf("Blaster face/transformed = %v/%v, want back/true (still has a counter)", blaster.Face, blaster.Transformed)
	}
}
