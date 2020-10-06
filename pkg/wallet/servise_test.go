package wallet

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/khushbakhtmahkamov/wallet/pkg/types"
)

func TestServise_Reject_success(t *testing.T) {
	s := &Service{}
	phone := types.Phone("+992000000000")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("can`t register account, error =%v", err)
		return
	}

	err = s.Deposit(account.ID, 10000)
	if err != nil {
		t.Errorf("can`t deposit account, error =%v", err)
		return
	}

	payment, err := s.Pay(account.ID, 10000, "auto")
	if err != nil {
		t.Errorf("can`t create pay, error =%v", err)
		return
	}

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

	_, err = s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject() can`t faind account by id, error =%v", err)
		return
	}

}

func TestServise_FindPaymentByID_success(t *testing.T) {
	s := &Service{}
	phone := types.Phone("+992000000000")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("can`t register account, error =%v", err)
		return
	}

	err = s.Deposit(account.ID, 10000)
	if err != nil {
		t.Errorf("can`t deposit account, error =%v", err)
		return
	}

	payment, err := s.Pay(account.ID, 10000, "auto")
	if err != nil {
		t.Errorf("can`t create pay, error =%v", err)
		return
	}

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
	s := &Service{}
	phone := types.Phone("+992000000000")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("can`t register account, error =%v", err)
		return
	}

	err = s.Deposit(account.ID, 10000)
	if err != nil {
		t.Errorf("can`t deposit account, error =%v", err)
		return
	}

	_, err = s.Pay(account.ID, 10000, "auto")
	if err != nil {
		t.Errorf("can`t create pay, error =%v", err)
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

	pay, err = svc.Repeat(pay.ID)
	if err != nil {
		t.Errorf("Repeat(): Error(): can't pay for an account(%v): %v", pay.ID, err)
	}
}
