/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"oneserve/utils"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type dirWrapper struct {
	dir string
}

type LogResponseWriter struct {
	http.ResponseWriter
	status int
	size   string
}

func (w *LogResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *LogResponseWriter) Write(p []byte) (n int, err error) {
	w.size = w.Header().Get("Content-Length")
	return w.ResponseWriter.Write(p)
}

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start a server using the HTTP protocol",
	Long: `Start a server using the HTTP protocol
Example: oneserve http -d /var/tmp --port 8080 --basicauth admin:test -u uploadfile`,
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("address")
		port, _ := cmd.Flags().GetString("port")
		directory, _ := cmd.Flags().GetString("directory")
		basicauth, _ := cmd.Flags().GetString("basicauth")
		upload, _ := cmd.Flags().GetString("upload")
		cert, _ := cmd.Flags().GetString("cert")
		key, _ := cmd.Flags().GetString("key")
		tls, _ := cmd.Flags().GetBool("tls")
		colour, _ := cmd.Flags().GetBool("colour")
		httpServe(addr, port, directory, cert, key, tls, basicauth, upload, colour)
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().StringP(
		"port", "p", "8000", "Port on which the server will listen",
	)
	httpCmd.Flags().StringP(
		"address", "a", "0.0.0.0", "Address on which the server will listen",
	)
	httpCmd.Flags().StringP(
		"directory", "d", ".", "Directory to serve",
	)
	httpCmd.Flags().StringP(
		"basicauth", "b", "", "Basic auth credentials, separated by a colon (:)",
	)
	httpCmd.Flags().BoolP(
		"tls", "T", false, "Enables TLS",
	)
	httpCmd.Flags().StringP(
		"cert", "C", "server.crt", "TLS Certificate to use",
	)
	httpCmd.Flags().StringP(
		"key", "K", "server.key", "Private key to use",
	)
	httpCmd.Flags().BoolP(
		"colour", "c", false, "Enables colour output",
	)
	httpCmd.Flags().StringP(
		"upload", "u", "", "Enables uploading files",
	)
}

func (dw dirWrapper) httpUploadHandler(w http.ResponseWriter, r *http.Request) {
	//func httpUploadHandler(dir string) http.HandlerFunc {
	//return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		log.Printf("%s | %s | %s %s | %s\n", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourStatusCode(http.StatusForbidden), utils.ColourHTTPMethod(r.Method), r.URL.String(), "-")
	} else {
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			return
		}

		defer file.Close()

		log.Printf("%s | %s | %s %s | %s | %s\n", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourStatusCode(http.StatusOK), utils.ColourHTTPMethod(r.Method), r.URL.String(), utils.HumanReadableBytes(strconv.FormatInt(header.Size, 10)), utils.Colorise(fmt.Sprintf("UPLOADED %s!", header.Filename)))

		f, err := os.Create(fmt.Sprintf("%s%s", dw.dir, header.Filename))
		defer f.Close()
		if err != nil {
			log.Println("Could not create file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(f, file); err != nil {
			log.Println("Could not write to file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Uploaded"))
	}
}

func httpLogWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &LogResponseWriter{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(lw, r)
		log.Printf("%s | %s | %s %s | %s\n", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourStatusCode(lw.status), utils.ColourHTTPMethod(r.Method), r.URL.String(), utils.HumanReadableBytes(lw.size))
	})
}

func httpServe(addr string, port string, directory string, cert string, key string, tls bool, basicauth string, upload string, colour bool) {
	handlers := dirWrapper{dir: utils.TrailingSlash(directory)}
	if !colour {
		color.NoColor = true
	}
	if tls {
		if _, err := os.Stat(cert); errors.Is(err, os.ErrNotExist) {
			log.Println("Please provide both a valid certificate and key to use the TLS option")
			os.Exit(3)
		}
		if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
			log.Println("Please provide both a valid certificate and key to use the TLS option")
			os.Exit(3)
		}
	}
	if basicauth == "" {
		http.Handle("/", httpLogWrapper(http.FileServer(http.Dir(directory))))
		if upload != "" {
			http.HandleFunc(fmt.Sprintf("/%s", upload), handlers.httpUploadHandler)
		}
	} else {
		split := strings.Split(basicauth, ":")
		user := split[0]
		pass := split[1]
		http.Handle("/", basicAuth(httpLogWrapper(http.FileServer(http.Dir(directory))), user, pass))
		if upload != "" {
			http.Handle(fmt.Sprintf("/%s", upload), basicAuth(http.HandlerFunc(handlers.httpUploadHandler), user, pass))
		}
	}
	if !tls {
		log.Printf("Serving %s via HTTP on %s:%s\n", directory, addr, port)
		log.Fatal(http.ListenAndServe(addr+":"+port, nil))
	} else {
		log.Printf("Serving %s via HTTPS on %s:%s\n", directory, addr, port)
		log.Fatal(http.ListenAndServeTLS(addr+":"+port, cert, key, nil))
	}
}
