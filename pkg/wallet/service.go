package wallet

import (
	"sync"
	// "strings"
	// "io"
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
// func (s *Service)  ExportToFile(path string) error {
// 	file, err :=os.Create(path)	
// 	if err != nil {
// 		log.Print(err)
// 		return ErrFileNotFound
// 	}
	
// 	defer func () {
// 		if cerr := file.Close(); cerr!=nil{
// 			log.Print(cerr)
// 		}
// 	}()
//     data := ""
// 	for _, account := range s.accounts {
// 		id := strconv.Itoa(int(account.ID))+";"
// 		phone:=string(account.Phone)+";"
// 		balance := strconv.Itoa(int(account.Balance))

// 		data +=id
// 		data += phone 
// 		data +=balance+"|"
// 	}

// 	_, err = file.Write([]byte(data))
// 	if err!=nil {
// 		log.Print(err)
// 		return ErrFileNotFound
// 	}
// 	return nil
// }

// //ImportFromFile - импортирует все записи из файла
// func (s *Service) ImportFromFile(path string) error {
// 	file, err := os.Open(path)

// 	if err != nil {
// 		log.Print(err)
// 		return ErrFileNotFound
// 	}
// 	defer func(){
// 		if cerr := file.Close(); cerr != nil {
// 			log.Print(cerr)
// 		}
// 	}()
	
// 	content :=make([]byte, 0)
// 	buf := make([]byte, 4)
// 	for {
// 		read, err := file.Read(buf)
// 		if err == io.EOF {
// 			break
// 		}

// 		if err!=nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
// 		content = append(content, buf[:read]...)
// 	}

// 	data:=string(content)
	
// 	accounts :=strings.Split(data, "|")
// 	accounts = accounts[:len(accounts)-1]
// 	// if accounts == nil {
// 	// 	return ErrAccountNotFound
// 	// }
// 	for _, account := range accounts {
		
// 		value := strings.Split(account, ";")
// 		id,err := strconv.Atoi(value[0])
// 		if err!=nil {
// 			return err
// 		}
// 		phone :=types.Phone(value[1])
// 		balance, err := strconv.Atoi(value[2])
// 		if err!=nil {
// 			return err
// 		}
// 		editAccount := &types.Account {
// 			ID: int64(id),
// 			Phone: phone,
// 			Balance: types.Money(balance),
// 		}

// 		s.accounts = append(s.accounts, editAccount)
// 		log.Print(account)
// 	}
// 	return nil
// }

// //Export(dir string) error
// func (s *Service) Export(dir string) error {
// 	lenAccounts := len(s.accounts)

// 	if lenAccounts!=0 {
// 		fileDir := dir+"/accounts.dump"
// 		file, err :=os.Create(fileDir)	
// 		if err != nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
		
// 		defer func () {
// 			if cerr := file.Close(); cerr!=nil{
// 				log.Print(cerr)
// 			}
// 		}()
// 		data := ""
// 		for _, account := range s.accounts {
// 			id := strconv.Itoa(int(account.ID))+";"
// 			phone:=string(account.Phone)+";"
// 			balance := strconv.Itoa(int(account.Balance))

// 			data +=id
// 			data += phone 
// 			data +=balance+"|"
// 		}

// 		_, err = file.Write([]byte(data))
// 		if err!=nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
// 	}

// 	lenPayments := len(s.payments)

// 	if lenPayments!=0 {
// 		fileDir := dir+"/payments.dump"

// 		file, err :=os.Create(fileDir)	
// 		if err != nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
		
// 		defer func () {
// 			if cerr := file.Close(); cerr!=nil{
// 				log.Print(cerr)
// 			}
// 		}()
// 		data := ""
// 		for _, payment := range s.payments {
// 			idPayment := string(payment.ID)+";"
// 			idPaymnetAccountId := strconv.Itoa(int(payment.AccountID))+";"
// 			amountPayment :=strconv.Itoa(int(payment.Amount))+";"
// 			categoryPayment :=string(payment.Category)+";"
// 			statusPayment := string(payment.Status)

// 			data +=idPayment
// 			data += idPaymnetAccountId 
// 			data+=amountPayment
// 			data+=categoryPayment
// 			data +=statusPayment+"|"
// 		}

// 		_, err = file.Write([]byte(data))
// 		if err!=nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
// 	}

// 	lenFavorites := len(s.favorites)

// 	if lenFavorites!=0 {
// 		fileDir := dir+"/favorites.dump"
// 		file, err :=os.Create(fileDir)	
// 		if err != nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
		
// 		defer func () {
// 			if cerr := file.Close(); cerr!=nil{
// 				log.Print(cerr)
// 			}
// 		}()
// 		data := ""
// 		for _, favorite := range s.favorites {
// 			idFavorite := string(favorite.ID)+";"
// 			idFavoriteAccountId := strconv.Itoa(int(favorite.AccountID))+";"
// 			nameFavorite := string(favorite.Name)+";"
// 			amountFavorite :=strconv.Itoa(int(favorite.Amount))+";"
// 			categoryFavorite :=string(favorite.Category)

// 			data +=idFavorite
// 			data+=idFavoriteAccountId
// 			data += nameFavorite 
// 			data+=amountFavorite
// 			data +=categoryFavorite+"|"
// 		}
// 		_, err = file.Write([]byte(data))
// 		if err!=nil {
// 			log.Print(err)
// 			return ErrFileNotFound
// 		}
// 	}
// 	return nil
// }


// // Import(dir string) error
// func (s *Service) Import(dir string) error {
// 	dirAccount := dir + "/accounts.dump"
// 	file, err := os.Open(dirAccount)

// 	if err != nil {
// 		log.Print(err)
// 		// return ErrFileNotFound
// 		err = ErrFileNotFound
// 	}
// 	if err!=ErrFileNotFound{
// 		defer func(){
// 			if cerr := file.Close(); cerr != nil {
// 				log.Print(cerr)
// 			}
// 		}()
		
// 		content :=make([]byte, 0)
// 		buf := make([]byte, 4)
// 		for {
// 			read, err := file.Read(buf)
// 			if err == io.EOF {
// 				break
// 			}

// 			if err!=nil {
// 				log.Print(err)
// 				//log.Print(dirAccount, " 3333")
// 				return ErrFileNotFound
// 			}
// 			content = append(content, buf[:read]...)
// 		}

// 		data:=string(content)

// 		accounts :=strings.Split(data, "|")
// 		accounts = accounts[:len(accounts)-1]
// 		// if accounts == nil {
// 		// 	return ErrAccountNotFound
// 		// }

// 		for _, account := range accounts {

// 			value := strings.Split(account, ";")

// 			id,err := strconv.Atoi(value[0])
// 			if err!=nil {
// 				return err
// 			}
// 			phone :=types.Phone(value[1])
// 			balance, err := strconv.Atoi(value[2])
// 			if err!=nil {
// 				return err
// 			}
// 			editAccount := &types.Account {
// 				ID: int64(id),
// 				Phone: phone,
// 				Balance: types.Money(balance),
// 			}
// 			//log.Print(editAccount, " read")

// 			s.accounts = append(s.accounts, editAccount)
// 		}
// 	}

// 	dirPaymnet := dir + "/payments.dump"
// 	filePayment, err := os.Open(dirPaymnet)

// 	if err != nil {
// 		log.Print(err)
// 		// return ErrFileNotFound
// 		err = ErrFileNotFound
// 	}
// 	if err!=ErrFileNotFound {
// 		defer func(){
// 			if cerr := filePayment.Close(); cerr != nil {
// 				log.Print(cerr)
// 			}
// 		}()
		
// 		contentPayment :=make([]byte, 0)
// 		buf := make([]byte, 4)
// 		for {
// 			readPayment, err := filePayment.Read(buf)
// 			if err == io.EOF {
// 				break
// 			}

// 			if err!=nil {
// 				log.Print(err)
// 				return ErrFileNotFound
// 			}
// 			contentPayment = append(contentPayment, buf[:readPayment]...)
// 		}

// 		data := string(contentPayment)
		
// 		payments :=strings.Split(data, "|")
// 		payments = payments[:len(payments)-1]
// 		//log.Print(favorites, " fav")
// 		for _, payment := range payments {
			
// 			value := strings.Split(payment, ";")
// 			idPayment := string(value[0])
			
// 			accountIdPeyment, err :=strconv.Atoi(value[1])
// 			if err!=nil {
// 				return err
// 			}

// 			amountPayment, err:=strconv.Atoi(value[2])
// 			if err!=nil {
// 				return err
// 			}
// 			categoryPayment:=types.PaymentCategory(value[3])
			
// 			statusPayment :=types.PaymentStatus(value[4])
// 			newPayment := &types.Payment {
// 				ID: idPayment,
// 				AccountID: int64(accountIdPeyment),
// 				Amount: types.Money(amountPayment),
// 				Category: categoryPayment,
// 				Status: statusPayment,
// 			}

// 			s.payments = append(s.payments, newPayment)
// 			//log.Print(payment)
			
// 		}
// 	}

// 	dirfavorite := dir + "/favorites.dump"
// 	fileFavorite, err := os.Open(dirfavorite)
	
// 	if err != nil {
// 		log.Print(err)
// 		// return ErrFileNotFound
// 		err = ErrFileNotFound
// 	}
// 	if err!=ErrFileNotFound{
// 		defer func(){
// 			if cerr := fileFavorite.Close(); cerr != nil {
// 				log.Print(cerr)
// 			}
// 		}()
		
// 		contentFavorite :=make([]byte, 0)
// 		buf := make([]byte, 4)
// 		for {
// 			readFavorite, err := fileFavorite.Read(buf)
// 			if err == io.EOF {
// 				break
// 			}

// 			if err!=nil {
// 				log.Print(err)
// 				return ErrFileNotFound
// 			}
// 			contentFavorite = append(contentFavorite, buf[:readFavorite]...)
// 		}

// 		data := string(contentFavorite)
// 		//log.Print(dirfavorite, " fav ", data)
// 		favorites :=strings.Split(data, "|")
// 		favorites = favorites[:len(favorites)-1]

// 		for _, favorite := range favorites {
			
// 			valueFavorite := strings.Split(favorite, ";")
// 			idFavorite := string(valueFavorite[0])
// 			accountIdFavorite, err :=strconv.Atoi(valueFavorite[1])
// 			if err!=nil {
// 				return err
// 			}
// 			nameFavorite :=string(valueFavorite[2])

// 			amountFavorite, err:=strconv.Atoi(valueFavorite[3])
// 			if err!=nil {
// 				return err
// 			}
// 			categoryPayment:=types.PaymentCategory(valueFavorite[4])
			
// 			newFavorite := &types.Favorite {
// 				ID: idFavorite,
// 				AccountID: int64(accountIdFavorite),
// 				Name: nameFavorite,
// 				Amount: types.Money(amountFavorite),
// 				Category: categoryPayment,
// 			}

// 			s.favorites = append(s.favorites, newFavorite)
// 			//log.Print(favorite)
// 		}
// 	}
	
// 	return nil
// }

//ExportAccountHistory
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error){
	var paymentFound []types.Payment

	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			paymentFound = append(paymentFound, *payment)
		}
	}
	if paymentFound == nil {
		return nil, ErrAccountNotFound
	}
	return paymentFound, nil
}

//HistoryToFiles
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) > 0 {
		if len(payments) <= records {
			file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			defer func(){
				if cerr := file.Close(); cerr != nil {
					log.Print(cerr)
				}
			}()

			var str string
			for _, payment := range payments {
				// str += fmt.Sprint(payment.ID) + ";" + fmt.Sprint(payment.AccountID) + ";" + fmt.Sprint(payment.Amount) + ";" + fmt.Sprint(payment.Category) + ";" + fmt.Sprint(payment.Status) + "\n"
				idPayment := string(payment.ID)+";"
				idPaymnetAccountId := strconv.Itoa(int(payment.AccountID))+";"
				amountPayment :=strconv.Itoa(int(payment.Amount))+";"
				categoryPayment :=string(payment.Category)+";"
				statusPayment := string(payment.Status)

				str +=idPayment
				str += idPaymnetAccountId 
				str+=amountPayment
				str+=categoryPayment
				str +=statusPayment+"\n"
			}
			_, err := file.WriteString(str)
			if err!=nil {
				log.Print(err)
			}
		}else{
			var str string
			k:=0
			t:=1
			var file *os.File
			for _, payment := range payments {
				if k==0{
					file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
				}
				k++
				str = string(payment.ID)+";" + strconv.Itoa(int(payment.AccountID))+";" + strconv.Itoa(int(payment.Amount))+";" + string(payment.Category)+";" + string(payment.Status) + "\n"
				_, err := file.WriteString(str)
				if err!=nil {
					log.Print(err)
				}
				if k == records{
					str=""
					t++
					k=0;
					// fmt.Println(t, " = t")
					file.Close()
				}
			}
		}
	}
	return nil
}

// func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
// 	var paymentFound []types.Payment

// 	for _, payment := range s.payments {
// 		if payment.AccountID == accountID {
// 			paymentFound = append(paymentFound, *payment)
// 		}
// 	}
// 	if paymentFound == nil {
// 		return nil, ErrAccountNotFound
// 	}
// 	return paymentFound, nil
// }

// func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
// 	// if payments == nil {
// 	// 	log.Print(ErrPaymentNotFound)
// 	// 	return nil
// 	// }
// 	if len(payments)>0 {
// 		if len(payments) <= records {
// 			filePayments := dir + "/payments.dump"
// 				file, err := os.Create(filePayments)
// 				if err != nil {
// 					log.Print(err)
// 					// return ErrFileNotFound
// 				}
// 				defer func () {
// 					if cerr := file.Close(); cerr != nil {
// 						log.Print(cerr)
// 					}
// 				}()
// 				data_payment := ""
// 				for _, payment := range payments {
// 					id := payment.ID + ";"
// 					accountID := strconv.Itoa(int(payment.AccountID)) + ";"
// 					amount := strconv.Itoa(int(payment.Amount)) + ";"
// 					category := string(payment.Category) + ";"
// 					status := string(payment.Status)

// 					data_payment += id
// 					data_payment += accountID
// 					data_payment += amount
// 					data_payment += category
// 					data_payment += status + "\n"
// 				}
// 				_, pay_err := file.Write([]byte(data_payment))
// 				if pay_err != nil {
// 					log.Print(pay_err)
// 					// return ErrFileNotFound
// 				}
// 		} else {
			
// 			var str string
// 			k:=0
// 			t:=1
// 			var file *os.File
// 			for _, payment := range payments {
// 				if k==0{
// 					file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
// 				}
// 				k++
// 				str = string(payment.ID)+";" + strconv.Itoa(int(payment.AccountID))+";" + strconv.Itoa(int(payment.Amount))+";" + string(payment.Category)+";" + string(payment.Status) + "\n"
// 				_, err := file.WriteString(str)
// 				if err!=nil {
// 					log.Print(err)
// 				}
// 			    if k == records{
// 					str=""
// 					t++
// 					k=0;
// 				    fmt.Println(t, " = t")
// 					file.Close()
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

//SumPayments ...
func (s *Service) SumPayments(goroutines int) types.Money {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	sum := int64(0)
	kol := 0
	i := 0
	if goroutines == 0 {
		kol = len(s.payments)
	} else {
		kol = int(len(s.payments) / goroutines)
	}
	for i = 0; i < goroutines-1; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			val := int64(0)
			payments := s.payments[index*kol : (index+1)*kol]
			for _, payment := range payments {
				val += int64(payment.Amount)
			}
			mu.Lock()
			sum += val
			mu.Unlock()

		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		val := int64(0)
		payments := s.payments[i*kol:]
		for _, payment := range payments {
			val += int64(payment.Amount)
		}
		mu.Lock()
		sum += val
		mu.Unlock()

	}()
	wg.Wait()
	return types.Money(sum)
}