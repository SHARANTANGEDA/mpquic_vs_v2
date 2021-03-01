package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/SHARANTANGEDA/mpquic_vs_v2/common"
	"github.com/SHARANTANGEDA/mpquic_vs_v2/constants"

	quic "github.com/SHARANTANGEDA/mp-quic"
	mpqConstants "github.com/SHARANTANGEDA/mp-quic/constants"
)

// This is a mpquic server which receives frames and saves them to jpeg files.

func main() {

	videoDir := os.Getenv(constants.PROJECT_HOME_DIR) + "/vid"
	fmt.Println("Saving Video in: ", videoDir)

	serverPort, cfgServer := initializeServerArguments()
	serverAddr := constants.SERVER_HOST + ":" + serverPort
	tlsConfig := common.GenerateTLSConfig()
	listener, err := quic.ListenAddr(serverAddr, tlsConfig, cfgServer)
	if err != nil {
		fmt.Println(err)
		fmt.Println("There was an error in socket")
		return
	}
	fmt.Println("Listening on " + serverAddr)
	fmt.Println("Press ^C to shutdown the server")

	fmt.Println("Server started! Waiting for streams from client...")

	session, err := listener.Accept()
	if err != nil {
		log.Fatal("Error accepting: ", err.Error())
	}

	fmt.Println("session created: ", session.RemoteAddr())

	stream, err := session.AcceptStream()

	if err != nil {
		log.Fatal("Error opening application stream, Error: ", err.Error())
	}
	defer stream.Close()
	defer session.Close(err)

	fmt.Println("stream created: ", stream.StreamID())

	frameCounter := 0

	// Infinite loop which receives frames from go sender peer and then saves them as jpeg files.
	for {
		frameSize := common.ReadDataWithQUIC(stream, 20)
		size, _ := strconv.ParseInt(strings.Trim(string(frameSize), ":"), 10, 64)

		if size == 0 {
			fmt.Println("Got zero frame, will keep trying")
			continue
		}

		fmt.Println("frame size: ", size)

		frameContent := common.ReadDataWithQUIC(stream, size)
		jpegFile, err := os.Create(videoDir + "/img" + strconv.Itoa(frameCounter) + ".jpg")
		if err != nil {
			fmt.Println("Error saving the received frame, Error: ", err.Error())
			break
		}
		frameCounter += 1

		_, err = jpegFile.Write(frameContent)
		if err != nil {
			fmt.Println("Error saving the received frame, Error: ", err.Error())
		}
		_ = jpegFile.Close()
	}
}

func initializeServerArguments() (string, *quic.Config) {
	if os.Getenv(constants.PROJECT_HOME_DIR) == "" {
		panic("`PROJECT_HOME_DIR` Env variable not found")
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
	dumpExperiences := flag.Bool(constants.DUMP_EXPERIENCES_PARAM, false, "a bool(true, false), default: false")
	epsilon := flag.Float64(constants.EPSILON_PARAM, 0.01, "a float64, default: 0 for epsilon value")
	allowedCongestion := flag.Int(constants.ALLOWED_CONGESTION_PARAM, 2500, "a Int, default: 2500")
	flag.Parse()

	return serverPort, &quic.Config{
		WeightsFile:       weightsFile,
		Scheduler:         *scheduler,
		CreatePaths:       true,
		Training:          true,
		DumpExperiences:   *dumpExperiences,
		Epsilon:           *epsilon,
		AllowedCongestion: *allowedCongestion,
	}
}
