package main

import (
	"fmt"
	"math/rand"

	"agent-dou-dizhu/internal/tiandi/fsm"
)

func main() {
	machine := fsm.NewMachine(rand.New(rand.NewSource(42)))
	must(machine.Start())

	state := machine.Snapshot()
	fmt.Printf("phase=%s start=%s di_laizi=%s\n", state.Phase, state.StartingBidder, state.Laizi.Di)

	must(machine.Apply(fsm.Action{Seat: state.CurrentActor, Kind: fsm.ActionJiaoDiZhu}))
	state = machine.Snapshot()
	fmt.Printf("phase=%s actor=%s candidate=%s\n", state.Phase, state.CurrentActor, state.Candidate)

	must(machine.Apply(fsm.Action{Seat: state.CurrentActor, Kind: fsm.ActionQiangDiZhu}))
	state = machine.Snapshot()
	must(machine.Apply(fsm.Action{Seat: state.CurrentActor, Kind: fsm.ActionBuQiang}))
	state = machine.Snapshot()
	must(machine.Apply(fsm.Action{Seat: state.CurrentActor, Kind: fsm.ActionWoQiang}))

	state = machine.Snapshot()
	fmt.Printf(
		"phase=%s landlord=%s multiplier=%d tian=%s bottom=%d\n",
		state.Phase,
		state.Landlord,
		state.Multiplier,
		state.Laizi.Tian,
		len(state.Bottom),
	)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
