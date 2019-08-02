package pkgsign

import (
    "encoding/base64"
    "fmt"
    "golang.org/x/crypto/openpgp"
    "io/ioutil"
    "os"
    "strings"
)

var SignatureVerificationPublicKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - http://gpgtools.org

mQGiBEPspSsRBADdguKAxMQbA32vTQrCyONR6Zs/YGdvau2Zrr3SSSSR0Ge4FMjZ
4tzwpf6+32m4Bsf7YIwdLl0H5hI1CgT5gDl9kXvfaFUehFnwR+FDyiBRiyHjUpGF
4dgkQfWy9diYeWGtsvszsvWHXtED4SXb322StX4MfJj+YesA1iEdTiXK6wCg1QDa
RucfjC+kx4zPsJwkJOgYpyMEAMTiXtNwQcke6nIFb/lb5374NjwwVAuuMTrRWLyq
5HodugEIHaw3EitQWtnFfXNkXTJZzS6t2HAGv29UTfhiBzKdkydgCkOk2MLWISOV
fqcg0tNIp5ZJCmUHg3s+OFNSH4oUi65u+FyDseUid3OKtPI+ZhIk8N+DjOIg2Kvo
/UALA/9q+WfBd7re+W3iUtU7TutUcwbKsjP+jpaJeUHg2ChOBxVfQKt4YlPHVdrR
iCrfNi90Z8qbsZ0iAXuqexrfMq20pAPmpHRpe54mmP1CMT5m+Gq71eKIfkUrb3LC
/zv08dLG2vm9oghd242wbcifaX+t7AhNAIpe/WTvQsB0gpdO4LQmSGlyYW0gQ2hp
cmlubyA8aGlyYW1AaGlyYW1jaGlyaW5vLmNvbT6IWwQTEQIAGwUCQ+ylKwYLCQgH
AwIDFQIDAxYCAQIeAQIXgAAKCRCf8lmA9bp+T/G/AKDM1QDs7il/CJhTycgDvE3c
EOgUBwCfelsVK4sgBCooZptoaCCDgVtt71GIRgQQEQIABgUCRO3MrwAKCRDs3+o8
tEk7lPoGAJ4qoY6sQPRCmVAvygftCnkHzOsc/gCeLoG4wCuTSnH1EjJoPdMHya0e
udGIWwQTEQIAGwUCQ+ylKwYLCQgHAwIDFQIDAxYCAQIeAQIXgAAKCRCf8lmA9bp+
T/G/AKCM+2vI3pYagtmNxdamMgJZ/AWIeQCff0OpzpKQNf5P0Hn+wVCzW2YbRai5
AQ0EQ+ylLhAEAJD25AWgwcNgBFKYsvExQaGIojIGJyn4Cf/5U30cui/K7fIU7Jty
NhKcfZdCrh2hKx+x3H/dTF6e0SrRhzKV7Dx0j76yhHHB1Ak25kjRxoU4Jk+CG0m+
bRNTF9xz9k1ALSm3Y+A5RqNU10K6e/5KsPuXMGSGoQgJ1H6g/i80Wf8PAAMFA/9m
Ixu7lMaqE1OE7EeAsHgLslNbi0h9pjDUVNv8bc1Os2gBPaJD8B89EeheTHw6NMNI
e75HVOpKk4UA0gvOBrxJqCr18yFJBM5sIlaEmuJwZOW4dDGOR1oS5qgE9NzpmyKh
E+fu/S1wmy0coL667+1xZcnrPbUFD4i7/aD1r8qJhohGBBgRAgAGBQJD7KUuAAoJ
EJ/yWYD1un5Pth0An0QEUs5cxpl8zL5kZCj7c8MN8YZDAKDR9LTb6woveul50+uG
tUl2fIH1uA==
=7BPT
-----END PGP PUBLIC KEY BLOCK-----
`

var keyring openpgp.EntityList = nil

func init() {
    k, err := openpgp.ReadArmoredKeyRing(strings.NewReader(SignatureVerificationPublicKey))
    if err != nil {
        panic("Invalid SignatureVerificationPublicKey: " + err.Error())
    }
    keyring = k
}

func Base64Decode(base64String string) (string, error) {
    decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64String))
    decoded, err := ioutil.ReadAll(decoder)
    if err != nil {
        return "", err
    }
    return string(decoded), nil
}

func CheckSignature(base64Sig string, file string) error {
    signed, err := os.Open(file)
    if err != nil {
        return fmt.Errorf("read file: " + err.Error())
    }

    signature, err := Base64Decode(base64Sig)
    if err != nil {
        return fmt.Errorf("invalid signature: " + err.Error())
    }

    entity, err := openpgp.CheckDetachedSignature(keyring, signed, strings.NewReader(signature))
    if err != nil {
        return fmt.Errorf("signature check failure: " + err.Error())
    }

    fmt.Printf("entity: %v\n", entity)
    return nil
}
