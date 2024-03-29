package prometheus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	http      *http.Server
	runErr    error
	readiness bool
}

func New(port int) *Service {
	return &Service{
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler(),
		},
	}
}

func (s *Service) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("prometheus srv: begin run")

	go func() {
		defer wg.Done()
		log.Debug("prometheus srv addr:", s.http.Addr)
		err := s.http.ListenAndServe()
		s.runErr = err
		log.Info("prometheus srv: end run (", err, ")")
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("prometheus srv shutdown (", err, ")")
		}
	}()

	s.readiness = true
}

func handler() http.Handler {
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	return handler
}

func (s *Service) Check() error {
	if !s.readiness {
		return errors.New("prometheus srv is't ready yet")
	}
	if s.runErr != nil {
		return errors.New("run prometheus srv issue")
	}
	return nil
}
