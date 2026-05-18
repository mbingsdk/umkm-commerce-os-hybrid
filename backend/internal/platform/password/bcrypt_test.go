package password

import "testing"

func TestBcryptHasherHashAndCompare(t *testing.T) {
	t.Parallel()

	hasher := NewBcryptHasher()

	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "password123" {
		t.Fatal("Hash() returned the raw password")
	}
	if err := hasher.Compare(hash, "password123"); err != nil {
		t.Fatalf("Compare() valid password error = %v", err)
	}
	if err := hasher.Compare(hash, "wrong-password"); err == nil {
		t.Fatal("Compare() accepted a wrong password")
	}
}
