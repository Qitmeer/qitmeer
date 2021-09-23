package opreturn

type ShowAmount struct {
}

func (a *ShowAmount) GetType() OPReturnType {
	return ShowAmountType
}

func (a *ShowAmount) Verify() error {
	return nil
}

func (a *ShowAmount) Deserialize(data []byte) error {

}

func (a *ShowAmount) Serialize() ([]byte, error) {

}

func NewShowAmount() *ShowAmount {
	return &ShowAmount{}
}
