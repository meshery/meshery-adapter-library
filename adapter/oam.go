package adapter

import "context"

// ProcessOAM processes OAM components. This is adapter specific and needs to be implemented by each adapter.
func (h *Adapter) ProcessOAM(context.Context, OAMRequest) error {
	return nil
}
