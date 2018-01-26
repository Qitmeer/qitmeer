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
	hashes := []giota.Trytes{
		//"SYLXZTEIBYKSYFELXJCFGRZNLUOLTARZYW9YJHRAVAUFYJLPYSEBTGZGENEMUVMEMMQQGIKHBUXC99999",
		//"TTSNKNIGDWYXWUWTZLDVGWRUDULMGPYQTTPTWLSOIRKDZYKVNU9SPPETIPUMV9P9QIZIDMUIWFCSZ9999",
		//"AEWQVVOKVKL9QPXVGVTN9VUIKSOYOBGSYBXAXZIXCKMIOVQMNOJDFATTDAVDOFMP9NTLJDJBAFYHZ9999",
		"WNOGGIYGFWHI9ZYWSDSOWHBDIVQMCMSTLBJKEOZ9AJCIODJPGEASTAQZINI9PXIBEXPKOZWUXLYNZ9999",
	}
	var resp *giota.GetTrytesResponse
	resp, _ = api.GetTrytes(hashes)
	fmt.Printf("%v\n", resp.Trytes[0])
	fmt.Printf("%s\n", resp.Trytes[0].Trytes())
}
