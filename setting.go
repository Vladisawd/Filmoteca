package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type setting struct {
	ServerHost string
	ServerPort string
	PgHost     string
	PgPort     string
	PgUser     string
	PgPassword string
	PgBase     string
}

func newConf() setting {

	file, err := os.Open("setting.cfg")
	if err != nil {
		panic(fmt.Sprintf("Не удалось открыть файл %s", err))
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(fmt.Sprintf("Не удалось прочитать информацию о файле %s", err))
	}

	fileByte := make([]byte, stat.Size())

	_, err = file.Read(fileByte)
	if err != nil {
		panic(fmt.Sprintf("Не удалось прочитать файл конфигурации %s", err))
	}

	var conf setting

	err = json.Unmarshal(fileByte, &conf)
	if err != nil {
		panic(fmt.Sprintf("Не считать данные %s", err))
	}

	return conf
}
