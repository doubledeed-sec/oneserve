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
	"log"
	"net/http"
	"oneserve/utils"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/net/webdav"
)

// webdavCmd represents the webdav command
var webdavCmd = &cobra.Command{
	Use:   "webdav",
	Short: "Start a server using the WebDAV protocol",
	Long: `Start a server using the WebDAV protocol
Example: oneserve webdav -d /etc --port 8181`,
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("address")
		port, _ := cmd.Flags().GetString("port")
		directory, _ := cmd.Flags().GetString("directory")
		cert, _ := cmd.Flags().GetString("cert")
		key, _ := cmd.Flags().GetString("key")
		tls, _ := cmd.Flags().GetBool("tls")
		colour, _ := cmd.Flags().GetBool("colour")
		prefix, _ := cmd.Flags().GetString("prefix")
		webdavServe(addr, port, directory, cert, key, tls, colour, prefix)
	},
}

func init() {
	rootCmd.AddCommand(webdavCmd)
	webdavCmd.Flags().StringP(
		"port", "p", "8000", "Port on which the server will listen",
	)
	webdavCmd.Flags().StringP(
		"address", "a", "0.0.0.0", "Address on which the server will listen",
	)
	webdavCmd.Flags().StringP(
		"directory", "d", ".", "Directory to serve",
	)
	webdavCmd.Flags().BoolP(
		"tls", "T", false, "Enables TLS",
	)
	webdavCmd.Flags().StringP(
		"cert", "C", "server.crt", "TLS Certificate to use",
	)
	webdavCmd.Flags().StringP(
		"key", "K", "server.key", "Private key to use",
	)
	webdavCmd.Flags().BoolP(
		"colour", "c", false, "Enables colour output",
	)
	webdavCmd.Flags().StringP(
		"prefix", "P", "webdav", "Prefix for webdav server",
	)
}

func webdavServe(addr string, port string, directory string, cert string, key string, tls bool, colour bool, prefix string) {
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

	ws := &webdav.Handler{
		Prefix:     fmt.Sprintf("/%s", prefix),
		FileSystem: webdav.Dir(directory),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("%s | %s | %s", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourWebDAVMethod(r.Method), r.URL)
			} else {
				log.Printf("%s | %s | %s", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourWebDAVMethod(r.Method), r.URL)
			}
		},
	}
	http.Handle("/", ws)

	if !tls {
		log.Printf("Serving %s via WebDAV on %s:%s/%s/\n", directory, addr, port, prefix)
		log.Fatal(http.ListenAndServe(addr+":"+port, nil))
	} else {
		log.Printf("Serving %s via WebDAV (TLS) on %s:%s/%s/\n", directory, addr, port, prefix)
		log.Fatal(http.ListenAndServeTLS(addr+":"+port, cert, key, nil))
	}
}
