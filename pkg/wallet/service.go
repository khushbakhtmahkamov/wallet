package wallet

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/khushbakhtmahkamov/wallet/pkg/types"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var ErrPhoneRegistered = errors.New("Phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than 0")
var ErrAccountNotFound = errors.New("account not found")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrNotEnoughBalance = errors.New("not enough balance")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
}

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992000000000",
	balance: 10000,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 100, category: "auto"},
	},
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can`t register account, error =%v", err)
	}

	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can`t deposit account, error =%v", err)
	}

	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can`t make payment, error =%v", err)
		}
	}

	return account, payments, nil
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {

	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can`t register account, error =%v", err)
	}

	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can`t deposit account, error =%v", err)
	}

	return account, nil
}

func RegisterAccount(service *Service, phone types.Phone) {
	for _, account := range service.accounts {
		if account.Phone == phone {
			return
		}
	}

	service.nextAccountID++
	service.accounts = append(service.accounts, &types.Account{
		ID:      service.nextAccountID,
		Phone:   phone,
		Balance: 0,
	})
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

	account, err:= s.FindAccountByID(payment.AccountID)
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

func (s *testService) Repeat(paymentID string) (*types.Payment,error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	account, err:= s.FindAccountByID(payment.AccountID)
	if err != nil {
		return nil, err
	}

	account.Balance -= payment.Amount
	paymentID = uuid.New().String()
	payment_new := &types.Payment{
		ID:        paymentID,
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Category:  payment.Category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment_new)
	return payment_new,nil
}
