package native

// RunOwnedState uses the same C ABI, state checks and code/memory entry pins as
// RunState. Its owner must serialize every Context operation, including Close,
// except Cancel, and drain cancellation delivery before rebinding or closing.
// State and backing storage remain borrowed only for this call. A VM Store's
// exclusive operation/workspace ownership supplies this contract; ordinary
// users with concurrent Context access must use RunState.
func (c *Context) RunOwnedState(l *Lease, entry uint32, state *State) (Exit, error) {
	if state == nil {
		return 0, ErrInvalid
	}
	address, err := c.prepareEntry(l, entry, state)
	if err != nil {
		return 0, err
	}
	status := c.runEntry(l, address, c.record)
	*state = c.record.State
	if status < Return || status > Move {
		return 0, ErrInvalid
	}
	return status, nil
}
