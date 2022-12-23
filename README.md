# CSV2FolderGo

Leer CSV y generar automáticamente directorios con los datos de una columna del archivo, tambien permite tomar una columna del CSV que contiene links de Google Drive y descargar los archivos.

### Ejecutar

```bash
git clone git@github.com:juliandavidmr/CSV2FolderGo.git
cd CSV2FolderGo
go run main.go ./MyFile.csv 3 94
# 3     => Column number index to read from CSV file and create folders
# 94    => Column number index to read from CSV file and download Google Drive files
```

### Compilar

```bash
# Compilar para Windows
GOOS=windows GOARCH=amd64 go build -o bin/app-amd64.exe main.go
```
