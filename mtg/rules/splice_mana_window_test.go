package rules

import (
	"slices"
	"strings"
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardg "github.com/natefinch/council4/mtg/cards/g"
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

// Source review: Engine gap #3635. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Desperate Ritual
// 3620ab8f-b693-4225-bb96-5ca3cb3fa465 {1}{R} Instant — Arcane "Add {R}{R}{R}.
// Splice onto Arcane {1}{R}"; Glacial Ray e65a8ab8-84a7-4fe2-b009-ceb87bec6b4e
// {1}{R} Instant — Arcane "Glacial Ray deals 2 damage to any target. Splice
// onto Arcane {1}{R}"; Waterlogged Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa
// and Silent Clearing fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life:
// Add {G} or {U}." / "{W} or {B}." (mana abilities the payment planner never
// activates on its own, so only the caster's payment-time activations can
// pay with them); Thalia, Guardian of Thraben
// 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells cost {1} more to
// cast." Desperate Ritual rulings (2013-06-07): all splice cards are revealed
// at once and their splice costs paid; a card can't be spliced onto itself;
// targets for spliced text are chosen after revealing. Millikin, named by the
// issue, isn't used: under 605.1a its ability is not a mana ability (#3690).
// Comprehensive Rules effective 2026-09-25: 601.2b a player who wishes to
// splice reveals the cards as the spell is announced; 601.2f the total cost
// (mana cost or alternative cost plus additional costs and increases) is
// locked; 601.2g "If the total cost includes a mana payment, the player then
// has a chance to activate mana abilities"; 601.2h "Partial payments are not
// allowed."; 605.3a mana abilities may be activated while casting a spell that
// requires a mana payment; 702.47a splice "you may reveal this card from your
// hand and pay [cost] as an additional cost ... follows rules 601.2b and
// 601.2f-h"; 702.47b splice costs are paid with the host; 733.1 an action that
// can't be legally completed is reversed, and the player may reverse mana
// abilities activated while making it.

const splicePrompt = "Choose a card to splice onto this Arcane spell"

// spliceManaAgent records the splice cards offered at each splice prompt and
// chooses splices by name in order (an empty name or running out declines),
// announces each Phyrexian symbol with life or mana, activates the mana
// abilities whose labels contain each of mana in order while paying, targets
// target when offered, and accepts optional casts.
type spliceManaAgent struct {
	splices   []string
	life      bool
	mana      []string
	target    opt.V[game.PlayerID]
	offered   [][]string
	activated int
}

func (*spliceManaAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *spliceManaAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case strings.HasPrefix(request.Prompt, splicePrompt):
		labels := make([]string, len(request.Options))
		for i, option := range request.Options {
			labels[i] = option.Label
		}
		round := len(a.offered)
		a.offered = append(a.offered, labels)
		if round >= len(a.splices) || a.splices[round] == "" {
			return nil
		}
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.splices[round]) {
				return []int{option.Index}
			}
		}
		return nil
	case request.Kind == game.ChoicePayment && strings.HasSuffix(request.Prompt, "/P}"):
		want := 0
		if a.life {
			want = 1
		}
		return []int{want}
	case request.Prompt == "Activate a mana ability while paying?":
		if a.activated < len(a.mana) {
			for _, option := range request.Options {
				if strings.Contains(option.Label, a.mana[a.activated]) {
					a.activated++
					return []int{option.Index}
				}
			}
		}
	case request.Kind == game.ChoiceTarget && a.target.Exists:
		for _, option := range request.Options {
			if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer && option.Targets[0].PlayerID == a.target.Val {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// spliceOffered reports whether any splice prompt offered the card named name.
func (a *spliceManaAgent) spliceOffered(name string) bool {
	for _, labels := range a.offered {
		if slices.ContainsFunc(labels, func(label string) bool { return strings.Contains(label, name) }) {
			return true
		}
	}
	return false
}

func spliceManaAgents(agent PlayerAgent) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	return agents
}

func addSpliceLands(t *testing.T, g *game.Game, names ...string) {
	t.Helper()
	for _, name := range names {
		addCombatPermanent(g, game.Player1, registryCardDef(t, name))
	}
}

func requireSplicedHost(t *testing.T, g *game.Game, hostID id.ID, spliced int) *game.StackObject {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != hostID {
		t.Fatalf("stack top = %+v, want the cast host %v", obj, hostID)
	}
	if len(obj.SplicedContent) != spliced {
		t.Fatalf("spliced content = %d, want %d", len(obj.SplicedContent), spliced)
	}
	return obj
}

// TestSpliceOfferedWhenPayableWithPaymentTimeManaAbilities casts Desperate
// Ritual from hand with priority while a second Desperate Ritual is in hand.
// The host and splice cost together, plus any Thalia tax, can be paid only by
// activating every listed land's mana ability while paying (CR 601.2g,
// 605.3a), which the payment planner never does on its own. With one land
// fewer the total can't be paid and the splice isn't offered; with all of them
// it is offered, the lands are activated while paying, and the spliced effect
// resolves with the host's.
func TestSpliceOfferedWhenPayableWithPaymentTimeManaAbilities(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		thalia bool
		pool   []mana.Color
		lands  []string
		order  []string
	}{
		{name: "one land", pool: []mana.Color{mana.R, mana.R, mana.C}, lands: []string{"Waterlogged Grove"}, order: []string{groveMana}},
		{name: "two lands with Thalia's tax", thalia: true, pool: []mana.Color{mana.R, mana.R, mana.C},
			lands: []string{"Waterlogged Grove", "Silent Clearing"}, order: []string{groveMana, clearingMana}},
	} {
		for lands := len(tc.lands) - 1; lands <= len(tc.lands); lands++ {
			payable := lands == len(tc.lands)
			name := tc.name
			if !payable {
				name += " short"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				g, e := newSpliceGame()
				hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
				spliceID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
				if tc.thalia {
					addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
				}
				addSpliceLands(t, g, tc.lands[:lands]...)
				p := g.Players[game.Player1]
				life := p.Life
				for _, c := range tc.pool {
					p.ManaPool.Add(c, 1)
				}
				agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}, mana: tc.order[:lands]}

				outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

				if outcome != actionApplied {
					t.Fatalf("outcome = %v, want the host cast", outcome)
				}
				if got := agent.spliceOffered("Desperate Ritual"); got != payable {
					t.Fatalf("splice offered = %t, want %t (offers %v)", got, payable, agent.offered)
				}
				if !payable {
					requireSplicedHost(t, g, hostID, 0)
					return
				}
				requireSplicedHost(t, g, hostID, 1)
				for _, land := range tc.lands {
					if !reboundPermanentNamed(t, g, land).Tapped {
						t.Fatalf("%s is untapped, want it activated while paying", land)
					}
				}
				if p.Life != life-len(tc.lands) || p.ManaPool.Total() != 0 || agent.activated != len(tc.lands) {
					t.Fatalf("life = %d, pool = %d, activated = %d; want each land paid for the total", p.Life, p.ManaPool.Total(), agent.activated)
				}
				if !p.Hand.Contains(spliceID) {
					t.Fatal("the spliced card left the hand; splice only reveals it")
				}

				resolveStackWithTriggers(e, g, spliceManaAgents(agent))

				if got := p.ManaPool.Amount(mana.R); got != 6 {
					t.Fatalf("red mana after resolving = %d, want the host's and spliced {R}{R}{R}", got)
				}
			})
		}
	}
}

// TestSpliceDeclinedWithPaymentTimeManaAbilityLeavesLandUnused offers the
// splice payable only with Waterlogged Grove, but the caster declines it: the
// host is cast alone for its own cost and the Grove is never activated.
func TestSpliceDeclinedWithPaymentTimeManaAbilityLeavesLandUnused(t *testing.T) {
	t.Parallel()
	g, e := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addSpliceLands(t, g, "Waterlogged Grove")
	p := g.Players[game.Player1]
	life := p.Life
	p.ManaPool.Add(mana.R, 2)
	p.ManaPool.Add(mana.C, 1)
	agent := &spliceManaAgent{splices: []string{""}}

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied || !agent.spliceOffered("Desperate Ritual") {
		t.Fatalf("outcome = %v, offers = %v; want the splice offered and the host cast", outcome, agent.offered)
	}
	requireSplicedHost(t, g, hostID, 0)
	requireLandsUnused(t, g, life, "Waterlogged Grove")
	if p.ManaPool.Total() != 1 {
		t.Fatalf("pool = %d, want the host's {1}{R} paid from the pool", p.ManaPool.Total())
	}
}

// TestUnpaidSpliceReversesCast chooses the splice payable only with
// Waterlogged Grove but leaves the Grove unused while paying. The locked total
// can't be partly paid (CR 601.2h), so the whole cast is reversed (CR 733.1):
// the host returns to hand, the pool is restored, and nothing is spliced.
func TestUnpaidSpliceReversesCast(t *testing.T) {
	t.Parallel()
	g, e := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	spliceID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addSpliceLands(t, g, "Waterlogged Grove")
	p := g.Players[game.Player1]
	life := p.Life
	p.ManaPool.Add(mana.R, 2)
	p.ManaPool.Add(mana.C, 1)
	agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}}

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

	if outcome != actionReversed || !agent.spliceOffered("Desperate Ritual") {
		t.Fatalf("outcome = %v, offers = %v; want the offered splice's cast reversed", outcome, agent.offered)
	}
	if g.Stack.Size() != 0 || !p.Hand.Contains(hostID) || !p.Hand.Contains(spliceID) {
		t.Fatalf("stack = %d, host in hand = %t; want the cast undone", g.Stack.Size(), p.Hand.Contains(hostID))
	}
	requireLandsUnused(t, g, life, "Waterlogged Grove")
	if p.ManaPool.Amount(mana.R) != 2 || p.ManaPool.Amount(mana.C) != 1 {
		t.Fatalf("pool = %d, want {R}{R}{C} restored", p.ManaPool.Total())
	}
}

// TestSpliceTargetsPaidWithPaymentTimeManaAbility splices Glacial Ray onto
// Desperate Ritual: the spliced text's target is chosen after revealing it
// (Desperate Ritual's 2013-06-07 splice rulings), and its splice cost is paid with
// Waterlogged Grove's mana while paying.
func TestSpliceTargetsPaidWithPaymentTimeManaAbility(t *testing.T) {
	t.Parallel()
	g, e := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	addSpliceLands(t, g, "Waterlogged Grove")
	p := g.Players[game.Player1]
	life := p.Life
	p.ManaPool.Add(mana.R, 2)
	p.ManaPool.Add(mana.C, 1)
	opponentLife := g.Players[game.Player2].Life
	agent := &spliceManaAgent{splices: []string{"Glacial Ray"}, mana: []string{groveMana}, target: opt.Val(game.Player2)}

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied {
		t.Fatalf("outcome = %v, offers = %v; want Glacial Ray spliced", outcome, agent.offered)
	}
	requireSplicedHost(t, g, hostID, 1)
	if !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || p.Life != life-1 {
		t.Fatal("Waterlogged Grove wasn't activated to pay the splice cost")
	}

	resolveStackWithTriggers(e, g, spliceManaAgents(agent))

	if got := g.Players[game.Player2].Life; got != opponentLife-2 {
		t.Fatalf("opponent life = %d, want 2 less from the spliced Glacial Ray", got)
	}
	if got := p.ManaPool.Amount(mana.R); got != 3 {
		t.Fatalf("red mana = %d, want the host's {R}{R}{R}", got)
	}
}

// phyrexianArcane is an Arcane instant costing {1}{R/P} that adds {R}.
func phyrexianArcane() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Phyrexian Arcane Test",
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.PhyrexianMana(mana.R)}),
		Colors:   []color.Color{color.Red},
		Types:    []types.Card{types.Instant},
		Subtypes: []types.Sub{types.Arcane},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.R},
		}}}.Ability()),
	}}
}

// TestSpliceJudgedAgainstAnnouncedPhyrexianPayment casts an Arcane host whose
// {R/P} payment is announced before splicing (CR 601.2b). Desperate Ritual's
// splice is then judged against the host's total as announced: paying the
// symbol with life leaves {1} plus {1}{R}, payable from {R}{C} and Waterlogged
// Grove's mana while paying, so the splice is offered and paid; paying it with
// mana needs one more mana than those make, so it isn't offered.
func TestSpliceJudgedAgainstAnnouncedPhyrexianPayment(t *testing.T) {
	t.Parallel()
	for _, payLife := range []bool{true, false} {
		name := "mana"
		if payLife {
			name = "life"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g, e := newSpliceGame()
			hostID := addCardToHand(g, game.Player1, phyrexianArcane())
			addCardToHand(g, game.Player1, cardd.DesperateRitual())
			addSpliceLands(t, g, "Waterlogged Grove")
			p := g.Players[game.Player1]
			life := p.Life
			p.ManaPool.Add(mana.R, 1)
			p.ManaPool.Add(mana.C, 1)
			agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}, life: payLife}
			if payLife {
				agent.mana = []string{groveMana}
			}

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied {
				t.Fatalf("outcome = %v, want the host cast", outcome)
			}
			if got := agent.spliceOffered("Desperate Ritual"); got != payLife {
				t.Fatalf("splice offered = %t, want %t (offers %v)", got, payLife, agent.offered)
			}
			if !payLife {
				requireSplicedHost(t, g, hostID, 0)
				requireLandsUnused(t, g, life, "Waterlogged Grove")
				return
			}
			requireSplicedHost(t, g, hostID, 1)
			if !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || p.Life != life-3 || p.ManaPool.Total() != 0 {
				t.Fatalf("life = %d, pool = %d; want 2 life for {R/P} and the Grove paid", p.Life, p.ManaPool.Total())
			}
		})
	}
}

// TestFreeCastSpliceOfferedWithPaymentTimeManaAbility casts Desperate Ritual
// from hand without paying its mana cost. Its splice cost is still an
// additional cost to pay (CR 702.47a, 601.2f), and with only {R} floating it
// is payable only with Waterlogged Grove's mana while paying, so the splice is
// offered only with the Grove, and is then paid with it.
func TestFreeCastSpliceOfferedWithPaymentTimeManaAbility(t *testing.T) {
	t.Parallel()
	for _, grove := range []bool{false, true} {
		name := "without Grove"
		if grove {
			name = "with Grove"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			e, g := newEffectiveCascadeGame()
			hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
			addCardToHand(g, game.Player1, cardd.DesperateRitual())
			if grove {
				addSpliceLands(t, g, "Waterlogged Grove")
			}
			p := g.Players[game.Player1]
			life := p.Life
			p.ManaPool.Add(mana.R, 1)
			agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}, mana: []string{groveMana}}

			if !e.chooseAndCastFreeSpellFromSource(g, p, game.Player1, hostID, zone.Hand, game.Selection{}, opt.V[int]{}, false, spliceManaAgents(agent), &TurnLog{}) {
				t.Fatal("the free cast of Desperate Ritual failed")
			}
			if got := agent.spliceOffered("Desperate Ritual"); got != grove {
				t.Fatalf("splice offered = %t, want %t (offers %v)", got, grove, agent.offered)
			}
			if !grove {
				requireSplicedHost(t, g, hostID, 0)
				if p.ManaPool.Total() != 1 {
					t.Fatalf("pool = %d, want the free host to leave {R}", p.ManaPool.Total())
				}
				return
			}
			requireSplicedHost(t, g, hostID, 1)
			if !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || p.Life != life-1 || p.ManaPool.Total() != 0 {
				t.Fatalf("life = %d, pool = %d; want the splice cost paid with {R} and the Grove", p.Life, p.ManaPool.Total())
			}
		})
	}
}

// TestUnpaidFreeCastSpliceReversesCast chooses the splice on a free cast but
// leaves Waterlogged Grove unused while paying. The locked total can't be
// partly paid (CR 601.2h), so the cast is reversed (CR 733.1): the host returns
// to hand with the Grove, life and pool untouched. As on the other free routes
// whose cost has no Phyrexian payment to revise (such as rebound), the
// reversed free cast isn't offered again.
func TestUnpaidFreeCastSpliceReversesCast(t *testing.T) {
	t.Parallel()
	e, g := newEffectiveCascadeGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	spliceID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addSpliceLands(t, g, "Waterlogged Grove")
	p := g.Players[game.Player1]
	life := p.Life
	p.ManaPool.Add(mana.R, 1)
	agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}}

	if e.chooseAndCastFreeSpellFromSource(g, p, game.Player1, hostID, zone.Hand, game.Selection{}, opt.V[int]{}, false, spliceManaAgents(agent), &TurnLog{}) {
		t.Fatal("the free cast completed without paying the splice cost")
	}
	if len(agent.offered) != 1 {
		t.Fatalf("splice offers = %v, want the splice offered once", agent.offered)
	}
	if g.Stack.Size() != 0 || !p.Hand.Contains(hostID) || !p.Hand.Contains(spliceID) {
		t.Fatalf("stack = %d, host in hand = %t; want the cast undone", g.Stack.Size(), p.Hand.Contains(hostID))
	}
	requireLandsUnused(t, g, life, "Waterlogged Grove")
	if p.ManaPool.Amount(mana.R) != 1 || p.ManaPool.Total() != 1 {
		t.Fatalf("pool = %d, want {R} restored", p.ManaPool.Total())
	}
}

// TestSpliceManaWindowCastInCloneLeavesOriginal casts with a splice paid by
// Waterlogged Grove in a clone of the game, leaving the original unchanged.
func TestSpliceManaWindowCastInCloneLeavesOriginal(t *testing.T) {
	t.Parallel()
	g, e := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addSpliceLands(t, g, "Waterlogged Grove")
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	life := g.Players[game.Player1].Life
	clone := g.Clone()
	agent := &spliceManaAgent{splices: []string{"Desperate Ritual"}, mana: []string{groveMana}}

	outcome, _ := e.attemptActionWithChoices(clone, game.Player1, action.CastSpell(hostID, nil, 0, nil), spliceManaAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied {
		t.Fatalf("outcome = %v, want the clone's cast spliced", outcome)
	}
	requireSplicedHost(t, clone, hostID, 1)
	if g.Stack.Size() != 0 || !g.Players[game.Player1].Hand.Contains(hostID) || g.Players[game.Player1].ManaPool.Total() != 3 {
		t.Fatal("casting in the clone changed the original game")
	}
	requireLandsUnused(t, g, life, "Waterlogged Grove")
}
