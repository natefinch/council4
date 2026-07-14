package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// underworldBreachDef mirrors the CardDef that cardgen lowers Underworld Breach
// to: a static ability granting escape to each nonland card in the controller's
// graveyard, with the computed escape cost equal to the card's own mana cost
// plus exiling three other cards from the controller's graveyard.
func underworldBreachDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Underworld Breach",
		Types:    []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectGrantGraveyardCardKeyword,
				AffectedPlayer: game.PlayerYou,
				CardSelection:  game.Selection{ExcludedTypes: []types.Card{types.Land}},
				GrantedKeyword: game.Escape,
				GraveyardCastCost: game.GraveyardCastGrantCost{
					UseCardManaCost: true,
					AdditionalCosts: []cost.Additional{{
						Kind:          cost.AdditionalExile,
						Text:          "Exile three other cards from your graveyard",
						Amount:        3,
						Source:        zone.Graveyard,
						ExcludeSource: true,
					}},
				},
			}},
		}},
	}}
}

// retraceGrantDef mirrors the CardDef that cardgen lowers Six to: a static
// ability granting retrace to nonland permanent cards in the controller's
// graveyard during the controller's turn, carrying no computed cost (retrace's
// cost is intrinsic: the card's mana cost plus discarding a land card).
func retraceGrantDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Retrace Source",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                           game.RuleEffectGrantGraveyardCardKeyword,
				AffectedPlayer:                 game.PlayerYou,
				RestrictedDuringControllerTurn: true,
				CardSelection: game.Selection{
					RequiredTypesAny: []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle},
					ExcludedTypes:    []types.Card{types.Land},
				},
				GrantedKeyword: game.Retrace,
			}},
		}},
	}}
}

func graveyardNonlandSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
	}}
}

func graveyardLandCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Land",
		Types: []types.Card{types.Land},
	}}
}

// TestGrantedGraveyardEscapeSynthesizesComputedCost verifies the runtime
// synthesizes an escape alternative that uses the graveyard card's own mana cost
// (CR 702.139) plus the grant's computed additional cost.
func TestGrantedGraveyardEscapeSynthesizesComputedCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	spell := graveyardNonlandSpell()

	alternatives := grantedGraveyardCastAlternatives(g, game.Player1, spell)
	if len(alternatives) != 1 {
		t.Fatalf("granted alternatives = %d, want 1", len(alternatives))
	}
	alt := alternatives[0]
	if alt.Mechanic != cost.AlternativeMechanicEscape {
		t.Fatalf("mechanic = %v, want escape", alt.Mechanic)
	}
	if !alt.ManaCost.Exists || !reflect.DeepEqual(alt.ManaCost.Val, spell.ManaCost.Val) {
		t.Fatalf("escape mana cost = %+v, want the card's own %+v", alt.ManaCost, spell.ManaCost.Val)
	}
	want := []cost.Additional{{
		Kind:          cost.AdditionalExile,
		Text:          "Exile three other cards from your graveyard",
		Amount:        3,
		Source:        zone.Graveyard,
		ExcludeSource: true,
	}}
	if !reflect.DeepEqual(alt.AdditionalCosts, want) {
		t.Fatalf("additional costs = %+v, want %+v", alt.AdditionalCosts, want)
	}
}

// TestGrantedGraveyardEscapeAddsEscapePermission verifies the graveyard-cast
// permission set gains escape for a card that only a grant makes escapable.
func TestGrantedGraveyardEscapeAddsEscapePermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	cardID := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())

	perms := castPermissionsForZone(g, game.Player1, cardID, zone.Graveyard, game.FaceFront)
	if !slices.Contains(perms, payment.SpellCastPermissionEscape) {
		t.Fatalf("permissions = %v, want escape present", perms)
	}
}

// TestGrantedGraveyardEscapeExcludesLandCards verifies the nonland selection: a
// land card in the graveyard is not granted escape.
func TestGrantedGraveyardEscapeExcludesLandCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, underworldBreachDef())

	if alternatives := grantedGraveyardCastAlternatives(g, game.Player1, graveyardLandCard()); len(alternatives) != 0 {
		t.Fatalf("granted alternatives for land = %d, want 0", len(alternatives))
	}
}

// TestGrantedGraveyardGrantAppliesOnlyToController verifies the "your graveyard"
// scoping: only the grant's controller may cast their graveyard cards, not an
// opponent from theirs.
func TestGrantedGraveyardGrantAppliesOnlyToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	spell := graveyardNonlandSpell()

	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, spell)); got != 1 {
		t.Fatalf("controller alternatives = %d, want 1", got)
	}
	if got := len(grantedGraveyardCastAlternatives(g, game.Player2, spell)); got != 0 {
		t.Fatalf("opponent alternatives = %d, want 0", got)
	}
}

// TestGrantedGraveyardGrantStopsWhenSourceLeaves verifies the grant is
// continuous: once its source permanent leaves the battlefield the alternative
// is no longer offered.
func TestGrantedGraveyardGrantStopsWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	breach := addPermanentForSBA(g, game.Player1, underworldBreachDef())
	spell := graveyardNonlandSpell()

	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, spell)); got != 1 {
		t.Fatalf("alternatives while source present = %d, want 1", got)
	}
	if !movePermanentToZone(g, breach, zone.Graveyard) {
		t.Fatal("moving grant source to graveyard failed")
	}
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, spell)); got != 0 {
		t.Fatalf("alternatives after source left = %d, want 0", got)
	}
}

// TestGrantedGraveyardEscapeFailsClosedWithoutComputedCost verifies an escape
// grant that carries no computed cost synthesizes nothing rather than an
// unpayable or free cast.
func TestGrantedGraveyardEscapeFailsClosedWithoutComputedCost(t *testing.T) {
	def := underworldBreachDef()
	def.StaticAbilities[0].RuleEffects[0].GraveyardCastCost = game.GraveyardCastGrantCost{}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, def)

	if alternatives := grantedGraveyardCastAlternatives(g, game.Player1, graveyardNonlandSpell()); len(alternatives) != 0 {
		t.Fatalf("granted alternatives = %d, want 0 (fail closed)", len(alternatives))
	}
}

// TestGrantedEscapeCastsFromGraveyardPayingComputedCost is the end-to-end
// Underworld Breach path: a nonland card is cast from the graveyard for its own
// mana cost plus exiling three other graveyard cards, and — like native escape —
// returns to the graveyard on resolution so it can be recast.
func TestGrantedEscapeCastsFromGraveyardPayingComputedCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	cardID := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())
	fuel := []id.ID{
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}}),
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}}),
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Three"}}),
	}
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted escape cast from graveyard failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escaping card was not removed from the graveyard when cast")
	}
	for _, f := range fuel {
		if !g.Players[game.Player1].Exile.Contains(f) {
			t.Fatalf("fuel card %d was not exiled as part of the escape cost", f)
		}
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want granted escape graveyard cast", obj)
	}
	if obj.Flashback {
		t.Fatal("granted escape must not be marked flashback (it is not exiled on resolution)")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("granted escape card was exiled on resolution; escape returns it to the graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("granted escape card did not return to the graveyard after resolving")
	}
}

// TestGrantedEscapeRequiresThreeOtherCards verifies the exclude-source and
// insufficient-graveyard behavior: with only two other cards the escape cost is
// unpayable because the escaping card cannot pay its own exile cost.
func TestGrantedEscapeRequiresThreeOtherCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	cardID := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted escape succeeded with only two other cards; the escaping card must not pay its own exile cost")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escaping card left the graveyard despite an unpayable cost")
	}
}

// TestGrantedEscapeObjectIdentityExcludesOnlySourceCopy verifies the exile cost
// excludes the escaping card by object identity, so another card sharing its
// definition is still valid fuel.
func TestGrantedEscapeObjectIdentityExcludesOnlySourceCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	cardID := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())
	sibling := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted escape cast failed with a same-name sibling available as fuel")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("the escaping card paid its own exile cost")
	}
	if !g.Players[game.Player1].Exile.Contains(sibling) {
		t.Fatal("a distinct card sharing the escaping card's definition should be valid fuel")
	}
}

// TestGrantedEscapeCounteredGoesToGraveyard verifies a countered granted-escape
// spell goes to its owner's graveyard, not exile (it is not a flashback cast).
func TestGrantedEscapeCounteredGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	cardID := addCardToGraveyard(g, game.Player1, graveyardNonlandSpell())
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Three"}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted escape cast from graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("granted escape spell was not put on the stack")
	}
	if !counterStackObject(g, obj.ID) {
		t.Fatal("counterStackObject(granted escape) = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("countered granted-escape spell did not go to the graveyard")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("countered granted-escape spell was exiled")
	}
}

// TestGrantedEscapeFizzledGoesToGraveyard verifies a granted-escape spell whose
// only target became illegal is countered on resolution to the graveyard, not
// exile.
func TestGrantedEscapeFizzledGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, underworldBreachDef())
	spell := graveyardNonlandSpell()
	spell.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target creature",
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
		}},
		Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
	}.Ability())
	cardID := addCardToGraveyard(g, game.Player1, spell)
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Three"}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	target := addCreaturePermanent(g, game.Player2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("granted escape cast with a target failed")
	}
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("removing the spell's target failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("fizzled granted-escape spell did not go to the graveyard")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("fizzled granted-escape spell was exiled")
	}
}

// TestGrantedRetraceCastsPermanentFromGraveyard verifies the retrace-style grant
// (Six) is no longer a runtime no-op: a nonland permanent card can be cast from
// the graveyard by paying its mana cost and discarding a land card, entering the
// battlefield on resolution.
func TestGrantedRetraceCastsPermanentFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, retraceGrantDef())
	creature := &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Bear",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
	}}
	cardID := addCardToGraveyard(g, game.Player1, creature)
	addCardToHand(g, game.Player1, graveyardLandCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted retrace cast from graveyard failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("retraced card was not removed from the graveyard when cast")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want granted retrace graveyard cast", obj)
	}
	if obj.Flashback {
		t.Fatal("granted retrace must not be marked flashback")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !permanentForCardOnBattlefield(g, cardID) {
		t.Fatal("retraced permanent did not enter the battlefield on resolution")
	}
}

// TestGrantedRetraceRequiresLandToDiscard verifies the retrace grant is
// unpayable without a land card to discard.
func TestGrantedRetraceRequiresLandToDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, retraceGrantDef())
	creature := &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Bear",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
	}}
	cardID := addCardToGraveyard(g, game.Player1, creature)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Not A Land", Types: []types.Card{types.Instant}}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("granted retrace cast succeeded without a land card to discard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("retrace card left the graveyard despite an unpayable discard cost")
	}
}

func permanentForCardOnBattlefield(g *game.Game, cardID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return true
		}
	}
	return false
}
