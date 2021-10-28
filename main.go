package main

import (
	"flag"
	"log"
	"sync"

	"github.com/dtcap/MailHog-IMAP/config"
	"github.com/dtcap/MailHog-IMAP/imap"
	"github.com/dtcap/backends/auth"
	sconfig "github.com/dtcap/backends/config"
	"github.com/dtcap/backends/mailbox"
	"github.com/dtcap/backends/resolver"
)

var conf *config.Config
var wg sync.WaitGroup

func configure() {
	config.RegisterFlags()
	flag.Parse()
	conf = config.Configure()
}

func main() {
	configure()

	for _, s := range conf.Servers {
		wg.Add(1)
		go func(s *config.Server) {
			defer wg.Done()
			err := newServer(conf, s)
			if err != nil {
				log.Fatal(err)
			}
		}(s)
	}

	wg.Wait()
}

func newServer(cfg *config.Config, server *config.Server) error {
	var a, m, r sconfig.BackendConfig
	var err error

	if server.Backends.Auth != nil {
		a, err = server.Backends.Auth.Resolve(cfg.Backends)
		if err != nil {
			return err
		}
	}
	if server.Backends.Mailbox != nil {
		m, err = server.Backends.Mailbox.Resolve(cfg.Backends)
		if err != nil {
			return err
		}
	}
	if server.Backends.Resolver != nil {
		r, err = server.Backends.Resolver.Resolve(cfg.Backends)
		if err != nil {
			return err
		}
	}

	res := resolver.Load(r, *cfg)

	s := &imap.Server{
		BindAddr:        server.BindAddr,
		Hostname:        server.Hostname,
		PolicySet:       server.PolicySet,
		AuthBackend:     auth.Load(a, *cfg),
		MailboxBackend:  mailbox.Load(m, *cfg, res),
		ResolverBackend: res,
		Config:          cfg,
		Server:          server,
	}

	return s.Listen()
}
