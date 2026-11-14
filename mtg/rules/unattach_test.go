package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#3873): pinned corpus 6da9ced0 Unexpected Request
// (f4497af5-8b81-46b3-bb52-eaed1999809b, printing 0265fd20) {2}{R} Sorcery:
// "Gain control of target creature until end of turn. Untap that creature. It
// gains haste until end of turn. You may attach an Equipment you control to
// that creature. If you do, unattach it at the beginning of the next end step."
// Scryfall rulings: none. Ogre Geargrabber's matching delayed "unattach it"
// ruling (2011-01-01): the Equipment becomes unattached from whatever creature
// it happens to be attached to at that time. CR 20261106 read: 301.5c-d
// (Equipment attaches only to creatures; its controller is separate from the
// equipped creature's), 400.7 (a permanent that leaves and returns is a new
// object), 603.7c (a delayed trigger referring to a particular object affects
// it only while it is still that object), 608.2c, 611.2a-b, 613.7e (only
// attaching gives a new timestamp), 701.3a-d (attach; "unattach" moves an
// Equipment away from the creature it is attached to, and it remains on the
// battlefield). These scenarios use a test-local shape of the card's text; a
// later card unit admits it.

const (
	unexpectedRequestAttached game.LinkedKey = "test/unexpected-request/attached"
	unexpectedRequestAttach   game.ResultKey = "test/unexpected-request/attach"
)

func unexpectedRequestTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
	}
}

func unexpectedRequestAttachEquipment() game.Attach {
	return game.Attach{
		ChooseAttachmentFrom: game.PlayerControlledGroup(game.ControllerReference(), game.Selection{
			SubtypesAny: []types.Sub{types.Equipment},
		}),
		Target:        game.TargetPermanentReference(0),
		Optional:      true,
		Prompt:        "You may attach an Equipment you control to that creature",
		PublishLinked: unexpectedRequestAttached,
	}
}

func unexpectedRequestUnattachAtEndStep() game.CreateDelayedTrigger {
	return game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing:         game.DelayedAtBeginningOfNextEndStep,
		CapturedObject: opt.Val(game.LinkedObjectReference(string(unexpectedRequestAttached))),
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Unattach{Attachment: game.CapturedObjectReference()},
		}}}.Ability(),
	}}
}

func unexpectedRequestStealUntapHaste() []game.Instruction {
	target := game.TargetPermanentReference(0)
	return []game.Instruction{
		{Primitive: game.ApplyContinuous{
			Object: opt.Val(target),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:         game.LayerControl,
				NewController: opt.Val(game.Player1),
			}},
			Duration: game.DurationUntilEndOfTurn,
		}},
		{Primitive: game.Untap{Object: target}},
		{Primitive: game.ApplyContinuous{
			Object: opt.Val(target),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				AddKeywords: []game.Keyword{game.Haste},
			}},
			Duration: game.DurationUntilEndOfTurn,
		}},
	}
}

func unexpectedRequestShape() *game.CardDef {
	sequence := append(unexpectedRequestStealUntapHaste(),
		game.Instruction{
			Primitive:     unexpectedRequestAttachEquipment(),
			PublishResult: unexpectedRequestAttach,
			LocalProducts: game.LocalProducts{
				Results: []game.ResultKey{unexpectedRequestAttach},
				Links:   []game.LinkedKey{unexpectedRequestAttached},
			},
		},
		game.Instruction{
			Primitive:  unexpectedRequestUnattachAtEndStep(),
			ResultGate: opt.Val(game.InstructionResultGate{Key: unexpectedRequestAttach, Succeeded: game.TriTrue}),
		},
	)
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Unexpected Request",
		Types:    []types.Card{types.Sorcery},
		Colors:   []color.Color{color.Red},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{unexpectedRequestTarget()},
			Sequence: sequence,
		}.Ability()),
	}}
}

func castUnexpectedRequest(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, creature *game.Permanent, agent attachChoiceAgent) {
	t.Helper()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("casting Unexpected Request failed")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
}

// unexpectedRequestSetup casts the shape on a tapped Player2 creature, attaching
// a Player1 Equipment, and returns the stolen creature and the Equipment.
func unexpectedRequestSetup(t *testing.T, def *game.CardDef, engine *Engine) (*game.Game, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	creature := addCreatureWithManaValue(g, game.Player2, 3)
	creature.Tapped = true
	equipment := addEquipmentPermanent(g, game.Player1)
	agent := newAttachChoiceAgent(creature, equipment)
	castUnexpectedRequest(t, g, engine, def, creature, agent)
	requireOffered(t, agent, 0, equipment)
	requireAttachedTo(t, equipment, creature)
	if got := effectiveController(g, creature); got != game.Player1 || creature.Tapped || !hasKeyword(g, creature, game.Haste) {
		t.Fatalf("stolen creature controller %v tapped %v haste %v, want untapped hasty under Player1",
			got, creature.Tapped, hasKeyword(g, creature, game.Haste))
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want the end-step unattach", len(g.DelayedTriggers))
	}
	return g, creature, equipment
}

func requireUnattachedOnBattlefield(t *testing.T, g *game.Game, equipment, formerHost *game.Permanent, controller game.PlayerID) {
	t.Helper()
	if _, ok := permanentByObjectID(g, equipment.ObjectID); !ok {
		t.Fatal("unattached Equipment left the battlefield")
	}
	if equipment.AttachedTo.Exists {
		t.Fatalf("Equipment still attached to %v", equipment.AttachedTo.Val)
	}
	for _, attachmentID := range formerHost.Attachments {
		if attachmentID == equipment.ObjectID {
			t.Fatal("former host still lists the unattached Equipment")
		}
	}
	if got := effectiveController(g, equipment); got != controller {
		t.Fatalf("Equipment controller = %v, want %v", got, controller)
	}
}

func TestUnexpectedRequestUnattachesEquipmentAtBeginningOfNextEndStep(t *testing.T) {
	g, creature, equipment := unexpectedRequestSetup(t, unexpectedRequestShape(), NewEngine(nil))
	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	requireUnattachedOnBattlefield(t, g, equipment, creature, game.Player1)
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("creature controller after the turn = %v, want Player2", got)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after the end step = %d, want 0", len(g.DelayedTriggers))
	}
}

func TestUnexpectedRequestDecliningAttachCreatesNoUnattachTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	creature := addCreatureWithManaValue(g, game.Player2, 3)
	equipment := addEquipmentPermanent(g, game.Player1)
	agent := newAttachChoiceAgent(creature, nil)
	castUnexpectedRequest(t, g, engine, unexpectedRequestShape(), creature, agent)

	requireOffered(t, agent, 0, equipment)
	if equipment.AttachedTo.Exists {
		t.Fatal("declined Equipment was attached")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want none when nothing was attached", len(g.DelayedTriggers))
	}
}

// The delayed trigger tracks the Equipment, not the attachment relationship: a
// re-equipped Equipment is the same object and is unattached from whatever it
// is attached to when the trigger resolves (Ogre Geargrabber ruling, CR 603.7c).
func TestUnexpectedRequestUnattachesReEquippedEquipmentFromItsCurrentCreature(t *testing.T) {
	g, creature, equipment := unexpectedRequestSetup(t, unexpectedRequestShape(), NewEngine(nil))
	other := addCreatureWithManaValue(g, game.Player1, 1)
	if !attachPermanent(g, equipment, other) {
		t.Fatal("re-equipping failed")
	}
	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	requireUnattachedOnBattlefield(t, g, equipment, other, game.Player1)
	if len(creature.Attachments) != 0 {
		t.Fatalf("stolen creature attachments = %v, want none", creature.Attachments)
	}
}

// An Equipment that left the battlefield is a new object afterward (CR 400.7),
// so the delayed trigger does nothing, and other Equipment on the creature is
// never touched.
func TestUnexpectedRequestUnattachIgnoresDepartedEquipmentAndOtherAttachments(t *testing.T) {
	g, creature, equipment := unexpectedRequestSetup(t, unexpectedRequestShape(), NewEngine(nil))
	if !movePermanentToZone(g, equipment, zone.Graveyard) {
		t.Fatal("moving the Equipment to the graveyard failed")
	}
	other := addEquipmentPermanent(g, game.Player1)
	if !attachPermanent(g, other, creature) {
		t.Fatal("attaching the other Equipment failed")
	}
	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	requireAttachedTo(t, other, creature)
}

func TestUnexpectedRequestUnattachTriggerIsCloneSafe(t *testing.T) {
	g, creature, equipment := unexpectedRequestSetup(t, unexpectedRequestShape(), NewEngine(nil))
	clone := g.Clone()
	NewEngine(nil).runEndingPhase(clone, [game.NumPlayers]PlayerAgent{})

	clonedEquipment, ok := permanentByObjectID(clone, equipment.ObjectID)
	if !ok {
		t.Fatal("cloned Equipment missing")
	}
	clonedCreature, ok := permanentByObjectID(clone, creature.ObjectID)
	if !ok {
		t.Fatal("cloned creature missing")
	}
	requireUnattachedOnBattlefield(t, clone, clonedEquipment, clonedCreature, game.Player1)
	requireAttachedTo(t, equipment, creature)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("original delayed triggers = %d, want 1 after the clone resolved its copy", len(g.DelayedTriggers))
	}

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	requireUnattachedOnBattlefield(t, g, equipment, creature, game.Player1)
}

func TestUnexpectedRequestProgramSchedulesUnattachOnlyAfterAttaching(t *testing.T) {
	reference := game.ProgramReference{ID: "test/unexpected-request", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			for _, instruction := range unexpectedRequestStealUntapHaste() {
				if _, err := ctx.Execute(instruction.Primitive); err != nil {
					return err
				}
			}
			attached, err := ctx.Execute(unexpectedRequestAttachEquipment())
			if err != nil || !attached.Succeeded {
				return err
			}
			_, err = ctx.Execute(unexpectedRequestUnattachAtEndStep())
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	def := unexpectedRequestShape()
	def.SpellAbility = opt.Val(game.Mode{Targets: []game.TargetSpec{unexpectedRequestTarget()}, Program: reference}.Ability())

	g, creature, equipment := unexpectedRequestSetup(t, def, NewEngineWithPrograms(nil, catalog))
	NewEngineWithPrograms(nil, catalog).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	requireUnattachedOnBattlefield(t, g, equipment, creature, game.Player1)
}

// Unattach succeeds only when it moves an Equipment off its creature; an
// unattached permanent or an Aura is left alone.
func TestUnattachProgramOutcome(t *testing.T) {
	reference := game.ProgramReference{ID: "test/unattach", Version: "1"}
	var outcome cardprogram.Outcome
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			var err error
			outcome, err = ctx.Execute(game.Unattach{Attachment: game.TargetPermanentReference(0)})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Unattach",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
			Program: reference,
		}.Ability()),
	}}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngineWithPrograms(nil, catalog)
	host := addCreatureWithManaValue(g, game.Player2, 2)
	equipment := addEquipmentPermanent(g, game.Player2)
	aura := addAuraPermanent(g, game.Player2)
	if !attachPermanent(g, equipment, host) || !attachPermanent(g, aura, host) {
		t.Fatal("setup attach failed")
	}
	unattach := func(target *game.Permanent) bool {
		t.Helper()
		outcome = cardprogram.Outcome{}
		cardID := addCardToHand(g, game.Player1, def)
		if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
			t.Fatal("casting the unattach spell failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		return outcome.Succeeded
	}

	if !unattach(equipment) {
		t.Fatal("unattaching an attached Equipment did not succeed")
	}
	requireUnattachedOnBattlefield(t, g, equipment, host, game.Player2)
	if unattach(equipment) {
		t.Fatal("unattaching an unattached Equipment succeeded")
	}
	if unattach(aura) {
		t.Fatal("unattaching an Aura succeeded")
	}
	requireAttachedTo(t, aura, host)
}
