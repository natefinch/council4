package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#3872): pinned corpus 6da9ced0 Aura Graft
// (00317e4f-0228-469a-9789-fb6c39976860, printing 55552a2b 10e) {1}{U}
// Instant: "Gain control of target Aura that's attached to a permanent. Attach
// it to another permanent it can enchant." Rulings (2007-07-15, 2004-10-04):
// the control change has no duration; an Aura you already control may be
// targeted; with no legal place to move it, it doesn't move but you still
// control it. Yuffie, Materia Hunter (acdcbb49-9d71-4c36-8ccb-650fb11c3a58,
// printing 117bb5db fic) {2}{R} Legendary Creature — Human Ninja 3/3:
// "Ninjutsu {1}{R}. When Yuffie enters, gain control of target noncreature
// artifact for as long as you control Yuffie. Then you may attach an Equipment
// you control to Yuffie." Rulings (2025-06-06): an illegal target means no
// Equipment is attached; you can't attach an Equipment that can't legally be
// attached to Yuffie; if Yuffie leaves before resolution you gain no control.
// CR 20250404 read: 301.5c (Equipment can't equip itself or a noncreature),
// 303.4j (an Aura doesn't move to an object it can't legally enchant), 608.2c
// (instructions in order; the chosen object is not a target), 701.3a-c (attach;
// an illegal attach doesn't move the attachment; a new timestamp), 702.16c-d
// (protection prevents enchanting and equipping). These scenarios use
// test-local shapes of the cards' own text; later card units admit the cards.

func auraGraftTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target Aura that's attached to a permanent",
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			SubtypesAny:              []types.Sub{types.Aura},
			MatchAttachedToPermanent: true,
		}),
	}
}

func auraGraftShape() *game.CardDef {
	target := game.TargetPermanentReference(0)
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Aura Graft",
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Blue},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{auraGraftTarget()},
			Sequence: []game.Instruction{
				{Primitive: game.ApplyContinuous{
					Object: opt.Val(target),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:         game.LayerControl,
						NewController: opt.Val(game.Player1),
					}},
					Duration: game.DurationPermanent,
				}},
				{Primitive: game.Attach{
					Attachment:       target,
					ChooseTargetFrom: game.BattlefieldGroupExcluding(game.Selection{}, game.TargetAttachedPermanentReference(0)),
					Prompt:           "Attach it to another permanent it can enchant",
				}},
			},
		}.Ability()),
	}}
}

func yuffieShape() *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Test Yuffie",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Subtypes:   []types.Sub{types.Human, types.Ninja},
		Colors:     []color.Color{color.Red},
		ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R}),
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "When Yuffie enters, gain control of target noncreature artifact for as long as you control Yuffie. Then you may attach an Equipment you control to Yuffie.",
			Trigger: game.TriggerCondition{
				Type:    game.TriggerWhen,
				Pattern: game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target noncreature artifact",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Artifact},
						ExcludedTypes:    []types.Card{types.Creature},
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						}},
						Duration: game.DurationForAsLongAsYouControlSource,
					}},
					{Primitive: game.Attach{
						ChooseAttachmentFrom: game.PlayerControlledGroup(game.ControllerReference(), game.Selection{
							SubtypesAny: []types.Sub{types.Equipment},
						}),
						Target:   game.SourcePermanentReference(),
						Optional: true,
						Prompt:   "You may attach an Equipment you control to Yuffie",
					}},
				},
			}.Ability(),
		}},
	}}
}

// attachChoiceAgent targets target, then answers each resolution choice by
// picking the option for pick (declining when pick is zero) and records the
// cards each resolution choice offered.
type attachChoiceAgent struct {
	target  game.Target
	pick    id.ID
	offered *[][]id.ID
	minimum *[]int
}

func (attachChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a attachChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceTarget:
		for _, option := range request.Options {
			if slices.Contains(option.Targets, a.target) {
				return []int{option.Index}
			}
		}
	case game.ChoiceResolution:
		cards := make([]id.ID, 0, len(request.Options))
		for _, option := range request.Options {
			cards = append(cards, option.Card.Val.CardID)
		}
		*a.offered = append(*a.offered, cards)
		*a.minimum = append(*a.minimum, request.MinChoices)
		for _, option := range request.Options {
			if a.pick != 0 && option.Card.Val.CardID == a.pick {
				return []int{option.Index}
			}
		}
		return []int{}
	}
	return request.DefaultSelection
}

func newAttachChoiceAgent(target *game.Permanent, pick *game.Permanent) attachChoiceAgent {
	agent := attachChoiceAgent{offered: &[][]id.ID{}, minimum: &[]int{}}
	if target != nil {
		agent.target = game.PermanentTarget(target.ObjectID)
	}
	if pick != nil {
		agent.pick = pick.CardInstanceID
	}
	return agent
}

func requireOffered(t *testing.T, agent attachChoiceAgent, wantMin int, want ...*game.Permanent) {
	t.Helper()
	if len(*agent.offered) != 1 {
		t.Fatalf("resolution choices = %d, want 1", len(*agent.offered))
	}
	wantIDs := make([]id.ID, 0, len(want))
	for _, permanent := range want {
		wantIDs = append(wantIDs, permanent.CardInstanceID)
	}
	got := slices.Clone((*agent.offered)[0])
	slices.Sort(got)
	slices.Sort(wantIDs)
	if !slices.Equal(got, wantIDs) {
		t.Fatalf("offered cards = %v, want %v", got, wantIDs)
	}
	if (*agent.minimum)[0] != wantMin {
		t.Fatalf("minimum choices = %d, want %d", (*agent.minimum)[0], wantMin)
	}
}

func requireAttachedTo(t *testing.T, attachment, host *game.Permanent) {
	t.Helper()
	if !attachment.AttachedTo.Exists || attachment.AttachedTo.Val != host.ObjectID ||
		!slices.Contains(host.Attachments, attachment.ObjectID) {
		t.Fatalf("attachment %v attached to %v, want %v", attachment.ObjectID, attachment.AttachedTo, host.ObjectID)
	}
}

func castAuraGraft(t *testing.T, g *game.Game, engine *Engine, aura *game.Permanent, agent attachChoiceAgent) id.ID {
	t.Helper()
	if issues := game.ValidateCardDef(auraGraftShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	cardID := addCardToHand(g, game.Player1, auraGraftShape())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(aura.ObjectID)}, 0, nil)) {
		t.Fatal("casting Aura Graft failed")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	return cardID
}

func TestAuraGraftMovesAuraToChosenPermanentItCanEnchant(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	host := addCreatureWithManaValue(g, game.Player2, 2)
	aura := addAuraPermanent(g, game.Player2)
	if !attachPermanent(g, aura, host) {
		t.Fatal("setup attach failed")
	}
	mine := addCreatureWithManaValue(g, game.Player1, 1)
	theirs := addCreatureWithManaValue(g, game.Player3, 3)
	addProtectionFromTypesPermanent(g, game.Player3, types.Enchantment)
	addEquipmentPermanent(g, game.Player2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Land", Types: []types.Card{types.Land}}})
	before := aura.Timestamp()

	agent := newAttachChoiceAgent(nil, theirs)
	castAuraGraft(t, g, engine, aura, agent)

	requireOffered(t, agent, 1, mine, theirs)
	requireAttachedTo(t, aura, theirs)
	if slices.Contains(host.Attachments, aura.ObjectID) {
		t.Fatal("Aura still listed on its original host")
	}
	if aura.Timestamp() <= before {
		t.Fatal("moved Aura kept its old timestamp (CR 701.3c)")
	}
	if got := effectiveController(g, aura); got != game.Player1 {
		t.Fatalf("Aura controller = %v, want Player1", got)
	}
}

func TestAuraGraftWithNoOtherLegalHostKeepsAuraButTakesControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	host := addCreatureWithManaValue(g, game.Player2, 2)
	aura := addAuraPermanent(g, game.Player2)
	if !attachPermanent(g, aura, host) {
		t.Fatal("setup attach failed")
	}
	addProtectionFromTypesPermanent(g, game.Player3, types.Enchantment)
	addEquipmentPermanent(g, game.Player1)

	agent := newAttachChoiceAgent(nil, nil)
	cardID := castAuraGraft(t, g, engine, aura, agent)

	if len(*agent.offered) != 0 {
		t.Fatalf("resolution choices = %v, want none without a legal host", *agent.offered)
	}
	requireAttachedTo(t, aura, host)
	if got := effectiveController(g, aura); got != game.Player1 {
		t.Fatalf("Aura controller = %v, want Player1", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Aura Graft did not finish resolving")
	}
}

func castYuffieAndTrigger(t *testing.T, g *game.Game, engine *Engine, agent attachChoiceAgent) *game.Permanent {
	t.Helper()
	if issues := game.ValidateCardDef(yuffieShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	cardID := addCardToHand(g, game.Player1, yuffieShape())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Yuffie failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	yuffie := permanentForCard(g, cardID)
	if yuffie == nil {
		t.Fatal("Yuffie did not enter")
	}
	agents := simPassPolicies()
	agents[game.Player1] = agent
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Yuffie's enters trigger was not put on the stack")
	}
	return yuffie
}

func TestYuffieMayAttachAnEquipmentYouControlIncludingTheStolenOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	theirCreature := addCreatureWithManaValue(g, game.Player2, 2)
	stolen := addEquipmentPermanent(g, game.Player2)
	if !attachPermanent(g, stolen, theirCreature) {
		t.Fatal("setup attach failed")
	}
	mine := addEquipmentPermanent(g, game.Player1)
	addEquipmentPermanent(g, game.Player3)
	addAuraPermanent(g, game.Player1)

	agent := newAttachChoiceAgent(stolen, stolen)
	yuffie := castYuffieAndTrigger(t, g, engine, agent)
	declined := g.Clone()
	declineAgent := newAttachChoiceAgent(stolen, nil)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	requireOffered(t, agent, 0, stolen, mine)
	requireAttachedTo(t, stolen, yuffie)
	if slices.Contains(theirCreature.Attachments, stolen.ObjectID) {
		t.Fatal("stolen Equipment still listed on its previous creature")
	}
	if mine.AttachedTo.Exists {
		t.Fatal("unchosen Equipment became attached")
	}

	engine.resolveTopOfStackWithChoices(declined, [game.NumPlayers]PlayerAgent{game.Player1: declineAgent}, &TurnLog{})
	requireOffered(t, declineAgent, 0, stolen, mine)
	declinedStolen, _ := permanentByObjectID(declined, stolen.ObjectID)
	declinedCreature, _ := permanentByObjectID(declined, theirCreature.ObjectID)
	requireAttachedTo(t, declinedStolen, declinedCreature)
	if got := effectiveController(declined, declinedStolen); got != game.Player1 {
		t.Fatalf("declined branch controller = %v, want Player1", got)
	}

	if !movePermanentToZone(g, yuffie, zone.Graveyard) {
		t.Fatal("moving Yuffie to the graveyard failed")
	}
	engine.applyStateBasedActions(g)
	if got := effectiveController(g, stolen); got != game.Player2 {
		t.Fatalf("controller after Yuffie left = %v, want Player2", got)
	}
	if stolen.AttachedTo.Exists {
		t.Fatal("Equipment stayed attached to a departed creature")
	}
}

func TestAuraGraftProgramAttachesChosenHostAndPublishesAttachment(t *testing.T) {
	const graftedLink game.LinkedKey = "test/aura-graft/grafted"
	reference := game.ProgramReference{ID: "test/aura-graft", Version: "1"}
	var outcome cardprogram.Outcome
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			aura := game.TargetPermanentReference(0)
			if _, err := ctx.Execute(game.ApplyContinuous{
				Object:            opt.Val(aura),
				ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerControl, NewController: opt.Val(game.Player1)}},
				Duration:          game.DurationPermanent,
			}); err != nil {
				return err
			}
			var err error
			if outcome, err = ctx.Execute(game.Attach{
				Attachment:       aura,
				ChooseTargetFrom: game.BattlefieldGroupExcluding(game.Selection{}, game.TargetAttachedPermanentReference(0)),
				PublishLinked:    graftedLink,
			}); err != nil || !outcome.Succeeded {
				return err
			}
			_, err = ctx.Execute(game.Tap{Object: game.LinkedObjectReference(string(graftedLink))})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	host := addCreatureWithManaValue(g, game.Player2, 2)
	aura := addAuraPermanent(g, game.Player2)
	if !attachPermanent(g, aura, host) {
		t.Fatal("setup attach failed")
	}
	newHost := addCreatureWithManaValue(g, game.Player1, 1)
	def := auraGraftShape()
	def.SpellAbility = opt.Val(game.Mode{Targets: []game.TargetSpec{auraGraftTarget()}, Program: reference}.Ability())
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(aura.ObjectID)}, 0, nil)) {
		t.Fatal("casting the program Aura Graft failed")
	}
	agent := newAttachChoiceAgent(nil, newHost)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})

	requireOffered(t, agent, 1, newHost)
	requireAttachedTo(t, aura, newHost)
	if !outcome.Succeeded || len(outcome.Objects) != 1 || outcome.Objects[0].ObjectID != aura.ObjectID {
		t.Fatalf("attach outcome = %+v, want the attached Aura", outcome)
	}
	if !aura.Tapped {
		t.Fatal("published linked attachment was not available to the next operation")
	}
}

func TestAttachChoiceValidation(t *testing.T) {
	group := game.BattlefieldGroup(game.Selection{})
	source := game.SourcePermanentReference()
	for name, prim := range map[string]game.Attach{
		"both sides chosen":         {ChooseAttachmentFrom: group, ChooseTargetFrom: group},
		"chosen and referenced":     {Attachment: source, ChooseAttachmentFrom: group, Target: source},
		"chosen object referenced":  {Attachment: source, Target: source, ChooseTargetFrom: group},
		"optional without a choice": {Attachment: source, Target: source, Optional: true},
	} {
		if err := game.ValidatePrimitive(prim, nil); err == nil {
			t.Errorf("%s: validation passed, want an error", name)
		}
	}
	if err := game.ValidatePrimitive(game.Attach{Attachment: source, ChooseTargetFrom: group, Optional: true}, nil); err != nil {
		t.Fatalf("chosen object attach: %v", err)
	}
}
