package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/giota"
)

const Host = "http://node03.iotatoken.nl:15265"

func main() {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	api := giota.NewAPI(Host, &client)
	resp, _ := api.GetTransactionsToApprove(2, giota.DefaultNumberOfWalks, "")
	fmt.Printf("%v\n", resp)
	//&{6360 MVCRKZSLVXVDCSEOASCQQVLKT9PAQKZRIFCFLCMHZMYBQRJABLSVNCBXXKPWLRWWLZOQOISXTVWS99999 YKKXOQRROPBXMXDKTVREXWXXDSQIUPKAZZEZW9LLGQRBTQIFZKPNSDCLMSRQNUMLUIAEMQPGETIB99999}
}
