package election

type Operation interface {
	TransferLeaderShip(id, address string) error
	Join(id, address string) error
	RemoveSvr(id string) error
}
