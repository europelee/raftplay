package election

type Operation interface {
	TransferLeaderShip(address string) error
}
