package confreader

import "encoding/json"
import "io/ioutil"
import "fmt"

type Conf struct {
	BaseUrl string
	Concurrency uint
	TickInterval uint
	DbDsn string
	LoginPass string
	JwtSecret string
}

func Read() Conf {
	content, err := ioutil.ReadFile("./config.json")
	if err != nil {
			fmt.Println("Error when opening file: ", err)
	}

	var ret Conf
	err = json.Unmarshal(content, &ret)
	if err != nil {
		fmt.Println("Error during Unmarshal(): ", err)
	}

	return ret
}