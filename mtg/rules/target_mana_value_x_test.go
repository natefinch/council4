package rules

import (
	"slices"
	"testing"

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

// manaValueAtMostXSpellDef builds a Dominate-like spell whose lone creature
// target spec bounds the target's mana value by the chosen X via
// ManaValueAtMostX.
func manaValueAtMostXSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Gain Control MV X",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets:       1,
				MaxTargets:       1,
				Constraint:       "target creature with mana value X or less",
				Allow:            game.TargetAllowPermanent,
				Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				ManaValueAtMostX: true,
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

func addCreatureWithManaValue(g *game.Game, controller game.PlayerID, manaValue int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "MV Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}})
}

// TestSpellTargetsSatisfyManaValueXBoundsTargetByX verifies the ManaValueAtMostX
// cast legality: a target is legal only when its mana value is at most the chosen
// X, so a low X cannot gain control of an expensive creature.
func TestSpellTargetsSatisfyManaValueXBoundsTargetByX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mv2 := addCreatureWithManaValue(g, game.Player2, 2)
	mv5 := addCreatureWithManaValue(g, game.Player2, 5)
	card := manaValueAtMostXSpellDef()

	cases := []struct {
		name   string
		target *game.Permanent
		xValue int
		want   bool
	}{
		{"mv2 X=2", mv2, 2, true},
		{"mv2 X=1", mv2, 1, false},
		{"mv2 X=5", mv2, 5, true},
		{"mv5 X=5", mv5, 5, true},
		{"mv5 X=4", mv5, 4, false},
		{"mv5 X=0", mv5, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targets := []game.Target{game.PermanentTarget(tc.target.ObjectID)}
			got := spellTargetsSatisfyManaValueX(g, game.Player1, card, nil, targets, tc.xValue, game.CastBranch{})
			if got != tc.want {
				t.Fatalf("spellTargetsSatisfyManaValueX(mv target, X=%d) = %v, want %v", tc.xValue, got, tc.want)
			}
		})
	}
}

// TestSpellTargetsSatisfyManaValueXIgnoresNonXSpecs verifies a spell without a
// ManaValueAtMostX spec is unaffected by the bound check.
func TestSpellTargetsSatisfyManaValueXIgnoresNonXSpecs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mv5 := addCreatureWithManaValue(g, game.Player2, 5)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Control",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
	if !spellTargetsSatisfyManaValueX(g, game.Player1, card, nil, []game.Target{game.PermanentTarget(mv5.ObjectID)}, 0, game.CastBranch{}) {
		t.Fatal("non-ManaValueAtMostX spell should pass the X-bound check regardless of X")
	}
}

// TestManaValueAtMostXResolutionRechecksBound verifies CR 608.2b re-checking: a
// target whose mana value exceeds the resolving spell's X is no longer legal at
// resolution and is deferred, while a within-bound target stays legal.
func TestManaValueAtMostXResolutionRechecksBound(t *testing.T) {
	specs := spellTargetSpecs(manaValueAtMostXSpellDef(), nil, game.CastBranch{})

	cases := []struct {
		name      string
		manaValue int
		xValue    int
		wantLegal bool
		wantDefer bool
	}{
		{"within bound", 2, 3, true, false},
		{"at bound", 3, 3, true, false},
		{"over bound", 5, 3, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCreatureWithManaValue(g, game.Player2, tc.manaValue)
			obj := &game.StackObject{
				Kind:       game.StackSpell,
				Controller: game.Player1,
				XValue:     tc.xValue,
				Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
			}
			gotLegal := stackObjectHasAnyLegalTargetsForSpecs(g, nil, 0, specs, obj)
			if gotLegal != tc.wantLegal {
				t.Fatalf("stackObjectHasAnyLegalTargetsForSpecs = %v, want %v", gotLegal, tc.wantLegal)
			}
			deferred := obj.Targets[0].Kind == game.TargetDeferred
			if deferred != tc.wantDefer {
				t.Fatalf("target deferred = %v, want %v", deferred, tc.wantDefer)
			}
		})
	}
}

func powerAtMostXCreatureTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
		Constraint:   "target creature with power X or less",
		Selection:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		PowerAtMostX: true,
	}
}

func killingGlareEngineFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Killing Glare", OracleText: "Destroy target creature with power X or less.",
		Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.X, cost.B}),
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{powerAtMostXCreatureTarget()},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
}

// Source review: Killing Glare, GTC 70, Oracle ID
// 7b47891f-130a-4d76-b262-53f5e3bb6b95, revision f7a4d87d-b844-4f20-8b14-4fd32c53dea5
// in the pinned corpus (SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4).
// {X}{B} Instant, "Destroy target creature with power X or less."
// https://api.scryfall.com/cards/f7a4d87d-b844-4f20-8b14-4fd32c53dea5/rulings
// (retrieved 2026-11-19): WotC 2013-01-24 ruling checks power both at
// announcement and at resolution; if above X at resolution, nothing happens.
// Wizards CR effective 2026-09-25, https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 107.3a, 115.1a, 601.2b–h bind X before target choice and payment; 208.1,
// 613.4a–d use modified power; 608.2b rechecks against that same X; 701.8a
// moves the destroyed creature to its owner's graveyard. Route tests also
// use 107.3b (free cast means X=0), 115.7d and 707.10/707.10c (new targets
// must be legal at the copied X), 707.12 (casting a card copy), 702.40a
// (storm copies aren't cast), and 702.33a, 702.34a, 702.35a (additional and
// alternative cost routes). These are unregistered Engine fixtures for #3763.
func TestKillingGlarePowerAtMostX(t *testing.T) {
	for _, pump := range []bool{false, true} {
		name := "at X is destroyed"
		if pump {
			name = "power rises above X in response"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			spell := addCardToHand(g, game.Player1, killingGlareEngineFixture())
			atBound := addCombatPermanent(g, game.Player2, vanillaCreature("Power two", 2, 4))
			above := addCombatPermanent(g, game.Player2, vanillaCreature("Already boosted", 2, 4))
			raiseTargetPower(engine, g, above.ObjectID)
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)

			legal := engine.legalActions(g, game.Player1)
			bad := action.CastSpell(spell, []game.Target{game.PermanentTarget(above.ObjectID)}, 2, nil)
			good := action.CastSpell(spell, []game.Target{game.PermanentTarget(atBound.ObjectID)}, 2, nil)
			assertPowerXCastRejected(t, engine, g, legal, bad)
			// Replay the same announcement against the same starting state.
			for _, replay := range []*game.Game{g, g.Clone()} {
				if !containsAction(legal, good) || !engine.applyAction(replay, game.Player1, good) {
					t.Fatal("could not cast Killing Glare at X=2 targeting the power-2 creature")
				}
				obj, ok := replay.Stack.Peek()
				if !ok || obj.XValue != 2 || !replay.Players[game.Player1].ManaPool.IsEmpty() {
					t.Fatal("Killing Glare did not retain the announced and paid X=2")
				}
				if pump {
					raiseTargetPower(engine, replay, atBound.ObjectID)
				}
				// Also clone the pending spell to exercise resolution from stored X.
				for _, pending := range []*game.Game{replay, replay.Clone()} {
					engine.resolveTopOfStack(pending, &TurnLog{})
					_, alive := permanentByObjectID(pending, atBound.ObjectID)
					if alive != pump || pending.Players[game.Player2].Graveyard.Contains(atBound.CardInstanceID) == pump ||
						!pending.Players[game.Player1].Graveyard.Contains(spell) {
						t.Fatalf("power boost=%v: target survived=%v or incorrect graveyard destinations", pump, alive)
					}
				}
			}
		})
	}
}

func assertPowerXCastRejected(t *testing.T, engine *Engine, g *game.Game, legal []action.Action, act action.Action) {
	t.Helper()
	if containsAction(legal, act) || engine.applyAction(g.Clone(), game.Player1, act) {
		t.Fatalf("illegal cast was offered or accepted: %+v", act)
	}
}

func raiseTargetPower(engine *Engine, g *game.Game, target id.ID) {
	addEffectSpellToStack(g, game.Player2, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerPowerToughnessModify, PowerDelta: 1,
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(target)})
	engine.resolveTopOfStack(g, &TurnLog{})
}

// Source review: Minamo Sightbender, DDI 5, Oracle ID
// c4f1d8c8-29e5-43fe-8281-7d42d0173dc5, pinned revision a598d08d-c9d4-4de5-bef5-998dd115b8b6.
// {1}{U}, Creature — Human Wizard, 1/2; "{X}, {T}: Target creature with power
// X or less can't be blocked this turn."
// https://api.scryfall.com/cards/a598d08d-c9d4-4de5-bef5-998dd115b8b6/rulings
// returned no rulings on 2026-11-19. Wizards CR 2026-09-25,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt:
// 107.3a, 115.1c, 601.2b–h and 602.2a–b bind X and the target before paying;
// 602.5a governs the tap cost; 208.1, 613.4a–d and 608.2b check current power;
// 509.1a–b disallows blocking the affected creature, until cleanup per 611.2a
// and 514.2. Unregistered Engine fixture, replacing the retained #3763 probe.
func TestMinamoSightbenderPowerAtMostX(t *testing.T) {
	const text = "{X}, {T}: Target creature with power X or less can't be blocked this turn."
	def := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Minamo Sightbender", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Wizard},
			Colors: []color.Color{color.Blue}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield, ManaCost: opt.Val(cost.Mana{cost.X}), AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Targets: []game.TargetSpec{powerAtMostXCreatureTarget()},
					Sequence: []game.Instruction{{Primitive: game.ApplyRule{
						Object:      opt.Val(game.TargetPermanentReference(0)),
						RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBeBlocked}},
						Duration:    game.DurationUntilEndOfTurn,
					}}},
				}.Ability(),
			}},
		},
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, def)
	inRange := addCombatPermanent(g, game.Player1, vanillaCreature("Power one", 1, 2))
	overRange := addCombatPermanent(g, game.Player1, vanillaCreature("Power two", 2, 2))
	blocker := addCombatPermanent(g, game.Player2, vanillaCreature("Blocker", 1, 2))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !canBlockAttacker(g, blocker, inRange) {
		t.Fatal("target can't be blocked before the ability resolves")
	}
	ability := def.ActivatedAbilityIndex(0)
	legal := engine.legalActions(g, game.Player1)
	bad := action.ActivateAbility(source.ObjectID, ability, []game.Target{game.PermanentTarget(overRange.ObjectID)}, 1)
	good := action.ActivateAbility(source.ObjectID, ability, []game.Target{game.PermanentTarget(inRange.ObjectID)}, 1)
	if containsAction(legal, bad) || engine.applyAction(g.Clone(), game.Player1, bad) {
		t.Fatal("engine accepted a power-2 target for Minamo Sightbender's X=1")
	}
	if !containsAction(legal, good) || !engine.applyAction(g, game.Player1, good) {
		t.Fatal("could not target the power-1 creature with Minamo Sightbender's X=1")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 1 || !source.Tapped || !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatal("Minamo Sightbender did not tap and pay X=1")
	}
	pumped := g.Clone()
	raiseTargetPower(engine, pumped, inRange.ObjectID)
	engine.resolveTopOfStack(pumped, &TurnLog{})
	pumpedTarget, _ := permanentByObjectID(pumped, inRange.ObjectID)
	pumpedBlocker, _ := permanentByObjectID(pumped, blocker.ObjectID)
	engine.resolveTopOfStack(g, &TurnLog{})
	if canBlockAttacker(g, blocker, inRange) || !canBlockAttacker(g, blocker, overRange) ||
		!canBlockAttacker(pumped, pumpedBlocker, pumpedTarget) {
		t.Fatal("Minamo Sightbender must make only its still-legal target unblockable")
	}
}

func TestPowerAtMostXStackCopyRoutes(t *testing.T) {
	for _, storm := range []bool{false, true} {
		name := "ordinary copy may only retarget in range"
		if storm {
			name = "storm copy keeps X and is not cast"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			spell := addCardToHand(g, game.Player1, killingGlareEngineFixture())
			first := game.PermanentTarget(addCombatPermanent(g, game.Player2, vanillaCreature("Original", 2, 3)).ObjectID)
			legal := game.PermanentTarget(addCombatPermanent(g, game.Player2, vanillaCreature("Legal replacement", 2, 3)).ObjectID)
			illegal := game.PermanentTarget(addCombatPermanent(g, game.Player2, vanillaCreature("Illegal replacement", 3, 4)).ObjectID)
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{first}, 2, nil)) {
				t.Fatal("could not cast the original spell")
			}
			original, _ := g.Stack.Peek()
			want := first
			if storm {
				resolveStormTriggerFor(engine, g, original, 1)
			} else {
				want = legal
				addEffectSpellToStack(g, game.Player1,
					game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
					[]game.Target{game.StackObjectTarget(original.ID)})
				var agents [game.NumPlayers]PlayerAgent
				agents[game.Player1] = crimeChoiceAgent(func(request game.ChoiceRequest) ([]int, bool) {
					for _, option := range request.Options {
						if slices.Contains(option.Targets, illegal) {
							t.Fatal("power greater than copied X was offered as a new target")
						}
					}
					return retargetTo(legal)(request)
				})
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			}
			copy, ok := g.Stack.Peek()
			if !ok || !copy.Copy || copy.XValue != 2 || !slices.Equal(copy.Targets, []game.Target{want}) ||
				spellCastEvents(g, original.ID) != 1 || spellCastEvents(g, copy.ID) != 0 {
				t.Fatalf("copy has wrong target, X, or cast provenance: %+v", copy)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, alive := permanentByObjectID(g, want.PermanentID); alive {
				t.Fatal("the stack copy did not destroy its in-range target")
			}
		})
	}
}

func TestPowerAtMostXAlternativeCastRoutes(t *testing.T) {
	t.Run("free cast offered from hand only at X zero", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spell := addCardToHand(g, game.Player1, killingGlareEngineFixture())
		addCombatPermanent(g, game.Player2, vanillaCreature("Power one", 1, 3))
		if castableHandSpell(g, game.Player1, spell) {
			t.Fatal("hand preflight offered a free cast with only a power-1 target")
		}
		addCombatPermanent(g, game.Player2, vanillaCreature("Power zero", 0, 3))
		if !castableHandSpell(g, game.Player1, spell) {
			t.Fatal("hand preflight missed a free cast with a power-0 target")
		}
	})
	t.Run("paid resolution cast finds power breakpoint", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		setMainPhasePriority(g, game.Player1)
		cardID := addCardToExile(g, game.Player1, killingGlareEngineFixture())
		card, _ := g.GetCardInstance(cardID)
		target := addCombatPermanent(g, game.Player2, vanillaCreature("Power three", 3, 4))
		g.Players[game.Player1].ManaPool.Add(mana.B, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 3)
		if !engine.castMadnessSpellWithChoices(g, game.Player1, card, cost.Mana{cost.X, cost.B}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("paid resolution cast missed X=3 as the first legal X")
		}
		obj, _ := g.Stack.Peek()
		if obj.XValue != 3 || !g.Players[game.Player1].ManaPool.IsEmpty() {
			t.Fatalf("resolution cast did not pay X=3: %+v", obj)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
			t.Fatal("paid resolution cast did not destroy the power-3 creature")
		}
	})
	for _, copyCard := range []bool{false, true} {
		name := "free original from exile"
		if copyCard {
			name = "free cast of a card copy from exile"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			spell := addCardToExile(g, game.Player1, killingGlareEngineFixture())
			above := addCombatPermanent(g, game.Player2, vanillaCreature("Power one", 1, 3))
			freeCast := func() bool {
				if copyCard {
					return engine.castFreeCopyOfCard(g, game.Player1, spell, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				}
				return engine.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, spell,
					zone.Exile, game.Selection{}, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			}
			if freeCast() {
				t.Fatal("free cast was legal with only a power-1 target")
			}
			zero := addCombatPermanent(g, game.Player2, vanillaCreature("Power zero", 0, 3))
			if !freeCast() {
				t.Fatal("free cast could not choose the power-0 target")
			}
			obj, _ := g.Stack.Peek()
			if obj.XValue != 0 || obj.CastCopy != copyCard || spellCastEvents(g, obj.ID) != 1 ||
				!slices.Equal(obj.Targets, []game.Target{game.PermanentTarget(zero.ObjectID)}) {
				t.Fatalf("free cast has wrong X, target or provenance: %+v", obj)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(zero.CardInstanceID) ||
				g.Players[game.Player2].Graveyard.Contains(above.CardInstanceID) ||
				g.Players[game.Player1].Exile.Contains(spell) != copyCard {
				t.Fatal("free cast resolved with the wrong target or source destination")
			}
		})
	}
}

func TestPowerAtMostXWithCombinedCosts(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		name := "decline optional mana"
		if kicked {
			name = "pay optional mana"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := killingGlareEngineFixture()
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			def.StaticAbilities = []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.G}}},
			}}
			spell := addCardToHand(g, game.Player1, def)
			fodder := addCardToHand(g, game.Player1, vanillaCreature("Discard payment", 1, 1))
			target := addCombatPermanent(g, game.Player2, vanillaCreature("At bound", 2, 4))
			above := addCombatPermanent(g, game.Player2, vanillaCreature("Above bound", 3, 4))
			targets := []game.Target{game.PermanentTarget(target.ObjectID)}
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			act := action.CastKickedSpell(spell, targets, 2, nil)
			assertPowerXCastRejected(t, engine, g, engine.legalActions(g, game.Player1), act)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			assertPowerXCastRejected(t, engine, g, engine.legalActions(g, game.Player1),
				action.CastKickedSpell(spell, []game.Target{game.PermanentTarget(above.ObjectID)}, 2, nil))
			if !kicked {
				act = action.CastSpell(spell, targets, 2, nil)
			}
			if !containsAction(engine.legalActions(g, game.Player1), act) || !engine.applyAction(g, game.Player1, act) {
				t.Fatal("the affordable in-range cast failed")
			}
			obj, _ := g.Stack.Peek()
			if obj.XValue != 2 || obj.KickerPaid != kicked ||
				g.Players[game.Player1].ManaPool.IsEmpty() != kicked ||
				!g.Players[game.Player1].Graveyard.Contains(fodder) {
				t.Fatal("X, optional mana and mandatory discard did not remain separate payments")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
				t.Fatal("the spell did not destroy the power-2 creature")
			}
		})
	}
}

func TestPowerAtMostXFlashbackDestinations(t *testing.T) {
	for _, outcome := range []string{"resolve", "fizzle", "counter"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := killingGlareEngineFixture()
			def.StaticAbilities = []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.X, cost.B}}},
			}}
			spell := addCardToGraveyard(g, game.Player1, def)
			target := addCombatPermanent(g, game.Player2, vanillaCreature("At bound", 1, 3))
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			targets := []game.Target{game.PermanentTarget(target.ObjectID)}
			legal := engine.legalActions(g, game.Player1)
			assertPowerXCastRejected(t, engine, g, legal, action.CastSpellFromZone(spell, zone.Graveyard, targets, 0, nil))
			act := action.CastSpellFromZone(spell, zone.Graveyard, targets, 1, nil)
			if !containsAction(legal, act) || !engine.applyAction(g, game.Player1, act) {
				t.Fatal("flashback could not cast at X=1")
			}
			obj, _ := g.Stack.Peek()
			if !obj.Flashback || obj.XValue != 1 {
				t.Fatalf("wrong flashback announcement: %+v", obj)
			}
			if outcome == "counter" {
				if !counterStackObject(g, obj.ID) {
					t.Fatal("could not counter the flashback spell")
				}
			} else {
				if outcome == "fizzle" {
					raiseTargetPower(engine, g, target.ObjectID)
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			if !g.Players[game.Player1].Exile.Contains(spell) ||
				g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) != (outcome == "resolve") {
				t.Fatalf("wrong spell or target destination after %s", outcome)
			}
		})
	}
}
