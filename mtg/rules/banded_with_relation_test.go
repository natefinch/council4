package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, retained under .cardwork), read 702.22a-m,
// 611.2c, 608.2h and 615.1. 702.22c-d: bands are announced with the attack
// declaration; 702.22e: a band lasts for the rest of combat even if members
// lose banding; 702.22f: a creature removed from combat leaves its band. 611.2c:
// a continuous effect from a resolving ability that changes characteristics
// affects only the objects determined as it resolves. 615.1: a static
// prevention applies while its condition holds. Card faces (pinned corpus,
// no rulings): Icatian Skirmishers "First strike; banding ... Whenever this
// creature attacks, all creatures banded with it gain first strike until end
// of turn."; Urza's Engine "Trample / {3}: This creature gains banding until
// end of turn. / {3}: Attacking creatures banded with this creature gain
// trample until end of turn."; Camel "Banding / As long as this creature is
// attacking, prevent all damage Deserts would deal to this creature and to
// creatures banded with this creature." The scenarios use test-local shapes of
// those abilities; the cards themselves are admitted by a later card unit.

var bandedWithCreatures = game.Selection{RequiredTypes: []types.Card{types.Creature}, BandedWithSource: true}

func bandedWithSkirmisherDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Banded-With Skirmisher",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.FirstStrike, game.Banding),
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever this creature attacks, all creatures banded with it gain first strike until end of turn.",
			Trigger: game.TriggerCondition{
				Type:    game.TriggerWhenever,
				Pattern: game.TriggerPattern{Event: game.EventAttackerDeclared, Source: game.TriggerSourceSelf},
			},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:    game.DurationPermanent,
					Group:       game.BattlefieldGroup(bandedWithCreatures),
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.FirstStrike},
				}},
				Duration: game.DurationUntilEndOfTurn,
			}}}}.Ability(),
		}},
	}}
}

func bandedWithEngineDef() *game.CardDef {
	sel := bandedWithCreatures
	sel.CombatState = game.CombatStateAttacking
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Banded-With Engine",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Trample, game.Banding),
		}},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: greenCost(),
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:    game.DurationPermanent,
					Group:       game.BattlefieldGroup(sel),
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.Trample},
				}},
				Duration: game.DurationUntilEndOfTurn,
			}}}}.Ability(),
		}},
	}}
}

func bandedWithCamelDef() *game.CardDef {
	deserts := game.Selection{SubtypesAny: []types.Sub{types.Desert}}
	attacking := opt.Val(game.Condition{
		Object:        opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{CombatState: game.CombatStateAttacking}),
	})
	prevention := func(text string, replacement game.ReplacementEffect) game.ReplacementAbility {
		replacement.DamagePreventAll = true
		replacement.DamageSourceSelection = &deserts
		replacement.Condition = attacking
		return game.ReplacementAbility{Text: text, Replacement: replacement}
	}
	recipients := bandedWithCreatures
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Banded-With Camel",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Banding),
		}},
		ReplacementAbilities: []game.ReplacementAbility{
			prevention("prevent all damage Deserts would deal to this creature", game.ReplacementEffect{DamageRecipientSelf: true}),
			prevention("and to creatures banded with this creature", game.ReplacementEffect{DamageRecipientSelection: &recipients}),
		},
	}}
}

func addSubtypedLandPermanent(g *game.Game, controller game.PlayerID, subtypes ...types.Sub) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name: "Banded-With Land", Types: []types.Card{types.Land}, Subtypes: subtypes,
	}})
}

func hasFirstStrike(g *game.Game, permanent *game.Permanent) bool {
	return hasKeyword(g, permanent, game.FirstStrike)
}

func bandedWithMembers(g *game.Game, source *game.Permanent, sel game.Selection) []id.ID {
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}
	return newReferenceResolver(g, obj).groupMembers(game.BattlefieldGroup(sel))
}

// TestBandedWithAttackTriggerGrantsFirstStrikeToBand plays the Icatian
// Skirmishers shape through a real combat. The skirmisher and a vanilla 2/2
// attack as a band beside a separate 2/2. The attack trigger gives first
// strike only to the banded 2/2. A 2/3 blocking the banded 2/2 blocks the
// skirmisher too (702.22h) and dies to their 3 first-strike damage before it
// deals any, while the separate 2/2 trades with its blocker in the regular step.
func TestBandedWithAttackTriggerGrantsFirstStrikeToBand(t *testing.T) {
	g := bandCombatGame()
	skirmisher := addCombatPermanent(g, game.Player1, bandedWithSkirmisherDef())
	member := bandingTestCreature(g, game.Player1, 2, 2)
	loner := bandingTestCreature(g, game.Player1, 2, 2)
	bandBlocker := bandingTestCreature(g, game.Player2, 2, 3)
	lonerBlocker := bandingTestCreature(g, game.Player2, 2, 2)
	attack := bandAttack(skirmisher, member)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))

	policies := simPassPolicies()
	policies[game.Player1] = &bandingDivisionAgent{}
	policies[game.Player2] = &bandBlockingAgent{blocks: []game.BlockDeclaration{
		{Blocker: bandBlocker.ObjectID, Blocking: member.ObjectID},
		{Blocker: lonerBlocker.ObjectID, Blocking: loner.ObjectID},
	}}
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attack, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	afterMember, _ := permanentByObjectID(after, member.ObjectID)
	if !hasFirstStrike(after, afterMember) {
		t.Fatal("the banded 2/2 did not gain first strike")
	}
	if afterLoner, ok := permanentByObjectID(after, loner.ObjectID); ok && hasFirstStrike(after, afterLoner) {
		t.Fatal("the non-banded attacker gained first strike")
	}
	if permanentWithCardOnBattlefield(after, bandBlocker.CardInstanceID) {
		t.Fatal("the band's blocker survived the band's 3 first-strike damage")
	}
	if markedDamage(t, after, skirmisher) != 0 || markedDamage(t, after, member) != 0 {
		t.Fatal("the band's blocker dealt damage after dying in the first-strike step")
	}
	if permanentWithCardOnBattlefield(after, loner.CardInstanceID) || permanentWithCardOnBattlefield(after, lonerBlocker.CardInstanceID) {
		t.Fatal("the non-banded attacker and its blocker should trade in the regular damage step")
	}
}

// TestBandedWithActivatedGrantFixedAtResolution plays the Urza's Engine shape:
// the engine, a banding 2/2 and a 2/2 attack as one band beside a separate 2/2
// attacker. The 2/2 removed from combat before the ability resolves has left the
// band (702.22f) and gets nothing; the remaining member gains trample; the
// separate attacker does not. The affected set is fixed as the ability resolves
// (611.2c), so removing the member from combat afterwards doesn't end its
// trample.
func TestBandedWithActivatedGrantFixedAtResolution(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, bandedWithEngineDef())
	member := bandingTestCreature(g, game.Player1, 2, 2, game.Banding)
	removed := bandingTestCreature(g, game.Player1, 2, 2)
	loner := bandingTestCreature(g, game.Player1, 2, 2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	attack := bandAttack(source, member, removed)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))
	declared, ok := engine.Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	declared.Turn.PriorityPlayer = game.Player1

	activate := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !engine.applyAction(declared, game.Player1, activate) {
		t.Fatal("activating the band trample ability failed")
	}
	removePermanentFromCombat(declared, removed.ObjectID)
	engine.resolveTopOfStack(declared, nil)

	trample := func(p *game.Permanent) bool {
		current, ok := permanentByObjectID(declared, p.ObjectID)
		return ok && hasKeyword(declared, current, game.Trample)
	}
	if !trample(member) {
		t.Fatal("the attacking band member did not gain trample")
	}
	if trample(removed) || trample(loner) {
		t.Fatalf("trample reached a creature outside the band: removed=%v loner=%v", trample(removed), trample(loner))
	}
	removePermanentFromCombat(declared, member.ObjectID)
	if got := bandedWithMembers(declared, source, bandedWithCreatures); len(got) != 0 {
		t.Fatalf("band after both members left combat = %v, want none", got)
	}
	if !trample(member) {
		t.Fatal("trample granted at resolution ended when the member left combat")
	}
}

// TestBandedWithDesertPreventionWhileAttacking plays the Camel shape: while the
// camel attacks, a Desert's damage to it and to its band members is prevented
// (615.1), but not a Desert's damage to a separate attacker or another land's
// damage to a member. A Desert that has left the battlefield is still a Desert
// source through last-known information (608.2h). A member removed from combat
// loses the shield; once the camel leaves combat the shield ends for the camel
// and its former band.
func TestBandedWithDesertPreventionWhileAttacking(t *testing.T) {
	g := bandCombatGame()
	camel := addCombatPermanent(g, game.Player1, bandedWithCamelDef())
	registerPermanentReplacementEffects(g, camel)
	first := bandingTestCreature(g, game.Player1, 2, 9, game.Banding)
	second := bandingTestCreature(g, game.Player1, 2, 9, game.Banding)
	third := bandingTestCreature(g, game.Player1, 2, 9)
	loner := bandingTestCreature(g, game.Player1, 2, 9)
	desert := addSubtypedLandPermanent(g, game.Player2, types.Desert)
	mountain := addSubtypedLandPermanent(g, game.Player2, types.Mountain)
	departedDesert := addSubtypedLandPermanent(g, game.Player2, types.Desert)

	damage := func(g *game.Game, from *game.Permanent, to *game.Permanent) int {
		t.Helper()
		target, ok := permanentByObjectID(g, to.ObjectID)
		if !ok {
			t.Fatalf("%v left the battlefield", to.ObjectID)
		}
		return dealPermanentDamage(g, from.CardInstanceID, from.ObjectID, game.Player2, target, 1, false)
	}
	if damage(g, desert, camel) != 1 || damage(g, desert, first) != 1 {
		t.Fatal("Desert damage was prevented before the camel attacked")
	}

	attack := bandAttack(camel, first, second, third)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))
	declared, ok := NewEngine(nil).Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	for _, protected := range []*game.Permanent{camel, first, second, third} {
		if got := damage(declared, desert, protected); got != 0 {
			t.Fatalf("Desert damage to %v while the camel attacks = %d, want 0", protected.ObjectID, got)
		}
	}
	if damage(declared, desert, loner) != 1 {
		t.Fatal("Desert damage to an attacker outside the band was prevented")
	}
	if damage(declared, mountain, first) != 1 {
		t.Fatal("non-Desert damage to a band member was prevented")
	}
	departed, _ := permanentByObjectID(declared, departedDesert.ObjectID)
	if !sacrificePermanent(declared, departed) {
		t.Fatal("sacrificing the second Desert failed")
	}
	if damage(declared, departedDesert, second) != 0 {
		t.Fatal("damage from a Desert that left the battlefield (last-known information) was not prevented")
	}

	removePermanentFromCombat(declared, third.ObjectID)
	if damage(declared, desert, third) != 1 {
		t.Fatal("Desert damage to a creature removed from the band was prevented")
	}
	if damage(declared, desert, first) != 0 {
		t.Fatal("Desert damage to a remaining band member was not prevented")
	}

	removePermanentFromCombat(declared, camel.ObjectID)
	if !slices.ContainsFunc(declared.Combat.Bands, func(band game.AttackingBand) bool {
		return slices.Equal(band.Members, []id.ID{first.ObjectID, second.ObjectID})
	}) {
		t.Fatalf("bands after the camel left combat = %+v, want the remaining pair", declared.Combat.Bands)
	}
	if damage(declared, desert, camel) != 1 || damage(declared, desert, first) != 1 || damage(declared, desert, second) != 1 {
		t.Fatal("Desert damage was still prevented after the camel left combat")
	}
}

// TestBandedWithStaticGroupIsLive checks a static "creatures banded with this
// creature get +1/+0" (a test-local static shape): it reaches only current band
// members and is re-read continuously, so a member removed from combat loses it
// at once (702.22f).
func TestBandedWithStaticGroupIsLive(t *testing.T) {
	g := bandCombatGame()
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Banded-With Leader",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			{KeywordAbilities: game.SimpleKeywords(game.Banding)},
			{
				Text: "Creatures banded with this creature get +1/+0.",
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:   game.DurationPermanent,
					Group:      game.BattlefieldGroup(bandedWithCreatures),
					Layer:      game.LayerPowerToughnessModify,
					PowerDelta: 1,
				}},
			},
		},
	}})
	member := bandingTestCreature(g, game.Player1, 2, 2)
	loner := bandingTestCreature(g, game.Player1, 2, 2)
	if effectivePower(g, source) != 1 || effectivePower(g, member) != 2 {
		t.Fatal("the static applied outside combat")
	}
	attack := bandAttack(source, member)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))
	declared, ok := NewEngine(nil).Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	power := func(p *game.Permanent) int {
		current, _ := permanentByObjectID(declared, p.ObjectID)
		return effectivePower(declared, current)
	}
	if power(source) != 1 || power(member) != 3 || power(loner) != 2 {
		t.Fatalf("powers source=%d member=%d loner=%d, want 1/3/2", power(source), power(member), power(loner))
	}
	removePermanentFromCombat(declared, member.ObjectID)
	if power(member) != 2 {
		t.Fatalf("member power after leaving combat = %d, want 2", power(member))
	}
}

// TestBandedWithMatchesOnlyCurrentBandMembers checks the relation directly: it
// matches nothing outside combat or for an attacker with no band, matches the
// other members of the source's band (never the source), survives every member
// losing banding (702.22e) and a game clone, and ends when combat ends.
func TestBandedWithMatchesOnlyCurrentBandMembers(t *testing.T) {
	g := bandCombatGame()
	source := bandingTestCreature(g, game.Player1, 1, 1, game.Banding)
	member := bandingTestCreature(g, game.Player1, 2, 2, game.Banding)
	loner := bandingTestCreature(g, game.Player1, 2, 2, game.Banding)
	bystander := bandingTestCreature(g, game.Player1, 2, 2)

	outside := g.Clone()
	outside.Combat = nil
	if got := bandedWithMembers(outside, source, bandedWithCreatures); len(got) != 0 {
		t.Fatalf("members outside combat = %v, want none", got)
	}

	attack := bandAttack(source, member)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))
	declared, ok := NewEngine(nil).Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	if got := bandedWithMembers(declared, loner, bandedWithCreatures); len(got) != 0 {
		t.Fatalf("members for an attacker with no band = %v, want none", got)
	}
	if got := bandedWithMembers(declared, bystander, bandedWithCreatures); len(got) != 0 {
		t.Fatalf("members for a non-attacking source = %v, want none", got)
	}
	want := []id.ID{member.ObjectID}
	if got := bandedWithMembers(declared, source, bandedWithCreatures); !slices.Equal(got, want) {
		t.Fatalf("members = %v, want %v", got, want)
	}

	for _, p := range []*game.Permanent{source, member} {
		declared.ContinuousEffects = append(declared.ContinuousEffects, game.ContinuousEffect{
			ID: declared.IDGen.Next(), Layer: game.LayerAbility, Duration: game.DurationUntilEndOfTurn,
			AffectedObjectID: p.ObjectID, RemoveKeywords: []game.Keyword{game.Banding},
			Timestamp: game.Timestamp(declared.IDGen.Next()),
		})
	}
	if current, _ := permanentByObjectID(declared, member.ObjectID); hasKeyword(declared, current, game.Banding) {
		t.Fatal("band member still has banding")
	}
	if got := bandedWithMembers(declared, source, bandedWithCreatures); !slices.Equal(got, want) {
		t.Fatalf("members after losing banding = %v, want %v", got, want)
	}
	if got := bandedWithMembers(declared.Clone(), source, bandedWithCreatures); !slices.Equal(got, want) {
		t.Fatalf("members in a clone = %v, want %v", got, want)
	}

	declared.Combat = nil
	if got := bandedWithMembers(declared, source, bandedWithCreatures); len(got) != 0 {
		t.Fatalf("members after combat = %v, want none", got)
	}
}
