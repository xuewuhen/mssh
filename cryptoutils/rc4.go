package cryptoutils

import (
     "crypto/rc4"
     "encoding/base64"
)

func Encrypt(src string) (string, error) {
     cipher, err := rc4.NewCipher(key)
     if err != nil {
          return "", err
     }

     encrypted := make([]byte, len([]byte(src)))
     cipher.XORKeyStream(encrypted, []byte(src))
     encryptedStr := base64.StdEncoding.EncodeToString(encrypted)

     return encryptedStr, nil
}

func Decrypt(src string) (string, error) {
     decodedSrc, err := base64.StdEncoding.DecodeString(src)
     if err != nil {
          return "", err
     }

     cipher, err := rc4.NewCipher(key)
     if err != nil {
          return "", err
     }

     decrypted := make([]byte, len(decodedSrc))
     cipher.XORKeyStream(decrypted, decodedSrc)

     return string(decrypted), nil
}

func SetKey(setkey string) string {
     if len(setkey) > 0 {
          key = []byte(setkey)
     }

     return string(key)
}
