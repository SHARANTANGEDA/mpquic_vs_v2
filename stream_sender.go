package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/SHARANTANGEDA/mpquic_vs_v2/common"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/SHARANTANGEDA/mpquic_vs_v2/constants"

	quic "github.com/SHARANTANGEDA/mp-quic"
	mpqConstants "github.com/SHARANTANGEDA/mp-quic/constants"
)

// This script receives frames via TCP from the stream generator, and then transmits it to
// the peer using mp-quic sockets.
// this script is acting as client for both tcp as well as mp-quic server.

// config
const tcpServerAddr = "localhost:8002"

func main() {
	// tcp connection
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcpServerAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	serverPort, serverIp, cfgClient := initializeClientArguments()
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	session, err := quic.DialAddr(serverIp+":"+serverPort, tlsConfig, cfgClient)
	if err != nil {
		log.Fatal("Error connecting to server: ", err.Error())
	}

	fmt.Println("session created: ", session.RemoteAddr())

	stream, err := session.OpenStreamSync()
	if err != nil {
		log.Fatal("Error Opening Write Stream: ", err)
	}
	defer stream.Close()

	// Infinite loop which takes frame from stream generator and then transmit it to the peer using
	// mpquic stream.
	for {
		// receive the frame size
		frameSizeContent := common.ReadDataWithTCP(conn)

		// fetch size value from the packet by removing trailing ':' symbols.
		size, _ := strconv.ParseInt(strings.Trim(frameSizeContent, ":"), 10, 64)

		// terminate the process when an empty packet is received.
		if size == 0 {
			// send the reply size of zero to terminate the peer.
			_, err := stream.Write([]byte(frameSizeContent))
			if err != nil {
				log.Fatal("Error sending frame size")
			}
			break
		}
		println("frame size: ", size)
		frameContent := common.ReadDataWithTCP(conn)

		// Send the frame size and frame using mpquic stream
		_, err = stream.Write([]byte(frameSizeContent))
		if err != nil {
			log.Fatal("Error sending frame size")
		}
		_, err = stream.Write([]byte(frameContent))
		if err != nil {
			log.Fatal("Error sending Photo Frame")
		}
	}
}

func initializeClientArguments() (string, string, *quic.Config) {
	if os.Getenv(constants.PROJECT_HOME_DIR) == "" {
		panic("`PROJECT_HOME_DIR` Env variable not found")
	}
	serverIp := os.Getenv(constants.SERVER_IP_ADDRESS)
	if serverIp == "" {
		panic("`SERVER_IP` Env variable not found")
	}
	serverPort := os.Getenv(constants.SERVER_PORT)
	if serverPort == "" {
		panic("`SERVER_PORT` Env variable not found")
	}

	weightsFile := os.Getenv(constants.TRAIN_WEIGHTS_FILE_PARAM)
	if weightsFile == "" {
		panic("`WEIGHTS_FILE_PATH` Env variable not found")
	}

	scheduler := flag.String(constants.SCHEDULER_PARAM, mpqConstants.SCHEDULER_ROUND_ROBIN, "Scheduler Name, a string")
	flag.Parse()

	return serverPort, serverIp, &quic.Config{
		WeightsFile: weightsFile,
		Scheduler:   *scheduler,
		CreatePaths: true,
		Training:    true,
	}
}
