package wallet

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/khushbakhtmahkamov/wallet/pkg/types"
)

type Error string

var err error

func (e Error) Error() string {
	return string(e)
}

var ErrPhoneRegistered = errors.New("Phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than 0")
var ErrAccountNotFound = errors.New("account not found")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
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

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {

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
	return account, nil
}

func (s *Service) Pay(accountID int64, amount types.Money, categoty types.PaymentCategory) (*types.Payment, error) {
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
		Category:  categoty,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
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

	account.Balance += payment.Amount
	payment.Status = types.PaymentStatusFail
	return nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {

	var payment *types.Payment
	for _, pay := range s.payments {
		if pay.ID == paymentID {
			payment = pay
			break
		}
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	payment_new, err := s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}

	return payment_new, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {

	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)

	return favorite, nil
}

//PayFromFavorite method
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {

	var favorite *types.Favorite
	for _, v := range s.favorites {
		if v.ID == favoriteID {
			favorite = v
			break
		}
	}
	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)

	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *Service) ExportToFile(path string) error {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var str string
	for _, v := range s.accounts {
		str += fmt.Sprint(v.ID) + ";" + string(v.Phone) + ";" + fmt.Sprint(v.Balance) + "|"
	}
	_, err = file.WriteString(str)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImportFromFile(path string) error {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	strArray := strings.Split(string(content), "|")
	if len(strArray) > 0 {
		strArray = strArray[:len(strArray)-1]
	}
	for _, v := range strArray {
		strArrAcount := strings.Split(v, ";")
		fmt.Println(strArrAcount)

		id, err := strconv.ParseInt(strArrAcount[0], 10, 64)
		if err != nil {
			return err
		}
		balance, err := strconv.ParseInt(strArrAcount[2], 10, 64)
		if err != nil {
			return err
		}
		account := &types.Account{
			ID:      id,
			Phone:   types.Phone(strArrAcount[1]),
			Balance: types.Money(balance),
		}
		s.accounts = append(s.accounts, account)
	}

	return nil
}

func (s *Service) Export(dir string) error {

	/* 	abs, err := filepath.Abs(dir)
	   	if err != nil {
	   		return err
	   	}
	   	err = os.Chdir(abs)
	   	if err != nil {
	   		return err
	   	} */

	if len(s.accounts) > 0 {
		file, _ := os.OpenFile(dir+"/accounts.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

		defer file.Close()

		var str string
		for _, v := range s.accounts {
			str += fmt.Sprint(v.ID) + ";" + string(v.Phone) + ";" + fmt.Sprint(v.Balance) + "\n"
		}
		_, err = file.WriteString(str)
	}

	if len(s.payments) > 0 {
		file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

		defer file.Close()

		var str string
		for _, v := range s.payments {
			str += fmt.Sprint(v.ID) + ";" + fmt.Sprint(v.AccountID) + ";" + fmt.Sprint(v.Amount) + ";" + fmt.Sprint(v.Category) + ";" + fmt.Sprint(v.Status) + "\n"
		}
		_, err = file.WriteString(str)
	}

	if len(s.favorites) > 0 {
		file, _ := os.OpenFile(dir+"/favorites.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

		defer file.Close()

		var str string
		for _, v := range s.favorites {
			str += fmt.Sprint(v.ID) + ";" + fmt.Sprint(v.AccountID) + ";" + fmt.Sprint(v.Amount) + ";" + fmt.Sprint(v.Category) + "\n"
		}
		_, err = file.WriteString(str)
	}

	return nil
}

func (s *Service) Import(dir string) error {

	_, err := os.Stat(dir + "/accounts.dump")

	if err == nil {
		content, err := ioutil.ReadFile(dir + "/accounts.dump")
		if err != nil {
			return err
		}

		strArray := strings.Split(string(content), "\n")
		if len(strArray) > 0 {
			strArray = strArray[:len(strArray)-1]
		}
		for _, v := range strArray {
			strArrAcount := strings.Split(v, ";")
			fmt.Println(strArrAcount)

			id, err := strconv.ParseInt(strArrAcount[0], 10, 64)
			if err != nil {
				return err
			}
			balance, err := strconv.ParseInt(strArrAcount[2], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.accounts {
				if v.ID == id {
					v.Phone = types.Phone(strArrAcount[1])
					v.Balance = types.Money(balance)
					flag = false
				}
			}
			if flag {
				account := &types.Account{
					ID:      id,
					Phone:   types.Phone(strArrAcount[1]),
					Balance: types.Money(balance),
				}
				s.accounts = append(s.accounts, account)
			}
		}
	}

	_, err1 := os.Stat(dir + "/payments.dump")

	if err1 == nil {
		content, err := ioutil.ReadFile(dir + "/payments.dump")
		if err != nil {
			return err
		}

		strArray := strings.Split(string(content), "\n")
		if len(strArray) > 0 {
			strArray = strArray[:len(strArray)-1]
		}
		for _, v := range strArray {
			strArrAcount := strings.Split(v, ";")
			fmt.Println(strArrAcount)

			id := strArrAcount[0]
			if err != nil {
				return err
			}
			aid, err := strconv.ParseInt(strArrAcount[1], 10, 64)
			if err != nil {
				return err
			}
			amount, err := strconv.ParseInt(strArrAcount[2], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.payments {
				if v.ID == id {
					v.AccountID = aid
					v.Amount = types.Money(amount)
					v.Category = types.PaymentCategory(strArrAcount[3])
					v.Status = types.PaymentStatus(strArrAcount[4])
					flag = false
				}
			}
			if flag {
				data := &types.Payment{
					ID:        id,
					AccountID: aid,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(strArrAcount[3]),
					Status:    types.PaymentStatus(strArrAcount[4]),
				}
				s.payments = append(s.payments, data)
			}
		}
	}

	_, err2 := os.Stat(dir + "/favorites.dump")

	if err2 == nil {
		content, err := ioutil.ReadFile(dir + "/favorites.dump")
		if err != nil {
			return err
		}

		strArray := strings.Split(string(content), "\n")
		if len(strArray) > 0 {
			strArray = strArray[:len(strArray)-1]
		}
		for _, v := range strArray {
			strArrAcount := strings.Split(v, ";")
			fmt.Println(strArrAcount)

			id := strArrAcount[0]
			if err != nil {
				return err
			}
			aid, err := strconv.ParseInt(strArrAcount[1], 10, 64)
			if err != nil {
				return err
			}
			amount, err := strconv.ParseInt(strArrAcount[2], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.favorites {
				if v.ID == id {
					v.AccountID = aid
					v.Amount = types.Money(amount)
					v.Category = types.PaymentCategory(strArrAcount[3])
					flag = false
				}
			}
			if flag {
				data := &types.Favorite{
					ID:        id,
					AccountID: aid,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(strArrAcount[3]),
				}
				s.favorites = append(s.favorites, data)
			}
		}
	}

	return nil
}

func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {

	account, err := s.FindAccountByID(accountID)

	if err != nil {
		return nil, err
	}

	var payments []types.Payment
	for _, v := range s.payments {
		if v.AccountID == account.ID {
			data := types.Payment{
				ID:        v.ID,
				AccountID: v.AccountID,
				Amount:    v.Amount,
				Category:  v.Category,
				Status:    v.Status,
			}
			payments = append(payments, data)
		}
	}
	return payments, nil
}

//HistoryToFiles ...
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {

	if len(payments) > 0 {
		if len(payments) <= records {
			file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			defer file.Close()

			var str string
			for _, v := range payments {
				str += fmt.Sprint(v.ID) + ";" + fmt.Sprint(v.AccountID) + ";" + fmt.Sprint(v.Amount) + ";" + fmt.Sprint(v.Category) + ";" + fmt.Sprint(v.Status) + "\n"
			}
			file.WriteString(str)
		} else {

			var str string
			k := 0
			t := 1
			var file *os.File
			for _, v := range payments {
				if k == 0 {
					file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
				}
				k++
				str = fmt.Sprint(v.ID) + ";" + fmt.Sprint(v.AccountID) + ";" + fmt.Sprint(v.Amount) + ";" + fmt.Sprint(v.Category) + ";" + fmt.Sprint(v.Status) + "\n"
				_, err = file.WriteString(str)
				if k == records {
					str = ""
					t++
					k = 0
					file.Close()
				}
			}

		}
	}

	return nil
}

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