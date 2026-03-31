package play

import "fmt"

func CanPass(hasCurrentTrick bool) bool {
	return hasCurrentTrick
}

func Beats(next, current ResolvedHand) (bool, error) {
	if current.Kind == "" {
		return true, nil
	}
	if next.IsBomb && !current.IsBomb {
		return true, nil
	}
	if !next.IsBomb && current.IsBomb {
		return false, nil
	}
	if next.IsBomb && current.IsBomb {
		return compareBombs(next, current), nil
	}

	if isAttachmentCore(next) && isAttachmentCore(current) {
		return next.MainRank > current.MainRank, nil
	}
	if next.Kind != current.Kind {
		return false, nil
	}

	switch next.Kind {
	case "single", "pair", "triple", "triple_with_single", "triple_with_pair", "four_with_two_singles", "four_with_two_pairs":
		return next.MainRank > current.MainRank, nil
	case "straight", "serial_pairs":
		if next.Length != current.Length {
			return false, nil
		}
		return next.MainRank > current.MainRank, nil
	case "plane_base", "plane_with_singles", "plane_with_pairs":
		if next.GroupCount != current.GroupCount {
			return false, nil
		}
		return next.MainRank > current.MainRank, nil
	default:
		return false, fmt.Errorf("unsupported comparison kind %s", next.Kind)
	}
}

func isAttachmentCore(hand ResolvedHand) bool {
	switch hand.Kind {
	case "triple_with_single", "triple_with_pair", "four_with_two_singles", "four_with_two_pairs":
		return true
	default:
		return false
	}
}

func compareBombs(next, current ResolvedHand) bool {
	if next.BombTier != current.BombTier {
		return next.BombTier < current.BombTier
	}
	if next.Length != current.Length {
		return next.Length > current.Length
	}
	return next.MainRank > current.MainRank
}
