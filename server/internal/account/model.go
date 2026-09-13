package account

type MissionPermissions struct {
	IsOwner         bool
	IsMissionPublic bool
}

type PairingRequest struct {
	EncryptionKey string
	AccountKey    string
}
