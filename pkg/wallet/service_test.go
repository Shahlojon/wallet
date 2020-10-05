package wallet

import (
	"reflect"
	"testing"

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

// func TestService_Reject_success(t *testing.T)  {
// 	s := &Service{}

// 	phone := types.Phone("+9920000001")
// 	account, err := s.RegisterAccount(phone)
// 	if err != nil {
// 		t.Errorf("Reject(): can't register account, error = %v", err)
// 		return
// 	}

// 	err = s.Deposit(account.ID, 10_000_00)
// 	if err != nil {
// 		t.Errorf("Reject(): can't deposit account, error = %v", err)
// 		return
// 	}

// 	payment, err := s.Pay(account.ID, 1000_00, "auto")
// 	if err != nil {
// 		t.Errorf("Reject(): can't create payment, error = %v", err)
// 		return
// 	}

// 	err = s.Reject(payment.ID)
// 	if err != nil {
// 		t.Errorf("Reject(): error = %v", err)
// 		return
// 	}
// }

// func TestService_FindPaymentByID_success(t *testing.T)  {
// 	s := &Service{}

// 	phone := types.Phone("+9920000001")
// 	account, err := s.RegisterAccount(phone)
// 	if err != nil {
// 		t.Errorf("Reject(): can't register account, error = %v", err)
// 		return
// 	}

// 	err = s.Deposit(account.ID, 10_000_00)
// 	if err != nil {
// 		t.Errorf("Reject(): can't deposit account, error = %v", err)
// 		return
// 	}

// 	payment, err := s.Pay(account.ID, 1000_00, "auto")
// 	if err != nil {
// 		t.Errorf("Reject(): can't create payment, error = %v", err)
// 		return
// 	}

// 	got, err := s.FindPaymentByID(payment.ID)
// 	if err != nil {
// 		t.Errorf("FindPaymentByID(): error = %v", err)
// 		return
// 	}

// 	if !reflect.DeepEqual(payment, got) {
// 		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
// 	}
// }

// func TestService_FindPaymentByID_fail(t *testing.T)  {
// 	//создаём сервис
// 	s := newTestService();

// 	account, err := s.addAccountWithBalnce("+99200000001", 10_000_00)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	//осушествляем платёж на его счёт
// 	payment, err := s.Pay(account.ID, 1000_00, "auto")
// 	if err != nil {
// 		t.Errorf("FindPaymentByID: can't create	payment, error = %v", err)
// 		return
// 	}
// 	//пробуем найти несушествующий платёж
// 	got, err := s.FindPaymentByID(uuid.New().String())
// 	if err != nil {
// 		t.Error("FindPaymentByID: must return error, returned nil")
// 		return
// 	}

// 	if err!=ErrPaymentNotFound {
// 		t.Errorf("FindPaymentByID(): must return ErrPaymnetNotFound, returned = %v", err)
// 	}
// }

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