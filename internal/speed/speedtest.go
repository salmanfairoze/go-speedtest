package speed

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/showwin/speedtest-go/speedtest"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type SpeedTest struct {
	client   *speedtest.Speedtest
	log      *log.Entry
	ctx      context.Context
	cancel   context.CancelFunc
	filePath string
}

func New(ctx context.Context, cancel context.CancelFunc, filePath string) SpeedTest {
	return SpeedTest{
		client:   speedtest.New(),
		log:      log.New().WithFields(log.Fields{"client": "speedtest_net"}),
		ctx:      ctx,
		cancel:   cancel,
		filePath: filePath,
	}
}

func (s SpeedTest) CloseSpeedTest() {
	s.log.Infof("stopping speedtest")
	s.cancel()
}

func (s SpeedTest) ExecuteSpeedTestAsync() {
	ctx := s.ctx
	s.ExecuteSpeedTest()

	// Create a ticker to trigger the speed test every 30 minutes
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Define a goroutine to perform the speed test every time the ticker ticks
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.ExecuteSpeedTest()
		}
	}
}

func (s SpeedTest) ExecuteSpeedTest() {
	// Create a new speed test Client
	client := s.client
	logger := s.log

	// Retrieve the closest speed.net server
	logger.Debug("fetching closest server...")
	closestServers, err := client.FetchServers()
	if err != nil {
		logger.Errorf("error fetching closest server list: %v", err)
		return
	}

	// Select the closest server
	closestServer := closestServers[0]

	logger = logger.WithFields(log.Fields{
		"server":   closestServer.Name,
		"country":  closestServer.Country,
		"distance": closestServer.Distance,
	})

	// Download speed test
	logger.Info("running all speed test...")
	err = closestServer.TestAll()
	if err != nil {
		logger.Errorf("error running all speed test: %v", err)
	}

	logger.Infof("download speed: %.2f Mbps", closestServer.DLSpeed)
	logger.Infof("upload speed: %.2f Mbps", closestServer.ULSpeed)
	logger.Infof("jitter: %s", closestServer.Jitter)

	logger.Debug("writing results to csv file...")

	// Append the results to the CSV file
	result := []string{
		time.Now().In(time.FixedZone("IST", 19800)).Format("2006-01-02 15:04:05"),
		closestServer.ID,
		closestServer.Name,
		closestServer.Host,
		fmt.Sprintf("%.2f", closestServer.DLSpeed),
		fmt.Sprintf("%.2f", closestServer.ULSpeed),
		closestServer.Jitter.String(),
		closestServer.Country,
		closestServer.Lat,
		closestServer.Lon,
		fmt.Sprintf("%.4f", closestServer.Distance),
		closestServer.Latency.String(),
		closestServer.MinLatency.String(),
		closestServer.MaxLatency.String(),
		closestServer.TestDuration.Total.String(),
		closestServer.Sponsor,
	}

	s.WriteToCSV(result)
}

func (s SpeedTest) WriteToCSV(entry []string) {
	logger := s.log.WithField("func", "WriteToCSV")
	fn := fmt.Sprintf("%s/speedtest_results.csv", s.filePath)
	file, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("Error opening CSV file: %v", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Errorf("error opening csv file: %v", err)
		}
	}(file)

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header if the file is empty
	if fileInfo, _ := file.Stat(); fileInfo.Size() == 0 {
		header := []string{"Time", "Server ID", "Server Name", "Host", "Download Speed (Mbps)", "Upload Speed (Mbps)", "Jitter", "Country", "Latitude", "Longitude", "Distance", "Latency", "Min Latency", "Max Latency", "Test Duration", "Sponsor"}
		err := writer.Write(header)
		if err != nil {
			logger.Errorf("error writing csv header: %v", err)
		}
	}

	err = writer.Write(entry)
	if err != nil {
		logger.Errorf("error writing csv file: %v", err)
		return
	}

	logger.Infof("speed test results appended to csv file")
}
