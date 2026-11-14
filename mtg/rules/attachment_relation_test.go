package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review (#3871): pinned corpus 6da9ced0 Awaken the Sleeper
// (d4df1710-7402-45a7-966f-b51ddc087fbe, printing 3b92f866 one) {3}{R}
// Sorcery: "Gain control of target creature until end of turn. Untap that
// creature. It gains haste until end of turn. If it's equipped, you may
// destroy all Equipment attached to that creature." Scryfall lists no rulings
// (retrieved 2026-11-14). Aura Graft (00317e4f-0228-469a-9789-fb6c39976860,
// printing 55552a2b 10e) {1}{U} Instant: "Gain control of target Aura that's
// attached to a permanent. Attach it to another permanent it can enchant."
// Rulings (2007-07-15, 2004-10-04): the control change has no duration; an
// Aura you already control may be targeted; with no legal new host it stays
// put but you still control it. Official CR effective 2026-11-06
// (MagicCompRules 20261106.txt) read: 301.5a/d (an Equipment is attached to the
// creature it equips; control of each is separate), 303.4b/h (an Aura may be
// attached to an object or a player), 608.2b (targets are rechecked on
// resolution), 701.3a/d (attach and become unattached). These scenarios use
// test-local spells; the cards themselves are authored by later card units.

func awakenTheSleeperShape() *game.CardDef {
	target := game.TargetPermanentReference(0)
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Awaken the Sleeper",
		Types:    []types.Card{types.Sorcery},
		Colors:   []color.Color{color.Red},
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{
				{Primitive: game.ApplyContinuous{
					Object: opt.Val(target),
					ContinuousEffects: []game.ContinuousEffect{{
						Duration:      game.DurationPermanent,
						Layer:         game.LayerControl,
						NewController: opt.Val(game.Player1),
					}},
					Duration: game.DurationUntilEndOfTurn,
				}},
				{Primitive: game.Untap{Object: target}},
				{Primitive: game.ApplyContinuous{
					Object: opt.Val(target),
					ContinuousEffects: []game.ContinuousEffect{{
						Duration:    game.DurationPermanent,
						Layer:       game.LayerAbility,
						AddKeywords: []game.Keyword{game.Haste},
					}},
					Duration: game.DurationUntilEndOfTurn,
				}},
				{
					Primitive: game.Destroy{Group: game.AttachmentsGroup(target, game.Selection{
						SubtypesAny: []types.Sub{types.Equipment},
					})},
					Optional: true,
					Condition: opt.Val(game.EffectCondition{
						Text: "If it's equipped",
						Condition: opt.Val(game.Condition{
							Object:        opt.Val(target),
							ObjectMatches: opt.Val(game.Selection{MatchEquipped: true}),
						}),
					}),
				},
			},
		}.Ability()),
	}}
}

// mayCountingAgent answers every may-choice with accept and counts them.
type mayCountingAgent struct {
	accept bool
	asked  *int
}

func (mayCountingAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a mayCountingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		*a.asked++
		if a.accept {
			return []int{1}
		}
	}
	return []int{0}
}

func castAwakenTheSleeper(t *testing.T, g *game.Game, engine *Engine, target *game.Permanent, accept bool) int {
	t.Helper()
	if issues := game.ValidateCardDef(awakenTheSleeperShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	cardID := addCardToHand(g, game.Player1, awakenTheSleeperShape())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("casting Awaken the Sleeper failed")
	}
	asked := 0
	agents := [game.NumPlayers]PlayerAgent{game.Player1: mayCountingAgent{accept: accept, asked: &asked}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	return asked
}

func requireAttachmentOnBattlefield(t *testing.T, g *game.Game, permanent *game.Permanent, want bool, what string) {
	t.Helper()
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok != want {
		t.Fatalf("%s on battlefield = %v, want %v", what, ok, want)
	}
}

func TestAwakenTheSleeperDestroysOnlyEquipmentAttachedToTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	stolen := addCreatureWithManaValue(g, game.Player2, 3)
	stolen.Tapped = true
	sword := addEquipmentPermanent(g, game.Player2)
	axe := addEquipmentPermanent(g, game.Player3)
	aura := addAuraPermanent(g, game.Player2)
	other := addCreatureWithManaValue(g, game.Player2, 2)
	otherGear := addEquipmentPermanent(g, game.Player2)
	looseGear := addEquipmentPermanent(g, game.Player2)
	for _, gear := range []*game.Permanent{sword, axe, aura} {
		if !attachPermanent(g, gear, stolen) {
			t.Fatal("setup attach failed")
		}
	}
	if !attachPermanent(g, otherGear, other) {
		t.Fatal("setup attach failed")
	}

	if asked := castAwakenTheSleeper(t, g, engine, stolen, true); asked != 1 {
		t.Fatalf("destroy choices offered = %d, want 1", asked)
	}

	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller = %v, want Player1", got)
	}
	if stolen.Tapped || !hasKeyword(g, stolen, game.Haste) {
		t.Fatalf("stolen creature tapped=%v haste=%v, want untapped with haste", stolen.Tapped, hasKeyword(g, stolen, game.Haste))
	}
	requireAttachmentOnBattlefield(t, g, sword, false, "attached Equipment")
	requireAttachmentOnBattlefield(t, g, axe, false, "attached Equipment controlled by another player")
	requireAttachmentOnBattlefield(t, g, aura, true, "attached Aura")
	requireAttachmentOnBattlefield(t, g, otherGear, true, "Equipment attached elsewhere")
	requireAttachmentOnBattlefield(t, g, looseGear, true, "unattached Equipment")

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := effectiveController(g, stolen); got != game.Player2 {
		t.Fatalf("controller after cleanup = %v, want Player2", got)
	}
	if hasKeyword(g, stolen, game.Haste) {
		t.Fatal("haste survived cleanup")
	}
}

func TestAwakenTheSleeperOffersNoDestroyForUnequippedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	stolen := addCreatureWithManaValue(g, game.Player2, 3)
	aura := addAuraPermanent(g, game.Player2)
	if !attachPermanent(g, aura, stolen) {
		t.Fatal("setup attach failed")
	}
	looseGear := addEquipmentPermanent(g, game.Player2)

	if asked := castAwakenTheSleeper(t, g, engine, stolen, true); asked != 0 {
		t.Fatalf("destroy choices offered = %d, want 0", asked)
	}
	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller = %v, want Player1", got)
	}
	requireAttachmentOnBattlefield(t, g, aura, true, "attached Aura")
	requireAttachmentOnBattlefield(t, g, looseGear, true, "unattached Equipment")
}

func auraGraftTargetShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Aura Graft Target",
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Blue},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target Aura that's attached to a permanent",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					SubtypesAny:              []types.Sub{types.Aura},
					MatchAttachedToPermanent: true,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: game.DurationPermanent,
			}}},
		}.Ability()),
	}}
}

func TestAuraAttachedToPermanentTargetLegality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	if issues := game.ValidateCardDef(auraGraftTargetShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	cardID := addCardToHand(g, game.Player1, auraGraftTargetShape())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	creature := addCreatureWithManaValue(g, game.Player2, 2)
	creatureAura := addAuraPermanent(g, game.Player2)
	playerAura := addEnchantPlayerAura(g, game.Player2, game.PlayerAny)
	looseGear := addEquipmentPermanent(g, game.Player2)
	attachedGear := addEquipmentPermanent(g, game.Player2)
	if !attachPermanent(g, creatureAura, creature) || !attachAuraToPlayer(g, playerAura, game.Player3) ||
		!attachPermanent(g, attachedGear, creature) {
		t.Fatal("setup attach failed")
	}

	legal := engine.legalActions(g, game.Player1)
	castAt := func(p *game.Permanent) action.Action {
		return action.CastSpell(cardID, []game.Target{game.PermanentTarget(p.ObjectID)}, 0, nil)
	}
	if !containsAction(legal, castAt(creatureAura)) {
		t.Fatal("Aura attached to a creature was not a legal target")
	}
	for name, p := range map[string]*game.Permanent{
		"Aura attached to a player":        playerAura,
		"unattached Equipment":             looseGear,
		"Equipment attached to a creature": attachedGear,
	} {
		if containsAction(legal, castAt(p)) {
			t.Fatalf("%s was offered as a legal target", name)
		}
	}

	if !engine.applyAction(g, game.Player1, castAt(creatureAura)) {
		t.Fatal("cast at an attached Aura failed")
	}
	detachPermanent(g, creatureAura)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, creatureAura); got != game.Player2 {
		t.Fatalf("controller of an Aura unattached before resolution = %v, want Player2", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Stack.Size() != 0 {
		t.Fatal("spell with its only target illegal did not leave the stack for the graveyard")
	}
}

func TestAttachmentsGroupWithUnattachedAnchorIsEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCreatureWithManaValue(g, game.Player1, 2)
	addEquipmentPermanent(g, game.Player1)
	resolver := newReferenceResolverWithControllerAndSource(g, game.Player1, creature)
	group := game.AttachmentsGroup(game.SourcePermanentReference(), game.Selection{})
	if !group.Valid() {
		t.Fatalf("Validate = %v", group.Validate())
	}
	if members := resolver.groupMembers(group); len(members) != 0 {
		t.Fatalf("members = %v, want none", members)
	}
}
