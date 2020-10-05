package main

import (
	"github.com/Shahlojon/wallet/pkg/wallet"
	"fmt"
)

func main() {
	fmt.Println("hello")
	svc :=&wallet.Service{}
	account, err :=svc.RegisterAccount("+992000000001")
	if err !=nil {
		fmt.Println(account, err)
		return
	}

	err = svc.Deposit(account.ID, 10)
	if err !=nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("The summa must be positive")
		case wallet.ErrAccountNotFound:
			fmt.Println("The account user not found")
		}
		return
	}
	fmt.Println(account.Balance)
	// svc.RegisterAccount("+992000000001")
	// svc.Deposit(1,10)
	//wallet.RegisterAccount(svc, "+992000000001")
	// wallet.RegisterAccount(svc, "+992000000001")
	//svc.RegisterAccount("+992000000001")
}