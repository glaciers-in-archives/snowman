package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var port int
var serverInterface string

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves your site through an HTTP server.",
	Long:  `This command will serve your site through Snowman's built-in webserver. It's intended only for usage during development.`,
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat("site"); err != nil {
			fmt.Println("No site found. Did your run snowman build?")
		}

		fs := http.FileServer(http.Dir("site/"))
		address := serverInterface + ":" + strconv.Itoa(port)
		fmt.Println("Serving site at http://" + address + ". Hold ctrl+c to exit.")

		srv := &http.Server{
			Addr:    address,
			Handler: loggingHandler(fs),
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Println(err)
			}
		}()

		<-stop
		fmt.Println("\nShutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&port, "port", "p", 8000, "Port on which the server will listen.")
	serverCmd.Flags().StringVarP(&serverInterface, "address", "a", "127.0.0.1", "Address to which the server will bind.")
}
