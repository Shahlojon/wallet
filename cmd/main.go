package main

import (
	//"path/filepath"
	"os"
	"log"
	"github.com/Shahlojon/wallet/pkg/wallet"
	"fmt"
)

func main() {
	//fmt.Println("hello")
//	svc :=&wallet.Service{}
	//svc.RegisterAccount("+992000000001")
	//svc.Deposit(1, 10)
	//svc.RegisterAccount("+992000000002")
	//svc.ExportToFile("data/export.txt")
	//svc.ImportFromFile("data/export.txt")
	// account, err :=svc.RegisterAccount("+992000000001")
	// if err !=nil {
	// 	fmt.Println(account, err)
	// 	return
	// }

	// err = svc.Deposit(account.ID, 10)
	// if err !=nil {
	// 	switch err {
	// 	case wallet.ErrAmountMustBePositive:
	// 		fmt.Println("The summa must be positive")
	// 	case wallet.ErrAccountNotFound:
	// 		fmt.Println("The account user not found")
	// 	}
	// 	return
	// }
	// fmt.Println(account.Balance)
	// svc.RegisterAccount("+992000000001")
	// svc.Deposit(1,10)
	//wallet.RegisterAccount(svc, "+992000000001")
	// wallet.RegisterAccount(svc, "+992000000001")
	//svc.RegisterAccount("+992000000001")


	svc := &wallet.Service{}
	accountTest , err := svc.RegisterAccount("+992000000001")
	if err != nil {
		fmt.Println(err)
		return
	} 

	err = svc.Deposit(accountTest.ID, 100_000_00)
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")		
		}		
		return
	}
	fmt.Println(accountTest.Balance)

	err = svc.Deposit(accountTest.ID, 200_000_00)
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")		
		}		
		return
	}
	fmt.Println(accountTest.Balance)


	newPay, err := svc.Pay(accountTest.ID,10_000_00,"auto")
	fmt.Println(accountTest.Balance)
	fmt.Println(newPay)
	fmt.Println(err)

	fav, errFav := svc.FavoritePayment(newPay.ID, "Babilon")
	fmt.Println(errFav)
	fmt.Println(fav)
   
	wd, err := os.Getwd()
	if err != nil {
		log.Print(err)
		return
	}

	err = svc.Import(wd)
	if err != nil {
	 	log.Print(err)
	 	return
	}
}