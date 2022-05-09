package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ineverbee/earth-gif/giff"
)

var (
	client   = http.DefaultClient
	baseURL  = "https://api.nasa.gov/EPIC/api/natural"
	start, _ = time.Parse("2006-01-02", "2015-06-13")
	end      = time.Now()
)

type ImgData struct {
	Image string `json:"image"`
	Date  string `json:"date"`
	Bytes []byte
}

func FieldSlice[T string | []byte](imgs []*ImgData, fieldName string) []T {
	slice := make([]T, 0, len(imgs))
	for _, v := range imgs {
		r := reflect.ValueOf(v)
		f := reflect.Indirect(r).FieldByName(fieldName)
		if fieldName == "Date" {
			slice = append(slice, T(f.String()))
		} else if fieldName == "Bytes" {
			slice = append(slice, T(f.Bytes()))
		} else {
			return nil
		}
	}
	return slice
}

func main() {
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatal("There is no API_KEY.")
	}

	date := flag.String("date", "", fmt.Sprintf("Usage example (between %s and %s): --date=2022-01-01", start, end))
	flag.Parse()

	if *date != "" {
		d, err := time.Parse("2006-01-02", *date)
		if err != nil {
			log.Fatal("error: not a valid date")
		} else if !(d.After(start) && d.Before(end)) {
			log.Fatalf("error: date is not between %s and %s", start, end)
		}
		dates, err := GetAllDates(key)
		if err != nil {
			log.Fatal(err)
		}
		for i, v := range dates {
			if v.Date == *date {
				break
			} else if v.Date < *date {
				fmt.Printf("There is no images on that date.\nChoose between these 2 options: (0) %s, (1) %s\n", dates[i-1].Date, v.Date)
				for {
					fmt.Println("Type 0 or 1:")
					buf := bufio.NewReader(os.Stdin)
					bytes, err := buf.ReadBytes('\n')
					choice := strings.TrimRight(string(bytes), "\n")
					if err != nil {
						log.Fatal(err)
					}
					if choice == "0" {
						*date = dates[i-1].Date
						break
					} else if choice == "1" {
						*date = v.Date
						break
					}
				}
				break
			}
		}
	}

	imgs, err := GetImgNames(key, *date)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for _, img := range imgs {
		wg.Add(1)
		go func(im *ImgData) {
			im.Bytes, err = GetPNG(key, im.Image, strings.ReplaceAll(strings.Split(im.Date, " ")[0], "-", "/"))
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}(img)
	}
	log.Println("Retrieving all PNGs..")
	wg.Wait()
	log.Println("Done!")
	log.Println("Creating GIF..")
	err = giff.CreateGIF(
		FieldSlice[[]byte](imgs, "Bytes"),
		FieldSlice[string](imgs, "Date"),
	)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done!")
	log.Println("Now you can go to http://localhost:8080/gif and see the result!")
	log.Println("Or use this command to download GIF: 'curl -v -X POST http://localhost:8080/gif > temp.gif'")
	log.Println("(Ctrl+C to quit)")
	http.Handle("/gif", http.HandlerFunc(gifHandler))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// curl -v -X POST http://localhost:8080/gif > temp.gif
func gifHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "image/gif")
	bytes, err := ioutil.ReadFile("earth.gif")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Write(bytes)
}

func GetImgNames(key, date string) ([]*ImgData, error) {
	if date != "" {
		date = "/date/" + date
	}
	url := baseURL + date + "?api_key=" + key

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var imgSlice []*ImgData
	err = json.Unmarshal(body, &imgSlice)
	if err != nil {
		return nil, err
	}
	return imgSlice, nil
}

func GetAllDates(key string) ([]ImgData, error) {
	url := baseURL + "/all?api_key=" + key

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var datesSlice []ImgData
	err = json.Unmarshal(body, &datesSlice)
	if err != nil {
		return nil, err
	}
	return datesSlice, nil
}

func GetPNG(key, png, date string) ([]byte, error) {
	url := "https://api.nasa.gov/EPIC/archive/natural/" + date + "/png/" + png + ".png?api_key=" + key

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "image/png")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
