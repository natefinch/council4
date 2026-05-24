package game

// CostModifierKind identifies which costs a modifier applies to.
type CostModifierKind int

const (
	CostModifierSpell CostModifierKind = iota
	CostModifierAbility
	CostModifierAttack
)

// CostModifier is a generic-cost increase/reduction/set effect.
type CostModifier struct {
	Kind             CostModifierKind
	Controller       PlayerID
	MatchCardType    bool
	CardType         CardType
	GenericIncrease  int
	GenericReduction int
	SetGeneric       *int
	MinimumGeneric   int
}

// AttackTax is an additional generic mana cost to attack a player.
type AttackTax struct {
	DefendingPlayer PlayerID
	Amount          int
}
