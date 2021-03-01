package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/teris-io/shortid"

	quic "github.com/SHARANTANGEDA/mp-quic"
)

func GenerateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func WriteBytesWithQUIC(session quic.Session, bytesToSend []byte, logTime bool) {
	stream, err := session.OpenStreamSync()
	logStartTime := time.Now()
	if err != nil {
		log.Fatal("Error Opening Write Stream: ", err)
	}
	sentBytes, _ := stream.Write(bytesToSend)
	fmt.Println("Sent Bytes: ", sentBytes)
	if logTime {
		fmt.Printf("Time Taken to Send File: %f sec\n", time.Now().Sub(logStartTime).Seconds())
	}
	_ = stream.Close()
}

func SendStringWithQUIC(session quic.Session, message string) {
	bytesToSend := []byte(message)
	WriteBytesWithQUIC(session, bytesToSend, false)
}

func SendFileWithQUIC(session quic.Session, filePath string) error {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %v", err.Error())
	}
	WriteBytesWithQUIC(session, fileBytes, true)
	return nil
}

func ReadDataWithQUIC(stream quic.Stream, frameSize int64) []byte {
	// Make a buffer to hold incoming data.
	buf := make([]byte, frameSize)
	// Read the incoming connection into the buffer.
	readLen, err := io.ReadFull(stream, buf)
	if err != nil {
		log.Fatal("Error reading: ", err.Error(), readLen)
	} else {
		fmt.Println("Read: ", readLen)
	}
	return buf
}

func ReadDataWithTCP(connection *net.TCPConn, frameSize int64) []byte {
	// Make a buffer to hold incoming data.
	buf := make([]byte, frameSize)
	// Read the incoming connection into the buffer.
	readLen, err := io.ReadFull(connection, buf)

	if err != nil {
		log.Fatal("Error reading: ", err.Error(), readLen)
	}
	return buf
}

func StoreFile(fileName, dirPath, fileData string) {
	if fileData == "" {
		log.Fatal("Error: [FILE_RECEIVE] File doesn't exist or is empty")
	}
	uniqueId, err := shortid.Generate()
	file, err := os.Create(dirPath + "/" + uniqueId + "_" + fileName)
	if err != nil {
		log.Fatal("Error writing the file: ", err)
	}
	_, _ = file.WriteString(fileData)
	_ = file.Close()
}
