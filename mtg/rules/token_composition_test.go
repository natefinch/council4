package rules

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestTokenCompositionNoncreatureAddendsDoNotEnterCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	addReplacementPermanent(t, g, game.Player1, crossTypeAddendCardDef())
	engine := NewEngine(nil)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateToken{
		Source: game.TokenDef(attackingTokenDef()), Amount: game.Fixed(1), EntryTapped: true, EntryAttacking: true,
	}, &TurnLog{})
	if len(g.Combat.Attackers) != 2 || countTokenPermanentsNamed(g, "Food") != 1 {
		t.Fatalf("replacement attack entry = %d attackers, %d Food; want 2/1",
			len(g.Combat.Attackers), countTokenPermanentsNamed(g, "Food"))
	}
	for _, declaration := range g.Combat.Attackers {
		permanent, _ := permanentByObjectID(g, declaration.Attacker)
		if !permanentHasType(g, permanent, types.Creature) {
			t.Fatal("noncreature addend was recorded as attacking")
		}
	}
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Food" {
			g.Combat.Attackers = nil
			declareCreatedTokensAttacking(engine, g, game.Player1, []*game.Permanent{permanent}, simPassPolicies(), &TurnLog{})
			declareTokenAttackingDefender(g, game.Player1, permanent, game.Player2)
			declareTokenAttackingTarget(g, game.Player1, permanent, game.AttackTarget{Player: game.Player2})
			if len(g.Combat.Attackers) != 0 {
				t.Fatal("an attack-entry helper accepted a noncreature")
			}
		}
	}
}

func TestTokenCompositionSingularCreationPreservesAmbientChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.ReplacementEffects = append(g.ReplacementEffects,
		game.ReplacementEffect{ID: g.IDGen.Next(), Description: "double", MatchEvent: game.EventTokenCreated, TokenMultiplier: 2},
		game.ReplacementEffect{ID: g.IDGen.Next(), Description: "add", MatchEvent: game.EventTokenCreated, TokenAddend: 1},
	)
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: tokenCompositionPolicy("add")}, &TurnLog{})
	defer g.ClearChoiceContext()
	if _, ok := createTokenPermanent(g, game.Player1, treasureTokenCardDef()); !ok {
		t.Fatal("singular token creation failed")
	}
	if got := countTokenPermanentsNamed(g, "Treasure"); got != 4 {
		t.Fatalf("ambient replacement choice produced %d tokens, want add-then-double 4", got)
	}
}

func TestTokenCompositionMultipleManufactors(t *testing.T) {
	for count := 1; count <= 3; count++ {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range count {
				addReplacementPermanent(t, g, game.Player1, academyManufactorReplacementCardDef())
			}

			if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, treasureTokenCardDef(), 1, false,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
				t.Fatal("token creation failed")
			}
			want := 1
			for range count - 1 {
				want *= 3
			}
			for _, name := range []string{"Clue", "Food", "Treasure"} {
				if got := countTokenPermanentsNamed(g, name); got != want {
					t.Fatalf("%s tokens = %d, want %d", name, got, want)
				}
			}
		})
	}
}

func TestTokenCompositionQuantityOrder(t *testing.T) {
	for _, first := range []string{"add", "double"} {
		t.Run(first, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.ReplacementEffects = append(g.ReplacementEffects,
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "add", MatchEvent: game.EventTokenCreated,
					TokenAddend: 1, TokenRequiredSubtypes: []types.Sub{types.Treasure}},
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "double", MatchEvent: game.EventTokenCreated, TokenMultiplier: 2},
			)
			policy := tokenCompositionPolicy(first)
			if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, treasureTokenCardDef(), 1, false,
				[game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{}) {
				t.Fatal("token creation failed")
			}
			want := 3
			if first == "add" {
				want = 4
			}
			if got := countTokenPermanentsNamed(g, "Treasure"); got != want {
				t.Fatalf("Treasure tokens = %d, want %d", got, want)
			}
		})
	}
}

func TestTokenCompositionIdentityRecomputesTypeMatching(t *testing.T) {
	for _, first := range []string{"identity", "double"} {
		t.Run(first, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			soldier := vanillaCreature("Soldier", 1, 1)
			g.ReplacementEffects = append(g.ReplacementEffects,
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "identity", MatchEvent: game.EventTokenCreated,
					TokenRequiredTypes: []types.Card{types.Creature}, TokenReplaceDef: treasureTokenCardDef()},
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "double", MatchEvent: game.EventTokenCreated,
					TokenRequiredTypes: []types.Card{types.Creature}, TokenMultiplier: 2},
			)
			if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, soldier, 1, false,
				[game.NumPlayers]PlayerAgent{game.Player1: tokenCompositionPolicy(first)}, &TurnLog{}) {
				t.Fatal("token creation failed")
			}
			want := 1
			if first == "double" {
				want = 2
			}
			if got := countTokenPermanentsNamed(g, "Treasure"); got != want || countTokenPermanentsNamed(g, "Soldier") != 0 {
				t.Fatalf("replacement Treasure tokens = %d, want %d without original Soldiers", got, want)
			}
		})
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.ReplacementEffects = append(g.ReplacementEffects,
		game.ReplacementEffect{ID: g.IDGen.Next(), MatchEvent: game.EventTokenCreated,
			TokenRequiredTypes: []types.Card{types.Artifact}, TokenReplaceDef: vanillaCreature("Soldier", 1, 1)},
		game.ReplacementEffect{ID: g.IDGen.Next(), MatchEvent: game.EventTokenCreated,
			TokenRequiredTypes: []types.Card{types.Creature}, TokenMultiplier: 2},
	)
	createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, treasureTokenCardDef(), 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if countTokenPermanentsNamed(g, "Soldier") != 2 {
		t.Fatal("initially inapplicable creature replacement was lost")
	}
}

func TestTokenCompositionCrossTypeAddendParticipatesOnce(t *testing.T) {
	for _, first := range []string{"add", "double"} {
		t.Run(first, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			food := &game.CardDef{CardFace: game.CardFace{Name: "Food", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Food}}}
			g.ReplacementEffects = append(g.ReplacementEffects,
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "add", MatchEvent: game.EventTokenCreated,
					TokenAddend: 1, TokenAddendDef: food},
				game.ReplacementEffect{ID: g.IDGen.Next(), Description: "double", MatchEvent: game.EventTokenCreated, TokenMultiplier: 2},
			)
			createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, vanillaCreature("Soldier", 1, 1), 1, true,
				[game.NumPlayers]PlayerAgent{game.Player1: tokenCompositionPolicy(first)}, &TurnLog{})
			wantFood := 1
			if first == "add" {
				wantFood = 2
			}
			if countTokenPermanentsNamed(g, "Soldier") != 2 || countTokenPermanentsNamed(g, "Food") != wantFood {
				t.Fatalf("cross-type composition = %d Soldiers/%d Food, want 2/%d",
					countTokenPermanentsNamed(g, "Soldier"), countTokenPermanentsNamed(g, "Food"), wantFood)
			}
			for _, permanent := range g.Battlefield {
				if !permanent.Tapped {
					t.Fatal("replacement-created token lost tapped-entry instructions")
				}
			}
		})
	}
}

func tokenCompositionPolicy(first string) *controlScript {
	return &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceReplacement {
			for _, option := range request.Options {
				if option.Label == first {
					return []int{option.Index}
				}
			}
		}
		return request.DefaultSelection
	}}
}

func TestTokenCompositionPreparesEntryBeforeNewReplacementSources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := vanillaCreature("Counter-granting token", 2, 2)
	token.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersGroupReplacement("Other creatures enter with vigilance.",
			&game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true},
			game.CounterPlacement{Kind: counter.Vigilance, Amount: 1}),
	}
	created, ok := createTokenPermanentsCollectingWithChoices(NewEngine(nil), g, game.Player1, token, 2, false,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok || len(created) != 2 {
		t.Fatal("token batch failed")
	}
	for _, permanent := range created {
		if permanent.Counters.Get(counter.Vigilance) != 0 {
			t.Fatal("a simultaneously entering token altered another token's entry")
		}
	}
}

func TestTokenCompositionAuthoredReceiptsAndResolvedSimulator(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 2 {
		addReplacementPermanent(t, g, game.Player1, academyManufactorReplacementCardDef())
	}
	engine := NewEngine(nil)
	ctx := &programContext{engine: engine, game: g, obj: &game.StackObject{Controller: game.Player1},
		active: true, log: &TurnLog{}}
	receipts, err := ctx.CreateTokens(treasureTokenCardDef(), cardprogram.TokenRecipient{Player: game.Player1, Count: 1})
	if err != nil || len(receipts) != 9 {
		t.Fatalf("actual authored receipts = %d, %v; want 9", len(receipts), err)
	}
	for _, receipt := range receipts {
		if _, found := permanentByObjectID(g, receipt.ObjectID); !found {
			t.Fatal("receipt does not identify an actual entrant")
		}
	}
	base := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 2 {
		addReplacementPermanent(t, base, game.Player1, academyManufactorReplacementCardDef())
	}
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Create a Treasure", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
			Source: game.TokenDef(treasureTokenCardDef()), Amount: game.Fixed(1),
		}}}}.Ability()),
	}}
	addImplementationSpellToStack(base, game.Player1, spell, nil)
	setMainPhasePriority(base, game.Player1)
	first := engine.Simulator().ResolvePriority(base, simPassPolicies())
	second := NewEngine(nil).Simulator().ResolvePriority(base, simPassPolicies())
	if !reflect.DeepEqual(first.Events, second.Events) || countTokenPermanentsNamed(base, "Treasure") != 0 {
		t.Fatal("token replacement simulation was not independent/reproducible")
	}
	for _, state := range []*game.Game{first, second} {
		for _, name := range []string{"Clue", "Food", "Treasure"} {
			if countTokenPermanentsNamed(state, name) != 3 {
				t.Fatal("resolved spell did not compose both Manufactors")
			}
		}
	}
}
