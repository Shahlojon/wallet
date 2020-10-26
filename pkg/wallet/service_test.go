package wallet

import (
	"os"
	"reflect"
	"testing"
	"log"
	"github.com/Shahlojon/wallet/pkg/types"
	"github.com/google/uuid"
)

func TestService_RegisterAccount_unsuccess(t *testing.T) {
	vc := Service{}

	accounts := []types.Account{
		{ID: 1, Phone: "+992000000001", Balance: 2_000_00},
		{ID: 2, Phone: "+992000000002", Balance: 3_000_00},
		{ID: 3, Phone: "+992000000003", Balance: 4_000_00},
		{ID: 4, Phone: "+992000000004", Balance: 5_000_00},
		{ID: 5, Phone: "+992000000005", Balance: 6_000_00},
		{ID: 6, Phone: "+992000000006", Balance: 7_000_00},
	}
	result, err := vc.RegisterAccount("+992000000007")
	for _, account := range accounts {
		if account.Phone == result.Phone {
			t.Errorf("invalid result, expected: %v, actual: %v", err, result)
			break
		}
	}
}

func TestService_FindAccoundById_Method_NotFound(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(3)
	if err == nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestService_FindPaymentByID_success(t *testing.T) {
	//создаем сервис
	s := newTestServiceUser()
	_, payments, err := s.addAccountUser(defaultTestAccountUser)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платеж
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	//создаем сервис
	s := newTestServiceUser()
	_, _, err := s.addAccountUser(defaultTestAccountUser)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти несуществующий платеж
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID: must return error, returned nil")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymnetNotFound, returned = %v", err)
		return
	}
}

func TestService_Reject_fail(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	editPayID := pay.ID + "l"
	err = svc.Reject(editPayID)
	if err == nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}
}

func TestService_Reject_success(t *testing.T) {
	//создаем сервис
	s := newTestServiceUser()
	_, payments, err := s.addAccountUser(defaultTestAccountUser)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем отменить платёж
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): status didn't changed, paymnet = %v", savedPayment)
		return
	}
	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccountUser.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}
}


func TestService_Repeat_success_user(t *testing.T) {
	//создаем сервис
	s := newTestServiceUser()
	s.RegisterAccount("+9922000000")
	account, err :=s.FindAccountByID(1)
	if err != nil {
		t.Error(err)
		return
	}
	//пополняем баланс
	err = s.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}
	//pay
	payment, err := s.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err = s.Repeat(pay.ID)
	if err != nil {
		t.Errorf("Repeat(): can't payment for an account(%v), error(%v)",pay.ID, err)
	}
}

func TestService_Export_success_user(t *testing.T) {
	var svc Service

	svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")
	svc.RegisterAccount("+992000000003")

	err := svc.ExportToFile("export.txt")
	if err != nil {
		t.Errorf("method ExportToFile returned not nil error, err => %v", err)
	}

}

func TestService_Import_success_user(t *testing.T) {
	var svc Service


	err := svc.ImportFromFile("export.txt")
	
	if err != nil {
		t.Errorf("method ExportToFile returned not nil error, err => %v", err)
	}

}



func TestService_ExportImport_success_user(t *testing.T) {
	var svc Service

	svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")
	svc.RegisterAccount("+992000000003")
	svc.RegisterAccount("+992000000004")

	wd,_:= os.Getwd()
	err := svc.Export(wd)
	if err != nil {
		t.Errorf("method Export returned not nil error, err => %v", err)
	}
	err = svc.Import(wd)
	if err != nil {
		t.Errorf("method Import returned not nil error, err => %v", err)
	}
}


func TestService_ExportHistory_success_user(t *testing.T) {
	var svc Service

	account, err := svc.RegisterAccount("+992000000001")

	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	_, err = svc.Pay(account.ID, 1, "Cafe")
	_, err = svc.Pay(account.ID, 2, "Cafe")
	_, err = svc.Pay(account.ID, 3, "Cafe")
	_, err = svc.Pay(account.ID, 4, "Cafe")
	_, err = svc.Pay(account.ID, 5, "Cafe")
	_, err = svc.Pay(account.ID, 6, "Cafe")
	_, err = svc.Pay(account.ID, 7, "Cafe")
	_, err = svc.Pay(account.ID, 8, "Cafe")
	_, err = svc.Pay(account.ID, 9, "Cafe")
	_, err = svc.Pay(account.ID, 10, "Cafe")
	_, err = svc.Pay(account.ID, 11, "Cafe")
	if err != nil {
		t.Errorf("method Pay returned not nil error, err => %v", err)
	}

	payments, err := svc.ExportAccountHistory(account.ID)

	if err != nil {
		t.Errorf("method ExportAccountHistory returned not nil error, err => %v", err)
	}
	err = svc.HistoryToFiles(payments, "data", 4)

	if err != nil {
		t.Errorf("method HistoryToFiles returned not nil error, err => %v", err)
	}

}

func TestService_ExportHistory(t *testing.T) {
	svc := &Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
	}

	svc.Pay(account.ID, 10_00, "auto")

	payment, err := svc.ExportAccountHistory(1)
	if err != nil {
		t.Error(err)
	}
	err = svc.HistoryToFiles(payment, "data", 2)
	if err != nil {
		t.Error(err)
	}
}
func TestService_FavoritePayment_success_user(t *testing.T) {
	//создаем сервис
	var s Service

	account, err := s.RegisterAccount("+9922000000")
	if err != nil {
		t.Errorf("method RegisterAccount return not nil error, account=>%v", account)
		return
	}
	//пополняем баланс
	err = s.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("method Deposit return not nil error, error=>%v", err)
	}
	//pay
	payment, err := s.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("method Pay return not nil error, account=>%v", account)
	}
	//edit favorite
	favorite, err := s.FavoritePayment(payment.ID, "auto")
	if err != nil {
		t.Errorf("method FavoritePayment returned not nil error, favorite=>%v", favorite)
	}

	paymentFavorite, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("method PayFromFavorite returned nil, paymentFavorite(%v)", paymentFavorite)
	}
}

func BenchmarkSumPayment(b *testing.B){
	var svc Service

	account, err := svc.RegisterAccount("+992000000001")

	if err != nil {
		b.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		b.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	_, err = svc.Pay(account.ID, 1, "Cafe")
	_, err = svc.Pay(account.ID, 2, "Cafe")
	_, err = svc.Pay(account.ID, 3, "Cafe")
	_, err = svc.Pay(account.ID, 4, "Cafe")
	_, err = svc.Pay(account.ID, 5, "Cafe")
	_, err = svc.Pay(account.ID, 6, "Cafe")
	_, err = svc.Pay(account.ID, 7, "Cafe")
	_, err = svc.Pay(account.ID, 8, "Cafe")
	_, err = svc.Pay(account.ID, 9, "Cafe")
	_, err = svc.Pay(account.ID, 10, "Cafe")
	_, err = svc.Pay(account.ID, 11, "Cafe")
	if err != nil {
		b.Errorf("method Pay returned not nil error, err => %v", err)
	}

	want := types.Money(66)

	got := svc.SumPayments(2)
	if want != got{
		b.Errorf(" error, want => %v got => %v", want, got)
	}
}

func BenchmarkFilterPayments(b *testing.B) {
	svc := &Service{}

	account, err := svc.RegisterAccount("+992000000000")
	account1, err := svc.RegisterAccount("+992000000001")
	account2, err := svc.RegisterAccount("+992000000002")
	account3, err := svc.RegisterAccount("+992000000003")
	account4, err := svc.RegisterAccount("+992000000004")
	acc, err := svc.RegisterAccount("+992000000005")
	if err != nil {
	}
	svc.Deposit(acc.ID, 100)
	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
	}

	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account2.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account3.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account3.ID, 10_00, "auto")
	svc.Pay(account2.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")

	a, err := svc.FilterPayments(account.ID, 5)
	if err != nil {
		b.Error(err)
	}
	log.Println(len(a))
}

func BenchmarkService_FilterPaymentsByFn(b *testing.B) {
	svc := &Service{}
	filter := func(payment types.Payment) bool {
		for _, value := range svc.payments {
			if payment.ID == value.ID {
				return true
			}
		}
		return false
	}
	account, err := svc.RegisterAccount("+992000000000")
	account1, err := svc.RegisterAccount("+992000000001")
	account2, err := svc.RegisterAccount("+992000000002")
	account3, err := svc.RegisterAccount("+992000000003")
	account4, err := svc.RegisterAccount("+992000000004")
	acc, err := svc.RegisterAccount("+992000000005")
	if err != nil {
	}
	svc.Deposit(acc.ID, 100)
	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
	}

	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account2.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account3.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account1.ID, 10_00, "auto")
	svc.Pay(account3.ID, 10_00, "auto")
	svc.Pay(account2.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	svc.Pay(account4.ID, 10_00, "auto")
	a, err := svc.FilterPaymentsByFn(filter, 4)
	if err != nil {
		b.Error(err)
	}
	log.Println(a)
}

func BenchmarkSumPaymentsWithProgress_user(b *testing.B) {
	var svc Service

	account, err := svc.RegisterAccount("+992000000001")

	if err != nil {
		b.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 10000000_0000000)
	if err != nil {
		b.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	// for i := 0; i < 10000; i++ {
	// 	svc.Pay(account.ID, types.Money(i), "Cafe")
	// } 

	// ch := svc.SumPaymentsWithProgress()

	

	sum :=svc.SumPayments(2)
	//  s, ok := <-ch
	total := types.Money(0)
	for i := range svc.SumPaymentsWithProgress() {
		total += i.Result
	}
	if sum != total{
		b.Error("Errorrrorororo")
	}
	// if !ok {
	// 	b.Errorf(" method SumPaymentsWithProgress ok not closed => %v", ok)
	// }

	// log.Println("=======>>>>>",s) 
}

func TestService_SumPayments(b *testing.T) {
	svc := &Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
	}

	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	svc.Pay(account.ID, 10_00, "auto")
	want := types.Money(7000)

	got := svc.SumPayments(5)
	if want != got {
		b.Errorf(" error, want => %v got => %v", want, got)
	}
}

