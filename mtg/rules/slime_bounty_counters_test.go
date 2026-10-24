package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Proving scenarios for the slime and bounty marker counters (#2283). Both
// fixtures are unregistered test declarations, not card admissions.
//
// Sources: supplied corpus (SHA256 6da9ced0…a1b4); Scryfall rulings for
// revisions a32be366-d86e-48aa-8257-ca8afb87ba18 (two 2021-09-24 rulings:
// becoming a Horror ends the effect while other P/T modifiers still apply, and
// slime counters from any source count) and
// 5db59164-f0b1-4b5a-b821-ad0b2e1612d1 (no rulings); official CR effective
// 2026-09-25: 122.1-122.2 (counters are markers that cease to exist on a zone
// change), 400.7, 603.2, 603.10a, 608.2b and 613.1/613.4-613.8.

const (
	sludgeMonsterTriggerText = "Whenever this creature enters or attacks, put a slime counter on up to one other target creature."
	sludgeMonsterStaticText  = "Non-Horror creatures with slime counters on them lose all abilities and have base power and toughness 2/2."
	terminationActivatedText = "{T}: Put a bounty counter on target creature or planeswalker. Activate only as a sorcery."
	terminationTriggerText   = "Whenever a creature or planeswalker an opponent controls with a bounty counter on it is dealt damage, destroy it."
)

// sludgeMonsterFixture mirrors Sludge Monster (oracle
// 2802fc92-49a0-43c8-bc10-24882eeee3f1, revision
// a32be366-d86e-48aa-8257-ca8afb87ba18): {3}{U}{U}, 5/5 Horror.
func sludgeMonsterFixture() *game.CardDef {
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes:   []types.Card{types.Creature},
		ExcludedSubtype: types.Horror,
		MatchCounter:    true,
		RequiredCounter: counter.Slime,
	})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Sludge Monster",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U, cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Horror},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			OracleText: sludgeMonsterTriggerText + "\n" + sludgeMonsterStaticText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: sludgeMonsterTriggerText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventPermanentEnteredBattlefield,
						Source:     game.TriggerSourceSelf,
						UnionEvent: game.EventAttackerDeclared,
					},
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one other target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
					}},
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Slime,
					}}},
				}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{{
				Text:           sludgeMonsterStaticText,
				ZoneOfFunction: zone.Battlefield,
				ContinuousEffects: []game.ContinuousEffect{
					{Layer: game.LayerAbility, Group: group, RemoveAllAbilities: true},
					{Layer: game.LayerPowerToughnessSet, Group: group, SetPower: opt.Val(game.PT{Value: 2}), SetToughness: opt.Val(game.PT{Value: 2})},
				},
			}},
		},
	}
}

// terminationFacilitatorFixture mirrors Termination Facilitator (oracle
// 56edd306-8676-49c4-b3c3-5ff840536c3a, revision
// 5db59164-f0b1-4b5a-b821-ad0b2e1612d1): {1}{B}, 1/3 Human Assassin. The
// printed trigger has no intervening "if" clause, so none is declared.
func terminationFacilitatorFixture() *game.CardDef {
	creatureOrPlaneswalker := []types.Card{types.Creature, types.Planeswalker}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Termination Facilitator",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: terminationActivatedText + "\n" + terminationTriggerText,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:            terminationActivatedText,
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
				ZoneOfFunction:  zone.Battlefield,
				Timing:          game.SorceryOnly,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: creatureOrPlaneswalker}),
					}},
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Bounty,
					}}},
				}.Ability(),
			}},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: terminationTriggerText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventDamageDealt,
						Controller: game.TriggerControllerOpponent,
						SubjectSelection: game.Selection{
							RequiredTypesAny: creatureOrPlaneswalker,
							MatchCounter:     true,
							RequiredCounter:  counter.Bounty,
						},
						DamageRecipient: game.DamageRecipientPermanent,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Destroy{
					Object: game.EventPermanentReference(),
				}}}}.Ability(),
			}},
		},
	}
}

func TestSlimeBountyFixturesValidate(t *testing.T) {
	for _, def := range []*game.CardDef{sludgeMonsterFixture(), terminationFacilitatorFixture()} {
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("%s fixture issues: %+v", def.Name, issues)
		}
	}
}

// slimeTargetAgent picks the target option matching want exactly (an empty
// want selects the zero-target option) and records each offered option set.
type slimeTargetAgent struct {
	want    []game.Target
	offered [][]game.Target
}

func (*slimeTargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *slimeTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		a.offered = append(a.offered, option.Targets)
	}
	for _, option := range request.Options {
		if slices.Equal(option.Targets, a.want) {
			return []int{option.Index}
		}
	}
	return nil
}

func slimeVanilla(name string, pt int, subtypes ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, Subtypes: subtypes,
		Power: opt.Val(game.PT{Value: pt}), Toughness: opt.Val(game.PT{Value: pt}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}}
}

func assertSlimeValues(t *testing.T, g *game.Game, p *game.Permanent, power, toughness int, flying bool) {
	t.Helper()
	values := effectivePermanentValues(g, p)
	if values.power != power || values.toughness != toughness || hasKeyword(g, p, game.Flying) != flying {
		t.Fatalf("%s: P/T %d/%d flying=%v, want %d/%d flying=%v",
			values.name, values.power, values.toughness, hasKeyword(g, p, game.Flying), power, toughness, flying)
	}
}

func castSludgeMonster(t *testing.T, g *game.Game, e *Engine) *game.Permanent {
	t.Helper()
	card := addCardToHand(g, game.Player1, sludgeMonsterFixture())
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if !e.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("Sludge Monster fixture could not be cast")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	monster := permanentByCardID(g, card)
	if monster == nil {
		t.Fatal("Sludge Monster fixture did not resolve onto the battlefield")
	}
	return monster
}

func countersAddedEvents(g *game.Game, kind counter.Kind) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCountersAdded && event.CounterKind == kind {
			events = append(events, event)
		}
	}
	return events
}

// TestSlimeAndBountyPlacementRemovalAndProliferate places each marker through
// the ordinary resolving AddCounter operation, with a kind-filtered placement
// replacement, then removes and proliferates them. Neither marker changes
// characteristics or causes destruction by itself.
func TestSlimeAndBountyPlacementRemovalAndProliferate(t *testing.T) {
	g, e := setupBestowMain(t)
	recipient := addCombatPermanent(g, game.Player2, slimeVanilla("Marked creature", 3))
	bystander := addCombatPermanent(g, game.Player2, slimeVanilla("Bystander", 3))
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventCountersAdded, MatchCounterKind: true,
		CounterKindFilter: counter.Slime, CounterMultiplier: 2, Duration: game.DurationPermanent,
	})
	targets := []game.Target{game.PermanentTarget(recipient.ObjectID)}

	addEffectSpellToStack(g, game.Player3, game.AddCounter{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Slime}, targets)
	e.resolveTopOfStack(g, &TurnLog{})
	addEffectSpellToStack(g, game.Player3, game.AddCounter{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Bounty}, targets)
	e.resolveTopOfStack(g, &TurnLog{})

	if recipient.Counters.Get(counter.Slime) != 2 || recipient.Counters.Get(counter.Bounty) != 1 || !bystander.Counters.IsEmpty() {
		t.Fatalf("recipient counters = %v, bystander = %v; want doubled slime and one bounty on the recipient only",
			recipient.Counters.All(), bystander.Counters.All())
	}
	for _, check := range []struct {
		kind   counter.Kind
		name   string
		amount int
	}{{counter.Slime, "slime", 2}, {counter.Bounty, "bounty", 1}} {
		events := countersAddedEvents(g, check.kind)
		if len(events) != 1 {
			t.Fatalf("%s placement events = %d, want 1", check.name, len(events))
		}
		event := events[0]
		if event.PermanentID != recipient.ObjectID || event.Amount != check.amount || event.PreviousCounterAmount != 0 ||
			event.Controller != game.Player3 || event.CounterKind.String() != check.name {
			t.Fatalf("%s placement event = %+v; want recipient, replacement-adjusted amount %d and actor Player3", check.name, event, check.amount)
		}
	}
	e.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, recipient.ObjectID); !ok {
		t.Fatal("marker counters destroyed the permanent without any declaration reading them")
	}
	assertSlimeValues(t, g, recipient, 3, 3, true)

	addEffectSpellToStack(g, game.Player1, game.RemoveCounter{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Slime}, targets)
	e.resolveTopOfStack(g, &TurnLog{})
	if recipient.Counters.Get(counter.Slime) != 1 || recipient.Counters.Get(counter.Bounty) != 1 {
		t.Fatalf("after removal counters = %v; want one slime removed and bounty untouched", recipient.Counters.All())
	}

	var labels []string
	recorder := &recordingProliferateAgent{pick: "bounty", labels: &labels}
	addEffectSpellToStack(g, game.Player1, game.Proliferate{}, nil)
	log := TurnLog{}
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: recorder}, &log)
	if !slices.Equal(labels, []string{"bounty", "slime"}) {
		t.Fatalf("proliferate options = %v; want public slime and bounty kinds", labels)
	}
	if recipient.Counters.Get(counter.Bounty) != 2 || recipient.Counters.Get(counter.Slime) != 1 {
		t.Fatalf("after proliferating bounty counters = %v; want bounty 2 and slime unchanged", recipient.Counters.All())
	}
	recorder.pick = "slime"
	addEffectSpellToStack(g, game.Player1, game.Proliferate{}, nil)
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: recorder}, &log)
	if recipient.Counters.Get(counter.Slime) != 3 {
		t.Fatalf("proliferated slime = %d; want slime replacement to double the proliferated counter", recipient.Counters.Get(counter.Slime))
	}
	if last := countersAddedEvents(g, counter.Slime); last[len(last)-1].Controller != game.Player1 || last[len(last)-1].Amount != 2 {
		t.Fatalf("proliferated slime event = %+v; want actor Player1 and doubled amount", last[len(last)-1])
	}
}

type recordingProliferateAgent struct {
	pick   string
	labels *[]string
}

func (*recordingProliferateAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *recordingProliferateAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	*a.labels = (*a.labels)[:0]
	chosen := -1
	for _, option := range request.Options {
		*a.labels = append(*a.labels, option.Label)
		if option.Label == a.pick {
			chosen = option.Index
		}
	}
	return []int{chosen}
}

// TestSludgeMonsterEntersOrAttacksTargetsUpToOneOtherCreature proves the
// trigger offers zero or one other creature, never the Monster itself, and
// places slime on the chosen creature for both enter and attack events.
func TestSludgeMonsterEntersOrAttacksTargetsUpToOneOtherCreature(t *testing.T) {
	for _, test := range []struct {
		name   string
		choose string
	}{{"one target", "opponent"}, {"zero targets", "none"}} {
		t.Run(test.name, func(t *testing.T) {
			g, e := setupBestowMain(t)
			opponent := addCombatPermanent(g, game.Player2, slimeVanilla("Opponent flyer", 4))
			own := addCombatPermanent(g, game.Player1, slimeVanilla("Own flyer", 4))
			agent := &slimeTargetAgent{}
			if test.choose == "opponent" {
				agent.want = []game.Target{game.PermanentTarget(opponent.ObjectID)}
			}
			monster := castSludgeMonster(t, g, e)
			log := TurnLog{}
			if !e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log) {
				t.Fatal("Sludge Monster's enter trigger did not trigger")
			}
			sawEmpty := false
			for _, targets := range agent.offered {
				if len(targets) == 0 {
					sawEmpty = true
				}
				if slices.Contains(targets, game.PermanentTarget(monster.ObjectID)) {
					t.Fatal("Sludge Monster offered itself as an other target creature")
				}
			}
			if !sawEmpty || len(agent.offered) != 3 {
				t.Fatalf("offered targets = %v; want zero, opponent and own creature options", agent.offered)
			}
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log)
			if test.choose == "none" {
				if len(countersAddedEvents(g, counter.Slime)) != 0 || !opponent.Counters.IsEmpty() || !own.Counters.IsEmpty() {
					t.Fatal("zero-target trigger placed a slime counter")
				}
				return
			}
			if opponent.Counters.Get(counter.Slime) != 1 || monster.Counters.Has(counter.Slime) || own.Counters.Has(counter.Slime) {
				t.Fatal("enter trigger did not place exactly one slime counter on the chosen creature")
			}
			if event := countersAddedEvents(g, counter.Slime)[0]; event.Controller != game.Player1 || event.PermanentID != opponent.ObjectID || event.Amount != 1 {
				t.Fatalf("slime placement event = %+v; want one slime on the target placed by Sludge Monster's controller", event)
			}
			assertSlimeValues(t, g, opponent, 2, 2, false)
			assertSlimeValues(t, g, own, 4, 4, true)
			assertSlimeValues(t, g, monster, 5, 5, false)

			monster.SummoningSick = false
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareAttackers
			g.Combat = &game.CombatState{}
			attack, ok := action.DeclareAttackers([]game.AttackDeclaration{{
				Attacker: monster.ObjectID, Target: game.AttackTarget{Player: game.Player2},
			}}).DeclareAttackersPayload()
			if !ok || !e.applyDeclareAttackers(g, game.Player1, attack) {
				t.Fatal("Sludge Monster could not attack")
			}
			agent.want = []game.Target{game.PermanentTarget(own.ObjectID)}
			if !e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log) {
				t.Fatal("Sludge Monster's attack trigger did not trigger")
			}
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log)
			if own.Counters.Get(counter.Slime) != 1 || opponent.Counters.Get(counter.Slime) != 1 {
				t.Fatal("attack trigger did not place slime on the newly chosen creature")
			}
			assertSlimeValues(t, g, own, 2, 2, false)
		})
	}
}

// TestSludgeMonsterStaticHorrorStatusModifiersAndDeparture follows the 2021
// rulings: slime from any source counts, Horror status ends the effect, other
// P/T modifiers still apply, later grants work, and the source leaving restores
// values while the counters remain.
func TestSludgeMonsterStaticHorrorStatusModifiersAndDeparture(t *testing.T) {
	g, e := setupBestowMain(t)
	monster := addCombatPermanent(g, game.Player1, sludgeMonsterFixture())
	victim := addCombatPermanent(g, game.Player2, slimeVanilla("Pumped flyer", 4))
	horror := addCombatPermanent(g, game.Player2, slimeVanilla("Horror flyer", 4, types.Horror))
	unslimed := addCombatPermanent(g, game.Player2, slimeVanilla("Unslimed flyer", 4))
	victim.Counters.Add(counter.PlusOnePlusOne, 1)
	victim.TemporaryPowerModifier, victim.TemporaryToughnessModifier = 1, 1
	for _, p := range []*game.Permanent{victim, horror} {
		addEffectSpellToStack(g, game.Player3, game.AddCounter{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.Slime},
			[]game.Target{game.PermanentTarget(p.ObjectID)})
		e.resolveTopOfStack(g, &TurnLog{})
	}
	assertSlimeValues(t, g, victim, 4, 4, false)
	assertSlimeValues(t, g, horror, 4, 4, true)
	assertSlimeValues(t, g, unslimed, 4, 4, true)
	assertSlimeValues(t, g, monster, 5, 5, false)

	horrorEffect := game.ContinuousEffect{ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: victim.ObjectID, AddSubtypes: []types.Sub{types.Horror}}
	g.ContinuousEffects = append(g.ContinuousEffects, horrorEffect)
	assertSlimeValues(t, g, victim, 6, 6, true)
	g.ContinuousEffects = slices.DeleteFunc(g.ContinuousEffects, func(effect game.ContinuousEffect) bool { return effect.ID == horrorEffect.ID })
	assertSlimeValues(t, g, victim, 4, 4, false)

	aura := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Later flying Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility, Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
			AddKeywords: []game.Keyword{game.Flying},
		}}}},
	}})
	aura.AttachedTo = opt.Val(victim.ObjectID)
	assertSlimeValues(t, g, victim, 4, 4, true)
	movePermanentToZone(g, aura, zone.Graveyard)
	assertSlimeValues(t, g, victim, 4, 4, false)

	clone := g.Clone()
	if !movePermanentToZone(g, monster, zone.Graveyard) {
		t.Fatal("Sludge Monster could not leave the battlefield")
	}
	assertSlimeValues(t, g, victim, 6, 6, true)
	if victim.Counters.Get(counter.Slime) != 1 || horror.Counters.Get(counter.Slime) != 1 {
		t.Fatal("slime counters disappeared when only their generating source left")
	}
	clonedVictim, _ := permanentByObjectID(clone, victim.ObjectID)
	assertSlimeValues(t, clone, clonedVictim, 4, 4, false)
}

func addBountyTestPlaneswalker(g *game.Game, controller game.PlayerID, loyalty int) *game.Permanent {
	walker := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name: "Bounty walker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(loyalty),
	}})
	walker.Counters.Add(counter.Loyalty, loyalty)
	return walker
}

func activateFacilitator(t *testing.T, g *game.Game, e *Engine, facilitator, target *game.Permanent) bool {
	t.Helper()
	facilitator.Tapped = false
	return e.applyAction(g, facilitator.Controller, action.ActivateAbility(facilitator.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0))
}

// TestTerminationFacilitatorPlacesBountyAtSorcerySpeed proves the tap ability
// puts exactly one bounty counter on a creature or planeswalker with the
// activating player as the placement actor, and honors its timing restriction.
func TestTerminationFacilitatorPlacesBountyAtSorcerySpeed(t *testing.T) {
	g, e := setupBestowMain(t)
	g.Turn.ActivePlayer = game.Player1
	facilitator := addCombatPermanent(g, game.Player1, terminationFacilitatorFixture())
	creature := addCombatPermanent(g, game.Player2, slimeVanilla("Bountied creature", 4))
	walker := addBountyTestPlaneswalker(g, game.Player3, 4)
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Not a target", Types: []types.Card{types.Enchantment}}})

	for _, target := range []*game.Permanent{creature, walker} {
		if !activateFacilitator(t, g, e, facilitator, target) {
			t.Fatal("Termination Facilitator could not be activated at sorcery speed")
		}
		if !facilitator.Tapped {
			t.Fatal("activation did not pay the tap cost")
		}
		if e.applyAction(g, game.Player1, action.ActivateAbility(facilitator.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)) {
			t.Fatal("tapped Termination Facilitator activated again")
		}
		e.resolveTopOfStack(g, &TurnLog{})
		if target.Counters.Get(counter.Bounty) != 1 {
			t.Fatalf("bounty counters on %d = %d, want 1", target.ObjectID, target.Counters.Get(counter.Bounty))
		}
		events := countersAddedEvents(g, counter.Bounty)
		if event := events[len(events)-1]; event.PermanentID != target.ObjectID || event.Amount != 1 || event.Controller != game.Player1 {
			t.Fatalf("bounty event = %+v; want activator Player1 placing one counter", event)
		}
	}
	if activateFacilitator(t, g, e, facilitator, enchantment) {
		t.Fatal("bounty ability accepted a non-creature, non-planeswalker target")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	if activateFacilitator(t, g, e, facilitator, creature) {
		t.Fatal("sorcery-speed bounty ability activated during combat")
	}
	if creature.Counters.Get(counter.Bounty) != 1 || walker.Counters.Get(counter.Loyalty) != 4 {
		t.Fatal("bounty placement changed other characteristics")
	}
}

func damageAndTrigger(g *game.Game, e *Engine, source, recipient *game.Permanent, amount int) bool {
	dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, source.Controller, recipient, amount, false)
	e.applyStateBasedActions(g)
	return e.putTriggeredAbilitiesOnStack(g)
}

// TestTerminationFacilitatorDestroysDamagedOpponentBountyPermanents proves the
// damage trigger reads the bounty marker on opponent-controlled creatures and
// planeswalkers, and its negative controls.
func TestTerminationFacilitatorDestroysDamagedOpponentBountyPermanents(t *testing.T) {
	setup := func() (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
		g, e := setupBestowMain(t)
		facilitator := addCombatPermanent(g, game.Player1, terminationFacilitatorFixture())
		source := addCombatPermanent(g, game.Player4, slimeVanilla("Damage source", 1))
		return g, e, facilitator, source
	}
	t.Run("opponent creature and planeswalker", func(t *testing.T) {
		g, e, _, source := setup()
		creature := addCombatPermanent(g, game.Player2, slimeVanilla("Bountied creature", 4))
		walker := addBountyTestPlaneswalker(g, game.Player3, 4)
		for _, recipient := range []*game.Permanent{creature, walker} {
			recipient.Counters.Add(counter.Bounty, 1)
			if !damageAndTrigger(g, e, source, recipient, 1) {
				t.Fatalf("damage to bountied %d did not trigger", recipient.ObjectID)
			}
			e.resolveTopOfStack(g, &TurnLog{})
			if _, ok := permanentByObjectID(g, recipient.ObjectID); ok {
				t.Fatalf("bountied %d survived the destroy trigger", recipient.ObjectID)
			}
		}
	})
	for _, negative := range []struct {
		name  string
		setup func(g *game.Game) *game.Permanent
	}{
		{"wrong marker", func(g *game.Game) *game.Permanent {
			p := addCombatPermanent(g, game.Player2, slimeVanilla("Slimed creature", 4))
			p.Counters.Add(counter.Slime, 1)
			p.Counters.Add(counter.Ribbon, 1)
			return p
		}},
		{"no marker", func(g *game.Game) *game.Permanent {
			return addCombatPermanent(g, game.Player2, slimeVanilla("Clean creature", 4))
		}},
		{"own permanent", func(g *game.Game) *game.Permanent {
			p := addCombatPermanent(g, game.Player1, slimeVanilla("Own bountied creature", 4))
			p.Counters.Add(counter.Bounty, 1)
			return p
		}},
		{"prevented damage", func(g *game.Game) *game.Permanent {
			p := addCombatPermanent(g, game.Player2, slimeVanilla("Shielded bountied creature", 4))
			p.Counters.Add(counter.Bounty, 1)
			p.Counters.Add(counter.Shield, 1)
			return p
		}},
	} {
		t.Run(negative.name, func(t *testing.T) {
			g, e, _, source := setup()
			recipient := negative.setup(g)
			if damageAndTrigger(g, e, source, recipient, 1) {
				t.Fatal("Termination Facilitator triggered for a non-qualifying damage event")
			}
			if _, ok := permanentByObjectID(g, recipient.ObjectID); !ok {
				t.Fatal("non-qualifying recipient left the battlefield")
			}
		})
	}
}

// TestTerminationFacilitatorUsesDamagedObjectIdentity proves a recipient that
// dies from the damage still triggers (the trigger is matched when the damage
// is dealt, before state-based actions move it) and its "destroy it" affects no
// other object, and that "destroy it" never follows a recipient that left and
// returned as a new object without its counters. Departed-object marker reads
// are proven separately by TestSlimeBountyLiveAndDepartedSelection.
func TestTerminationFacilitatorUsesDamagedObjectIdentity(t *testing.T) {
	t.Run("dies from damage", func(t *testing.T) {
		g, e := setupBestowMain(t)
		addCombatPermanent(g, game.Player1, terminationFacilitatorFixture())
		source := addCombatPermanent(g, game.Player4, slimeVanilla("Damage source", 1))
		victim := addCombatPermanent(g, game.Player2, slimeVanilla("Fragile bountied creature", 2))
		victim.Counters.Add(counter.Bounty, 1)
		if !damageAndTrigger(g, e, source, victim, 3) {
			t.Fatal("lethal damage to a bountied creature did not trigger")
		}
		if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
			t.Fatal("lethally damaged creature survived state-based actions")
		}
		snapshot, ok := lastKnownObject(g, victim.ObjectID)
		if !ok || snapshot.Counters.Get(counter.Bounty) != 1 {
			t.Fatal("departed recipient's last-known bounty counter was not retained")
		}
		if cardCounters := g.CardCounters[victim.CardInstanceID]; !cardCounters.IsEmpty() {
			t.Fatalf("graveyard card retained counters %v", cardCounters.All())
		}
		before := len(g.Battlefield)
		e.resolveTopOfStack(g, &TurnLog{})
		if len(g.Battlefield) != before || !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
			t.Fatal("destroy trigger for a departed recipient affected another object")
		}
	})
	t.Run("leaves and returns", func(t *testing.T) {
		g, e := setupBestowMain(t)
		addCombatPermanent(g, game.Player1, terminationFacilitatorFixture())
		source := addCombatPermanent(g, game.Player4, slimeVanilla("Damage source", 1))
		victim := addCombatPermanent(g, game.Player2, slimeVanilla("Blinked bountied creature", 4))
		victim.Counters.Add(counter.Bounty, 1)
		if !damageAndTrigger(g, e, source, victim, 1) {
			t.Fatal("damage to a bountied creature did not trigger")
		}
		blink := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2,
			SourceID: addCardInstance(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Blink"}}),
			Targets:  []game.Target{game.PermanentTarget(victim.ObjectID)},
		}
		instructions := immediateBlinkInstructions(game.PutOnBattlefield{})
		for i := range instructions {
			e.resolveInstructionWithChoices(g, blink, &instructions[i], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}
		returned := permanentByCardID(g, victim.CardInstanceID)
		if returned == nil || returned.ObjectID == victim.ObjectID {
			t.Fatal("blinked recipient did not return as a new object")
		}
		if !returned.Counters.IsEmpty() {
			t.Fatalf("returned object kept counters %v across a zone change", returned.Counters.All())
		}
		e.resolveTopOfStack(g, &TurnLog{})
		if _, ok := permanentByObjectID(g, returned.ObjectID); !ok {
			t.Fatal("destroy trigger followed the recipient to its new object")
		}
		if damageAndTrigger(g, e, source, returned, 1) {
			t.Fatal("returned object without a bounty counter triggered")
		}
	})
}

// TestSlimeBountyLiveAndDepartedSelection proves exact-kind marker selection
// reads each marker from the live permanent and, after it leaves the
// battlefield, from its last-known snapshot (CR 608.2h), while the moved card
// itself keeps no counters (CR 122.2) and a returned new object does not
// inherit the departed object's markers (CR 400.7).
func TestSlimeBountyLiveAndDepartedSelection(t *testing.T) {
	for _, marker := range []counter.Kind{counter.Slime, counter.Bounty} {
		other := counter.Bounty
		if marker == counter.Bounty {
			other = counter.Slime
		}
		for _, ev := range []struct {
			name string
			kind game.EventKind
		}{{"damage", game.EventDamageDealt}, {"died", game.EventPermanentDied}} {
			eventKind := ev.kind
			t.Run(marker.String()+"/"+ev.name, func(t *testing.T) {
				g, _ := setupBestowMain(t)
				victim := addCombatPermanent(g, game.Player2, slimeVanilla("Marked creature", 4))
				victim.Counters.Add(marker, 2)
				event := game.Event{Kind: eventKind, PermanentID: victim.ObjectID, CardID: victim.CardInstanceID, Controller: game.Player2}
				matches := func(selection game.Selection) bool {
					return triggerSelectionMatches(g, game.Player1, event, victim.ObjectID, &selection, 0)
				}
				exact := game.Selection{MatchCounter: true, RequiredCounter: marker}
				wrong := game.Selection{MatchCounter: true, RequiredCounter: other}
				twoOrMore := game.Selection{RequiredCounter: marker, RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})}
				three := game.Selection{RequiredCounter: marker, RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}
				without := game.Selection{MatchExcludedCounter: true, ExcludedCounter: marker}
				withoutOther := game.Selection{MatchExcludedCounter: true, ExcludedCounter: other}
				check := func(phase string) {
					t.Helper()
					if !matches(exact) || !matches(twoOrMore) || !matches(withoutOther) {
						t.Fatalf("%s %s marker not matched", phase, marker)
					}
					if matches(wrong) || matches(three) || matches(without) {
						t.Fatalf("%s %s marker confused with another kind or count", phase, marker)
					}
				}
				check("live")
				if !movePermanentToZone(g, victim, zone.Graveyard) {
					t.Fatal("zone move failed")
				}
				if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
					t.Fatal("departed permanent still on the battlefield")
				}
				if cardCounters := g.CardCounters[victim.CardInstanceID]; !cardCounters.IsEmpty() {
					t.Fatalf("graveyard card retained counters %v", cardCounters.All())
				}
				check("departed")

				returned := addCombatPermanent(g, game.Player2, slimeVanilla("Fresh creature", 4))
				fresh := game.Event{Kind: eventKind, PermanentID: returned.ObjectID, CardID: returned.CardInstanceID, Controller: game.Player2}
				if triggerSelectionMatches(g, game.Player1, fresh, returned.ObjectID, &exact, 0) {
					t.Fatalf("unmarked new object matched %s selection", marker)
				}
			})
		}
	}
}

// TestSlimeBountyInvalidCounterKindRejected proves checked declarations reject
// an unrecognized counter kind in a damage-trigger subject filter, including a
// nested alternative, rather than silently matching nothing.
func TestSlimeBountyInvalidCounterKindRejected(t *testing.T) {
	direct := terminationFacilitatorFixture()
	direct.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection.RequiredCounter = counter.Bounty + 1
	if issues := game.ValidateCardDef(direct); len(issues) == 0 {
		t.Fatal("invalid counter kind accepted in damage-trigger subject filter")
	}
	nested := terminationFacilitatorFixture()
	nested.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection = game.Selection{AnyOf: []game.Selection{
		{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.Bounty},
		{RequiredTypes: []types.Card{types.Planeswalker}, MatchCounter: true, RequiredCounter: counter.Bounty + 1},
	}}
	if issues := game.ValidateCardDef(nested); len(issues) == 0 {
		t.Fatal("invalid counter kind accepted in nested damage-trigger subject alternative")
	}
	nested.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection.AnyOf[1].RequiredCounter = counter.Bounty
	if issues := game.ValidateCardDef(nested); len(issues) != 0 {
		t.Fatalf("valid nested bounty alternatives rejected: %+v", issues)
	}
}

// TestSlimeBountyLiveReadsCopiesClonesAndReplay covers live marker selection,
// recorded-choice replay, copied-trigger independence and game cloning.
func TestSlimeBountyLiveReadsCopiesClonesAndReplay(t *testing.T) {
	setup := func() (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
		g, e := setupBestowMain(t)
		first := addCombatPermanent(g, game.Player2, slimeVanilla("First flyer", 4))
		second := addCombatPermanent(g, game.Player3, slimeVanilla("Second flyer", 4))
		return g, e, first, second
	}
	trigger := func(g *game.Game, e *Engine, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *game.Permanent {
		monster := castSludgeMonster(t, g, e)
		if !e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
			t.Fatal("Sludge Monster's enter trigger did not trigger")
		}
		return monster
	}

	g, e, first, second := setup()
	pick := &slimeTargetAgent{want: []game.Target{game.PermanentTarget(second.ObjectID)}}
	recorded := TurnLog{}
	trigger(g, e, [game.NumPlayers]PlayerAgent{game.Player1: pick}, &recorded)
	if len(recorded.Choices) != 1 || recorded.Choices[0].Request.Kind != game.ChoiceTarget {
		t.Fatalf("recorded choices = %+v; want one trigger target choice", recorded.Choices)
	}

	rg, re, rFirst, rSecond := setup()
	trigger(rg, re, replayAgentsFromTurnLog(&recorded), &TurnLog{})
	re.resolveTopOfStack(rg, &TurnLog{})
	if rSecond.Counters.Get(counter.Slime) != 1 || rFirst.Counters.Has(counter.Slime) {
		t.Fatal("recorded target choice did not replay to the same slime recipient")
	}

	original, ok := g.Stack.Peek()
	if !ok || original.Kind != game.StackTriggeredAbility || !slices.Equal(original.Targets, pick.want) {
		t.Fatalf("top of stack = %+v; want the Sludge Monster trigger targeting the chosen creature", original)
	}
	copied := game.NewStackObjectCopy(original, g.IDGen.Next())
	copied.Targets = []game.Target{game.PermanentTarget(first.ObjectID)}
	g.Stack.Push(copied)
	clone := g.Clone()

	e.resolveTopOfStack(g, &TurnLog{})
	if first.Counters.Get(counter.Slime) != 1 || second.Counters.Has(counter.Slime) {
		t.Fatal("copied trigger did not resolve independently against its own target")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if first.Counters.Get(counter.Slime) != 1 || second.Counters.Get(counter.Slime) != 1 {
		t.Fatal("original trigger was affected by its resolved copy")
	}

	slimed := game.Selection{MatchCounter: true, RequiredCounter: counter.Slime}
	bountied := game.Selection{MatchCounter: true, RequiredCounter: counter.Bounty}
	second.Counters.Add(counter.Bounty, 1)
	if !matchSelectionForPermanent(g, game.Player1, slimed, first) || matchSelectionForPermanent(g, game.Player1, bountied, first) ||
		!matchSelectionForPermanent(g, game.Player1, bountied, second) {
		t.Fatal("live marker selection did not distinguish slime and bounty counters")
	}
	assertSlimeValues(t, g, first, 2, 2, false)

	clonedFirst, _ := permanentByObjectID(clone, first.ObjectID)
	clonedSecond, _ := permanentByObjectID(clone, second.ObjectID)
	if clonedFirst.Counters.Has(counter.Slime) || clonedSecond.Counters.Has(counter.Slime) ||
		clonedSecond.Counters.Has(counter.Bounty) || clone.Stack.Size() != 2 {
		t.Fatal("resolving in the original mutated the cloned game")
	}
	assertSlimeValues(t, clone, clonedFirst, 4, 4, true)
	e.resolveTopOfStack(clone, &TurnLog{})
	e.resolveTopOfStack(clone, &TurnLog{})
	clonedFirst.Counters.Add(counter.Slime, 2)
	if clonedFirst.Counters.Get(counter.Slime) != 3 || clonedSecond.Counters.Get(counter.Slime) != 1 ||
		first.Counters.Get(counter.Slime) != 1 {
		t.Fatal("cloned game counters were not independent of the original")
	}
}
