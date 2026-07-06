package match

import "github.com/flykby/anonimus_chat/internal/shared"

type Route string

const (
	RouteAI  Route = "ai"
	RouteP2P Route = "p2p"
)

type Result struct {
	Route      Route
	MatchRoute string
}

func Resolve(gender, seeking shared.Gender) Result {
	switch {
	case gender == shared.GenderMale && seeking == shared.GenderFemale:
		return Result{Route: RouteAI, MatchRoute: "m_seeks_f"}
	case gender == shared.GenderMale && seeking == shared.GenderMale:
		return Result{Route: RouteP2P, MatchRoute: "m_seeks_m"}
	case gender == shared.GenderFemale && seeking == shared.GenderFemale:
		return Result{Route: RouteAI, MatchRoute: "f_seeks_f"}
	case gender == shared.GenderFemale && seeking == shared.GenderMale:
		return Result{Route: RouteP2P, MatchRoute: "f_seeks_m"}
	default:
		return Result{Route: RouteAI, MatchRoute: "unknown"}
	}
}
