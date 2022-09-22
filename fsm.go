package fsm

import (
	"fmt"
	"sort"
	"sync"
)

func New() *StateMachine {
	return &StateMachine{
		states: map[State]*stateDescriptor{},
	}
}

type StateMachine struct {
	mu     sync.Mutex
	state  State
	states map[State]*stateDescriptor
}

func (sm *StateMachine) SetState(state State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.state == state {
		return nil
	}
	ns, ok := sm.states[state]
	if !ok {
		return fmt.Errorf("%q does not exist", state)
	}

	if sm.state != "" {
		cs := sm.states[sm.state]
		cs.onExit()
	}
	sm.state = state
	ns.onEntry()
	return nil
}

func (sm *StateMachine) State() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

func (sm *StateMachine) States() States {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	states := map[State]bool{}
	for source, s := range sm.states {
		states[source] = true
		for _, target := range s.transitions {
			states[target] = true
		}
	}
	var res []State
	for s := range states {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

func (sm *StateMachine) OutStates(source State) States {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, ok := sm.states[source]
	if !ok {
		return nil
	}
	states := map[State]bool{}
	for _, target := range s.transitions {
		states[target] = true
	}
	var res []State
	for s := range states {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

func (sm *StateMachine) AddTransition(source State, input Event, target State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.getOrCreateState(target)
	s := sm.getOrCreateState(source)
	s.transitions[input] = target
	return nil
}

func (sm *StateMachine) SetOnEntry(state State, action Action) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s := sm.getOrCreateState(state)
	if action != nil {
		s.onEntry = action
	}
}

func (sm *StateMachine) SetOnExit(state State, action Action) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s := sm.getOrCreateState(state)
	if action != nil {
		s.onExit = action
	}
}

func (sm *StateMachine) ValidEvents(state State) []Event {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, ok := sm.states[state]
	if !ok {
		return nil
	}
	var res []Event
	for e := range s.transitions {
		res = append(res, e)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

func (sm *StateMachine) IsValidEvent(state State, event Event) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, ok := sm.states[state]
	if !ok {
		return false
	}
	_, ok = s.transitions[event]
	return ok
}

func (sm *StateMachine) Data(state State) Data {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s := sm.getOrCreateState(state)
	return s.data
}

func (sm *StateMachine) Process(event Event) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.state == "" {
		return fmt.Errorf("no current state is set")
	}
	s, ok := sm.states[sm.state]
	if !ok {
		return fmt.Errorf("%q does not exist", sm.state)
	}
	if len(s.transitions) == 0 {
		return fmt.Errorf("%q does not have transitions", s.name)
	}
	target, ok := s.transitions[event]
	if !ok {
		return fmt.Errorf("%q does not accept event %q", s.name, event)
	}
	t, ok := sm.states[target]
	if !ok {
		return fmt.Errorf("%q does not exist", target)
	}

	s.onExit()
	sm.state = target
	t.onEntry()

	return nil
}

func (sm *StateMachine) getOrCreateState(state State) *stateDescriptor {
	if s, ok := sm.states[state]; ok {
		return s
	}
	s := newStateDescriptor(state)
	sm.states[state] = s
	return s
}

type State string

type States []State

func (ss States) Contains(s State) bool {
	for _, cs := range ss {
		if cs == s {
			return true
		}
	}
	return false
}

type Event string

type Action func()

type Data map[string]any

func (d Data) String(key string) string {
	v, ok := d[key]
	if !ok {
		return ""
	}
	res, ok := v.(string)
	if !ok {
		return ""
	}
	return res
}

func (d Data) Strings(key string) []string {
	v, ok := d[key]
	if !ok {
		return nil
	}
	res, ok := v.([]string)
	if !ok {
		return nil
	}
	return res
}

func (d Data) Int(key string) int {
	v, ok := d[key]
	if !ok {
		return 0
	}
	res, ok := v.(int)
	if !ok {
		return 0
	}
	return res
}

func (d Data) Ints(key string) []int {
	v, ok := d[key]
	if !ok {
		return nil
	}
	res, ok := v.([]int)
	if !ok {
		return nil
	}
	return res
}

func (d Data) Bool(key string) bool {
	v, ok := d[key]
	if !ok {
		return false
	}
	res, ok := v.(bool)
	if !ok {
		return false
	}
	return res
}

func NOOP() {
}

func newStateDescriptor(name State) *stateDescriptor {
	return &stateDescriptor{
		name:        name,
		transitions: map[Event]State{},
		data:        map[string]any{},
		onEntry:     NOOP,
		onExit:      NOOP,
	}
}

type stateDescriptor struct {
	name        State
	transitions map[Event]State
	data        map[string]any
	onEntry     Action
	onExit      Action
}
