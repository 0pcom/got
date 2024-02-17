package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"github.com/spf13/cobra"
	"github.com/dustin/go-humanize"
	"github.com/0pcom/got"
	"gitlab.com/poldi1405/go-ansi"
	"gitlab.com/poldi1405/go-indicators/progress"
	"golang.org/x/crypto/ssh/terminal"
)

var (
HeaderSlice []got.GotHeader
output string
dir string
file string
size uint64
concurrency uint
headers []string
agent string
)

func init() {
	RootCmd.Flags().StringVarP(&output, "output", "o", "", "Download `path`, if dir passed the path witll be `dir + output`.")
	RootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Save downloaded file to a `directory`.")
	RootCmd.Flags().StringVarP(&file, "file", "f", "", "Batch download from list of urls in a `file`.")
	RootCmd.Flags().Uint64VarP(&size, "size", "s", 0, "Chunk size in `bytes` to split the file.")
	RootCmd.Flags().UintVarP(&concurrency, "concurrency", "n", 0, "Chunks that will be downloaded concurrently.")
	RootCmd.Flags().StringSliceVarP(&headers, "header", "H", []string{}, `Set these HTTP-Headers on the requests. The format has to be: -H "Key: Value"`)
	RootCmd.Flags().StringVarP(&agent, "agent", "u", "", `Set user agent for got HTTP requests.`)

}
var RootCmd = &cobra.Command{
	Use:   "got",
	Short: "download files faster",
	Long: ``,
	SilenceErrors:         true,
	SilenceUsage:          true,
	DisableSuggestions:    true,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

		go func() {
			<-interruptChan
			cancel()
			signal.Stop(interruptChan)
			log.Fatal(got.ErrDownloadAborted)
		}()

		var (
			g *got.Got           = got.NewWithContext(ctx)
			p *progress.Progress = new(progress.Progress)
		)

		// Set progress style.
		p.SetStyle(progressStyle)

		// Progress func.
		g.ProgressFunc = func(d *got.Download) {
			p.Width = getWidth() - 55

			perc, err := progress.GetPercentage(float64(d.Size()), float64(d.TotalSize()))
			if err != nil {
				perc = 100
			}

			var bar string
			if getWidth() <= 46 {
				bar = ""
			} else {
				bar = r + color(p.GetBar(perc, 100)) + l
			}

			fmt.Printf(
				" %6.2f%% %s %s/%s @ %s/s%s\r",
				perc,
				bar,
				humanize.Bytes(d.Size()),
				humanize.Bytes(d.TotalSize()),
				humanize.Bytes(d.Speed()),
				ansi.ClearRight(),
			)
		}

		info, err := os.Stdin.Stat()

		if err != nil {
			log.Fatal(err)
		}

		// Create directory if not exists.
		if dir != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.MkdirAll(dir, os.ModePerm)
			}
		}

		// Set default user agent.
		if agent != "" {
			got.UserAgent = agent
		}

		// Piped stdin
		if info.Mode()&os.ModeNamedPipe > 0 || info.Size() > 0 {
			if err := multiDownload(ctx, g, bufio.NewScanner(os.Stdin)); err != nil {
				log.Fatal(err)
			}
		}

		// Batch file.
		if file != "" {
			file, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			if err := multiDownload(ctx, g, bufio.NewScanner(file)); err != nil {
				log.Fatal(err)
			}
		}

		if len(headers) > 0 {
			for _, h := range headers {
				split := strings.SplitN(h, ":", 2)
				if len(split) == 1 {
					log.Fatal(errors.New("malformatted header " + h))
				}
				HeaderSlice = append(HeaderSlice, got.GotHeader{Key: split[0], Value: strings.TrimSpace(split[1])})
			}
		}

		// Download from args.
		for _, url := range args {
			if err := download(ctx, g, url); err != nil {
				log.Fatal(err)
			}
			fmt.Print(ansi.ClearLine())
			fmt.Println(fmt.Sprintf("✔ %s", url))
		}
	},


}


func getWidth() int {
	if width, _, err := terminal.GetSize(0); err == nil && width > 0 {
		return width
	}
	return 80
}

func multiDownload(ctx context.Context, g *got.Got, scanner *bufio.Scanner) error {
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" {
			continue
		}
		if err := download(ctx, g, url); err != nil {
			return err
		}
		fmt.Print(ansi.ClearLine())
		fmt.Println(fmt.Sprintf("✔ %s", url))
	}
	return nil
}

func download(ctx context.Context, g *got.Got, url string) (err error) {
	if url, err = getURL(url); err != nil {
		return err
	}
	return g.Do(&got.Download{
		URL:         url,
		Dir:         dir,
		Dest:        output,
		Header:      HeaderSlice,
		Interval:    150,
		ChunkSize:   size,
		Concurrency: concurrency,
	})
}

func getURL(URL string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	// Fallback to https by default.
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	return u.String(), nil
}
