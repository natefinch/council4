package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine proving fixtures for Sneak (CR 702.190a), not the
// production registry definitions, from the pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Donatello's Technique: oracle 194284f7-cf92-4d9f-a8ce-e4dc3505ab7a, TMT 39
// printing e570082f-8129-44d2-b471-ec7f46a98dbd. Kitsune's Technique: oracle
// 3e0f8ab3-75a6-49a2-a228-fa720cbde3d8, TMT 42 printing
// 9ca5327e-df90-483c-875f-73a23781f56d. The Last Ronin's Technique: oracle
// 6a51bb21-0c01-4ad8-99d7-17685db90b69, TMT 12 printing
// dfb18239-d373-4795-8598-c82abae2cb62. Read the complete Scryfall rulings
// responses (3, 3 and 5 WotC rulings, all 2026-01-27): a sneak cost is usable
// any time you could cast an instant during the declare blockers step on your
// turn, after blockers are declared; tokens entering attacking were never
// declared as attackers; each token's defender is chosen separately. Verified
// official CR (effective 2026-11-06 copy) read: 702.190a-b (Sneak; permanent
// spells entering attacking are proven in sneak_permanent_test.go), 117.1a (instant
// timing), 118.9a (one alternative cost), 509.1h and 506.4 (unblocked
// creatures; a creature returned to hand leaves combat), 601.2b/f/h (announce,
// total cost including modifiers, pay) and 707.10 (a copy copies the
// alternative cost decision).
const (
	donatellosTechniqueText = "Sneak {U} (You may cast this spell for {U} if you also return an unblocked attacker you control to hand during the declare blockers step.)\nDraw two cards."
	kitsunesTechniqueText   = "Sneak {1}{U} (You may cast this spell for {1}{U} if you also return an unblocked attacker you control to hand during the declare blockers step.)\nTarget opponent mills half their library, rounded up."
	lastRoninsTechniqueText = "Sneak {1}{W} (You may cast this spell for {1}{W} if you also return an unblocked attacker you control to hand during the declare blockers step.)\nCreate three 1/1 white Ninja Turtle Spirit creature tokens. If this spell's sneak cost was paid, they enter tapped and attacking."
)

var lastRoninsTechniqueProgram = game.ProgramReference{ID: "test/the-last-ronins-technique/spell/0", Version: "1"}

func donatellosTechniqueFixture() *game.CardDef {
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Blue), CardFace: game.CardFace{
		Name: "Donatello's Technique", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
		Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Blue},
		OracleText:       donatellosTechniqueText,
		AlternativeCosts: []cost.Alternative{cost.SneakAlternative(cost.Mana{cost.U})},
		SpellAbility: opt.Val(game.Mode{Text: "Draw two cards.", Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(2), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

func kitsunesTechniqueFixture() *game.CardDef {
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Blue), CardFace: game.CardFace{
		Name: "Kitsune's Technique", ManaCost: opt.Val(cost.Mana{cost.O(4), cost.U, cost.U}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
		OracleText:       kitsunesTechniqueText,
		AlternativeCosts: []cost.Alternative{cost.SneakAlternative(cost.Mana{cost.O(1), cost.U})},
		SpellAbility: opt.Val(game.Mode{
			Text: "Target opponent mills half their library, rounded up.",
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "target opponent", Allow: game.TargetAllowPlayer,
				Selection: opt.Val(game.Selection{Player: game.PlayerOpponent}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.MoveTopOfLibrary{
					Destination: zone.Graveyard,
					Amount: game.Dynamic(game.DynamicAmount{
						Kind: game.DynamicAmountCountCardsInZone, Divisor: 2, RoundUp: true,
						Player:   func() *game.PlayerReference { ref := game.TargetPlayerReference(0); return &ref }(),
						CardZone: zone.Library, Selection: &game.Selection{},
					}),
					Player: game.TargetPlayerReference(0),
				},
			}},
		}.Ability()),
	}}
}

var ninjaTurtleSpiritToken = &game.CardDef{CardFace: game.CardFace{
	Name: "Ninja Turtle Spirit", Types: []types.Card{types.Creature},
	Subtypes: []types.Sub{types.Ninja, types.Turtle, types.Spirit}, Colors: []color.Color{color.White},
	Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
}}

func lastRoninsTechniqueFixture() *game.CardDef {
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "The Last Ronin's Technique", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.White},
		OracleText:       lastRoninsTechniqueText,
		AlternativeCosts: []cost.Alternative{cost.SneakAlternative(cost.Mana{cost.O(1), cost.W})},
		SpellAbility: opt.Val(game.Mode{
			Text:    "Create three 1/1 white Ninja Turtle Spirit creature tokens. If this spell's sneak cost was paid, they enter tapped and attacking.",
			Program: lastRoninsTechniqueProgram,
		}.Ability()),
	}}
}

func sneakEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{
		{
			Reference: lastRoninsTechniqueProgram,
			Resolve: func(ctx cardprogram.Context) error {
				sneaked := ctx.View().Source().SneakPaid
				_, err := ctx.Execute(game.CreateToken{
					Amount: game.Fixed(3), Source: game.TokenDef(ninjaTurtleSpiritToken),
					Recipient:   opt.Val(game.ControllerReference()),
					EntryTapped: sneaked, EntryAttacking: sneaked,
				})
				return err
			},
		},
		{
			Reference: reiterateFixtureProgram,
			Resolve: func(ctx cardprogram.Context) error {
				_, err := ctx.Copy(game.CopyStackObject{
					Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
				}, ctx.View().Source().Controller)
				return err
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

// sneakAgent pays with the payment option labeled cost, recording every
// payment option label it was offered.
type sneakAgent struct {
	cost    string
	offered [][]string
}

func (*sneakAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *sneakAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoicePayment {
		labels := make([]string, 0, len(request.Options))
		for _, option := range request.Options {
			labels = append(labels, option.Label)
		}
		a.offered = append(a.offered, labels)
		for _, option := range request.Options {
			if option.Label == a.cost {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

type sneakFixture struct {
	game     *game.Game
	engine   *Engine
	attacker *game.Permanent
}

// newSneakFixture is Player1's declare blockers step with priority, after
// Player2 declared no blockers against Player1's lone attacker.
func newSneakFixture(t *testing.T) sneakFixture {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	g.Combat.BlockDeclarationsComplete[game.Player2] = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player1
	for range 5 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card", Types: []types.Card{types.Land}}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Library Card", Types: []types.Card{types.Land}}})
	}
	return sneakFixture{game: g, engine: sneakEngine(t), attacker: attacker}
}

func (f sneakFixture) cast(t *testing.T, agent *sneakAgent, cast action.Action) {
	t.Helper()
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if !f.engine.applyActionWithChoices(f.game, game.Player1, cast, agents, &TurnLog{}) {
		t.Fatalf("cast %+v failed; offered payments %v", cast, agent.offered)
	}
}

func (f sneakFixture) resolve(agent *sneakAgent) {
	resolveStackWithTriggers(f.engine, f.game, [game.NumPlayers]PlayerAgent{game.Player1: agent})
}

func (f sneakFixture) assertAttackerReturned(t *testing.T) {
	t.Helper()
	if !f.game.Players[game.Player1].Hand.Contains(f.attacker.CardInstanceID) {
		t.Fatal("sneak cost did not return the unblocked attacker to its owner's hand")
	}
	if permanentForCard(f.game, f.attacker.CardInstanceID) != nil {
		t.Fatal("returned attacker is still on the battlefield")
	}
	if slices.ContainsFunc(f.game.Combat.Attackers, func(attack game.AttackDeclaration) bool {
		return attack.Attacker == f.attacker.ObjectID
	}) {
		t.Fatal("returned attacker is still in combat")
	}
}

// TestDonatellosTechniqueSneaksDuringDeclareBlockers casts a sorcery for its
// sneak cost in the declare blockers step: only Sneak is offered, it pays {U}
// and returns the unblocked attacker, the cast triggers cast abilities, and the
// spell resolves for its full effect.
func TestDonatellosTechniqueSneaksDuringDeclareBlockers(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	spellID := addCardToHand(g, game.Player1, donatellosTechniqueFixture())
	g.Players[game.Player1].ManaPool.Add(mana.U, 4)
	watcher := watchCasts(g)
	cast := action.CastSpell(spellID, nil, 0, nil)
	if !actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions omit the sorcery's sneak cast in the declare blockers step")
	}

	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, cast)

	// Sneak is the only payment outside sorcery timing (CR 117.1a,
	// 702.190a), so no payment choice is offered.
	if len(agent.offered) != 0 {
		t.Fatalf("offered payments = %v, want only Sneak with no payment choice", agent.offered)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 3 {
		t.Fatalf("mana left = %d, want 3 after paying the {U} sneak cost", got)
	}
	f.assertAttackerReturned(t)
	if triggers := resolveCastWatcherTriggers(t, g, f.engine, watcher); triggers != 1 {
		t.Fatalf("cast triggers = %d, want 1 for a sneak cast", triggers)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != spellID || !obj.Sneaked {
		t.Fatalf("stack object = %+v, want sneaked Donatello's Technique", obj)
	}
	handBefore := g.Players[game.Player1].Hand.Size()
	f.resolve(agent)
	if got := g.Players[game.Player1].Hand.Size() - handBefore; got != 2 {
		t.Fatalf("cards drawn = %d, want 2", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved sorcery is not in its owner's graveyard")
	}
}

// TestSneakRequiresYourDeclareBlockersStepAndUnblockedAttacker proves the
// sorcery has no cast outside the Sneak window and no Sneak cost elsewhere.
func TestSneakRequiresYourDeclareBlockersStepAndUnblockedAttacker(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(sneakFixture)
	}{
		{name: "main phase", mutate: func(f sneakFixture) {
			f.game.Combat = nil
			setMainPhasePriority(f.game, game.Player1)
		}},
		{name: "declare attackers step", mutate: func(f sneakFixture) { f.game.Turn.Step = game.StepDeclareAttackers }},
		{name: "combat damage step", mutate: func(f sneakFixture) { f.game.Turn.Step = game.StepCombatDamage }},
		{name: "blockers not yet declared", mutate: func(f sneakFixture) {
			f.game.Combat.BlockDeclarationsComplete[game.Player2] = false
		}},
		{name: "attacker was blocked", mutate: func(f sneakFixture) {
			f.game.Combat.BlockedAttackers = map[game.ObjectID]bool{f.attacker.ObjectID: true}
		}},
		{name: "opponent's turn", mutate: func(f sneakFixture) { f.game.Turn.ActivePlayer = game.Player2 }},
		{name: "without priority", mutate: func(f sneakFixture) { f.game.Turn.PriorityPlayer = game.Player2 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSneakFixture(t)
			spellID := addCardToHand(f.game, game.Player1, donatellosTechniqueFixture())
			f.game.Players[game.Player1].ManaPool.Add(mana.U, 2)
			tt.mutate(f)
			if actionsContain(f.engine.legalActions(f.game, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
				t.Fatal("sorcery was castable for {U} outside its Sneak window")
			}
		})
	}
}

// TestKitsunesTechniqueOffersSneakBesideNormalCost proves an instant in the
// Sneak window may be cast for either cost (CR 118.9a): Sneak mills the target
// opponent half their library rounded up after returning the attacker, while a
// normal cast leaves the attacker in combat.
func TestKitsunesTechniqueOffersSneakBesideNormalCost(t *testing.T) {
	for _, paid := range []string{"Sneak", "Normal cost"} {
		t.Run(paid, func(t *testing.T) {
			f := newSneakFixture(t)
			g := f.game
			spellID := addCardToHand(g, game.Player1, kitsunesTechniqueFixture())
			g.Players[game.Player1].ManaPool.Add(mana.U, 6)
			agent := &sneakAgent{cost: paid}
			f.cast(t, agent, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil))

			if len(agent.offered) != 1 || !slices.Contains(agent.offered[0], "Sneak") || !slices.Contains(agent.offered[0], "Normal cost") {
				t.Fatalf("offered payments = %v, want Normal cost and Sneak", agent.offered)
			}
			obj, _ := g.Stack.Peek()
			if obj.Sneaked != (paid == "Sneak") {
				t.Fatalf("Sneaked = %v after paying %s", obj.Sneaked, paid)
			}
			wantLeft := 4
			if paid == "Sneak" {
				f.assertAttackerReturned(t)
			} else {
				wantLeft = 0
				if permanentForCard(g, f.attacker.CardInstanceID) == nil || len(g.Combat.Attackers) != 1 {
					t.Fatal("normal cast removed the attacker")
				}
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != wantLeft {
				t.Fatalf("mana left = %d, want %d", got, wantLeft)
			}
			f.resolve(agent)
			if got := g.Players[game.Player2].Library.Size(); got != 2 {
				t.Fatalf("opponent library = %d, want 2 after milling 3 of 5", got)
			}
			if got := g.Players[game.Player2].Graveyard.Size(); got != 3 {
				t.Fatalf("opponent graveyard = %d, want 3", got)
			}
		})
	}
}

func ninjaTurtleSpiritTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef == ninjaTurtleSpiritToken {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func attackingPlayer(g *game.Game, permanent *game.Permanent, defender game.PlayerID) bool {
	return g.Combat != nil && slices.ContainsFunc(g.Combat.Attackers, func(attack game.AttackDeclaration) bool {
		return attack.Attacker == permanent.ObjectID && attack.Target.Player == defender
	})
}

// TestLastRoninsTechniqueSneakTokensEnterTappedAndAttacking proves "if this
// spell's sneak cost was paid" reads the resolving spell's paid Sneak cost.
func TestLastRoninsTechniqueSneakTokensEnterTappedAndAttacking(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	spellID := addCardToHand(g, game.Player1, lastRoninsTechniqueFixture())
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(spellID, nil, 0, nil))
	f.assertAttackerReturned(t)

	clone := g.Clone()
	if obj, ok := clone.Stack.Peek(); !ok || !obj.Sneaked {
		t.Fatalf("cloned stack object = %+v, want Sneaked preserved", obj)
	}

	f.resolve(agent)
	tokens := ninjaTurtleSpiritTokens(g)
	if len(tokens) != 3 {
		t.Fatalf("tokens = %d, want 3", len(tokens))
	}
	for _, token := range tokens {
		if !token.Tapped || !attackingPlayer(g, token, game.Player2) {
			t.Fatalf("token %+v is not tapped and attacking; attackers %+v", token, g.Combat.Attackers)
		}
	}
}

// TestLastRoninsTechniqueNormalCastTokensDoNotAttack casts the instant for
// its mana cost in the main phase, so its sneak cost was not paid.
func TestLastRoninsTechniqueNormalCastTokensDoNotAttack(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	g.Combat = nil
	setMainPhasePriority(g, game.Player1)
	spellID := addCardToHand(g, game.Player1, lastRoninsTechniqueFixture())
	g.Players[game.Player1].ManaPool.Add(mana.W, 4)
	agent := &sneakAgent{cost: "Normal cost"}
	f.cast(t, agent, action.CastSpell(spellID, nil, 0, nil))
	if len(agent.offered) != 0 {
		t.Fatalf("offered payments = %v, want only the normal cost outside the Sneak window", agent.offered)
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Sneaked {
		t.Fatalf("stack object = %+v, want a normal cast", obj)
	}

	f.resolve(agent)
	tokens := ninjaTurtleSpiritTokens(g)
	if len(tokens) != 3 {
		t.Fatalf("tokens = %d, want 3", len(tokens))
	}
	for _, token := range tokens {
		if token.Tapped {
			t.Fatalf("token %+v entered tapped without its sneak cost paid", token)
		}
	}
}

// TestCopyOfSneakedSpellKeepsSneakCostPaid proves a copy copies the paid Sneak
// alternative cost (CR 707.10), so its tokens also enter attacking.
func TestCopyOfSneakedSpellKeepsSneakCostPaid(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	spellID := addCardToHand(g, game.Player1, lastRoninsTechniqueFixture())
	copierID := addCardToHand(g, game.Player1, reiterateFixture())
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(spellID, nil, 0, nil))
	original, _ := g.Stack.Peek()

	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	agent.cost = "Normal cost"
	f.cast(t, agent, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil))
	f.engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.ID == original.ID || !copied.Sneaked {
		t.Fatalf("top of stack = %+v, want a Sneaked copy", copied)
	}

	f.resolve(agent)
	tokens := ninjaTurtleSpiritTokens(g)
	if len(tokens) != 6 {
		t.Fatalf("tokens = %d, want 6", len(tokens))
	}
	for _, token := range tokens {
		if !token.Tapped || !attackingPlayer(g, token, game.Player2) {
			t.Fatalf("token %+v is not tapped and attacking", token)
		}
	}
}

// TestSneakCostIncludesCostIncreases proves the total cost of a sneak cast
// starts from the sneak cost and still adds cost increases (CR 601.2f).
func TestSneakCostIncludesCostIncreases(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 1})
	spellID := addCardToHand(g, game.Player1, donatellosTechniqueFixture())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cast := action.CastSpell(spellID, nil, 0, nil)
	if actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("sneak cast is legal with {U} under a {1} spell tax")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	f.cast(t, &sneakAgent{cost: "Sneak"}, cast)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left = %d, want 0 after paying {1}{U}", got)
	}
	f.assertAttackerReturned(t)
}

// TestSneakCombinesWithOtherSpellCosts pays Sneak alongside an optional mana
// cost (Buyback) and a nonmana additional cost choosing another creature: the
// unblocked attacker pays only Sneak and the sacrifice takes a different
// creature (CR 118.9d, 601.2h). Declining Buyback still offers Sneak.
func TestSneakCombinesWithOtherSpellCosts(t *testing.T) {
	for _, tt := range []struct {
		name    string
		buyback bool
	}{{name: "with buyback", buyback: true}, {name: "without buyback"}} {
		t.Run(tt.name, func(t *testing.T) {
			f := newSneakFixture(t)
			g := f.game
			def := donatellosTechniqueFixture()
			def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
				Text: "Buyback {3}", ZoneOfFunction: zone.Stack,
				KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{Cost: cost.Mana{cost.O(3)}}},
			})
			def.AdditionalCosts = []cost.Additional{{
				Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1,
				MatchPermanentType: true, PermanentType: types.Creature,
			}}
			spellID := addCardToHand(g, game.Player1, def)
			bystander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Bystander", Types: []types.Card{types.Creature},
				Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			}})
			// The cost planner commits greedily to the first sacrifice
			// candidate (#3917), so the bystander precedes the attacker.
			slices.Reverse(g.Battlefield)
			g.Players[game.Player1].ManaPool.Add(mana.U, 4)
			var offered [][]string
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoicePayment {
					return request.DefaultSelection
				}
				var labels []string
				chosen := -1
				for _, option := range request.Options {
					labels = append(labels, option.Label)
					if option.Label == permanentChoiceLabel(g, bystander) ||
						strings.Contains(option.Label, "Sneak") && strings.Contains(option.Label, "Buyback") == tt.buyback {
						chosen = option.Index
					}
				}
				offered = append(offered, labels)
				if chosen < 0 {
					t.Fatalf("no wanted option in %v", labels)
				}
				return []int{chosen}
			}}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			if !f.engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatalf("sneak cast with other costs failed; offered %v", offered)
			}
			f.assertAttackerReturned(t)
			if permanentForCard(g, bystander.CardInstanceID) != nil || !g.Players[game.Player1].Graveyard.Contains(bystander.CardInstanceID) {
				t.Fatal("the other creature was not sacrificed")
			}
			obj, _ := g.Stack.Peek()
			if !obj.Sneaked || obj.BuybackPaid != tt.buyback {
				t.Fatalf("Sneaked = %v, BuybackPaid = %v; want true, %v", obj.Sneaked, obj.BuybackPaid, tt.buyback)
			}
			wantLeft := 3
			if tt.buyback {
				wantLeft = 0
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != wantLeft {
				t.Fatalf("mana left = %d, want %d", got, wantLeft)
			}
			resolveStackWithTriggers(f.engine, g, agents)
			if got := g.Players[game.Player1].Hand.Contains(spellID); got != tt.buyback {
				t.Fatalf("spell in hand after resolution = %v, want %v", got, tt.buyback)
			}
		})
	}
}

// TestStormCopiesOfSneakedSpellKeepSneakCostPaid proves storm's separate copy
// route keeps the paid Sneak alternative cost (CR 702.40a, 707.10).
func TestStormCopiesOfSneakedSpellKeepSneakCostPaid(t *testing.T) {
	f := newSneakFixture(t)
	g := f.game
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	def := lastRoninsTechniqueFixture()
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	spellID := addCardToHand(g, game.Player1, def)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(firstID, nil, 0, nil))
	f.engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})

	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	f.cast(t, agent, action.CastSpell(spellID, nil, 0, nil))
	f.engine.putTriggeredAbilitiesOnStack(g)
	f.engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.Sneaked {
		t.Fatalf("top of stack = %+v, want a Sneaked storm copy", copied)
	}

	f.resolve(agent)
	tokens := ninjaTurtleSpiritTokens(g)
	if len(tokens) != 6 {
		t.Fatalf("tokens = %d, want 6", len(tokens))
	}
	for _, token := range tokens {
		if !token.Tapped || !attackingPlayer(g, token, game.Player2) {
			t.Fatalf("token %+v is not tapped and attacking", token)
		}
	}
}
