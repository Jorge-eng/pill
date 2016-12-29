package command

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	errInvalidSha  = errors.New("Invalid sha")
	errInvalidBlob = errors.New("Invalid blob")
)

// const endPoint = "https://canary-api.hello.is/v1/provision/blob/pill/"
const endPoint = "https://api.hello.is/v1/provision/blob/pill/"

// const endPoint = "http://localhost:9999/v1/provision/blob/pill/"

type InfoBlob struct {
	DeviceId string
	HexKey   string
	RawBlob  string
}

func (i *InfoBlob) String() string {
	return fmt.Sprintf("InfoBlob{DeviceId: %X}", i.DeviceId)
}

func decrypt_aes_cfb(encrypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	plain := make([]byte, len(encrypted))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plain, encrypted)
	return plain, nil
}

func encrypt_aes_cfb(raw, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	encrypted := make([]byte, len(raw))

	stream := cipher.NewCBCEncrypter(block, raw[0:8])
	stream.CryptBlocks(encrypted, raw)
	return encrypted, nil
}

func parse(encrypted []byte, key string) (*InfoBlob, error) {

	nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	copy(nonce, encrypted[8:16])

	dataStart := encrypted[16 : 308+16]
	factoryKey, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	out, err := decrypt_aes_cfb(dataStart, factoryKey, nonce)
	if err != nil {
		return nil, err
	}
	deviceId := hex.EncodeToString(out[:8])
	hardwareKey := hex.EncodeToString(out[32+128 : 32+128+16])

	infoBlob := InfoBlob{
		DeviceId: strings.ToUpper(deviceId),
		HexKey:   strings.ToUpper(hardwareKey),
		RawBlob:  hex.EncodeToString(out),
	}
	return &infoBlob, nil
}

type BlobCheck interface {
	IsPill(filename string) bool
}

type DVTBlobCheck struct {
}

func (c *DVTBlobCheck) IsPill(filename string) bool {
	t := len(filename) == len("e14903f226c0e4e1559deb8a20a370c087b968f3")
	fmt.Println(filename, t)
	return t
}

type PVTBlobCheck struct {
}

type PillOneCheck struct {
}

func (c *PillOneCheck) IsPill(filename string) bool {
	return isPillBlob(filename)
}

func (c *PVTBlobCheck) IsPill(filename string) bool {
	return isPillBlobPVT(filename)
}

func isPillBlob(filename string) bool {
	return len(filename) == len("90500007A01152103843") && strings.HasPrefix(filename, "90500")
}

func isPillBlobPVT(filename string) bool {
	return len(filename) == len("905000071101164003302") && strings.HasPrefix(filename, "905000071")
}

func checker(manufacturingStage string) BlobCheck {
	switch manufacturingStage {
	case "dvt":
		return &DVTBlobCheck{}
	case "pvt":
		return &PVTBlobCheck{}
	case "one":
		return &PillOneCheck{}
	}
	panic(manufacturingStage)
	return nil
}

func check(archive, sn, key string) ([]string, error) {
	reader, err := zip.OpenReader(archive)
	res := make([]string, 0)
	if err != nil {
		return res, err
	}

	for _, file := range reader.File {
		fname := file.FileInfo().Name()
		// fmt.Println(fname)
		if isPillBlob(fname) || isPillBlobPVT(fname) {

			if fname == sn {

				fileReader, err := file.Open()
				if err != nil {
					return res, err
				}
				defer fileReader.Close()

				buff, err := ioutil.ReadAll(fileReader)
				if err != nil {
					return res, err
				}

				blob, err := parse(buff, key)
				if err != nil {
					return res, err
				}

				fmt.Printf("device_id: %s\n", blob.DeviceId)
				fmt.Printf("key: %s\n", blob.HexKey)
				fmt.Printf("raw: %s\n", blob.RawBlob)
				fmt.Println("")

				resp, upErr := upload(buff, fname)
				if upErr != nil {
					return res, upErr
				}

				res = append(res, blob.DeviceId)

				if strings.Contains(resp, blob.DeviceId) {
					fmt.Println("All good", fname, blob.DeviceId)
				}
			}
		}
	}

	return res, nil
}

func checkLogs(archive, deviceId string) ([]string, error) {
	reader, err := zip.OpenReader(archive)
	res := make([]string, 0)
	if err != nil {
		return res, err
	}

	for _, file := range reader.File {
		fname := file.FileInfo().Name()
		// fmt.Println(fname)
		if strings.HasSuffix(fname, ".htm") {

			fileReader, err := file.Open()
			if err != nil {
				return res, err
			}
			defer fileReader.Close()

			buff, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return res, err
			}

			if strings.Contains(string(buff), deviceId) {
				res = append(res, fname)
			}
		}
	}

	return res, nil
}

func search(archive, deviceId, key string) (string, error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		fname := file.FileInfo().Name()
		if isPillBlob(fname) || isPillBlobPVT(fname) {
			fileReader, err := file.Open()
			if err != nil {
				return "", err
			}

			defer fileReader.Close()

			buff, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return "", err
			}

			blob, err := parse(buff, key)

			if err != nil {
				return "", err
			}

			if blob.DeviceId == deviceId {
				return fname, nil
			}
		}
	}

	return "", nil
}

func upload(buff []byte, fname string) (string, error) {
	resp, err := http.Post(endPoint+fname, "text/plain", bytes.NewBuffer(buff))
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	respBuff, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(fname)
		log.Println(err)
		return "", err
	}

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode, fname)
		return "", errInvalidBlob
	}

	s := string(respBuff)
	return s, nil
}

func process(archive string, checker BlobCheck) ([]string, error) {
	fmt.Println("archive", archive)
	failedUploads := make([]string, 0)
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return failedUploads, err
	}

	i := 0

	for _, file := range reader.File {
		fmt.Println(file.FileInfo().Size())
		fname := file.FileInfo().Name()

		if checker.IsPill(fname) {

			fileReader, err := file.Open()
			if err != nil {
				return failedUploads, err
			}
			defer fileReader.Close()

			buff, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return failedUploads, err
			}

			s, uploadErr := upload(buff, fname)
			if uploadErr != nil {
				msg := fmt.Sprintf("%s : %v", fname, uploadErr)
				failedUploads = append(failedUploads, msg)
			}

			// if !strings.Contains(s, "EXISTS") {
			fmt.Println("resp", s)
			// }

			i += 1
			if i%30 == 0 {
				fmt.Println("Sleeping for 500ms...")
				time.Sleep(500 * time.Millisecond)
			}
			// return failedUploads, nil
		} else {
			fmt.Println(">>> is pill", fname)
		}

	}

	return failedUploads, nil
}
