package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

const (
	q128      = "128"
	q320      = "320"
	qm4a      = "m4a"
	qlossless = "lossless"
	qbest     = "best"

	extends_128      = "[128kbps_MP3].mp3"
	extends_320      = "[320kbps_MP3].mp3"
	extends_500      = "[500kbps_M4A].m4a"
	extends_lossless = "[Lossless_FLAC].flac"
)

type Media struct {
	url_128      string
	url_320      string
	url_m4a      string
	url_lossless string
	url_best     string
}

func parseAlbumUrl(url string) (mediaUrl []string) {

	resp, _ := http.Get(url)
	doc := html.NewTokenizer(resp.Body)

	for {
		tt := doc.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			fmt.Println("Parsed Album url")
			return

		case tt == html.StartTagToken:
			t := doc.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				for _, a := range t.Attr {
					if a.Key == "href" && strings.Contains(a.Val, "_download.html") {
						mediaUrl = append(mediaUrl, a.Val)
						break
					}
				}
			}
		}

	}
	return
}

func parseMediaUrl(url string, wg *sync.WaitGroup) (media Media) {
	resp, _ := http.Get(url)
	doc := html.NewTokenizer(resp.Body)

	for {
		tt := doc.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			wg.Done()
			return

		case tt == html.StartTagToken:
			t := doc.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				for _, a := range t.Attr {
					if a.Key == "href" && strings.Contains(a.Val, extends_128) {
						media.url_128 = a.Val
						if media.url_128 != "" {
							media.url_best = media.url_128
						}
						break
					}
					if a.Key == "href" && strings.Contains(a.Val, extends_320) {
						media.url_320 = a.Val
						if media.url_320 != "" {
							media.url_best = media.url_320
						}
						break
					}
					if a.Key == "href" && strings.Contains(a.Val, extends_500) {
						media.url_m4a = a.Val
						if media.url_m4a != "" {
							media.url_best = media.url_m4a
						}
						break
					}
					if a.Key == "href" && strings.Contains(a.Val, extends_lossless) {
						media.url_lossless = a.Val
						if media.url_lossless != "" {
							media.url_best = media.url_lossless
						}
						break
					}
				}
			}
		}

	}
	wg.Done()
	return

}

func getMedia(mediaUrl []string) (media []Media) {
	var wg sync.WaitGroup

	for i := 0; i < len(mediaUrl); i++ {
		wg.Add(1)
		media = append(media, parseMediaUrl(mediaUrl[i], &wg))
	}

	wg.Wait()
	return
}

func downloadFile(dir string, url string, wg *sync.WaitGroup) (err error) {

	tmp := strings.Split(url, "/")
	filename := strings.Replace(tmp[len(tmp)-1], "%20", " ", -1)

	path := dir + filename

	fmt.Println("started ", path)

	//create the file
	out, err := os.Create(path)
	if err != nil {
		wg.Done()
		return err
	}
	defer out.Close()

	//get the data
	resp, err := http.Get(url)
	if err != nil {
		wg.Done()
		return err
	}
	defer resp.Body.Close()

	//write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		wg.Done()
		return err
	}

	wg.Done()
	fmt.Println("done ", path)
	return nil
}

func main() {

	dir := os.Args[1]
	quality := os.Args[2]
	albumUrl := os.Args[3]

	/*
		"/home/manojax/Music/"
		"best"
		"http://chiasenhac.vn/nghe-album/amaranth~nightwish~ts370bw6qtnq9k.html"
	*/

	mediaUrl := parseAlbumUrl(albumUrl)
	medias := getMedia(mediaUrl)

	var wg sync.WaitGroup
	for i := 0; i < len(medias); i++ {
		wg.Add(1)
		switch quality {
		case qbest:
			go downloadFile(dir, medias[i].url_best, &wg)
		case qlossless:
			go downloadFile(dir, medias[i].url_lossless, &wg)
		case qm4a:
			go downloadFile(dir, medias[i].url_m4a, &wg)
		case q320:
			go downloadFile(dir, medias[i].url_320, &wg)
		case q128:
			go downloadFile(dir, medias[i].url_128, &wg)
		}

	}

	wg.Wait()
	fmt.Println("All files finished downloading")

}
