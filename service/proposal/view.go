package proposal

type (
	Item struct {
		Key    string
		Value  string
		Action string
	}

	Proposal struct {
		Number int64
		Action string
		Info   []Item
		Meta   []Item

		ApprovedCount int
		ApprovedBy    string
	}
)
