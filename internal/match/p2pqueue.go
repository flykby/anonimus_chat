package match

import "github.com/flykby/anonimus_chat/internal/shared"

type p2pQueueKind int

const (
	p2pQueueSameGender p2pQueueKind = iota
	p2pQueueHeteroFemale
)

type p2pQueueTarget struct {
	kind   p2pQueueKind
	gender shared.Gender
}

func p2pQueueFor(gender, seeking shared.Gender) p2pQueueTarget {
	switch {
	case gender == shared.GenderMale && seeking == shared.GenderMale:
		return p2pQueueTarget{kind: p2pQueueSameGender, gender: shared.GenderMale}
	case gender == shared.GenderFemale && seeking == shared.GenderMale:
		return p2pQueueTarget{kind: p2pQueueHeteroFemale, gender: shared.GenderFemale}
	default:
		return p2pQueueTarget{kind: p2pQueueSameGender, gender: gender}
	}
}
