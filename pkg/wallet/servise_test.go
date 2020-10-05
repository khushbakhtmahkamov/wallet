package wallet

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/khushbakhtmahkamov/wallet/pkg/types"
)

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
func TestServise_Reject_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]

	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject() can`t reject account, error =%v", err)
		return
	}
	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject() can`t faind payment by id, error =%v", err)
		return
	}

	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject() status did not chenged, error =%v", err)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject() can`t faind account by id, error =%v", err)
		return
	}

	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject() balanse did not chenged, account =%v", err)
		return
	}
}

func TestServise_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]

	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID() can`t pay account, error =%v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID() wrong payment returned =%v", err)
		return
	}
}

func TestServise_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID() can`t pay account, error =%v", err)
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID() mast returned ErrPaymentNotFound returned =%v", err)
		return
	}
}

func TestServise_Repeat_success(t *testing.T) {

	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject() can`t faind payment by id, error =%v", err)
		return
	}

	if savedPayment.Status != types.PaymentStatusInProgress {
		t.Errorf("Reject() status did not chenged, error =%v", err)
		return
	}

	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat() can`t reject account, error =%v", err)
		return
	}


}
