package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3954 (sub-issue of #2098). Official CR effective
// 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// read 301.5c, 604.2, 611.3a, 613.1f, 613.6, 613.7a and 613.8a-c. 604.2: a
// static ability creates continuous effects that are active while the
// permanent with the ability is on the battlefield. 613.7a: such an effect has
// the later of its object's timestamp and the timestamp of the effect that
// granted the ability. 613.8a: an effect that removes the ability generating
// another effect is applied first, so the granted ability's effect then does
// not exist. 613.6: an effect that began in a layer continues to the same
// objects in later layers. Cards (pinned corpus):
// Tenth District Hero (oracle 0953772c-9374-4616-9c04-5838aaf626c9, {1}{W}
// Creature — Human 2/3) "{1}{W}, Collect evidence 2: This creature becomes a
// Human Detective with base power and toughness 4/4 and gains vigilance. /
// {2}{W}, Collect evidence 4: If this creature is a Detective, it becomes a
// legendary creature named Mileva, the Stalwart, it has base power and
// toughness 5/5, and it gains "Other creatures you control have
// indestructible."" (rulings 2024-02-02: the effects have no duration and
// last until the game ends or it leaves the battlefield; the second ability
// checks whether it is a Detective on resolution; setting base P/T overrides
// only earlier setting effects). The fixture below follows that Oracle text
// and is not registered. Registered The Irencrag (oracle
// 8c89ba89-9bb3-4609-873c-6f5a961733ce) "{T}: Add {C}. / Whenever a legendary
// creature you control enters, you may have The Irencrag become a legendary
// Equipment artifact named Everflame, Heroes' Legacy. If you do, it gains
// equip {3} and "Equipped creature gets +3/+3" and loses all other abilities."
// (ruling 2023-09-01: the effect lasts indefinitely). Registered Wrath of God
// ("Destroy all creatures. They can't be regenerated."), Murder ("Destroy
// target creature.") and the #3953 ability-loss fixture supply the other
// effects.

func tenthDistrictHeroFixture() *game.CardDef {
	detective := types.Sub("Detective")
	evidence := func(amount int) []cost.Additional {
		return []cost.Additional{{
			Kind: cost.AdditionalCollectEvidence, Amount: amount, Source: zone.Graveyard,
			Text: "Collect evidence " + string(rune('0'+amount)),
		}}
	}
	becomeDetective := animateSelfAbility(
		"{1}{W}, Collect evidence 2: This creature becomes a Human Detective with base power and toughness 4/4 and gains vigilance.",
		cost.Mana{cost.O(1), cost.W},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerType,
			SetSubtypes: []types.Sub{types.Human, detective},
		},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
			SetPower: opt.Val(game.PT{Value: 4}), SetToughness: opt.Val(game.PT{Value: 4}),
		},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerAbility,
			AddKeywords: []game.Keyword{game.Vigilance},
		},
	)
	becomeDetective.AdditionalCosts = evidence(2)
	becomeMileva := animateSelfAbility(
		`{2}{W}, Collect evidence 4: If this creature is a Detective, it becomes a legendary creature named Mileva, the Stalwart, it has base power and toughness 5/5, and it gains "Other creatures you control have indestructible."`,
		cost.Mana{cost.O(2), cost.W},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerText, SetName: "Mileva, the Stalwart",
		},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerType,
			AddSupertypes: []types.Super{types.Legendary},
		},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
			SetPower: opt.Val(game.PT{Value: 5}), SetToughness: opt.Val(game.PT{Value: 5}),
		},
		game.ContinuousEffect{
			Duration: game.DurationPermanent, Layer: game.LayerAbility,
			AddAbilities: []game.Ability{&game.StaticAbility{
				Text: "Other creatures you control have indestructible.",
				ContinuousEffects: []game.ContinuousEffect{{
					Duration: game.DurationPermanent, Layer: game.LayerAbility,
					Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(),
						game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
					AddKeywords: []game.Keyword{game.Indestructible},
				}},
			}},
		},
	)
	becomeMileva.AdditionalCosts = evidence(4)
	becomeMileva.Content.Modes[0].Sequence[0].Condition = opt.Val(game.EffectCondition{
		Text: "If this creature is a Detective",
		Condition: opt.Val(game.Condition{
			Object:        opt.Val(game.SourcePermanentReference()),
			ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{detective}}),
		}),
	})
	// The abilities have no duration (ruling 2024-02-02).
	for _, ability := range []*game.ActivatedAbility{&becomeDetective, &becomeMileva} {
		apply := ability.Content.Modes[0].Sequence[0].Primitive.(game.ApplyContinuous)
		apply.Duration = game.DurationPermanent
		ability.Content.Modes[0].Sequence[0].Primitive = apply
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:               "Tenth District Hero",
		ManaCost:           opt.Val(cost.Mana{cost.O(1), cost.W}),
		Types:              []types.Card{types.Creature},
		Subtypes:           []types.Sub{types.Human},
		Power:              opt.Val(game.PT{Value: 2}),
		Toughness:          opt.Val(game.PT{Value: 3}),
		ActivatedAbilities: []game.ActivatedAbility{becomeDetective, becomeMileva},
	}}
}

func groupStaticCreature(name string, power, toughness int, supertypes ...types.Super) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		ManaCost:   opt.Val(cost.Mana{cost.G}),
		Supertypes: supertypes,
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: power}),
		Toughness:  opt.Val(game.PT{Value: toughness}),
	}}
}

func activateHeroAbility(t *testing.T, g *game.Game, engine *Engine, hero *game.Permanent, index, generic, evidence int) {
	t.Helper()
	addCardToGraveyard(g, game.Player1, evidenceCard("Evidence", evidence))
	g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player1
	activate := action.ActivateAbility(hero.ObjectID, index, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), activate) {
		t.Fatalf("Tenth District Hero ability %d was not offered", index)
	}
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatalf("activating Tenth District Hero ability %d failed", index)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].ManaPool.IsEmpty() || g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatal("fixture: the activation did not spend its mana and collect its evidence")
	}
}

// milevaBoard puts Tenth District Hero, another creature you control and an
// opponent's creature onto the battlefield and makes the Hero Mileva.
func milevaBoard(t *testing.T) (g *game.Game, engine *Engine, hero, ally, rival *game.Permanent) {
	t.Helper()
	g, engine = grantedStaticGame(game.Player1)
	hero = addCombatPermanent(g, game.Player1, tenthDistrictHeroFixture())
	ally = addCombatPermanent(g, game.Player1, groupStaticCreature("Ally Bear", 2, 2))
	rival = addCombatPermanent(g, game.Player2, groupStaticCreature("Rival Bear", 2, 2))

	// The second ability does nothing while the Hero isn't a Detective.
	activateHeroAbility(t, g, engine, hero, 1, 2, 4)
	if permanentEffectiveName(g, hero) != "Tenth District Hero" || hasKeyword(g, ally, game.Indestructible) {
		t.Fatal("the Mileva ability applied although the Hero isn't a Detective")
	}
	activateHeroAbility(t, g, engine, hero, 0, 1, 2)
	assertPowerToughness(t, g, hero, 4, 4)
	activateHeroAbility(t, g, engine, hero, 1, 2, 4)
	if permanentEffectiveName(g, hero) != "Mileva, the Stalwart" || !permanentHasSupertype(g, hero, types.Legendary) {
		t.Fatalf("the Hero is %q, want legendary Mileva, the Stalwart", permanentEffectiveName(g, hero))
	}
	assertPowerToughness(t, g, hero, 5, 5)
	if !hasKeyword(g, hero, game.Vigilance) {
		t.Fatal("Mileva lost the vigilance the first ability gave it")
	}
	return g, engine, hero, ally, rival
}

func assertIndestructible(t *testing.T, g *game.Game, want bool, permanents ...*game.Permanent) {
	t.Helper()
	for _, permanent := range permanents {
		if got := hasKeyword(g, permanent, game.Indestructible); got != want {
			t.Fatalf("%s indestructible = %v, want %v", permanentName(g, permanent), got, want)
		}
	}
}

func castWrathOfGod(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	spell := addCardToHand(g, game.Player1, registryCardDef(t, "Wrath of God"))
	g.Players[game.Player1].ManaPool.Add(mana.W, 4)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("casting Wrath of God failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

// TestTenthDistrictHeroMilevaGivesOtherCreaturesIndestructible is the #3954
// scenario: Mileva's granted static ability makes your other creatures
// survive Wrath of God while Mileva and your opponent's creatures die.
func TestTenthDistrictHeroMilevaGivesOtherCreaturesIndestructible(t *testing.T) {
	g, engine, hero, ally, rival := milevaBoard(t)
	assertIndestructible(t, g, true, ally)
	assertIndestructible(t, g, false, hero, rival)

	castWrathOfGod(t, g, engine)
	if _, ok := findPermanentByCardID(g, ally.CardInstanceID); !ok {
		t.Fatal("Wrath of God destroyed the creature Mileva gave indestructible")
	}
	for _, dead := range []*game.Permanent{hero, rival} {
		if _, ok := findPermanentByCardID(g, dead.CardInstanceID); ok {
			t.Fatalf("Wrath of God did not destroy %s", permanentName(g, dead))
		}
	}
	if _, ok := findPermanentByCardID(g, ally.CardInstanceID); ok && hasKeyword(g, ally, game.Indestructible) {
		t.Fatal("the surviving creature kept indestructible after Mileva left the battlefield")
	}
}

// TestMilevaGrantedStaticEndsWithItsSource covers the granted ability's
// effect ending when its permanent leaves the battlefield.
func TestMilevaGrantedStaticEndsWithItsSource(t *testing.T) {
	g, engine, hero, ally, _ := milevaBoard(t)
	castAndResolveRegistered(t, g, engine, game.Player2, "Murder", mana.B, 3, hero)
	if _, ok := findPermanentByCardID(g, hero.CardInstanceID); ok {
		t.Fatal("Murder did not destroy Mileva")
	}
	assertIndestructible(t, g, false, ally)
}

// TestMilevaGrantedStaticRemovedByLaterAbilityLoss covers CR 613.8a: a later
// effect that removes Mileva's abilities removes the granted ability before
// its effect applies, and the effect returns when the loss ends.
func TestMilevaGrantedStaticRemovedByLaterAbilityLoss(t *testing.T) {
	g, engine, hero, ally, _ := milevaBoard(t)
	spell := addCardToHand(g, game.Player2, loseAllAbilitiesFixture())
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(spell, []game.Target{game.PermanentTarget(hero.ObjectID)}, 0, nil)) {
		t.Fatal("casting the ability-loss fixture failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	assertIndestructible(t, g, false, ally)
	if hasKeyword(g, hero, game.Vigilance) {
		t.Fatal("Mileva kept vigilance after losing all abilities")
	}
	assertPowerToughness(t, g, hero, 5, 5)

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	assertIndestructible(t, g, true, ally)
}

// TestGrantedGroupStaticCloneIsIndependent covers clone safety: the granted
// ability's effects are derived from each game's own state.
func TestGrantedGroupStaticCloneIsIndependent(t *testing.T) {
	g, engine, hero, ally, rival := milevaBoard(t)
	clone := g.Clone()
	cloneAlly, ok := findPermanentByCardID(clone, ally.CardInstanceID)
	if !ok {
		t.Fatal("the clone lacks the ally")
	}
	cloneRival, _ := findPermanentByCardID(clone, rival.CardInstanceID)
	assertIndestructible(t, clone, true, cloneAlly)
	assertIndestructible(t, clone, false, cloneRival)

	castAndResolveRegistered(t, g, engine, game.Player2, "Murder", mana.B, 3, hero)
	assertIndestructible(t, g, false, ally)
	assertIndestructible(t, clone, true, cloneAlly)
}

// TestTheIrencragEverflameEquippedCreatureGetsPlusThree is the registered
// The Irencrag scenario: its granted "Equipped creature gets +3/+3" applies
// in layer 7c to the creature Everflame is attached to.
func TestTheIrencragEverflameEquippedCreatureGetsPlusThree(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	irencrag := addCombatPermanent(g, game.Player1, registryCardDef(t, "The Irencrag"))
	other := addCombatPermanent(g, game.Player1, groupStaticCreature("Bystander", 1, 1))

	legend := addCardToHand(g, game.Player1, groupStaticCreature("Legendary Bear", 2, 2, types.Legendary))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(legend, nil, 0, nil)) {
		t.Fatal("casting the legendary creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("The Irencrag did not trigger on a legendary creature entering")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	creature, ok := findPermanentByCardID(g, legend)
	if !ok {
		t.Fatal("the legendary creature is not on the battlefield")
	}
	if permanentEffectiveName(g, irencrag) != "Everflame, Heroes' Legacy" || !permanentHasSubtype(g, irencrag, types.Equipment) {
		t.Fatalf("The Irencrag is %q, want the Equipment Everflame, Heroes' Legacy", permanentEffectiveName(g, irencrag))
	}
	if abilities := permanentEffectiveAbilities(g, irencrag); len(abilities) != 2 ||
		slices.ContainsFunc(abilities, func(ability game.Ability) bool { _, mana := ability.(*game.ManaAbility); return mana }) {
		t.Fatalf("Everflame has %d abilities, want only equip {3} and its static ability", len(abilities))
	}
	assertPowerToughness(t, g, creature, 2, 2)

	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(irencrag.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)) {
		t.Fatal("Everflame's equip {3} could not be activated")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if irencrag.AttachedTo != opt.Val(creature.ObjectID) {
		t.Fatal("Everflame is not attached to the creature")
	}
	assertPowerToughness(t, g, creature, 5, 5)
	assertPowerToughness(t, g, other, 1, 1)
}
