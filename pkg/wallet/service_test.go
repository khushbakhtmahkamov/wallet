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

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	//регистрируем аккаунт
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// пополняем счет
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposity account, error = %v", err)
	}

	//выполняем платежи
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}

	return account, payments, nil

}

var defaultTestAccount = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func TestSerivce_FindPaymentByID_success(t *testing.T) {

	//create service
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//find payment
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return

	}
}

func TestSerivce_FindPaymentByID_fail(t *testing.T) {

	//create service
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//find payment
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID(): must return error, returned nill")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return

	}
}

func TestService_Reject_succes(t *testing.T) {
	//create service
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changet, account = %v", savedPayment)
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changet, account = %v", savedAccount)
	}

}

func TestService_Repeat_succes(t *testing.T) {
	//create service
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_addAccount_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, _, err = s.addAccount(defaultTestAccount)
	if err == nil {
		t.Errorf("ERRRRRRRROOOOORRR !!!! Add two accounts with one number %v", err)
		return
	}

}

func TestService_PayFromFavorite_OK(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	favorite, err := s.FavoritePayment(payment.ID, "First")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_FavoritePayment_OK(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, "First")
	if err == nil {
		return
	}

}

func TestService_ExportToFile(t *testing.T) {
	s := newTestService()
	err := s.ExportToFile("dump.txt")

	if err != nil {
		t.Fatalf("error %v", err)
	}
}

func TestService_ImportFromFile(t *testing.T) {
	s := newTestService()
	s.RegisterAccount("321")
	s.ExportToFile("dump.txt")

	err := s.ImportFromFile("dump.txt")
	if err != nil {
		t.Fatalf("error %v", err)
	}
}

func TestSerive_Export(t *testing.T) {
	s := newTestService()

	s.RegisterAccount("123")
	s.Deposit(1, 12345)
	pay, _ := s.Pay(1, 100, "1")
	s.FavoritePayment(pay.ID, "beha")
	err := s.Export("test")
	if err != nil {
		t.Fatalf("error %v", err)
	}

}

func TestService_Import(t *testing.T) {
	s := newTestService()

	s.RegisterAccount("123")
	s.Deposit(1, 12345)
	pay, _ := s.Pay(1, 100, "1")
	s.FavoritePayment(pay.ID, "beha")
	s.Export("test")

	err := s.Import("test")
	if err != nil {
		t.Fatalf("error %v", err)
	}
}

func BenchmarkSumPayments(b *testing.B) {
	s := newTestService()

	s.RegisterAccount("001")
	s.Deposit(1, 1000)
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")

	want := types.Money(40)

	for i := 0; i < b.N; i++ {
		result := s.SumPayments(3)
		if result != want {
			b.Fatalf("Invalid result, got %v, want %v", result, want)
		}
	}

}

func BenchmarkFilterPayments_one(b *testing.B) {
	s := newTestService()

	s.RegisterAccount("0001")
	s.Deposit(1, 1000)
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "2")
	s.Pay(1, 10, "3")
	s.Pay(1, 10, "4")
	s.Pay(1, 10, "5")
	for i := 0; i < b.N; i++ {
		p, err := s.FilterPayments(1, 1)
		if err != nil {
			b.Fatalf("Error %v", err)
		}
		if len(p) != 5 {
			b.Fatalf("Error %v,", p)
		}
	}
}
func BenchmarkFilterPayments_multiply(b *testing.B) {
	s := newTestService()

	s.RegisterAccount("0001")
	s.Deposit(1, 1000)
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "2")
	s.Pay(1, 10, "3")
	s.Pay(1, 10, "4")
	s.Pay(1, 10, "5")
	for i := 0; i < b.N; i++ {
		p, err := s.FilterPayments(1, 3)
		if err != nil {
			b.Fatalf("Error %v", err)
		}
		if len(p) != 5 {
			b.Fatalf("Error %v,", p)
		}
	}
}

func BenchmarkFilterPaymentsByFn(b *testing.B) {
	s := newTestService()

	s.RegisterAccount("0001")
	s.Deposit(1, 1000)
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	s.Pay(1, 10, "1")
	for i := 0; i < b.N; i++ {
		p, err := s.FilterPaymentsByFn(func(payment types.Payment) bool {
			if payment.AccountID == 1 {
				return true
			}
			return false
		}, 3)
		if err != nil {
			b.Fatalf("Error %v", err)
		}
		if len(p) != 5 {
			b.Fatalf("Error %v,", p)
		}
	}
}
