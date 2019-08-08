package election

type Operation interface {
	TransferLeaderShip(id, address string) error
}
