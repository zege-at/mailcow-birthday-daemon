package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Marco98/mailcow-birthday-daemon/pkg/mailcow"
	"github.com/emersion/go-webdav"
)

const (
	ConstUsertokenName = "Birthday Daemon"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Daemon struct {
	userTokens        map[string]string
	userTokensLock    *sync.RWMutex
	userTokensUnsaved bool
	httpClient        *http.Client
	baseURL           string
	mailcowClient     mailcow.Client
	statefile         string
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info("starting mcbdd", "version", version, "commit", commit, "date", date)
	d := &Daemon{
		userTokens:     make(map[string]string),
		userTokensLock: &sync.RWMutex{},
		baseURL:        os.Getenv("MAILCOW_BASE"),
		statefile:      os.Getenv("STATEFILE"),
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
	if len(d.statefile) == 0 {
		d.statefile = "state.json"
	}
	d.mailcowClient = mailcow.New(
		d.httpClient,
		d.baseURL,
		os.Getenv("MAILCOW_APIKEY"),
	)
	if err := d.LoadFromDisk(); err != nil {
		return err
	}
	d.daemonLoop()
	return nil
}

func (d *Daemon) daemonLoop() {
	for {
		if err := d.daemonRun(); err != nil {
			slog.Error("error while syncing birthdays", "err", err)
		}
		time.Sleep(time.Minute * 15)
	}
}

func (d *Daemon) daemonRun() error {
	mb, err := d.mailcowClient.GetMailboxes(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching mailboxes: %w", err)
	}
	eg := sync.WaitGroup{}
	for _, m := range mb {
		eg.Go(func() {
			ctx := context.Background()
			if err := d.processUser(ctx, m); err != nil {
				slog.ErrorContext(ctx, "error processing user", "err", err, "user", m.Username)
			}
		})
	}
	eg.Wait()
	if d.userTokensUnsaved {
		slog.Info("saving tokens to disk", "count", len(d.userTokens))
		if err := d.SaveToDisk(); err != nil {
			return err
		}
		d.userTokensUnsaved = false
	}
	return nil
}

func (d *Daemon) processUser(ctx context.Context, m mailcow.Mailbox) error {
	if !m.IsActive() {
		return nil
	}
	pass, err := d.getUserPass(ctx, m.Username)
	if err != nil {
		return fmt.Errorf("error getting userpass: %w", err)
	}
	davclient := webdav.HTTPClientWithBasicAuth(d.httpClient, m.Username, pass)
	bb, err := d.getBirthdays(ctx, davclient, m.Username)
	if err != nil {
		return fmt.Errorf("error getting birthdays from carddav: %w", err)
	}
	if err := d.ensureBirthdayCal(ctx, davclient, m.Username); err != nil {
		return fmt.Errorf("error creating birthday calendar in caldav: %w", err)
	}
	if err := d.syncBirthdaysToCal(ctx, davclient, m.Username, bb); err != nil {
		return fmt.Errorf("error syncing birthday events to caldav: %w", err)
	}
	return nil
}
