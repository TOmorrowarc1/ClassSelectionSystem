package logger

import (
	"os"
	"sync"
	"testing"
)

const LOGTHREADS = 10
const LOGCOUNT = 50
const LOGFILE = "test.log"

func TestLogger(t *testing.T) {
	test_instance := GetLogger()
	test_instance.SetLogLevel(INFO)
	err := test_instance.SetLogFile(LOGFILE)
	if err != nil {
		t.Errorf("Failed to set log file: %v", err)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < LOGTHREADS; i++ {
		wg.Add(1)
		go func(thread_id int) {
			defer wg.Done()
			level := DEBUG
			if thread_id%3 == 0 {
				level = INFO
			}
			for index := 0; index < LOGCOUNT; index++ {
				test_instance.Log(level, "Thread %d: log message %d", thread_id, index)
			}
		}(i)
	}
	wg.Wait()
	test_instance.Close()
	t.Log("Logging test completed. Please check the log file:", LOGFILE)
	os.Remove(LOGFILE)
}
