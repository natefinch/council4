package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #2446 (optional chosen-group tap of own permanents as
// an "if you do" gate; e.g. Civil Servant, Myr Battlesphere) and enlist
// (CR 702.154). Coalition Warbrute DMU 118 printing
// 554c422f-2a1b-47c0-9358-42b85cdbb84e (oracle
// 84d67d87-2696-4c2f-9f10-29d5832bfbb6): "Enlist (As this creature attacks,
// you may tap a nonattacking creature you control without summoning sickness.
// When you do, add its power to this creature's until end of turn.)";
// https://api.scryfall.com/cards/554c422f-2a1b-47c0-9358-42b85cdbb84e/rulings
// returned four WotC rulings dated 2022-09-09: the choice is made as attackers
// are declared, after they are tapped; the enlisted creature must be untapped,
// not attacking (even with vigilance), and have haste or have been controlled
// since the turn began; the bonus is a trigger put on the stack after
// attackers are declared; a creature can't pay two enlist costs. CR 2026-11-06
// from the verified official copy: 702.154a-d, 508.1f/g/j/m (enlist is an
// optional cost to attack, paid after attackers tap), 607.2h, 603.12
// (reflexive "when you do"), 302.6 (only {T} costs care about summoning
// sickness), and 608.2h (X is read on resolution).

type chosenTapPolicy struct {
	simPassPolicy

	may      bool
	pick     func(game.ChoiceRequest) []int
	requests []game.ChoiceRequest
}

func (p *chosenTapPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.requests = append(p.requests, request)
	if request.Kind == game.ChoiceMay {
		if p.may {
			return []int{1}
		}
		return []int{0}
	}
	if p.pick != nil {
		return p.pick(request)
	}
	return fallbackChoice(request)
}

func (p *chosenTapPolicy) tapRequests() []game.ChoiceRequest {
	var requests []game.ChoiceRequest
	for _, request := range p.requests {
		if request.Kind == game.ChoicePayment {
			requests = append(requests, request)
		}
	}
	return requests
}

func pickPermanentOption(target *game.Permanent) func(game.ChoiceRequest) []int {
	return func(request game.ChoiceRequest) []int {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == target.CardInstanceID {
				return []int{option.Index}
			}
		}
		return nil
	}
}

func optionCardIDs(request game.ChoiceRequest) []id.ID {
	ids := make([]id.ID, 0, len(request.Options))
	for _, option := range request.Options {
		ids = append(ids, option.Card.Val.CardID)
	}
	return ids
}

// "You may tap <amount> untapped creatures you control. If you do, draw a card."
func resolveChosenTapDrawSpell(t *testing.T, g *game.Game, policy *chosenTapPolicy, amount int) {
	t.Helper()
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Optional:      true,
			PublishResult: "tapped",
			Primitive: game.Tap{
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				Choose:        game.TapChoiceExactly,
				Amount:        game.Fixed(amount),
				PublishLinked: "tapped-creature",
			},
		},
		{
			ResultGate: opt.Val(game.InstructionResultGate{Key: "tapped", Succeeded: game.TriTrue}),
			Primitive:  game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		},
	}, nil)
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

func TestChosenTapGateDeclinedTapsNothingAndSkipsDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCreaturePermanent(g, game.Player1)
	addCardToLibraryNamed(g, game.Player1, "Library Card")
	policy := &chosenTapPolicy{may: false}

	resolveChosenTapDrawSpell(t, g, policy, 1)

	if creature.Tapped || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("declined chosen tap tapped=%v hand=%d", creature.Tapped, g.Players[game.Player1].Hand.Size())
	}
	if len(policy.tapRequests()) != 0 {
		t.Fatal("declined chosen tap still asked which creature to tap")
	}
}

func TestChosenTapGateWithTooFewUntappedCreaturesSkipsChoiceAndDraw(t *testing.T) {
	for _, tc := range []struct {
		name   string
		amount int
	}{
		{name: "none untapped", amount: 1},
		{name: "fewer than required", amount: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tapped := addCreaturePermanent(g, game.Player1)
			tapped.Tapped = true
			var untapped *game.Permanent
			if tc.amount > 1 {
				untapped = addCreaturePermanent(g, game.Player1)
			}
			opponents := addCreaturePermanent(g, game.Player2)
			addCardToLibraryNamed(g, game.Player1, "Library Card")
			policy := &chosenTapPolicy{may: true}

			resolveChosenTapDrawSpell(t, g, policy, tc.amount)

			if len(policy.tapRequests()) != 0 {
				t.Fatal("chosen tap prompted without enough untapped creatures you control")
			}
			if opponents.Tapped || untapped != nil && untapped.Tapped || g.Players[game.Player1].Hand.Size() != 0 {
				t.Fatalf("unpayable chosen tap tapped a creature or drew (hand=%d)", g.Players[game.Player1].Hand.Size())
			}
		})
	}
}

func TestChosenTapGateTapsOnlyTheChosenCreatureAndDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	alreadyTapped := addCreaturePermanent(g, game.Player1)
	alreadyTapped.Tapped = true
	first := addCreaturePermanent(g, game.Player1)
	chosen := addCreaturePermanent(g, game.Player1)
	opponents := addCreaturePermanent(g, game.Player2)
	addCardToLibraryNamed(g, game.Player1, "Library Card")
	policy := &chosenTapPolicy{may: true, pick: pickPermanentOption(chosen)}

	resolveChosenTapDrawSpell(t, g, policy, 1)

	requests := policy.tapRequests()
	if len(requests) != 1 {
		t.Fatalf("chosen tap requests = %d, want 1", len(requests))
	}
	want := []id.ID{first.CardInstanceID, chosen.CardInstanceID}
	if got := optionCardIDs(requests[0]); !slices.Equal(got, want) || requests[0].MinChoices != 1 || requests[0].MaxChoices != 1 {
		t.Fatalf("chosen tap options = %v min=%d max=%d, want untapped creatures you control %v exactly once",
			got, requests[0].MinChoices, requests[0].MaxChoices, want)
	}
	if !chosen.Tapped || first.Tapped || opponents.Tapped {
		t.Fatalf("tapped chosen=%v other=%v opponent=%v, want only the chosen creature", chosen.Tapped, first.Tapped, opponents.Tapped)
	}
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("hand = %d after paying the chosen tap, want 1 drawn card", g.Players[game.Player1].Hand.Size())
	}
	var linked []game.LinkedObjectRef
	for key, refs := range g.LinkedObjects {
		if key.LinkID == "tapped-creature" {
			linked = append(linked, refs...)
		}
	}
	if len(linked) != 1 || linked[0].ObjectID != chosen.ObjectID {
		t.Fatalf("linked tapped creatures = %+v, want only the chosen creature", linked)
	}
}

func addEnlistCreature(g *game.Game, controller game.PlayerID, power int) *game.Permanent {
	pt := game.PT{Value: power}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Enlist Test Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{game.EnlistStaticBody},
	}})
}

func newEnlistCombatGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	return g
}

func TestEnlistTapsChosenNonattackingCreatureAndAddsItsPower(t *testing.T) {
	g := newEnlistCombatGame()
	warbrute := addEnlistCreature(g, game.Player1, 3)
	vigilant := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Vigilance)
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	hasty := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Haste)
	hasty.SummoningSick = true
	sick := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	sick.SummoningSick = true
	tapped := addCombatCreaturePermanentWithPower(g, game.Player1, 6)
	tapped.Tapped = true
	addCombatCreaturePermanentWithPower(g, game.Player2, 8)

	policy := &chosenTapPolicy{pick: pickPermanentOption(enlisted)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy
	engine := NewEngine(nil)
	declaration := action.DeclareAttackersWithBands([]game.AttackDeclaration{
		attackAt(warbrute, game.AttackTarget{Player: game.Player2}),
		attackAt(vigilant, game.AttackTarget{Player: game.Player2}),
	}, nil)
	if !engine.applyActionWithChoices(g, game.Player1, declaration, agents, &TurnLog{}) {
		t.Fatal("declaration with an enlist attacker was rejected")
	}

	requests := policy.tapRequests()
	if len(requests) != 1 {
		t.Fatalf("enlist requests = %d, want 1 for the single enlist attacker", len(requests))
	}
	want := []id.ID{enlisted.CardInstanceID, hasty.CardInstanceID}
	if got := optionCardIDs(requests[0]); !slices.Equal(got, want) || requests[0].MinChoices != 0 || requests[0].MaxChoices != 1 {
		t.Fatalf("enlist options = %v min=%d max=%d, want optional choice among %v (not the vigilant attacker, sick, tapped or opponent's creatures)",
			got, requests[0].MinChoices, requests[0].MaxChoices, want)
	}
	if !enlisted.Tapped || vigilant.Tapped || hasty.Tapped {
		t.Fatalf("tapped enlisted=%v vigilant=%v hasty=%v, want only the enlisted creature", enlisted.Tapped, vigilant.Tapped, hasty.Tapped)
	}
	if power := effectivePower(g, warbrute); power != 3 {
		t.Fatalf("enlist bonus applied before its trigger resolved: power %d", power)
	}

	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d after declaring attackers, want the enlist trigger", g.Stack.Size())
	}
	enlisted.Counters.Add(counter.PlusOnePlusOne, 1)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	power := effectivePower(g, warbrute)
	toughness, _ := effectiveToughness(g, warbrute)
	if power != 8 || toughness != 4 {
		t.Fatalf("enlisting attacker is %d/%d, want 8/4 from the enlisted creature's power on resolution", power, toughness)
	}
}

func TestEnlistDeclinedOrWithoutCandidateQueuesNoTrigger(t *testing.T) {
	for _, tc := range []struct {
		name      string
		candidate bool
		decline   bool
	}{
		{name: "declined", candidate: true, decline: true},
		{name: "no eligible creature"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newEnlistCombatGame()
			warbrute := addEnlistCreature(g, game.Player1, 3)
			var other *game.Permanent
			if tc.candidate {
				other = addCombatCreaturePermanentWithPower(g, game.Player1, 4)
			}
			policy := &chosenTapPolicy{pick: func(game.ChoiceRequest) []int { return []int{} }}
			script := &controlScript{
				action: func(_ PlayerObservation, legal []action.Action) action.Action {
					for _, act := range legal {
						payload, ok := act.DeclareAttackersPayload()
						if ok && len(payload.Attackers) == 1 && payload.Attackers[0].Attacker == warbrute.ObjectID {
							return act
						}
					}
					t.Fatal("solo enlist attack missing")
					return action.Pass()
				},
				choice: policy.ChooseChoice,
			}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = script
			engine := NewEngine(nil)
			(combatEngine{engine}).declareAttackers(g, agents, &TurnLog{})

			if !warbrute.Tapped || len(g.Combat.Attackers) != 1 {
				t.Fatal("enlist attacker was not declared")
			}
			wantRequests := 0
			if tc.candidate {
				wantRequests = 1
			}
			if got := len(policy.tapRequests()); got != wantRequests {
				t.Fatalf("enlist requests = %d, want %d", got, wantRequests)
			}
			if other != nil && other.Tapped {
				t.Fatal("declined enlist tapped a creature")
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			if g.Stack.Size() != 0 || len(g.PendingReflexiveTriggers) != 0 {
				t.Fatalf("unpaid enlist queued a trigger (stack=%d)", g.Stack.Size())
			}
			if power := effectivePower(g, warbrute); power != 3 {
				t.Fatalf("unpaid enlist changed power to %d", power)
			}
		})
	}
}
