package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// veilConditionalDraw returns Veil of Summer's first clause as a resolving
// condition: "Draw a card if an opponent has cast a blue or black spell this
// turn." The instant has no permanent source, so the condition is evaluated
// with a nil source against the current turn's spell-cast event history.
func veilConditionalDraw() opt.V[game.Condition] {
	return opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerOpponent,
				CardSelection: game.Selection{ColorsAny: []color.Color{color.Blue, color.Black}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
}

// TestVeilConditionalDrawMatchesOpponentColoredCast proves the draw condition is
// satisfied only when an opponent cast a blue or black spell this turn.
func TestVeilConditionalDrawMatchesOpponentColoredCast(t *testing.T) {
	cases := []struct {
		name       string
		controller game.PlayerID
		colors     []color.Color
		want       bool
	}{
		{"opponent blue", game.Player2, []color.Color{color.Blue}, true},
		{"opponent black", game.Player2, []color.Color{color.Black}, true},
		{"opponent multicolor blue+red", game.Player2, []color.Color{color.Blue, color.Red}, true},
		{"opponent white", game.Player2, []color.Color{color.White}, false},
		{"opponent colorless", game.Player2, nil, false},
		{"own blue", game.Player1, []color.Color{color.Blue}, false},
	}
	cond := veilConditionalDraw()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			ctx := conditionContext{controller: game.Player1}
			if conditionSatisfied(g, ctx, cond) {
				t.Fatal("condition satisfied before any spell was cast")
			}
			emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: tc.controller, Colors: tc.colors})
			if got := conditionSatisfied(g, ctx, cond); got != tc.want {
				t.Fatalf("condition satisfied = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestVeilConditionalDrawIgnoresPriorTurns proves the current-turn window does
// not count spells cast on earlier turns.
func TestVeilConditionalDrawIgnoresPriorTurns(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	cond := veilConditionalDraw()

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2, Colors: []color.Color{color.Blue}})
	if !conditionSatisfied(g, ctx, cond) {
		t.Fatal("condition not satisfied after opponent cast a blue spell this turn")
	}

	g.Turn.TurnNumber++
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("condition satisfied from a spell cast on a prior turn")
	}
}

// applyPlayerHexproofFrom resolves Veil of Summer's player-scoped grant: "You
// gain hexproof from blue and from black until end of turn", controlled by
// controller and affecting the PlayerYou relation.
func applyPlayerHexproofFrom(engine *Engine, g *game.Game, controller game.PlayerID, colors ...color.Color) {
	obj := &game.StackObject{Controller: controller}
	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{
			{
				Kind:           game.RuleEffectPlayerHexproof,
				AffectedPlayer: game.PlayerYou,
				Protection:     game.ProtectionKeyword{FromColors: colors},
			},
		},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)
}

// TestVeilPlayerHexproofFromColors proves the player-scoped "hexproof from blue
// and black" grant blocks an opponent's blue or black source from targeting the
// player, allows other-colored opponent sources, and never blocks the player's
// own sources (CR 702.11e).
func TestVeilPlayerHexproofFromColors(t *testing.T) {
	cases := []struct {
		name          string
		sourceColors  []color.Color
		sourceOwner   game.PlayerID
		wantProtected bool
	}{
		{"opponent blue", []color.Color{color.Blue}, game.Player2, true},
		{"opponent black", []color.Color{color.Black}, game.Player2, true},
		{"opponent blue+white", []color.Color{color.Blue, color.White}, game.Player2, true},
		{"opponent white", []color.Color{color.White}, game.Player2, false},
		{"opponent colorless", nil, game.Player2, false},
		{"own blue", []color.Color{color.Blue}, game.Player1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			applyPlayerHexproofFrom(engine, g, game.Player1, color.Blue, color.Black)

			source := &game.CardDef{CardFace: game.CardFace{
				Name:   "Colored Spell",
				Types:  []types.Card{types.Instant},
				Colors: tc.sourceColors,
			}}
			target := game.PlayerTarget(game.Player1)
			got := targetProtectedFromSource(g, tc.sourceOwner, source, 0, target)
			if got != tc.wantProtected {
				t.Fatalf("targetProtectedFromSource = %v, want %v", got, tc.wantProtected)
			}
		})
	}
}

// TestVeilPlayerHexproofFromExpires proves the until-end-of-turn grant stops
// protecting once its rule effects are cleared.
func TestVeilPlayerHexproofFromExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	applyPlayerHexproofFrom(engine, g, game.Player1, color.Blue, color.Black)

	blueSource := &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Spell",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
	}}
	target := game.PlayerTarget(game.Player1)
	if !targetProtectedFromSource(g, game.Player2, blueSource, 0, target) {
		t.Fatal("opponent blue source could target the player while hexproof-from is active")
	}

	expireRuleEffects(g)
	if targetProtectedFromSource(g, game.Player2, blueSource, 0, target) {
		t.Fatal("hexproof-from should not persist past its until-end-of-turn duration")
	}
}

// addHexproofFromPermanent puts a creature with "hexproof from" the given colors
// onto the battlefield under controller, the ability Veil of Summer grants to
// each permanent its controller owns.
func addHexproofFromPermanent(g *game.Game, controller game.PlayerID, colors ...color.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Veiled Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.HexproofFromColorsStaticAbility(colors...)},
	}})
}

// TestVeilPermanentHexproofFromColors proves a permanent with "hexproof from
// blue and black" can't be targeted by an opponent's blue or black source,
// remains targetable by other-colored opponent sources, and is always
// targetable by its own controller's sources.
func TestVeilPermanentHexproofFromColors(t *testing.T) {
	cases := []struct {
		name          string
		sourceColors  []color.Color
		sourceOwner   game.PlayerID
		wantProtected bool
	}{
		{"opponent blue", []color.Color{color.Blue}, game.Player2, true},
		{"opponent black", []color.Color{color.Black}, game.Player2, true},
		{"opponent blue+green", []color.Color{color.Blue, color.Green}, game.Player2, true},
		{"opponent red", []color.Color{color.Red}, game.Player2, false},
		{"opponent colorless", nil, game.Player2, false},
		{"own blue", []color.Color{color.Blue}, game.Player1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			protected := addHexproofFromPermanent(g, game.Player1, color.Blue, color.Black)

			source := &game.CardDef{CardFace: game.CardFace{
				Name:   "Colored Spell",
				Types:  []types.Card{types.Instant},
				Colors: tc.sourceColors,
			}}
			target := game.PermanentTarget(protected.ObjectID)
			got := targetProtectedFromSource(g, tc.sourceOwner, source, 0, target)
			if got != tc.wantProtected {
				t.Fatalf("targetProtectedFromSource = %v, want %v", got, tc.wantProtected)
			}
		})
	}
}
