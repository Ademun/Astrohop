package account

type MissionPermissions struct {
	IsOwner bool
}

type PairingRequest struct {
	EncryptionKey string
	AccountKey    string
}
