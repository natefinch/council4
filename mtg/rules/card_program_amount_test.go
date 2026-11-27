package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: CR effective 2026-11-06 from the official copy
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// sections read: 608.2h (information from the game is determined only once,
// when the effect is applied), 608.2c (instructions are followed in order;
// replacement effects may modify them), 107.3c (X defined by the text),
// 121.2/121.2a (multiple draws are individual draws, each open to
// replacement), 121.4/704.5b (a draw from an empty library is attempted and the
// player loses at the next state-based check) and 119.3 (life loss adjusts the
// life total). So a shared X is read once and both halves use that number, even
// if a draw replacement changes the count or fewer cards can be drawn. Pinned
// Oracle corpus 6da9ced0...: Eldritch Pact 281ea9b6-055a-4274-90b3-c74311a6f863
// ("Target player draws X cards and loses X life, where X is the number of
// cards in their graveyard."; Scryfall rulings: none) and Painful Truths
// 58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2 ("Converge — You draw X cards and lose
// X life, where X is the number of colors of mana spent to cast this spell.";
// rulings 2015-08-25: colorless is not a color, copies spent no mana). Savanti
// Romero, Time's Exile 6074632b-4520-4ae8-bcdc-26b1eff71af9 ruling 2026-01-27:
// "The value of X is calculated only once." These test-local programs prove
// the shared View.Amount operation; the cards are admitted by a later card unit.

var (
	eldritchPactTestReference     = game.ProgramReference{ID: "test/eldritch-pact", Version: "1"}
	painfulTruthsTestReference    = game.ProgramReference{ID: "test/painful-truths", Version: "1"}
	untargetedAmountTestReference = game.ProgramReference{ID: "test/untargeted-amount", Version: "1"}
)

// drawAndLoseShared draws X cards and loses X life for player, reading X once.
func drawAndLoseShared(ctx cardprogram.Context, player game.PlayerReference, x game.DynamicAmount) error {
	amount, err := ctx.View().Amount(x)
	if err != nil {
		return err
	}
	if _, err := ctx.Execute(game.Draw{Amount: game.Fixed(amount), Player: player}); err != nil {
		return err
	}
	_, err = ctx.Execute(game.LoseLife{Amount: game.Fixed(amount), Player: player})
	return err
}

func sharedAmountTestPrograms(amountErr *error) []cardprogram.Program {
	target := game.TargetPlayerReference(0)
	return []cardprogram.Program{
		{
			Reference: eldritchPactTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				return drawAndLoseShared(ctx, target, game.DynamicAmount{
					Kind: game.DynamicAmountCountCardsInZone, Player: &target,
					CardZone: zone.Graveyard, Selection: &game.Selection{},
				})
			},
		},
		{
			Reference: painfulTruthsTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				return drawAndLoseShared(ctx, game.ControllerReference(),
					game.DynamicAmount{Kind: game.DynamicAmountColorsOfManaSpentToCast})
			},
		},
		{
			Reference: untargetedAmountTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				_, *amountErr = ctx.View().Amount(game.DynamicAmount{
					Kind: game.DynamicAmountCountCardsInZone, Player: &target,
					CardZone: zone.Graveyard, Selection: &game.Selection{},
				})
				return nil
			},
		},
	}
}

func newSharedAmountEngine(t *testing.T, amountErr *error) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(sharedAmountTestPrograms(amountErr))
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func eldritchPactTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Eldritch Pact",
		ManaCost: opt.Val(cost.Mana{cost.O(6), cost.B}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer}},
			Program: eldritchPactTestReference,
		}.Ability()),
	}}
}

func painfulTruthsTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Painful Truths",
		ManaCost:     opt.Val(cost.Mana{cost.O(2), cost.B}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: painfulTruthsTestReference}.Ability()),
	}}
}

// dredgingAgents have every player dredge whenever a draw offers it.
func dredgingAgents() [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	for player := range game.PlayerID(game.NumPlayers) {
		agents[player] = authoredChoicePolicy{
			simPassPolicy: agents[player].(simPassPolicy),
			choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceZoneSelection && len(request.Options) > 0 {
					return []int{request.Options[0].Index}
				}
				return append([]int(nil), request.DefaultSelection...)
			},
		}
	}
	return agents
}

func castSharedAmountSpell(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef, targets []game.Target, pool func(*mana.Pool)) id.ID {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	pool(&g.Players[game.Player1].ManaPool)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, targets, 0, nil)) {
		t.Fatalf("%s cast failed", def.Name)
	}
	return cardID
}

func TestEldritchPactReadsTargetGraveyardOnceForDrawAndLifeLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newSharedAmountEngine(t, nil)
	for range 6 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	}
	dredger := addCardToGraveyard(g, game.Player2, dredgeCardDef(3))
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Graveyard Card"}})
	life := g.Players[game.Player2].Life
	myLife := g.Players[game.Player1].Life

	castSharedAmountSpell(t, engine, g, eldritchPactTestDef(), []game.Target{game.PlayerTarget(game.Player2)}, func(pool *mana.Pool) {
		pool.Add(mana.B, 1)
		pool.Add(mana.C, 6)
	})
	engine.resolveTopOfStackWithChoices(g, dredgingAgents(), &TurnLog{})

	player := g.Players[game.Player2]
	// X is 2. The first draw is replaced by dredge 3 (graveyard now 4); the
	// second is a normal draw. Life loss still uses the X read before drawing.
	if !player.Hand.Contains(dredger) || player.Hand.Size() != 2 {
		t.Fatalf("hand = %v, want the dredged card and one drawn card", player.Hand.All())
	}
	if got := player.Graveyard.Size(); got != 4 {
		t.Fatalf("graveyard size = %d, want 4 after dredging three", got)
	}
	if got := player.Life; got != life-2 {
		t.Fatalf("target life = %d, want %d (X = 2 read once, not the post-dredge 4)", got, life-2)
	}
	if g.Players[game.Player1].Life != myLife || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("the caster drew or lost life instead of the target player")
	}
}

func TestEldritchPactLosesFullXWhenTheLibraryRunsOut(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newSharedAmountEngine(t, nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Last Card"}})
	for range 3 {
		addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Graveyard Card"}})
	}
	life := g.Players[game.Player2].Life

	castSharedAmountSpell(t, engine, g, eldritchPactTestDef(), []game.Target{game.PlayerTarget(game.Player2)}, func(pool *mana.Pool) {
		pool.Add(mana.B, 1)
		pool.Add(mana.C, 6)
	})
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	player := g.Players[game.Player2]
	if player.Hand.Size() != 1 || !g.FailedDraws[game.Player2] {
		t.Fatalf("hand = %d, failed draw = %v; want one card drawn and a failed draw", player.Hand.Size(), g.FailedDraws[game.Player2])
	}
	if got := player.Life; got != life-3 {
		t.Fatalf("target life = %d, want %d (X = 3 regardless of cards drawn)", got, life-3)
	}
}

func TestPainfulTruthsDrawsAndLosesColorsSpent(t *testing.T) {
	for _, tc := range []struct {
		name string
		pool func(*mana.Pool)
		want int
	}{
		{name: "three colors", want: 3, pool: func(pool *mana.Pool) {
			pool.Add(mana.W, 1)
			pool.Add(mana.U, 1)
			pool.Add(mana.B, 1)
		}},
		{name: "black and colorless", want: 1, pool: func(pool *mana.Pool) {
			pool.Add(mana.B, 1)
			pool.Add(mana.C, 2)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := newSharedAmountEngine(t, nil)
			for range 5 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
			}
			life := g.Players[game.Player1].Life

			spell := castSharedAmountSpell(t, engine, g, painfulTruthsTestDef(), nil, tc.pool)
			engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

			player := g.Players[game.Player1]
			if got := player.Hand.Size(); got != tc.want {
				t.Fatalf("hand size = %d, want %d", got, tc.want)
			}
			if got := player.Life; got != life-tc.want {
				t.Fatalf("life = %d, want %d", got, life-tc.want)
			}
			if !player.Graveyard.Contains(spell) {
				t.Fatal("Painful Truths did not finish resolving to the graveyard")
			}
		})
	}
}

func TestProgramAmountRejectsUndeclaredTarget(t *testing.T) {
	var amountErr error
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newSharedAmountEngine(t, &amountErr)
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Untargeted Amount",
		ManaCost:     opt.Val(cost.Mana{cost.B}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: untargetedAmountTestReference}.Ability()),
	}}
	castSharedAmountSpell(t, engine, g, def, nil, func(pool *mana.Pool) { pool.Add(mana.B, 1) })
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	if amountErr == nil || !strings.Contains(amountErr.Error(), "invalid authored amount") {
		t.Fatalf("Amount error = %v, want an undeclared target to be rejected", amountErr)
	}
}
