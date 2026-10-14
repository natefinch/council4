package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Oracle rulings (2006-02-01, 2009-10-01), official CR effective
// 2026-09-25: 613.1, 613.4, 613.6–613.8, 707.2 and 205.1b.
func challengeEnchantment(name string, mv int, subtypes ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, ManaCost: opt.Val(cost.Mana{cost.O(mv)}),
		Types: []types.Card{types.Enchantment}, Subtypes: subtypes,
		StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody},
	}}
}

func assertChallengeValues(t *testing.T, g *game.Game, p *game.Permanent, creature bool, pt int, abilities bool) {
	t.Helper()
	values := effectivePermanentValues(g, p)
	if slices.Contains(values.types, types.Creature) != creature ||
		creature && (!values.powerOK || !values.toughnessOK || values.power != pt || values.toughness != pt) ||
		(len(values.abilities) > 0) != abilities {
		t.Fatalf("%s: creature=%v P/T=%d/%d abilities=%d, want %v %d/%d abilities=%v",
			values.name, slices.Contains(values.types, types.Creature), values.power, values.toughness, len(values.abilities),
			creature, pt, pt, abilities)
	}
}

func TestHumilityOpalescenceAllTimestampOrders(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name                             string
		order                            []string
		humility, first, second, worship int
	}{
		{name: "Opalescence then Humility", order: []string{"O", "H"}, humility: 1, first: 0, worship: 1},
		{name: "Humility then Opalescence", order: []string{"H", "O"}, humility: 4, first: 0, worship: 4},
		{name: "Humility latest", order: []string{"O", "O", "H"}, humility: 1, first: 1, second: 1, worship: 1},
		{name: "Humility between", order: []string{"O", "H", "O"}, humility: 4, first: 4, second: 1, worship: 4},
		{name: "Humility earliest", order: []string{"H", "O", "O"}, humility: 4, first: 4, second: 4, worship: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e := setupBestowMain(t)
			worship := addCombatPermanent(g, game.Player2, challengeEnchantment("Worship shape", 4))
			var humility *game.Permanent
			var opals []*game.Permanent
			for _, card := range test.order {
				if card == "H" {
					humility = challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
				} else {
					opals = append(opals, challengeCastEnchantment(t, g, e, game.Player1, authored.Opalescence.New()))
				}
			}
			assertChallengeValues(t, g, humility, true, test.humility, false)
			assertChallengeValues(t, g, worship, true, test.worship, false)
			if len(opals) == 1 {
				assertChallengeValues(t, g, opals[0], false, 0, true)
			} else {
				assertChallengeValues(t, g, opals[0], true, test.first, false)
				assertChallengeValues(t, g, opals[1], true, test.second, false)
			}
			clone := g.Clone()
			for _, p := range g.Battlefield {
				cp, _ := permanentByObjectID(clone, p.ObjectID)
				before := effectivePermanentValues(g, p)
				after := effectivePermanentValues(clone, cp)
				if before.power != after.power || before.toughness != after.toughness ||
					!slices.Equal(before.types, after.types) || len(before.abilities) != len(after.abilities) {
					t.Fatal("clone changed static source identity, continuation or ordering")
				}
			}
		})
	}
}

func TestHumilityOpalescenceAurasModifiersCountersAndSourceDeparture(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Winged mana creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
		ManaAbilities:   []game.ManaAbility{game.TapManaAbility(mana.G)},
	}})
	creature.Counters.Add(counter.PlusOnePlusOne, 2)
	creature.TemporaryPowerModifier, creature.TemporaryToughnessModifier = 3, 3
	other := addCombatPermanent(g, game.Player2, challengeEnchantment("Other enchantment", 3))
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	opal := challengeCastEnchantment(t, g, e, game.Player1, authored.Opalescence.New())
	aura := addCombatPermanent(g, game.Player2, challengeEnchantment("Aura", 2, types.Aura))
	aura.AttachedTo = opt.Val(creature.ObjectID)
	assertChallengeValues(t, g, aura, false, 0, true)
	assertChallengeValues(t, g, creature, true, 6, false)
	assertChallengeValues(t, g, other, true, 3, false)
	assertChallengeValues(t, g, humility, true, 4, false)
	if e.applyAction(g, game.Player1, action.ActivateAbility(creature.ObjectID, 0, nil, 0)) {
		t.Fatal("Humility allowed the removed printed mana ability to activate")
	}
	clone := g.Clone()
	movePermanentToZone(g, humility, zone.Graveyard)
	assertChallengeValues(t, g, creature, true, 10, true)
	assertChallengeValues(t, g, other, true, 3, true)
	assertChallengeValues(t, g, opal, false, 0, true)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(creature.ObjectID, 0, nil, 0)) {
		t.Fatal("Humility departure did not restore the printed mana ability")
	}
	clonedCreature, _ := permanentByObjectID(clone, creature.ObjectID)
	assertChallengeValues(t, clone, clonedCreature, true, 6, false)
	movePermanentToZone(g, opal, zone.Graveyard)
	assertChallengeValues(t, g, other, false, 0, true)
	if len(g.ContinuousEffects) != 0 {
		t.Fatal("static evaluations leaked runtime effects into the game")
	}
}

func TestHumilityAbilityLayerDependenciesAndLaterExternalGrant(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	// The generator is a creature. Losing its ability changes the existence
	// of the grant, so its effect depends on Humility, regardless of timestamp.
	grant := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Creature ability anthem", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				AddKeywords: []game.Keyword{game.Flying},
			}},
		}},
	}})
	assertChallengeValues(t, g, target, true, 1, false)
	assertChallengeValues(t, g, grant, true, 1, false)
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("a removed static ability still generated a dependent grant")
	}
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Later ability Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility, Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddKeywords: []game.Keyword{game.Flying},
			}},
		}},
	}})
	aura.AttachedTo = opt.Val(target.ObjectID)
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("a later independent Aura could not grant an ability after removal")
	}
	movePermanentToZone(g, aura, zone.Graveyard)
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("a departed Aura left its grant behind")
	}
	movePermanentToZone(g, humility, zone.Graveyard)
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("source departure did not restore the creature generator")
	}
}

func TestOpalescenceCopiesUseCopiableManaValueAndPhysicalSourceExclusion(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	opal := challengeCastEnchantment(t, g, e, game.Player1, authored.Opalescence.New())
	blueprint := addCombatPermanent(g, game.Player1, challengeEnchantment("Five mana blueprint", 5))
	copied := addCombatPermanent(g, game.Player2, challengeEnchantment("Physical zero mana", 0))
	becomeStaticCopy(t, g, copied, blueprint, false)
	assertChallengeValues(t, g, copied, true, 5, true)
	becomeStaticCopy(t, g, copied, opal, false)
	assertChallengeValues(t, g, opal, true, 4, true)
	assertChallengeValues(t, g, copied, true, 4, true)
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	assertChallengeValues(t, g, opal, true, 1, false)
	assertChallengeValues(t, g, copied, true, 1, false)
	assertChallengeValues(t, g, humility, true, 1, false)
	movePermanentToZone(g, opal, zone.Graveyard)
	assertChallengeValues(t, g, copied, false, 0, true)
	assertChallengeValues(t, g, humility, true, 1, false)
}

func TestHumilitySuppressesLaterPowerToughnessAnthemAndAnimatedLandMana(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	land := addBasicLandWithManaPermanent(g, game.Player1, "Plains", types.Plains, mana.W)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: land.ObjectID, AddTypes: []types.Card{types.Creature}},
		game.ContinuousEffect{ID: g.IDGen.Next(), Layer: game.LayerPowerToughnessSet, AffectedObjectID: land.ObjectID,
			SetPower: opt.Val(game.PT{Value: 3}), SetToughness: opt.Val(game.PT{Value: 3})})
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Creature power anthem", Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:      game.LayerPowerToughnessModify,
				Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				PowerDelta: 2, ToughnessDelta: 2,
			}},
		}},
	}})
	assertChallengeValues(t, g, target, true, 1, false)
	assertChallengeValues(t, g, land, true, 1, false)
	movePermanentToZone(g, humility, zone.Graveyard)
	if effectivePower(g, target) != 5 || len(permanentEffectiveAbilitiesView(g, land)) == 0 {
		t.Fatal("Humility departure did not restore an anthem and intrinsic land mana")
	}
}

func TestHumilityKeywordCounterTimestampRefreshRemovalAndClone(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	addCountersToPermanent(g, creature, counter.Flying, 1)
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	if hasKeyword(g, creature, game.Flying) {
		t.Fatal("an earlier keyword counter applied after Humility")
	}
	addCountersToPermanent(g, creature, counter.Flying, 1)
	if !hasKeyword(g, creature, game.Flying) || effectivePower(g, creature) != 1 {
		t.Fatal("a new counter failed to refresh all counters of its kind")
	}
	clone := g.Clone()
	clonedCreature, _ := permanentByObjectID(clone, creature.ObjectID)
	creature.Counters.Remove(counter.Flying, 1)
	if !hasKeyword(g, creature, game.Flying) {
		t.Fatal("removing one counter refreshed or erased the remaining counter's timestamp")
	}
	creature.CounterTimestamps[counter.Flying] = creature.Timestamp()
	if hasKeyword(g, creature, game.Flying) || !hasKeyword(clone, clonedCreature, game.Flying) {
		t.Fatal("counter timestamps were shared with the clone")
	}
	movePermanentToZone(g, humility, zone.Graveyard)
	if !hasKeyword(g, creature, game.Flying) {
		t.Fatal("removing Humility did not restore an earlier keyword counter")
	}
}

func TestHumilityAnimatedChainsAndAbundanceLoseReplacementAbilities(t *testing.T) {
	t.Parallel()
	g, e := chainsGame(t, 1)
	challengeCastEnchantment(t, g, e, game.Player1, authored.Abundance.New())
	challengeCastEnchantment(t, g, e, game.Player1, authored.Opalescence.New())
	humility := challengeCastEnchantment(t, g, e, game.Player1, authored.Humility.New())
	top := abundanceLibrary(g, "LandA", "LandB", "Keep")
	a := &chainsChoices{}
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
	if len(a.requests) != 0 || !g.Players[game.Player1].Hand.Contains(top[0]) ||
		len(abundanceEvents(g, game.EventCardDiscarded)) != 0 {
		t.Fatal("animated enchantments kept their removed replacement abilities")
	}
	movePermanentToZone(g, humility, zone.Graveyard)
	a.order = []string{"that player discards"}
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDiscarded)) != 1 ||
		!g.Players[game.Player1].Hand.Contains(top[1]) || len(abundanceEvents(g, game.EventCardDrawn)) != 1 {
		t.Fatal("removing Humility failed to restore Chains followed by Abundance")
	}
}

func TestOpalescenceTypeDependenciesAndCrossLayerRecipientSet(t *testing.T) {
	t.Parallel()
	g, e := setupBestowMain(t)
	opal := challengeCastEnchantment(t, g, e, game.Player1, authored.Opalescence.New())
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Five mana artifact", Types: []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
	}})
	// This newer effect changes what the older animation applies to. CR
	// 613.8a requires Opalescence to wait for it, overriding timestamps.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()),
		Layer: game.LayerType, AffectedObjectID: artifact.ObjectID, AddTypes: []types.Card{types.Enchantment},
	})
	assertChallengeValues(t, g, artifact, true, 5, false)
	if !permanentHasType(g, artifact, types.Artifact) {
		t.Fatal("animation replaced rather than added types")
	}
	// Removing the Enchantment type after animation creates a dependency
	// loop, applied in timestamp order. The begun effect still sets base P/T.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()),
		Layer: game.LayerType, Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		RemoveTypes: []types.Card{types.Enchantment},
	})
	assertChallengeValues(t, g, artifact, true, 5, false)
	if permanentHasType(g, artifact, types.Enchantment) {
		t.Fatal("the later type removal did not apply")
	}
	opal.PhasedOut = true
	assertChallengeValues(t, g, artifact, false, 0, false)
	opal.PhasedOut = false
	assertChallengeValues(t, g, artifact, true, 5, false)
}
