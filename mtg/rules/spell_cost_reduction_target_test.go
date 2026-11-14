package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#2184): pinned corpus 6da9ced0 Oracle text for the
// target-gated reduction shapes these scenarios model:
//   - Animist's Might (4940783f-2534-41c1-b614-5dabb0dff405) {2}{G}: "This
//     spell costs {2} less to cast if it targets a legendary creature you
//     control. Target creature you control deals damage equal to twice its
//     power to target creature or planeswalker you don't control."
//   - Mystical Dispute (48390b8d-db41-48a5-9bac-316660ab7252) {2}{U}: "This
//     spell costs {2} less to cast if it targets a blue spell. Counter target
//     spell unless its controller pays {3}."
//   - No One Left Behind (afe8fd01-f0c7-46a5-9e9c-e8b9599cd19a) {4}{B}: "This
//     spell costs {3} less to cast if it targets a creature card with mana
//     value 3 or less. Return target creature card from your graveyard to the
//     battlefield."
//   - Swampsnare Trap (fc532d1d-6cf5-4b59-afb3-325d0bc0db77) {2}{B} Aura: "This
//     spell costs {1} less to cast if it targets a creature with flying.
//     Enchant creature. Enchanted creature gets -5/-3."
//   - Torch Breath (b6d61d96-a56a-4f4c-a196-05e03fe16ed2) {X}{R}: "This spell
//     costs {2} less to cast if it targets a blue permanent. ... Torch Breath
//     deals X damage to target creature or planeswalker."
//
// Scryfall rulings retrieved 2026-11-14: Animist's Might (2023-05-12) and
// Swampsnare Trap (2025-10-02) say the reduction changes only the total cost,
// never the mana cost or mana value; No One Left Behind (2022-10-14) says a
// targeted card's {X} is 0; Mystical Dispute and Torch Breath have none.
// Official CR effective 2026-11-06 (MagicCompRules 20261106.txt) read: 601.2c
// (targets are chosen before the total cost), 601.2f (the total cost adds
// increases and subtracts reductions, then is locked in), 118.7 and 118.7a (a
// generic reduction lowers only the generic component). The reduction applies
// when any chosen target qualifies, and Aura spells target per 303.4a and
// 601.2c. These scenarios use test-local definitions; the queued cards are
// authored by later card units.

// targetReductionSpell builds a spell whose own static ability reduces its
// generic cost by reduction when a chosen target matches qualifying.
func targetReductionSpell(name string, cardType types.Card, manaCost cost.Mana, targets []game.TargetSpec, reduction int, qualifying game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{cardType},
		ManaCost: opt.Val(manaCost),
		SpellAbility: opt.Val(game.Mode{
			Targets:  targets,
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				GenericReduction: reduction,
				ReductionTarget:  &qualifying,
			},
		}}}},
	}}
}

func targetReductionCreature(name string, supertypes []types.Super, colors []color.Color, keywords ...game.Keyword) *game.CardDef {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: supertypes,
		Types:      []types.Card{types.Creature},
		Colors:     colors,
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}}
	if len(keywords) > 0 {
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keywords...)}}
	}
	return def
}

func requireTargetReductionCastSpent(t *testing.T, g *game.Game, engine *Engine, act action.Action, wantLeft int) {
	t.Helper()
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("cast with a qualifying target failed")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != wantLeft {
		t.Fatalf("mana left after cast = %d, want %d", got, wantLeft)
	}
}

func requireTargetReductionCastRejected(t *testing.T, g *game.Game, engine *Engine, act action.Action, cardID id.ID) {
	t.Helper()
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("legal actions offered an unaffordable cast without a qualifying target")
	}
	before := g.Stack.Size()
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("cast without a qualifying target succeeded at the reduced cost")
	}
	if g.Stack.Size() != before || !g.Players[game.Player1].Hand.Contains(cardID) && !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("rejected cast moved the card or changed the stack")
	}
}

// animistsMightShape is "costs {2} less if it targets a legendary creature you
// control", with targets "creature you control" and "creature or planeswalker
// you don't control".
func animistsMightShape() *game.CardDef {
	return targetReductionSpell("Test Animist's Might", types.Sorcery, cost.Mana{cost.O(2), cost.G},
		[]game.TargetSpec{
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})},
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerNotYou})},
		},
		2,
		game.TargetSpec{Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Supertypes:    []types.Super{types.Legendary},
			Controller:    game.ControllerYou,
		})},
	)
}

func TestTargetGatedReductionAppliesWhenAnyPermanentTargetQualifies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := animistsMightShape()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	spellID := addCardToHand(g, game.Player1, def)
	legend := addCombatPermanent(g, game.Player1, targetReductionCreature("Own Legend", []types.Super{types.Legendary}, nil))
	ordinary := addCombatPermanent(g, game.Player1, targetReductionCreature("Own Bear", nil, nil))
	enemyLegend := addCombatPermanent(g, game.Player2, targetReductionCreature("Enemy Legend", []types.Super{types.Legendary}, nil))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)

	// An opponent's legendary creature is not "a legendary creature you
	// control", so the full {2}{G} is due and {G} alone cannot pay it.
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{
		game.PermanentTarget(ordinary.ObjectID), game.PermanentTarget(enemyLegend.ObjectID),
	}, 0, nil), spellID)

	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{
		game.PermanentTarget(legend.ObjectID), game.PermanentTarget(enemyLegend.ObjectID),
	}, 0, nil), 0)
}

func TestTargetGatedReductionMatchesTargetedSpellOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spellTargets := game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}}
	blueSpells := spellTargets
	blueSpells.SpellColors = []color.Color{color.Blue}
	def := targetReductionSpell("Test Mystical Dispute", types.Instant, cost.Mana{cost.O(2), cost.U},
		[]game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject, Predicate: spellTargets}},
		2,
		game.TargetSpec{Allow: game.TargetAllowStackObject, Predicate: blueSpells},
	)
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	spellID := addCardToHand(g, game.Player1, def)
	pushOpponentSpell := func(name string, c color.Color) *game.StackObject {
		sourceID := g.IDGen.Next()
		g.CardInstances[sourceID] = &game.CardInstance{ID: sourceID, Owner: game.Player2, Def: &game.CardDef{CardFace: game.CardFace{
			Name: name, Types: []types.Card{types.Sorcery}, Colors: []color.Color{c},
		}}}
		obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: sourceID, Controller: game.Player2}
		g.Stack.Push(obj)
		return obj
	}
	red := pushOpponentSpell("Red Spell", color.Red)
	blue := pushOpponentSpell("Blue Spell", color.Blue)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{game.StackObjectTarget(red.ID)}, 0, nil), spellID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{game.StackObjectTarget(blue.ID)}, 0, nil), 0)
}

func TestTargetGatedReductionMatchesTargetedGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	yourCreatureCards := game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}
	def := targetReductionSpell("Test No One Left Behind", types.Sorcery, cost.Mana{cost.O(4), cost.B},
		[]game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Selection: opt.Val(yourCreatureCards)}},
		3,
		game.TargetSpec{Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
		})},
	)
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	spellID := addCardToHand(g, game.Player1, def)
	creatureCard := func(name string, manaCost cost.Mana) *game.CardDef {
		c := targetReductionCreature(name, nil, nil)
		c.ManaCost = opt.Val(manaCost)
		return c
	}
	big := addCardToGraveyard(g, game.Player1, creatureCard("Big Creature", cost.Mana{cost.O(3), cost.B}))
	// A targeted card's {X} is 0 in the graveyard, so mana value 0 qualifies.
	xCreature := addCardToGraveyard(g, game.Player1, creatureCard("X Creature", cost.Mana{cost.X, cost.X, cost.B}))
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{game.CardTarget(big)}, 0, nil), spellID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{game.CardTarget(xCreature)}, 0, nil), 0)
}

func TestTargetGatedReductionAppliesToAuraEnchantTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	enchantCreature := game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Swampsnare Trap",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
		StaticAbilities: []game.StaticAbility{
			{RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:             game.CostModifierSpell,
					GenericReduction: 1,
					ReductionTarget: &game.TargetSpec{Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Keyword:       game.Flying,
					})},
				},
			}}},
			game.EnchantStaticAbility(&enchantCreature),
		},
	}}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	auraID := addCardToHand(g, game.Player1, def)
	walker := addCombatPermanent(g, game.Player2, targetReductionCreature("Ground Creature", nil, nil))
	flier := addCombatPermanent(g, game.Player2, targetReductionCreature("Flying Creature", nil, nil, game.Flying))
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(auraID, []game.Target{game.PermanentTarget(walker.ObjectID)}, 0, nil), auraID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(auraID, []game.Target{game.PermanentTarget(flier.ObjectID)}, 0, nil), 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	aura, ok := findPermanentByCardID(g, auraID)
	if !ok || !aura.AttachedTo.Exists || aura.AttachedTo.Val != flier.ObjectID {
		t.Fatal("reduced Aura did not resolve attached to the flying creature")
	}
}

// TestTargetGatedReductionLowersXSpellTotalCost proves the reduction applies to
// the generic mana X adds (CR 601.2f) and that legal-action enumeration offers
// X values that are only affordable with the reduction, for permanent, stack and
// graveyard-card targets alike.
func TestTargetGatedReductionLowersXSpellTotalCost(t *testing.T) {
	blueCreature := game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Blue}}
	tests := []struct {
		name   string
		target game.TargetSpec
		reduce game.TargetSpec
		setup  func(g *game.Game) (qualifying, other game.Target)
	}{
		{
			name:   "blue permanent",
			target: game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}})},
			reduce: game.TargetSpec{Allow: game.TargetAllowPermanent, Selection: opt.Val(blueCreature)},
			setup: func(g *game.Game) (game.Target, game.Target) {
				blue := addCombatPermanent(g, game.Player2, targetReductionCreature("Blue Creature", nil, []color.Color{color.Blue}))
				red := addCombatPermanent(g, game.Player2, targetReductionCreature("Red Creature", nil, []color.Color{color.Red}))
				return game.PermanentTarget(blue.ObjectID), game.PermanentTarget(red.ObjectID)
			},
		},
		{
			name:   "blue spell",
			target: game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject, Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}}},
			reduce: game.TargetSpec{Allow: game.TargetAllowStackObject, Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}, SpellColors: []color.Color{color.Blue}}},
			setup: func(g *game.Game) (game.Target, game.Target) {
				push := func(c color.Color) game.Target {
					sourceID := g.IDGen.Next()
					g.CardInstances[sourceID] = &game.CardInstance{ID: sourceID, Owner: game.Player2, Def: &game.CardDef{CardFace: game.CardFace{
						Name: "Opponent Spell", Types: []types.Card{types.Sorcery}, Colors: []color.Color{c},
					}}}
					obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: sourceID, Controller: game.Player2}
					g.Stack.Push(obj)
					return game.StackObjectTarget(obj.ID)
				}
				return push(color.Blue), push(color.Red)
			},
		},
		{
			name:   "small graveyard creature card",
			target: game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
			reduce: game.TargetSpec{Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Selection: opt.Val(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
			})},
			setup: func(g *game.Game) (game.Target, game.Target) {
				card := func(name string, manaCost cost.Mana) game.Target {
					def := targetReductionCreature(name, nil, nil)
					def.ManaCost = opt.Val(manaCost)
					return game.CardTarget(addCardToGraveyard(g, game.Player2, def))
				}
				return card("Small Creature", cost.Mana{cost.O(1), cost.B}), card("Big Creature", cost.Mana{cost.O(5), cost.B})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := targetReductionSpell("Test Torch Breath", types.Instant, cost.Mana{cost.X, cost.R}, []game.TargetSpec{test.target}, 2, test.reduce)
			if issues := game.ValidateCardDef(def); len(issues) != 0 {
				t.Fatalf("ValidateCardDef = %v", issues)
			}
			spellID := addCardToHand(g, game.Player1, def)
			qualifying, other := test.setup(g)
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)

			maxX := map[game.Target]int{}
			for _, act := range engine.legalActions(g, game.Player1) {
				payload, ok := act.CastSpellPayload()
				if !ok || payload.CardID != spellID || len(payload.Targets) != 1 {
					continue
				}
				key := payload.Targets[0]
				if key.Kind == game.TargetCard {
					// Offered card targets also pin the card's zone version.
					key = game.CardTarget(key.CardID)
				}
				maxX[key] = max(maxX[key], payload.XValue)
			}
			if maxX[qualifying] != 2 || maxX[other] != 0 {
				t.Fatalf("largest offered X = qualifying %d, other %d; want 2 and 0", maxX[qualifying], maxX[other])
			}
			requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{other}, 2, nil), spellID)
			requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{qualifying}, 2, nil), 0)
			if obj, ok := g.Stack.Peek(); !ok || obj.XValue != 2 {
				t.Fatalf("stack object = %+v, want X 2", obj)
			}
		})
	}
}

// tappedCreatureReductionSpell is "This spell costs {3} less to cast if it
// targets a tapped creature." on a {4}{W} spell.
func tappedCreatureReductionSpell() *game.CardDef {
	return targetReductionSpell("Test Tapped Reduction", types.Instant, cost.Mana{cost.O(4), cost.W},
		[]game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}},
		3,
		game.TargetSpec{Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Tapped:        game.TriTrue,
		})},
	)
}

// TestTargetGatedReductionCombinesWithKickerCost proves the reduction comes off
// the total cost including an additional kicker cost (CR 601.2f): {4}{W} plus
// kicker {1}{W} costs {2}{W}{W} with a tapped creature target.
func TestTargetGatedReductionCombinesWithKickerCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := tappedCreatureReductionSpell()
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.W}}},
	})
	spellID := addCardToHand(g, game.Player1, def)
	untapped := addCombatPermanent(g, game.Player2, targetReductionCreature("Untapped Creature", nil, nil))
	tapped := addCombatPermanent(g, game.Player2, targetReductionCreature("Tapped Creature", nil, nil))
	tapped.Tapped = true
	g.Players[game.Player1].ManaPool.Add(mana.W, 4)

	requireTargetReductionCastRejected(t, g, engine, action.CastKickedSpell(spellID, []game.Target{game.PermanentTarget(untapped.ObjectID)}, 0, nil), spellID)
	requireTargetReductionCastSpent(t, g, engine, action.CastKickedSpell(spellID, []game.Target{game.PermanentTarget(tapped.ObjectID)}, 0, nil), 0)
	if obj, ok := g.Stack.Peek(); !ok || !obj.KickerPaid {
		t.Fatalf("stack object = %+v, want KickerPaid", obj)
	}
}

// TestTargetGatedReductionAppliesToFlashbackCast proves the reduction also
// lowers an alternative cost paid while casting from the graveyard (CR
// 601.2f): flashback {3}{W} costs {W} with a tapped creature target.
func TestTargetGatedReductionAppliesToFlashbackCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := tappedCreatureReductionSpell()
	def.AlternativeCosts = []cost.Alternative{{
		Label:    "Flashback",
		Mechanic: cost.AlternativeMechanicFlashback,
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
	}}
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{KeywordAbilities: game.SimpleKeywords(game.Flashback)})
	spellID := addCardToGraveyard(g, game.Player1, def)
	untapped := addCombatPermanent(g, game.Player2, targetReductionCreature("Untapped Creature", nil, nil))
	tapped := addCombatPermanent(g, game.Player2, targetReductionCreature("Tapped Creature", nil, nil))
	tapped.Tapped = true
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpellFromZone(spellID, zone.Graveyard, []game.Target{game.PermanentTarget(untapped.ObjectID)}, 0, nil), spellID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpellFromZone(spellID, zone.Graveyard, []game.Target{game.PermanentTarget(tapped.ObjectID)}, 0, nil), 0)
	if obj, ok := g.Stack.Peek(); !ok || !obj.Flashback {
		t.Fatalf("stack object = %+v, want flashback cast", obj)
	}
}
