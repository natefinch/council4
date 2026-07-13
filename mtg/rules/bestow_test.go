package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// bestowTestCreatureEnchantTarget is the enchant-creature target a bestowed
// bestowTestCreature must satisfy while on the stack and while attached.
func bestowTestCreatureEnchantTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "creature",
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature},
		}),
	}
}

// bestowTestCreature is a hand-crafted enchantment creature with Bestow: a 1/1
// for {2} that may be cast bestowed for {1}. While bestowed it grants the
// enchanted creature +2/+2 and flying.
func bestowTestCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	enchantTarget := bestowTestCreatureEnchantTarget()
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Bestow Tester",
		Types:     []types.Card{types.Enchantment, types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{
			game.BestowStaticAbility(cost.Mana{cost.O(1)}, &enchantTarget),
			{
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:          game.LayerPowerToughnessModify,
						Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
						PowerDelta:     2,
						ToughnessDelta: 2,
					},
					{
						Layer:       game.LayerAbility,
						Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
						AddKeywords: []game.Keyword{game.Flying},
					},
				},
			},
		},
	}}
}

func setupBestowMain(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	return g, engine
}

// TestBestowCastAttachesAndGrantsToEnchantedCreature proves a spell cast
// bestowed becomes an Aura on the stack, resolves attached to its creature
// target, is no longer a creature, and grants the enchanted creature +2/+2 and
// flying (CR 702.103b).
func TestBestowCastAttachesAndGrantsToEnchantedCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Bestowed {
		t.Fatalf("stack object = %#v, want Bestowed spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	aura, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("bestowed permanent did not enter the battlefield")
	}
	if !aura.Bestowed {
		t.Fatal("bestowed permanent not marked Bestowed")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != target.ObjectID {
		t.Fatalf("bestowed permanent attached to = %v, want creature %v", aura.AttachedTo, target.ObjectID)
	}
	if permanentHasType(g, aura, types.Creature) {
		t.Fatal("bestowed permanent is still a creature, want Aura only")
	}
	if !permanentHasSubtype(g, aura, types.Aura) {
		t.Fatal("bestowed permanent is not an Aura")
	}
	if !isAuraPermanent(g, aura) {
		t.Fatal("bestowed permanent is not recognized as an Aura permanent")
	}
	if got := effectivePower(g, target); got != 4 {
		t.Fatalf("enchanted creature power = %d, want 4", got)
	}
	if got, _ := effectiveToughness(g, target); got != 4 {
		t.Fatalf("enchanted creature toughness = %d, want 4", got)
	}
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("enchanted creature did not gain flying")
	}
}

// TestNormalCastEntersAsCreatureNotAura proves the same card cast for its normal
// mana cost enters as a creature with no target and does not attach.
func TestNormalCastEntersAsCreatureNotAura(t *testing.T) {
	g, engine := setupBestowMain(t)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Bestowed {
		t.Fatalf("stack object = %#v, want non-bestowed creature spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("creature did not enter the battlefield")
	}
	if permanent.Bestowed {
		t.Fatal("normally cast creature marked Bestowed")
	}
	if permanent.AttachedTo.Exists {
		t.Fatal("normally cast creature is attached to something")
	}
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("normally cast permanent is not a creature")
	}
	if isAuraPermanent(g, permanent) {
		t.Fatal("normally cast permanent is an Aura")
	}
	if got := effectivePower(g, permanent); got != 1 {
		t.Fatalf("creature power = %d, want 1", got)
	}
}

// TestNormalCastCannotChooseCreatureTarget proves a normal cast requires no
// target and rejects a creature target, while a bestowed cast requires one.
func TestNormalCastCannotChooseCreatureTarget(t *testing.T) {
	g, engine := setupBestowMain(t)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil) {
		t.Fatal("normal cast accepted a creature target, want no target allowed")
	}
	if !engine.canCastSpell(g, game.Player1, spellID, nil, 0, nil) {
		t.Fatal("normal cast without a target was rejected")
	}
}

// TestBestowIllegalTargetResolvesAsCreature proves that when a bestowed Aura
// spell's target becomes illegal before resolution, it ceases to be bestowed and
// resolves as an ordinary creature instead of being countered (CR 702.103e).
func TestBestowIllegalTargetResolvesAsCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast failed")
	}
	// The enchant target leaves the battlefield before the spell resolves.
	movePermanentToZone(g, target, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("bestowed spell with an illegal target was countered, want creature on the battlefield")
	}
	if permanent.Bestowed {
		t.Fatal("permanent still marked Bestowed after resolving with an illegal target")
	}
	if permanent.AttachedTo.Exists {
		t.Fatal("permanent attached despite an illegal target")
	}
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("permanent that fell back from bestow is not a creature")
	}
	if isAuraPermanent(g, permanent) {
		t.Fatal("permanent that fell back from bestow is an Aura")
	}
}

// TestBestowedAuraBecomesCreatureWhenEnchantedCreatureLeaves proves that when
// the enchanted creature leaves, the bestowed Aura becomes unattached, ceases to
// be bestowed, and stays on the battlefield as a creature rather than being put
// into the graveyard as an illegal Aura (CR 702.103f).
func TestBestowedAuraBecomesCreatureWhenEnchantedCreatureLeaves(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	aura, ok := findPermanentByCardID(g, spellID)
	if !ok || !aura.Bestowed {
		t.Fatal("bestowed permanent did not attach")
	}

	movePermanentToZone(g, target, zone.Graveyard)
	engine.applyStateBasedActionsWithDeaths(g)

	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("bestowed permanent left the battlefield after its creature left, want it to survive as a creature")
	}
	if permanent.Bestowed {
		t.Fatal("permanent still marked Bestowed after becoming unattached")
	}
	if permanent.AttachedTo.Exists {
		t.Fatal("permanent still attached after its creature left")
	}
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("unattached bestowed permanent is not a creature")
	}
	if isAuraPermanent(g, permanent) {
		t.Fatal("unattached bestowed permanent is still an Aura")
	}
	if g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("bestowed permanent was put into the graveyard as an illegal Aura")
	}
}

// TestMultipleBestowsStackOnSameCreature proves two independent bestowed Auras
// can attach to the same creature and both grant their static bonuses, so the
// enchanted creature gets +2/+2 from each (+4/+4 total) and each bestowed
// permanent is an Aura rather than a creature.
func TestMultipleBestowsStackOnSameCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	firstID := addCardToHand(g, game.Player1, bestowTestCreature())
	secondID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(firstID, targets, 0, nil)) {
		t.Fatal("first bestowed cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(secondID, targets, 0, nil)) {
		t.Fatal("second bestowed cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	first, ok := findPermanentByCardID(g, firstID)
	if !ok || !first.Bestowed || first.AttachedTo.Val != target.ObjectID {
		t.Fatalf("first bestowed permanent = %#v, want attached Aura", first)
	}
	second, ok := findPermanentByCardID(g, secondID)
	if !ok || !second.Bestowed || second.AttachedTo.Val != target.ObjectID {
		t.Fatalf("second bestowed permanent = %#v, want attached Aura", second)
	}
	if permanentHasType(g, first, types.Creature) || permanentHasType(g, second, types.Creature) {
		t.Fatal("a bestowed permanent is still a creature, want Aura only")
	}
	if got := effectivePower(g, target); got != 6 {
		t.Fatalf("enchanted creature power = %d, want 6 (2 base + 2 + 2)", got)
	}
	if got, _ := effectiveToughness(g, target); got != 6 {
		t.Fatalf("enchanted creature toughness = %d, want 6", got)
	}
}

// TestUnattachedBestowCreatureIsSummoningSick proves a bestowed permanent that
// becomes a creature this turn (because its enchanted creature left) is
// summoning sick and cannot attack, exactly like any creature that entered the
// battlefield this turn.
func TestUnattachedBestowCreatureIsSummoningSick(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	movePermanentToZone(g, target, zone.Graveyard)
	engine.applyStateBasedActionsWithDeaths(g)

	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("bestowed permanent left the battlefield, want it to survive as a creature")
	}
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("unattached bestowed permanent is not a creature")
	}
	if !permanent.SummoningSick {
		t.Fatal("bestowed permanent that became a creature this turn is not summoning sick")
	}
	if canAttackWith(g, permanent, game.Player1) {
		t.Fatal("summoning-sick bestowed creature was allowed to attack")
	}
}
