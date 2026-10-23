package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Oracle SHA-256: 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// WotC rulings retrieved 2026-10-14: Twincast/Reiterate (2021-03-19),
// Reverberate (2010-08-15), Dualcaster Mage (2016-06-08),
// Cackling Counterpart (2024-11-08, 2025-01-24), Fated Infatuation (2013-09-15, 2014-02-01).
// Official CR 2026-09-25: 603.3/6a, 608.2b/h, 701.22, 702.8/27/34, 707.2/10.
// Reiterate remains wholly unregistered. The later bounded #170 Engine
// prerequisite is exercised independently in buyback_test.go.
type authoredSpellCopyCase struct {
	entry    cardset.Entry
	color    mana.Color
	creature bool
}

func authoredSpellCopyCases() []authoredSpellCopyCase {
	return []authoredSpellCopyCase{
		{entry: authored.Twincast, color: mana.U},
		{entry: authored.Reverberate, color: mana.R},
		{entry: authored.DualcasterMage, color: mana.R, creature: true},
	}
}

type authoredCopyChoiceAgent struct {
	choose func(game.ChoiceRequest) []int
}

func (*authoredCopyChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *authoredCopyChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return a.choose(request)
}

func castAuthoredCopy(t *testing.T, g *game.Game, engine *Engine, tc authoredSpellCopyCase, original *game.StackObject) id.ID {
	t.Helper()
	spell := addCardToHand(g, game.Player1, tc.entry.New())
	g.Players[game.Player1].ManaPool.Add(tc.color, 2)
	targets := []game.Target{game.StackObjectTarget(original.ID)}
	if tc.creature {
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		targets = nil
	}
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil)) {
		t.Fatalf("%s real cast failed", tc.entry.Name)
	}
	if tc.creature {
		engine.resolveTopOfStack(g, &TurnLog{})
		if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("Dualcaster Mage's real entry did not trigger")
		}
		trigger, _ := g.Stack.Peek()
		if trigger.Kind != game.StackTriggeredAbility || len(trigger.Targets) != 1 ||
			trigger.Targets[0].StackObjectID != original.ID {
			t.Fatal("Dualcaster Mage's trigger did not target the original spell")
		}
	}
	return spell
}

func castModalCopySubject(t *testing.T, g *game.Game, engine *Engine, targets []game.Target) *game.StackObject {
	t.Helper()
	x := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	spec := game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Modal X subject", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.X}),
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1, MaxModes: 1,
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(99)}}}},
				{Targets: []game.TargetSpec{spec, spec}, Sequence: []game.Instruction{
					{Primitive: game.Damage{Amount: x, Recipient: game.AnyTargetDamageRecipient(0)}},
					{Primitive: game.Damage{Amount: x, Recipient: game.AnyTargetDamageRecipient(1)}},
				}},
			},
		}),
	}}
	spell := addCardToHand(g, game.Player2, def)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(spell, targets, 3, []int{1})) {
		t.Fatal("real modal X cast failed")
	}
	original, _ := g.Stack.Peek()
	return original
}

func TestAuthoredSpellCopiesChoicesAndCasting(t *testing.T) {
	t.Parallel()
	for _, tc := range authoredSpellCopyCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, scenario := range []string{"keep", "retarget", "keep illegal slot"} {
				t.Run(scenario, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := choicesProgramEngine(t)
					a := addCombatPermanent(g, game.Player3, vanillaCreature("A", 6, 6))
					b := addCombatPermanent(g, game.Player3, vanillaCreature("B", 6, 6))
					c := addCombatPermanent(g, game.Player3, vanillaCreature("C", 6, 6))
					original := castModalCopySubject(t, g, engine, []game.Target{game.PermanentTarget(a.ObjectID), game.PermanentTarget(b.ObjectID)})
					castAuthoredCopy(t, g, engine, tc, original)
					want := []id.ID{a.ObjectID, b.ObjectID}
					if scenario != "keep" {
						want[1] = c.ObjectID
					}
					if scenario == "keep illegal slot" {
						movePermanentToZone(g, a, zone.Exile)
					}
					agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
						for _, option := range request.Options {
							if scenario == "keep" && option.Label == "Keep current targets" {
								return []int{option.Index}
							}
							if scenario != "keep" && len(option.Targets) == 2 &&
								option.Targets[0].PermanentID == want[0] && option.Targets[1].PermanentID == want[1] {
								return []int{option.Index}
							}
						}
						t.Fatalf("desired retarget choice missing: %+v", request)
						return nil
					}}
					log := TurnLog{}
					engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log)
					copyObj, _ := g.Stack.Peek()
					if !copyObj.Copy || copyObj.Controller != game.Player1 || !copyObj.CopyOwner.Exists ||
						copyObj.CopyOwner.Val != game.Player1 || copyObj.XValue != 3 ||
						!slices.Equal(copyObj.ChosenModes, []int{1}) ||
						len(copyObj.TargetCounts) != 0 && !slices.Equal(copyObj.TargetCounts, []int{1, 1}) ||
						copyObj.Targets[0].PermanentID != want[0] || copyObj.Targets[1].PermanentID != want[1] {
						t.Fatalf("copy lost its controller, targets, modes or X: %+v", copyObj)
					}
					if original.Targets[0].PermanentID != a.ObjectID || original.Targets[1].PermanentID != b.ObjectID ||
						len(log.Choices) != 1 || log.Choices[0].UsedFallback ||
						countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
						countEventsOfKind(g.Events, game.EventSpellCopied) != 1 {
						t.Fatal("original mutated, choice bypassed, or copy treated as cast")
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					if scenario == "keep" && (a.MarkedDamage != 3 || b.MarkedDamage != 3 || c.MarkedDamage != 0) ||
						scenario != "keep" && (b.MarkedDamage != 0 || c.MarkedDamage != 3) ||
						scenario == "retarget" && a.MarkedDamage != 3 {
						t.Fatal("copied selected mode/X/targets did not execute correctly")
					}
					if g.Players[game.Player1].Life != 40 || g.Stack.Size() != 1 {
						t.Fatal("unchosen mode ran or original left the stack")
					}
				})
			}
		})
	}
}

func TestAuthoredSpellCopiesIllegalTarget(t *testing.T) {
	t.Parallel()
	for _, tc := range authoredSpellCopyCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			addImplementationSpellToStack(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Subject", Types: []types.Card{types.Instant}}}, nil)
			original, _ := g.Stack.Peek()
			spell := castAuthoredCopy(t, g, engine, tc, original)
			effect := &game.StackObject{Targets: []game.Target{game.StackObjectTarget(original.ID)}}
			if !counterTargetStackObject(g, effect, 0, false, game.CounteredSpellGraveyard) {
				t.Fatal("failed to remove original spell")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Stack.IsEmpty() || countEventsOfKind(g.Events, game.EventSpellCopied) != 0 {
				t.Fatal("copied a departed target")
			}
			if !tc.creature && !g.Players[game.Player1].Graveyard.Contains(spell) {
				t.Fatal("all-illegal instant did not go to graveyard")
			}
		})
	}
}

func TestAuthoredSpellCopiesUntargeted(t *testing.T) {
	t.Parallel()
	for _, tc := range authoredSpellCopyCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Untargeted sorcery", Types: []types.Card{types.Sorcery},
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
				}}}.Ability()),
			}}
			spell := addCardToHand(g, game.Player2, def)
			setMainPhasePriority(g, game.Player2)
			if !engine.applyAction(g, game.Player2, action.CastSpell(spell, nil, 0, nil)) {
				t.Fatal("untargeted sorcery cast failed")
			}
			original, _ := g.Stack.Peek()
			castAuthoredCopy(t, g, engine, tc, original)
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Life != 43 || g.Players[game.Player2].Life != 43 {
				t.Fatal("untargeted copy or copy-relative 'you' failed")
			}
		})
	}
}

func TestAuthoredSpellCopiesTargetKinds(t *testing.T) {
	t.Parallel()
	for _, tc := range authoredSpellCopyCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, kind := range []string{"instant", "sorcery", "creature spell", "activated ability"} {
				t.Run(kind, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := choicesProgramEngine(t)
					cardType := types.Creature
					switch kind {
					case "instant":
						cardType = types.Instant
					case "sorcery":
						cardType = types.Sorcery
					default:
					}
					addImplementationSpellToStack(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
						Name: "Target kind", Types: []types.Card{cardType},
					}}, nil)
					original, _ := g.Stack.Peek()
					if kind == "activated ability" {
						original.Kind = game.StackActivatedAbility
						original.InlineActivated = &game.ActivatedAbility{Content: game.Mode{}.Ability()}
					}
					spell := addCardToHand(g, game.Player1, tc.entry.New())
					g.Players[game.Player1].ManaPool.Add(tc.color, 2)
					g.Players[game.Player1].ManaPool.Add(mana.C, 1)
					g.Turn.PriorityPlayer = game.Player1
					targets := []game.Target{game.StackObjectTarget(original.ID)}
					if tc.creature {
						targets = nil
					}
					cast := engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil))
					eligible := kind == "instant" || kind == "sorcery"
					switch {
					case tc.creature:
						if !cast {
							t.Fatal("Dualcaster Mage cast failed")
						}
						engine.resolveTopOfStack(g, &TurnLog{})
						placed := engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
						if placed != eligible {
							t.Fatal("Dualcaster trigger accepted the wrong stack-object kind")
						}
					case cast != eligible:
						t.Fatal("copy instant accepted the wrong stack-object kind")
					case !cast && (!g.Players[game.Player1].Hand.Contains(spell) || g.Stack.Size() != 1):
						t.Fatal("illegal copy cast changed zones")
					default:
					}
				})
			}
		})
	}
}

func TestAuthoredDualcasterMageFlashAndSourceDeparture(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	def := &game.CardDef{CardFace: game.CardFace{Name: "Instant subject", Types: []types.Card{types.Instant}}}
	addImplementationSpellToStack(g, game.Player2, def, nil)
	original, _ := g.Stack.Peek()
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	card := castAuthoredCopy(t, g, engine, authoredSpellCopyCases()[2], original)
	mage := permanentByCardID(g, card)
	if mage == nil || effectivePower(g, mage) != 2 {
		t.Fatal("flash creature did not enter with its printed characteristics")
	}
	movePermanentToZone(g, mage, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy || copyObj.Controller != game.Player1 || !g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("entry trigger depended on its departed source")
	}
}

func castCreatureCopy(t *testing.T, g *game.Game, engine *Engine, entry cardset.Entry, target *game.Permanent) id.ID {
	t.Helper()
	spell := addCardToHand(g, game.Player1, entry.New())
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatalf("%s real cast failed", entry.Name)
	}
	return spell
}

func TestAuthoredCreatureCopyTokens(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.CacklingCounterpart, authored.FatedInfatuation} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, sourceKind := range []string{"printed", "copy effect", "token", "face down"} {
				t.Run(sourceKind, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := choicesProgramEngine(t)
					g.Turn.ActivePlayer = game.Player2
					def := vanillaCreature("Flying subject", 2, 3)
					def.ManaCost = opt.Val(cost.Mana{cost.X, cost.G})
					def.Colors = []color.Color{color.Green}
					def.Subtypes = []types.Sub{types.Bird}
					def.StaticAbilities = []game.StaticAbility{game.FlyingStaticBody}
					def.ReplacementAbilities = []game.ReplacementAbility{
						game.EntersWithCountersReplacement("This creature enters with one +1/+1 counter.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
					}
					def.TriggeredAbilities = []game.TriggeredAbility{{
						Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf}},
						Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}}.Ability(),
					}}
					target := addCombatPermanent(g, game.Player1, def)
					switch sourceKind {
					case "copy effect":
						original := target
						target = addCombatPermanent(g, game.Player1, vanillaCreature("Clone subject", 1, 1))
						obj := &game.StackObject{SourceID: target.ObjectID, Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(original.ObjectID)}}
						r := newEffectResolver(engine, g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
						if !handleBecomeCopy(r, game.BecomeCopy{Object: game.TargetPermanentReference(0)}).succeeded {
							t.Fatal("copy setup failed")
						}
					case "token":
						addEffectSpellToStack(g, game.Player1, game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(def)}, nil)
						engine.resolveTopOfStack(g, &TurnLog{})
						target = g.Battlefield[len(g.Battlefield)-1]
						engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
						engine.resolveTopOfStack(g, &TurnLog{})
					case "face down":
						target.FaceDown = true
					default:
					}
					target.Tapped = true
					target.MarkedDamage = 1
					target.Counters.Add(counter.PlusOnePlusOne, 4)
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), Layer: game.LayerPowerToughnessSet, Duration: game.DurationPermanent,
						AffectedObjectID: target.ObjectID,
						SetPower:         opt.Val(game.PT{Value: 9}), SetToughness: opt.Val(game.PT{Value: 9}),
					})
					before := len(g.Battlefield)
					lifeBefore := g.Players[game.Player1].Life
					castCreatureCopy(t, g, engine, entry, target)
					engine.resolveTopOfStack(g, &TurnLog{})
					if len(g.Battlefield) != before+1 {
						t.Fatal("named spell did not create exactly one copy token")
					}
					token := g.Battlefield[len(g.Battlefield)-1]
					if !token.Token || token.Controller != game.Player1 || token.Owner != game.Player1 ||
						token.Tapped || token.MarkedDamage != 0 || !token.SummoningSick {
						t.Fatal("token copied noncopiable status or has wrong controller")
					}
					if sourceKind == "face down" {
						if effectivePower(g, token) != 2 || token.TokenDef.ManaCost.Exists ||
							len(token.TokenDef.StaticAbilities) != 0 || token.Counters.Get(counter.PlusOnePlusOne) != 0 {
							t.Fatal("face-down creature's concealed characteristics leaked into the copy")
						}
						return
					}
					value, valueKnown := effectivePermanentManaValue(g, token)
					if token.TokenDef.Name != def.Name || token.TokenDef.Power.Val.Value != 2 ||
						token.TokenDef.Toughness.Val.Value != 3 || !slices.Equal(token.TokenDef.Colors, def.Colors) ||
						!slices.Equal(token.TokenDef.Subtypes, def.Subtypes) || !valueKnown || value != 1 ||
						!hasKeyword(g, token, game.Flying) || token.Counters.Get(counter.PlusOnePlusOne) != 1 ||
						effectivePower(g, token) != 3 {
						t.Fatal("token failed to snapshot copiable values or apply its own entry replacement")
					}
					if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
						t.Fatal("copied entry ability did not trigger")
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					if g.Players[game.Player1].Life != lifeBefore+1 {
						t.Fatal("copied triggered ability did not execute")
					}
				})
			}
		})
	}
}

func TestAuthoredCreatureCopyTargetLegality(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.CacklingCounterpart, authored.FatedInfatuation} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, illegal := range []string{"opponent creature", "own artifact"} {
				t.Run(illegal, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := choicesProgramEngine(t)
					def := vanillaCreature("Illegal target", 2, 2)
					controller := game.Player2
					if illegal == "own artifact" {
						controller = game.Player1
						def.Types = []types.Card{types.Artifact}
					}
					target := addCombatPermanent(g, controller, def)
					card := addCardToHand(g, game.Player1, entry.New())
					g.Players[game.Player1].ManaPool.Add(mana.U, 3)
					g.Players[game.Player1].ManaPool.Add(mana.C, 1)
					g.Turn.PriorityPlayer = game.Player1
					if engine.applyAction(g, game.Player1, action.CastSpell(card, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) ||
						!g.Players[game.Player1].Hand.Contains(card) || !g.Stack.IsEmpty() {
						t.Fatal("named token-copy spell accepted an illegal announced target")
					}
				})
			}
			for _, change := range []string{"departure", "control", "type"} {
				t.Run(change, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := choicesProgramEngine(t)
					target := addCombatPermanent(g, game.Player1, vanillaCreature("Target", 2, 2))
					addCardToLibrary(g, game.Player1, vanillaCreature("Unseen", 2, 2))
					castCreatureCopy(t, g, engine, entry, target)
					switch change {
					case "departure":
						movePermanentToZone(g, target, zone.Exile)
					case "control":
						target.Controller = game.Player2
					default:
						g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
							ID: g.IDGen.Next(), Layer: game.LayerType, Duration: game.DurationPermanent,
							AffectedObjectID: target.ObjectID, RemoveTypes: []types.Card{types.Creature},
						})
					}
					log := TurnLog{}
					engine.resolveTopOfStack(g, &log)
					if countEventsOfKind(g.Events, game.EventTokenCreated) != 0 ||
						countEventsOfKind(g.Events, game.EventScry) != 0 || len(log.Choices) != 0 {
						t.Fatal("illegal target produced a token or scry")
					}
				})
			}
		})
	}
}

func TestAuthoredCreatureCopyRetargetedSpellAndClone(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.CacklingCounterpart, authored.FatedInfatuation} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			g.Turn.ActivePlayer = game.Player2
			a := addCombatPermanent(g, game.Player1, vanillaCreature("A", 2, 3))
			b := addCombatPermanent(g, game.Player1, vanillaCreature("B", 4, 5))
			castCreatureCopy(t, g, engine, entry, a)
			original, _ := g.Stack.Peek()
			castAuthoredCopy(t, g, engine, authoredSpellCopyCases()[0], original)
			clone := g.Clone()
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				for _, option := range request.Options {
					if len(option.Targets) == 1 && option.Targets[0].PermanentID == b.ObjectID {
						return []int{option.Index}
					}
				}
				t.Fatal("copy did not offer legal new controlled-creature target")
				return nil
			}}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.resolveTopOfStack(clone, &TurnLog{})
			engine.resolveTopOfStack(clone, &TurnLog{})
			if g.Battlefield[len(g.Battlefield)-1].TokenDef.Name != "B" ||
				clone.Battlefield[len(clone.Battlefield)-1].TokenDef.Name != "A" ||
				original.Targets[0].PermanentID != a.ObjectID {
				t.Fatal("authored body ignored copy retargeting or leaked state across clones")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Battlefield[len(g.Battlefield)-1].TokenDef.Name != "A" {
				t.Fatal("copy changed the original spell's target")
			}
		})
	}
}

func TestAuthoredCacklingCounterpartFlashback(t *testing.T) {
	t.Parallel()
	for _, ending := range []string{"resolve", "counter", "illegal target"} {
		t.Run(ending, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			target := addCombatPermanent(g, game.Player1, vanillaCreature("Subject", 2, 2))
			card := castCreatureCopy(t, g, engine, authored.CacklingCounterpart, target)
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Graveyard.Contains(card) {
				t.Fatal("normal cast did not go to the graveyard")
			}
			g.Players[game.Player1].ManaPool.Empty()
			g.Players[game.Player1].ManaPool.Add(mana.U, 2)
			g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			g.Turn.ActivePlayer = game.Player2
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareBlockers
			g.Turn.PriorityPlayer = game.Player1
			cast := action.CastSpellFromZone(card, zone.Graveyard, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
			if engine.applyAction(g, game.Player1, cast) {
				t.Fatal("flashback incorrectly paid only four generic mana")
			}
			if !g.Players[game.Player1].Graveyard.Contains(card) || !g.Stack.IsEmpty() {
				t.Fatal("failed flashback payment changed zones")
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			if !engine.applyAction(g, game.Player1, cast) {
				t.Fatal("instant flashback could not be cast on opponent's turn")
			}
			spell, _ := g.Stack.Peek()
			if !spell.Flashback || spell.SourceZone != zone.Graveyard {
				t.Fatal("flashback announcement lost its source or mechanic")
			}
			if ending == "counter" {
				obj := &game.StackObject{Targets: []game.Target{game.StackObjectTarget(spell.ID)}}
				if !counterTargetStackObject(g, obj, 0, false, game.CounteredSpellGraveyard) {
					t.Fatal("countering flashback failed")
				}
			} else {
				if ending == "illegal target" {
					movePermanentToZone(g, target, zone.Exile)
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			wantTokens := 1
			if ending == "resolve" {
				wantTokens++
			}
			if !g.Players[game.Player1].Exile.Contains(card) || g.Players[game.Player1].Graveyard.Contains(card) ||
				countEventsOfKind(g.Events, game.EventTokenCreated) != wantTokens {
				t.Fatal("flashback failed to exile or created an incorrect number of copies")
			}
		})
	}
}

func TestAuthoredFatedInfatuationScry(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"opponent turn", "split", "reverse top", "reverse bottom", "empty library"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			target := addCombatPermanent(g, game.Player1, vanillaCreature("Subject", 2, 2))
			var bottom, second, top id.ID
			if scenario != "empty library" {
				bottom = addCardToLibrary(g, game.Player1, vanillaCreature("Bottom", 2, 2))
				second = addCardToLibrary(g, game.Player1, vanillaCreature("Second", 2, 2))
				top = addCardToLibrary(g, game.Player1, vanillaCreature("Top", 2, 2))
			}
			if scenario == "opponent turn" {
				g.Turn.ActivePlayer = game.Player2
			}
			castCreatureCopy(t, g, engine, authored.FatedInfatuation, target)
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if countEventsOfKind(g.Events, game.EventTokenCreated) != 1 {
					t.Fatal("scry happened before creating the token")
				}
				if request.Kind == game.ChoiceOrder {
					return []int{1, 0}
				}
				if request.Kind != game.ChoiceScry || len(request.Options) != 2 ||
					!request.Options[0].Card.Exists || !request.Options[1].Card.Exists {
					t.Fatalf("scry did not disclose both looked-at cards: %+v", request)
				}
				if scenario == "split" {
					return []int{0}
				}
				if scenario == "reverse bottom" {
					return []int{0, 1}
				}
				return nil
			}}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log)
			want := []id.ID{top, second, bottom}
			switch scenario {
			case "split":
				want = []id.ID{second, bottom, top}
			case "reverse top":
				want = []id.ID{second, top, bottom}
			case "reverse bottom":
				want = []id.ID{bottom, second, top}
			case "empty library":
				want = nil
			default:
			}
			if !slices.Equal(g.Players[game.Player1].Library.All(), want) {
				t.Fatalf("scry library = %v, want %v", g.Players[game.Player1].Library.All(), want)
			}
			scryCount := 1
			if scenario == "opponent turn" {
				scryCount = 0
			}
			if countEventsOfKind(g.Events, game.EventScry) != scryCount ||
				countEventsOfKind(g.Events, game.EventTokenCreated) != 1 {
				t.Fatal("scry turn condition or token creation failed")
			}
			for _, choice := range log.Choices {
				if choice.UsedFallback {
					t.Fatal("valid scry choice was ignored")
				}
			}
		})
	}
}
