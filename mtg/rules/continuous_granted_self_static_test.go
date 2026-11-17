package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3953 (sub-issue of #2098). Official CR effective
// 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// read 604.3a, 611.3a, 613.1f, 613.4b, 613.5, 613.6, 613.7a and 613.8a.
// 604.3a(4): an ability an object gains from another effect is not a
// characteristic-defining ability, so a granted "power and toughness are each
// equal to" ability applies in layer 7b. 613.7a: a static ability's effect has
// the later of its object's timestamp and the timestamp of the effect that
// created the ability. 613.6: an effect that started applying continues in
// later layers even if its ability is removed. Cards (pinned corpus; the test
// fixtures below follow this Oracle text and are not registered):
// Chimeric Mass (oracle 98066238-28f8-41e3-903a-d367927f23ca, {X} artifact)
// "This artifact enters with X charge counters on it. / {1}: Until end of
// turn, this artifact becomes a Construct artifact creature with "This
// creature's power and toughness are each equal to the number of charge
// counters on it."" (rulings 2011-01-01: its P/T changes as charge counters
// are added or removed; activating it overrides earlier P/T-setting effects
// but not modifying effects; with three counters, Giant Growth makes it 6/6,
// a later "becomes 1/1" makes it 4/4, and activating it again makes it 6/6).
// Svogthos, the Restless Tomb (oracle a34a70b8-02e5-4e8c-a9e7-b21c5a11dddf,
// land) "{T}: Add {C}. / {3}{B}{G}: Until end of turn, this land becomes a
// black and green Plant Zombie creature with "This creature's power and
// toughness are each equal to the number of creature cards in your
// graveyard." It's still a land." (rulings 2005-10-01: its P/T changes as
// creature cards enter or leave your graveyard; 2009-10-01: activating it
// overrides earlier P/T-setting effects). Restless Spire (oracle
// 0ca4e80e-c19c-4b74-b531-c5a4dc5a8ba9, land) "This land enters tapped. /
// {T}: Add {U} or {R}. / {U}{R}: Until end of turn, this land becomes a 2/1
// blue and red Elemental creature with "During your turn, this creature has
// first strike." It's still a land. / Whenever this land attacks, scry 1."
// (rulings 2023-09-01: the attack trigger works if another effect animates it;
// summoning sickness applies). Registered Giant Growth ("Target creature gets
// +3/+3 until end of turn."), Diminish ("Target creature has base power and
// toughness 1/1 until end of turn.") and Murder ("Destroy target creature.")
// supply the other effects.

func chimericMassFixture() *game.CardDef {
	ptStatic := &game.StaticAbility{
		Text: "This creature's power and toughness are each equal to the number of charge counters on it.",
		ContinuousEffects: []game.ContinuousEffect{{
			Duration:       game.DurationPermanent,
			AffectedSource: true,
			Layer:          game.LayerPowerToughnessSet,
			SetPowerDynamic: opt.Val(game.DynamicAmount{
				Kind: game.DynamicAmountObjectCounters, Multiplier: 1,
				CounterKind: counter.Charge, Object: game.SourcePermanentReference(),
			}),
			SetToughnessDynamic: opt.Val(game.DynamicAmount{
				Kind: game.DynamicAmountObjectCounters, Multiplier: 1,
				CounterKind: counter.Charge, Object: game.SourcePermanentReference(),
			}),
		}},
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Chimeric Mass",
		ManaCost: opt.Val(cost.Mana{cost.X}),
		Types:    []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			`{1}: Until end of turn, this artifact becomes a Construct artifact creature with "This creature's power and toughness are each equal to the number of charge counters on it."`,
			cost.Mana{cost.O(1)},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerType,
				AddTypes: []types.Card{types.Artifact, types.Creature}, AddSubtypes: []types.Sub{types.Construct},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerAbility,
				AddAbilities: []game.Ability{ptStatic},
			},
		)},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("This artifact enters with X charge counters on it.",
				game.CounterPlacement{Kind: counter.Charge, AmountFromX: true}),
		},
	}}
}

func svogthosFixture() *game.CardDef {
	you := game.ControllerReference()
	creatureCards := game.DynamicAmount{
		Kind: game.DynamicAmountCountCardsInZone, Multiplier: 1, Player: &you,
		CardZone: zone.Graveyard, Selection: &game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
	ptStatic := &game.StaticAbility{
		Text: "This creature's power and toughness are each equal to the number of creature cards in your graveyard.",
		ContinuousEffects: []game.ContinuousEffect{{
			Duration: game.DurationPermanent, AffectedSource: true, Layer: game.LayerPowerToughnessSet,
			SetPowerDynamic: opt.Val(creatureCards), SetToughnessDynamic: opt.Val(creatureCards),
		}},
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:          "Svogthos, the Restless Tomb",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			`{3}{B}{G}: Until end of turn, this land becomes a black and green Plant Zombie creature with "This creature's power and toughness are each equal to the number of creature cards in your graveyard." It's still a land.`,
			cost.Mana{cost.O(3), cost.B, cost.G},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerColor,
				SetColors: []color.Color{color.Black, color.Green},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerType,
				AddTypes: []types.Card{types.Creature}, AddSubtypes: []types.Sub{types.Plant, types.Zombie},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerAbility,
				AddAbilities: []game.Ability{ptStatic},
			},
		)},
	}}
}

func restlessSpireFixture() *game.CardDef {
	firstStrike := &game.StaticAbility{
		Text:      "During your turn, this creature has first strike.",
		Condition: opt.Val(game.Condition{Text: "During your turn", SourceControllerTurn: true}),
		ContinuousEffects: []game.ContinuousEffect{{
			Duration: game.DurationPermanent, AffectedSource: true, Layer: game.LayerAbility,
			AddKeywords: []game.Keyword{game.FirstStrike},
		}},
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:          "Restless Spire",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaChoiceAbility(mana.U, mana.R)},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			`{U}{R}: Until end of turn, this land becomes a 2/1 blue and red Elemental creature with "During your turn, this creature has first strike." It's still a land.`,
			cost.Mana{cost.U, cost.R},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerColor,
				SetColors: []color.Color{color.Blue, color.Red},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerType,
				AddTypes: []types.Card{types.Creature}, AddSubtypes: []types.Sub{types.Elemental},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
				SetPower: opt.Val(game.PT{Value: 2}), SetToughness: opt.Val(game.PT{Value: 1}),
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerAbility,
				AddAbilities: []game.Ability{firstStrike},
			},
		)},
	}}
}

// loseAllAbilitiesFixture is a test-only instant: "Target creature loses all
// abilities until end of turn."
func loseAllAbilitiesFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Ability Drain",
		ManaCost: opt.Val(cost.Mana{cost.U}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{
			Modes: []game.Mode{{
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "target creature",
					Allow:     game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
					Object:   opt.Val(game.TargetPermanentReference(0)),
					Duration: game.DurationUntilEndOfTurn,
					ContinuousEffects: []game.ContinuousEffect{{
						Duration: game.DurationPermanent, Layer: game.LayerAbility, RemoveAllAbilities: true,
					}},
				}}},
			}},
			MinModes: 1, MaxModes: 1,
		}),
	}}
}

func animateSelfAbility(text string, manaCost cost.Mana, effects ...game.ContinuousEffect) game.ActivatedAbility {
	return game.ActivatedAbility{
		Text:           text,
		ManaCost:       opt.Val(manaCost),
		ZoneOfFunction: zone.Battlefield,
		Content: game.AbilityContent{
			Modes: []game.Mode{{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.SourcePermanentReference()),
				Duration:          game.DurationUntilEndOfTurn,
				ContinuousEffects: effects,
			}}}}},
			MinModes: 1, MaxModes: 1,
		},
	}
}

func grantedStaticGame(active game.PlayerID) (*game.Game, *Engine) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = active, active
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	return g, NewEngine(nil)
}

// activateAndResolve pays for and activates permanent's first activated
// ability through the real action path and resolves it.
func activateAndResolve(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, permanent *game.Permanent, payment ...mana.Color) {
	t.Helper()
	for _, c := range payment {
		g.Players[player].ManaPool.Add(c, 1)
	}
	g.Turn.PriorityPlayer = player
	activate := action.ActivateAbility(permanent.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, player), activate) {
		t.Fatalf("activating %s was not offered", permanentName(g, permanent))
	}
	if !engine.applyAction(g, player, activate) {
		t.Fatalf("activating %s failed", permanentName(g, permanent))
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[player].ManaPool.IsEmpty() {
		t.Fatal("fixture: mana is still floating after the activation")
	}
}

// castChimericMass casts the Chimeric Mass fixture for X through the real cast
// path and returns the resolved permanent.
func castChimericMass(t *testing.T, g *game.Game, engine *Engine, x int) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, chimericMassFixture())
	g.Players[game.Player1].ManaPool.Add(mana.C, x)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, x, nil)) {
		t.Fatal("casting Chimeric Mass failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	mass, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatal("Chimeric Mass is not on the battlefield")
	}
	if got := mass.Counters.Get(counter.Charge); got != x {
		t.Fatalf("Chimeric Mass entered with %d charge counters, want %d", got, x)
	}
	return mass
}

func assertPowerToughness(t *testing.T, g *game.Game, permanent *game.Permanent, power, toughness int) {
	t.Helper()
	gotToughness, _ := effectiveToughness(g, permanent)
	if gotPower := effectivePower(g, permanent); gotPower != power || gotToughness != toughness {
		t.Fatalf("%s is %d/%d, want %d/%d", permanentName(g, permanent), gotPower, gotToughness, power, toughness)
	}
}

// TestChimericMassGrantedStaticSetsPowerToughnessFromCharge is the #3953
// Chimeric Mass scenario: the granted P/T ability is not a CDA, tracks its
// charge counters live in layer 7b, is modified by a later +X/+X in 7c and
// ends with the animation at cleanup.
func TestChimericMassGrantedStaticSetsPowerToughnessFromCharge(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	mass := castChimericMass(t, g, engine, 3)
	if permanentHasType(g, mass, types.Creature) {
		t.Fatal("Chimeric Mass entered as a creature")
	}

	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	if !permanentHasType(g, mass, types.Creature) || !permanentHasType(g, mass, types.Artifact) ||
		!permanentHasSubtype(g, mass, types.Construct) {
		t.Fatal("Chimeric Mass did not become a Construct artifact creature")
	}
	assertPowerToughness(t, g, mass, 3, 3)

	mass.Counters.Remove(counter.Charge, 1)
	assertPowerToughness(t, g, mass, 2, 2)

	castAndResolveRegistered(t, g, engine, game.Player1, "Giant Growth", mana.G, 1, mass)
	assertPowerToughness(t, g, mass, 5, 5)

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentHasType(g, mass, types.Creature) || !permanentHasType(g, mass, types.Artifact) {
		t.Fatal("Chimeric Mass is still a creature after cleanup")
	}
	if _, ok := findPermanentByCardID(g, mass.CardInstanceID); !ok {
		t.Fatal("Chimeric Mass left the battlefield")
	}
}

// TestChimericMassRulingOverridesEarlierSetEffects plays the 2011-01-01
// ruling: Giant Growth makes the 3/3 Mass 6/6, a later Diminish sets its base
// to 1/1 (4/4), and activating it again makes the newer granted ability's 7b
// effect the latest one (CR 613.7a), so it is 6/6 again.
func TestChimericMassRulingOverridesEarlierSetEffects(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	mass := castChimericMass(t, g, engine, 3)
	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	castAndResolveRegistered(t, g, engine, game.Player1, "Giant Growth", mana.G, 1, mass)
	assertPowerToughness(t, g, mass, 6, 6)
	castAndResolveRegistered(t, g, engine, game.Player1, "Diminish", mana.U, 1, mass)
	assertPowerToughness(t, g, mass, 4, 4)
	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	assertPowerToughness(t, g, mass, 6, 6)
}

// TestChimericMassGrantedStaticRemovedByLaterAbilityLoss checks layer-6
// timestamp order. An activation that resolves after a "loses all abilities"
// effect grants a newer P/T ability the loss does not remove, so it is 3/3.
// Once a later loss applies, the granted ability is removed before it applies
// in layer 7b, so nothing sets its P/T (CR 208.5 treats it as 0/0).
func TestChimericMassGrantedStaticRemovedByLaterAbilityLoss(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	mass := castChimericMass(t, g, engine, 3)
	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	assertPowerToughness(t, g, mass, 3, 3)

	// Activate again and, in response, make the Mass lose all abilities. The
	// activation still resolves (CR 113.7a) after the loss.
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mass.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Chimeric Mass again failed")
	}
	drainMass := func() {
		t.Helper()
		spell := addCardToHand(g, game.Player1, loseAllAbilitiesFixture())
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		g.Turn.PriorityPlayer = game.Player1
		if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PermanentTarget(mass.ObjectID)}, 0, nil)) {
			t.Fatal("casting the ability-loss fixture failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	drainMass()
	assertPowerToughness(t, g, mass, 0, 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	assertPowerToughness(t, g, mass, 3, 3)

	drainMass()
	if !permanentHasType(g, mass, types.Creature) {
		t.Fatal("losing abilities ended the animation's type change")
	}
	if _, ok := effectiveToughness(g, mass); ok {
		t.Fatal("the removed granted ability still set Chimeric Mass's toughness")
	}
	assertPowerToughness(t, g, mass, 0, 0)
}

// TestSvogthosGrantedStaticCountsGraveyardLive is the #3953 Svogthos
// scenario: the granted ability counts creature cards in its controller's
// graveyard whenever its P/T is determined.
func TestSvogthosGrantedStaticCountsGraveyardLive(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	svogthos := addCombatPermanent(g, game.Player1, svogthosFixture())
	creatureCard := func(name string) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name: name, Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
		}}
	}
	addCardToGraveyard(g, game.Player1, creatureCard("Dead Bear"))
	addCardToGraveyard(g, game.Player1, creatureCard("Dead Elf"))
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dead Spell", Types: []types.Card{types.Sorcery}}})
	addCardToGraveyard(g, game.Player2, creatureCard("Opponent's Dead Bear"))
	victim := addCombatPermanent(g, game.Player1, creatureCard("Doomed Squire"))

	activateAndResolve(t, g, engine, game.Player1, svogthos, mana.C, mana.C, mana.C, mana.B, mana.G)
	if !permanentHasType(g, svogthos, types.Creature) || !permanentHasType(g, svogthos, types.Land) {
		t.Fatal("Svogthos did not become a land creature")
	}
	assertPowerToughness(t, g, svogthos, 2, 2)

	castAndResolveRegistered(t, g, engine, game.Player1, "Murder", mana.B, 3, victim)
	if !g.Players[game.Player1].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("Murder did not put the creature into its owner's graveyard")
	}
	assertPowerToughness(t, g, svogthos, 3, 3)
}

// TestRestlessSpireGrantedFirstStrikeOnlyDuringYourTurn is the #3953 Restless
// Spire scenario: the granted static's conditional layer-6 effect applies
// during its controller's turn and not during an opponent's turn.
func TestRestlessSpireGrantedFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	t.Run("your turn", func(t *testing.T) {
		g, engine := grantedStaticGame(game.Player1)
		spire := addCombatPermanent(g, game.Player1, restlessSpireFixture())
		activateAndResolve(t, g, engine, game.Player1, spire, mana.U, mana.R)
		if !permanentHasType(g, spire, types.Creature) || !permanentHasType(g, spire, types.Land) {
			t.Fatal("Restless Spire did not become a land creature")
		}
		assertPowerToughness(t, g, spire, 2, 1)
		if !hasKeyword(g, spire, game.FirstStrike) {
			t.Fatal("Restless Spire lacks first strike during its controller's turn")
		}
	})
	t.Run("opponent's turn", func(t *testing.T) {
		g, engine := grantedStaticGame(game.Player2)
		spire := addCombatPermanent(g, game.Player1, restlessSpireFixture())
		activateAndResolve(t, g, engine, game.Player1, spire, mana.U, mana.R)
		assertPowerToughness(t, g, spire, 2, 1)
		if hasKeyword(g, spire, game.FirstStrike) {
			t.Fatal("Restless Spire has first strike during an opponent's turn")
		}
	})
}

// TestGrantedSelfStaticCloneIsIndependent clones a game mid-turn while
// Chimeric Mass and Restless Spire are animated: the clone computes the same
// characteristics and later changes to either game do not leak.
func TestGrantedSelfStaticCloneIsIndependent(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	mass := castChimericMass(t, g, engine, 3)
	spire := addCombatPermanent(g, game.Player1, restlessSpireFixture())
	activateAndResolve(t, g, engine, game.Player1, mass, mana.C)
	activateAndResolve(t, g, engine, game.Player1, spire, mana.U, mana.R)

	clone := g.Clone()
	cloneMass, ok := findPermanentByCardID(clone, mass.CardInstanceID)
	if !ok {
		t.Fatal("the clone lacks Chimeric Mass")
	}
	cloneSpire, ok := findPermanentByCardID(clone, spire.CardInstanceID)
	if !ok {
		t.Fatal("the clone lacks Restless Spire")
	}
	assertPowerToughness(t, clone, cloneMass, 3, 3)
	assertPowerToughness(t, clone, cloneSpire, 2, 1)
	if !hasKeyword(clone, cloneSpire, game.FirstStrike) {
		t.Fatal("the cloned Restless Spire lacks first strike")
	}

	mass.Counters.Remove(counter.Charge, 2)
	clone.Turn.ActivePlayer = game.Player2
	assertPowerToughness(t, g, mass, 1, 1)
	assertPowerToughness(t, clone, cloneMass, 3, 3)
	if !hasKeyword(g, spire, game.FirstStrike) {
		t.Fatal("changing the clone's turn removed the original's first strike")
	}
	if hasKeyword(clone, cloneSpire, game.FirstStrike) {
		t.Fatal("the cloned Restless Spire has first strike during an opponent's turn")
	}
}

// TestGrantedUnconditionalSelfStaticGrantsKeyword covers a granted static
// whose own layer-6 effect is unconditional, which the per-permanent layer
// path applies at its CR 613.7a timestamp after the grant.
func TestGrantedUnconditionalSelfStaticGrantsKeyword(t *testing.T) {
	g, engine := grantedStaticGame(game.Player1)
	flier := &game.StaticAbility{
		Text: "This creature has flying.",
		ContinuousEffects: []game.ContinuousEffect{{
			Duration: game.DurationPermanent, AffectedSource: true, Layer: game.LayerAbility,
			AddKeywords: []game.Keyword{game.Flying},
		}},
	}
	statue := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Winged Statue",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			`{1}: Until end of turn, this artifact becomes a 1/1 artifact creature with "This creature has flying."`,
			cost.Mana{cost.O(1)},
			game.ContinuousEffect{Duration: game.DurationPermanent, Layer: game.LayerType, AddTypes: []types.Card{types.Creature}},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
				SetPower: opt.Val(game.PT{Value: 1}), SetToughness: opt.Val(game.PT{Value: 1}),
			},
			game.ContinuousEffect{Duration: game.DurationPermanent, Layer: game.LayerAbility, AddAbilities: []game.Ability{flier}},
		)},
	}})
	if needsAbilityLayerEvaluation(g, staticAbilitySources(g)) {
		t.Fatal("fixture: the unconditional grant should use the per-permanent layer path")
	}
	activateAndResolve(t, g, engine, game.Player1, statue, mana.C)
	if needsAbilityLayerEvaluation(g, staticAbilitySources(g)) {
		t.Fatal("fixture: the unconditional grant should use the per-permanent layer path")
	}
	if !hasKeyword(g, statue, game.Flying) {
		t.Fatal("the granted static did not give its permanent flying")
	}
}

// TestGrantedConditionalKeywordAbilitiesFollowCondition covers a granted static
// whose keyword abilities carry its condition, written as KeywordAbilities
// rather than an AddKeywords effect: the permanent has every such keyword,
// including protection, only while the condition holds (CR 611.3a), and a
// later ability loss removes them.
func TestGrantedConditionalKeywordAbilitiesFollowCondition(t *testing.T) {
	duringYourTurn := &game.StaticAbility{
		Text:      "During your turn, this creature has first strike and protection from red.",
		Condition: opt.Val(game.Condition{Text: "During your turn", SourceControllerTurn: true}),
		KeywordAbilities: []game.KeywordAbility{
			game.SimpleKeyword{Kind: game.FirstStrike},
			game.ProtectionKeyword{FromColors: []color.Color{color.Red}},
		},
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Turn-Bound Statue",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			`{1}: Until end of turn, this artifact becomes a 2/2 artifact creature with "During your turn, this creature has first strike and protection from red."`,
			cost.Mana{cost.O(1)},
			game.ContinuousEffect{Duration: game.DurationPermanent, Layer: game.LayerType, AddTypes: []types.Card{types.Creature}},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
				SetPower: opt.Val(game.PT{Value: 2}), SetToughness: opt.Val(game.PT{Value: 2}),
			},
			game.ContinuousEffect{Duration: game.DurationPermanent, Layer: game.LayerAbility, AddAbilities: []game.Ability{duringYourTurn}},
		)},
	}}
	red := sourceChars{colors: []color.Color{color.Red}}

	g, engine := grantedStaticGame(game.Player2)
	statue := addCombatPermanent(g, game.Player1, def)
	activateAndResolve(t, g, engine, game.Player1, statue, mana.C)
	if hasKeyword(g, statue, game.FirstStrike) || permanentProtectedFromChars(g, statue, red) {
		t.Fatal("the granted keywords apply during an opponent's turn")
	}

	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	if !hasKeyword(g, statue, game.FirstStrike) || !permanentProtectedFromChars(g, statue, red) {
		t.Fatal("the granted keywords do not apply during their controller's turn")
	}

	spell := addCardToHand(g, game.Player1, loseAllAbilitiesFixture())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PermanentTarget(statue.ObjectID)}, 0, nil)) {
		t.Fatal("casting the ability-loss fixture failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if hasKeyword(g, statue, game.FirstStrike) || permanentProtectedFromChars(g, statue, red) {
		t.Fatal("the granted keywords survived a later ability loss")
	}
	assertPowerToughness(t, g, statue, 2, 2)
}
