package keymanager_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/prysmaticlabs/prysm/validator/keymanager"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
)

func SetupWallet(t *testing.T) string {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	store := filesystem.New(filesystem.WithLocation(path))
	encryptor := keystorev4.New()

	// Create wallets with keys.
	w1, err := nd.CreateWallet("Wallet 1", store, encryptor)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}
	err = w1.Unlock(nil)
	if err != nil {
		t.Fatalf("Failed to unlock wallet: %v", err)
	}
	_, err = w1.CreateAccount("Account 1", []byte("foo"))
	if err != nil {
		t.Fatalf("Failed to create account 1: %v", err)
	}
	_, err = w1.CreateAccount("Account 2", []byte("bar"))
	if err != nil {
		t.Fatalf("Failed to create account 2: %v", err)
	}

	return path
}

func wallet(t *testing.T, opts string) keymanager.KeyManager {
	km, _, err := keymanager.NewWallet(opts)
	if err != nil {
		t.Fatal(err)
	}
	return km
}

func TestMultiplePassphrases(t *testing.T) {
	path := SetupWallet(t)
	defer os.RemoveAll(path)
	tests := []struct {
		name     string
		wallet   keymanager.KeyManager
		accounts int
	}{
		{
			name:     "Neither",
			wallet:   wallet(t, fmt.Sprintf(`{"location":%q,"accounts":["Wallet 1"],"passphrases":["neither"]}`, path)),
			accounts: 0,
		},
		{
			name:     "Foo",
			wallet:   wallet(t, fmt.Sprintf(`{"location":%q,"accounts":["Wallet 1"],"passphrases":["foo"]}`, path)),
			accounts: 1,
		},
		{
			name:     "Bar",
			wallet:   wallet(t, fmt.Sprintf(`{"location":%q,"accounts":["Wallet 1"],"passphrases":["bar"]}`, path)),
			accounts: 1,
		},
		{
			name:     "Both",
			wallet:   wallet(t, fmt.Sprintf(`{"location":%q,"accounts":["Wallet 1"],"passphrases":["foo","bar"]}`, path)),
			accounts: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keys, err := test.wallet.FetchValidatingKeys()
			if err != nil {
				t.Error(err)
			}
			if len(keys) != test.accounts {
				t.Errorf("Found %d keys; expected %d", len(keys), test.accounts)
			}
		})
	}
}
