package domain

import "fmt"

const PlayerCount = 3

type Seat int

const (
	Seat0 Seat = iota
	Seat1
	Seat2
)

func AllSeats() []Seat {
	return []Seat{Seat0, Seat1, Seat2}
}

func (s Seat) Next() Seat {
	return Seat((int(s) + 1) % PlayerCount)
}

func (s Seat) String() string {
	switch s {
	case Seat0:
		return "P0"
	case Seat1:
		return "P1"
	case Seat2:
		return "P2"
	default:
		return "UnknownSeat"
	}
}

func ParseSeat(raw string) (Seat, error) {
	switch raw {
	case "P0":
		return Seat0, nil
	case "P1":
		return Seat1, nil
	case "P2":
		return Seat2, nil
	default:
		return Seat0, fmt.Errorf("unknown seat %q", raw)
	}
}
