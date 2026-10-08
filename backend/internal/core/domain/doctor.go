package domain

// Doctor is an entry of the (fictitious) medical registry.
type Doctor struct {
	RegistryID         string
	Name               string
	Active             bool
	EnabledToPrescribe bool // habilitado para prescribir en Ecuador
}
