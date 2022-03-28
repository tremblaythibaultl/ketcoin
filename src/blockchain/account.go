package blockchain

type Account struct {
	Address string
	Balance uint
}

func (a *Account) add(amount uint) {
	a.Balance += amount
}
