package lib

import "testing"

func TestMustFail(t *testing.T) {
	e, err := Encrypts("The world wonders", "Nimitz")
	if err != nil {
		t.Fatal(err)
	}

	// try to decrypt with the wrong password.
	_, err = Decrypts(e, "Halsey")
	if err == nil {
		t.Fatal("Should have returned an error.")
	}
}

func TestEncrypt(t *testing.T) {
	e, err := Encrypts("The world wonders", "Nimitz")
	if err != nil {
		t.Fatal(err)
	}

	d, err := Decrypts(e, "Nimitz")
	if err != nil {
		t.Fatal(err)
	}

	if d != "The world wonders" {
		t.Fatal("Error decrypting")
	}
}

func TestUnique(t *testing.T) {
	a, err := Encrypts("The world wonders", "Nimitz")
	if err != nil {
		t.Fatal(err)
	}

	b, err := Encrypts("The world wonders", "Nimitz")
	if err != nil {
		t.Fatal(err)
	}

	if a == b {
		t.Fatal("Should be different", a, b)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, err := Encrypts("The world wonders", "Nimitz")
		if err != nil {
			b.Fatal(err)
		}
	}
}
