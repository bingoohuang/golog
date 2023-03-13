package rotate_test

import (
	"fmt"
	"os"

	"github.com/bingoohuang/golog/pkg/rotate"
	"github.com/sirupsen/logrus"
)

func ExampleNew() {
	logDir, err := os.MkdirTemp("", "rotate_test")
	if err != nil {
		fmt.Println("could not create log directory ", err)
		return
	}

	logPath := fmt.Sprintf("%s/test.log", logDir)

	for i := 0; i < 2; i++ {
		writer, err := rotate.New(logPath)
		if err != nil {
			fmt.Println("Could not open log file ", err)
			return
		}

		n, err := writer.Write(logrus.InfoLevel, []byte("test"))
		if err != nil || n != 4 {
			fmt.Println("Write failed ", err, " number written ", n)
			return
		}

		err = writer.Close()

		if err != nil {
			fmt.Println("Close failed ", err)
			return
		}
	}

	files, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Println("ReadDir failed ", err)
		return
	}

	for _, file := range files {
		info, _ := file.Info()
		fmt.Println(file.Name(), info.Size())
	}

	err = os.RemoveAll(logDir)
	if err != nil {
		fmt.Println("RemoveAll failed ", err)
		return
	}
}
