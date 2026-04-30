package cdp

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/iharee/websearch-mcp/internal/model"
)

const (
	defaultAddr        = "localhost:9222"
	navigationTimeout  = 30 * time.Second
	networkIdleTimeout = time.Second
)

// Provider implements fetcher.Provider using Chrome DevTools Protocol via go-rod.
type Provider struct {
	addr    string
	mu      sync.Mutex
	browser *rod.Browser
}

func NewProvider() *Provider {
	addr := os.Getenv("CHROME_DEBUG_ADDR")
	if addr == "" {
		addr = defaultAddr
	}
	return &Provider{addr: addr}
}

func (p *Provider) connect() error {
	u, err := launcher.ResolveURL(p.addr)
	if err != nil {
		return fmt.Errorf("cdp: cannot connect to Chrome at %s. Start Chrome with --remote-debugging-port=<port>, or use method=direct", p.addr)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("cdp: cannot connect to Chrome at %s. Start Chrome with --remote-debugging-port=<port>, or use method=direct", p.addr)
	}
	p.browser = browser
	return nil
}

func (p *Provider) disconnect() {
	if p.browser != nil {
		p.browser.Close()
		p.browser = nil
	}
}

func (p *Provider) Fetch(ctx context.Context, url string) (*model.FetchResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.browser == nil {
		if err := p.connect(); err != nil {
			return nil, err
		}
	}

	page, err := p.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		p.disconnect()
		if connectErr := p.connect(); connectErr != nil {
			return nil, connectErr
		}
		page, err = p.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
		if err != nil {
			return nil, fmt.Errorf("cdp: failed to open page: %w", err)
		}
	}
	defer page.Close()

	page = page.Context(ctx).Timeout(navigationTimeout)

	if err := page.Navigate(url); err != nil {
		return nil, fmt.Errorf("cdp: navigate: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("cdp: wait load: %w", err)
	}

	page.WaitRequestIdle(networkIdleTimeout, nil, nil, nil)()

	page = page.CancelTimeout()

	info, err := page.Info()
	if err != nil {
		return nil, fmt.Errorf("cdp: page info: %w", err)
	}

	result, err := page.Eval("() => document.body?.innerText || ''")
	if err != nil {
		return nil, fmt.Errorf("cdp: eval body: %w", err)
	}

	return &model.FetchResult{
		URL:     info.URL,
		Title:   info.Title,
		Content: result.Value.Str(),
	}, nil
}
