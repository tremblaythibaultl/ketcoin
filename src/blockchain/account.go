package blockchain

type Account struct {
	address [32]byte
	balance uint
}

func (a *Account) add(amount uint) {
	a.balance += amount
}
