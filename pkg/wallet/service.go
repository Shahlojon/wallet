package wallet

import (
	"strconv"
	"os"
	"log"
	"errors"
	"fmt"

	"github.com/Shahlojon/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrFavoriteNotFound = errors.New("favorite not found")
var ErrAmountMustBePositive = errors.New("amount must be greater than 0")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("balance is null")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFileNotFound = errors.New("file not found")

type Service struct {
	nextAccountID int64 //Для генерации уникального номера аккаунта
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account
	for _, accounts := range s.accounts {
		if accounts.ID == accountID {
			account = accounts
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}


func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	
	account, err := s.FindAccountByID(payment.AccountID)

	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

//FindPaymentByID
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

type testServiceUser struct {
	*Service
}

//Функция конструктор
func newTestServiceUser() *testServiceUser {
	return &testServiceUser{Service: &Service{}}
}

type testAccountUser struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccountUser = testAccountUser{
	phone:   "+99200000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func (s *testServiceUser) addAccountUser(data testAccountUser) (*types.Account, []*types.Payment, error) {
	//регистрируем там пользователя
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	//пополняем его счет
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposity account, error = %v", err)
	}

	//выпоняем платежи
	//можем создать слайс сразу
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		//тогда здесь работаем просто через index, а не через append
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, nil
}

//Repeat-позволяет по идентификатору повторить платёж - т.е.
//создать новый, у которого все данные, кроме идентификатора - те же самые, что в
//оригинальном платеже.
func (s *Service) Repeat(paymentID string) (*types.Payment, error){
	pay, err := s.FindPaymentByID(paymentID)
	if err!=nil {
		return nil, err
	}

	payment, err :=s.Pay(pay.AccountID, pay.Amount, pay.Category)
	if err!=nil {
		return nil, err
	}

	return payment, err
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	pay, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: pay.AccountID,
		Amount:    pay.Amount,
		Category:  pay.Category,
		Name:      name,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, err

}

//PayFromFavorite
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	var favorite *types.Favorite
	for _, favorites := range s.favorites {
		if favorites.ID == favoriteID {
			favorite = favorites
			break
		}
	}

	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}

	pay, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return pay, nil
}

//ExportToFile - экспортирует все аккаунты
func (s *Service)  ExportToFile(path string) error {
	file, err :=os.Create(path)	
	if err != nil {
		log.Print(err)
	}
	
	defer func () {
		if cerr := file.Close(); cerr!=nil{
			log.Print(cerr)
		}
	}()
    data := ""
	for _, account := range s.accounts {
		id := strconv.Itoa(int(account.ID))+";"
		balance := strconv.Itoa(int(account.Balance))+";"
		phone:=string(account.Phone)

		data +=id
		data +=balance
		data += phone + "|"
	}

	_, err = file.Write([]byte(data))
	if err!=nil {
		log.Print(err)
		return ErrFileNotFound
	}
	return err
}


