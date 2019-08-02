package pkgsign

import (
    "testing"
)

// generate signature with: gpg --output - --detach-sig test.txt| base64
var test_txt_signature = `iF0EABECAB0WIQTluCR6+KYZoo+Q/fyf8lmA9bp+TwUCXURptQAKCRCf8lmA9bp+T+IQAJ4wNdhr+GBr4AUGXlVMAX2Uy04GUgCgqoPgE3WE2dzVHSQ28wo7o0CTH+o=`

func TestUntamperedFile(t *testing.T) {
    err := CheckSignature(test_txt_signature, "test.txt")
    if err!=nil {
        t.Fail()
    }
}
func TestTamperedFile(t *testing.T) {
    err := CheckSignature(test_txt_signature, "test2.txt")
    if err==nil {
        t.Fail()
    }
}