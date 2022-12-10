package main

import (
	"bytes"
	"database/sql"
	"image"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/C-STYR/optimizer/optimizerdb"
	"github.com/chai2010/webp"

	_ "github.com/mattn/go-sqlite3"
)

type WebpImage = bytes.Buffer

var imageBytes []WebpImage = make([]WebpImage, 4)
var imagePaths []string = make([]string, 4)
var randomInts []int = make([]int, 50000)
var mainIndex int = 0

// creates slice of filenames from directory
func getImageNames(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

// reads raw img file and outputs a slice of bytes
func loadImg(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

/*
parseImg() converts a slice of bytes into an image.Image
bytes.NewReader() returns a *Reader
image.Decode() decodes an image that has been encoded in a registered format
it takes an io.Reader and returns Image, string, error (string is format name)
*/
func parseImg(rawImg []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(rawImg))
	if err != nil {
		log.Fatal(err)
	}
	return img
}

// converts an image to WebP format
func imgToWebp(img image.Image) WebpImage {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: true}); err != nil {
		log.Fatal(err)
	}
	return buf
}

func convertAndCache() {

	// create waitgroup and increment
	var wg sync.WaitGroup
	// create empty slices

	wg.Add(1)
	go func() {
		defer wg.Done()

		// convert all images to wepb
		pngs := getImageNames("data")

		for i, e := range pngs {
			// create a slice of filenames "imagePaths"
			imgName := e.Name()
			imgPath := path.Join("data", imgName)

			// convert to webp
			rawImg := loadImg(imgPath)
			parsedImg := parseImg(rawImg)
			webpImg := imgToWebp(parsedImg)

			// create fileName for database entries
			fileName := imgName[:len(imgName)-4] + ".webp"

			imageBytes[i] = webpImg
			imagePaths[i] = fileName
			log.Println("file created:", fileName)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// populate slice of ints
		for i := 0; i < 50000; i++ {
			randomInts[i] = rand.Intn(4)
		}
	}()
	wg.Wait()
}

func handleRoot(db *sql.DB) http.Handler {
	i := rand.Intn(3)
	go func() {
		// separately from main thread, increment in db
		optimizerdb.IncrementHitCount(db, imagePaths[i])
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		// writes the byte slice to the http reply
		w.Write(imageBytes[i].Bytes())
		// mainIndex++
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("Using database 'optimizer.db'")

	db, err := sql.Open("sqlite3", "optimizer.db")
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		optimizerdb.TryCreate(db)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		convertAndCache()
	}()

	wg.Wait()
	log.Println("imagePaths:", imagePaths)
	http.Handle("/", handleRoot(db))
	log.Println("Listening on localhost:8099")
	err = http.ListenAndServe("localhost:8099", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
