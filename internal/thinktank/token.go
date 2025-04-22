// Package architect contains the core application logic for the thinktank tool
package thinktank

// This file previously contained the TokenManager implementation.
// It was removed as part of task T032 to remove token handling from the application.
// We now rely on provider APIs to enforce their own token limits instead of
// pre-checking token counts at the application level.
