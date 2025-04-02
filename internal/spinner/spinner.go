// Package spinner provides a visual loading indicator for CLI operations
package spinner

import (
	"io"
	"time"

	"github.com/briandowns/spinner"
	"github.com/phrazzld/architect/internal/logutil"
)

// Default constants
const (
	defaultCharSet     = 14 // Default character set (circle dots)
	defaultRefreshRate = time.Millisecond * 100
	defaultPrefix      = " "
	defaultSuffix      = " "
)

// Spinner wraps the spinner library for architect's needs
type Spinner struct {
	spinner  *spinner.Spinner
	enabled  bool
	logger   logutil.LoggerInterface
	message  string
	isActive bool
}

// Options configures the spinner behavior
type Options struct {
	Enabled     bool
	CharSet     int
	RefreshRate time.Duration
	Output      io.Writer
	Prefix      string
	Suffix      string
	Color       string
}

// New creates a new spinner instance
func New(logger logutil.LoggerInterface, options *Options) *Spinner {
	// Default options if nil
	if options == nil {
		options = &Options{
			Enabled:     true,
			CharSet:     defaultCharSet,
			RefreshRate: defaultRefreshRate,
			Output:      nil, // Will use os.Stdout by default
			Prefix:      defaultPrefix,
			Suffix:      defaultSuffix,
		}
	}

	s := &Spinner{
		enabled: options.Enabled,
		logger:  logger,
	}

	// Only initialize the spinner if enabled
	if s.enabled {
		// Create the spinner with specified options
		s.spinner = spinner.New(
			spinner.CharSets[options.CharSet],
			options.RefreshRate,
		)

		// Configure spinner details
		if options.Output != nil {
			s.spinner.Writer = options.Output
		}

		if options.Prefix != "" {
			s.spinner.Prefix = options.Prefix
		} else {
			s.spinner.Prefix = defaultPrefix
		}

		if options.Suffix != "" {
			s.spinner.Suffix = options.Suffix
		} else {
			s.spinner.Suffix = defaultSuffix
		}

		// Set color if specified
		if options.Color != "" {
			s.spinner.Color(options.Color)
		}
	}

	return s
}

// Start begins the spinner animation with the given message
func (s *Spinner) Start(message string) {
	s.message = message

	// If spinner is disabled, just log the message
	if !s.enabled {
		s.logger.Info(message)
		return
	}

	// Update spinner message and start
	s.spinner.Suffix = defaultSuffix + message
	s.spinner.Start()
	s.isActive = true

	// Also log at debug level
	s.logger.Debug(message)
}

// Stop ends the spinner animation and logs completion with the given message
func (s *Spinner) Stop(message string) {
	// If spinner is disabled or not active, just log the message
	if !s.enabled || !s.isActive {
		s.logger.Info(message)
		return
	}

	// Stop spinner
	s.spinner.Stop()
	s.isActive = false

	// Log completion message
	s.logger.Info(message)
}

// StopFail ends the spinner for an error condition
func (s *Spinner) StopFail(message string) {
	// If spinner is disabled or not active, just log the message
	if !s.enabled || !s.isActive {
		s.logger.Error(message)
		return
	}

	// Stop spinner
	s.spinner.Stop()
	s.isActive = false

	// Log error message
	s.logger.Error(message)
}

// UpdateMessage updates the spinner message while it's running
func (s *Spinner) UpdateMessage(message string) {
	s.message = message

	// If spinner is disabled or not active, just log the message
	if !s.enabled || !s.isActive {
		s.logger.Info(message)
		return
	}

	s.spinner.Suffix = defaultSuffix + message

	// Also log at debug level
	s.logger.Debug(message)
}

// IsActive returns whether the spinner is currently running
func (s *Spinner) IsActive() bool {
	return s.isActive
}

// IsEnabled returns whether the spinner is enabled
func (s *Spinner) IsEnabled() bool {
	return s.enabled
}
