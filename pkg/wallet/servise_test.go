package wallet

import (
	"github.com/google/uuid"
	"reflect"
	"testing"

	"github.com/khushbakhtmahkamov/wallet/pkg/types"
)

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
