package image

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

type Renderer struct {
	source, dest string
}

func New(source, dest string) *Renderer {
	return &Renderer{
		source: source,
		dest:   dest,
	}
}

// Render a PNG image as a screenshot from a HTML input file.
func (r *Renderer) Render() (err error) {
	var dest io.WriteCloser

	if r.dest == "-" {
		dest = os.Stdout
	} else {
		dest, err = os.Create(r.dest)
		if err != nil {
			return fmt.Errorf("creating png output file %s: %w", r.dest, err)
		}
		defer func() {
			_ = dest.Close()
		}()
	}

	source, err := os.Open(r.source)
	if err != nil {
		return fmt.Errorf("opening html source %s: %w", r.source, err)
	}
	defer func() {
		_ = source.Close()
	}()

	screenshot, err := r.screenshot(source)
	if err != nil {
		return fmt.Errorf("taking screenshot: %w", err)
	}

	_, err = dest.Write(screenshot)
	if err != nil {
		return fmt.Errorf("writing screenshot: %w", err)
	}

	return nil
}

func (r *Renderer) screenshot(reader io.Reader) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
		// chromedp.WithBrowserOption(opts ...chromedp.BrowserOption)
	)
	defer cancel()
	var screenshot []byte
	// capture entire browser viewport, returning png with quality=90
	// localURL := fmt.Sprintf(`file://./%s`, file)
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Emulate(device.Info{
			Height:    1080,
			Width:     1920,
			Landscape: true,
		}),
		chromedp.Navigate("data:text/html,"+string(content)),
		// chromedp.WaitVisible(`canvas`, chromedp.ByQueryAll),
		// chromedp.WaitReady(`script  _, opts ...chromedp.QueryOption),
		chromedp.Sleep(time.Second),
		chromedp.FullScreenshot(&screenshot, 100), // 100 to force PNG
	)
	if err != nil {
		return nil, err
	}

	return screenshot, nil
}
