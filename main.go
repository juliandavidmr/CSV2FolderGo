package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Lista de posibles valores del UserAgent
var userAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",
}

// Genera un UserAgent aleatorio
func generateRandomUserAgent() string {
	// Genera un número aleatorio entre 0 y la longitud de la lista de posibles valores
	idx := rand.Intn(len(userAgentList))
	// Devuelve el valor en la posición indicada por el número aleatorio generado
	return userAgentList[idx]
}

func DownloadDriveImage(driveImageURLs, destinationFolder string) int {
	urlBase := "https://drive.google.com/uc?export=download&id="
	imageUrls := strings.Split(driveImageURLs, ",")
	downloadCounter := 0

	for _, imageUrl := range imageUrls {
		imageURL := strings.TrimSpace(imageUrl)
		if imageURL != "" && strings.HasPrefix(imageURL, "https") {
			pattern := regexp.MustCompile("id=(.*)&?")
			match := pattern.FindStringSubmatch(imageUrl)
			if len(match) > 0 {
				// Id encoded in the URL
				encodedId := match[1]
				imageToDownload := urlBase + encodedId
				log.Println("Downloading: " + imageToDownload)

				dir := filepath.Join(destinationFolder, encodedId)
				// Check if file exist and skip
				if _, err := os.Stat(dir); !os.IsNotExist(err) {
					log.Println("File already exist: ", dir)
					continue
				}

				err := downloadImage(imageToDownload, dir)
				if err != nil {
					log.Println("Error downloadling image", err)
					continue
				}
				log.Println("Downloaded:", imageToDownload)
			} else {
				log.Println("NO MATCH")
			}
		}
	}

	return downloadCounter
}

func downloadImage(url, filepath string) error {
	// Crea una nueva solicitud HTTP
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Maneja el error aquí
		return err
	}

	// Establece la cabecera User-Agent en la solicitud
	// para que sea reconocido como el navegador Chrome
	req.Header.Set("User-Agent", generateRandomUserAgent())

	// Realiza la solicitud HTTP
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Maneja el error aquí
		return err
	}
	if resp.StatusCode != 200 {
		// Crear mensaje de error personalizado
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()

	// Obtiene el tipo MIME de la imagen
	mimeType := resp.Header.Get("Content-Type")

	// Determina la extensión del archivo de destino en función del tipo MIME
	ext := strings.Split(mimeType, "/")[1]
	if ext == "" {
		// Si no se puede determinar la extensión, usa una extensión predeterminada
		ext = ".jpg"
	} else {
		// Agrega un punto al principio de la extensión
		ext = "." + ext
	}

	// Crea un archivo en la ruta especificada
	file, err := os.Create(filepath + ext)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribe el contenido de la respuesta en el archivo
	_, err = io.Copy(file, resp.Body)

	if err != nil {
		log.Println("Error writing file: ", err)
		return err
	}

	log.Println("File created: ", filepath+ext)

	return nil
}

func main() {
	// Comprobar si se proporcionaron los argumentos correctos
	if len(os.Args) != 4 {
		log.Fatal("Usage: go run main.go <csv_file> <column_index> <column_images>")
	}

	// Obtener el nombre del archivo CSV y el índice de la columna desde los argumentos
	fileName := os.Args[1]
	columnIndex, err := strconv.Atoi(os.Args[2])
	columnIndexImages, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	// Abrir el archivo CSV
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Crear un nuevo lector CSV que lea desde el archivo
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	reader.Comma = ';'

	// Leer todas las filas del archivo
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Tomar la columna especificada
	column := make([]string, 0)
	columnLinks := make([]string, 0)
	for _, row := range rows {
		// Validar que el índice de la columna esté dentro del rango
		if columnIndex >= len(row) {
			log.Println("Column index out of range", len(row))
			continue
		}

		column = append(column, row[columnIndex])

		if columnIndexImages >= len(row) {
			log.Println("Column index out of range links", len(row))
			continue
		}

		columnLinks = append(columnLinks, row[columnIndexImages])
	}

	// Crear un directorio por cada celda en la columna
	for _, cell := range column {
		// Validar que el directorio no exista
		if _, err := os.Stat(cell); os.IsNotExist(err) {
			// Crear el directorio
			os.Mkdir(cell, os.ModePerm)
			log.Println("Created directory", cell)
		} else {
			log.Println("Directory already exists", cell)
		}
	}

	// Descargar las imágenes
	for i, cell := range columnLinks {
		if cell != "" {
			DownloadDriveImage(cell, column[i])
		}
	}
}
